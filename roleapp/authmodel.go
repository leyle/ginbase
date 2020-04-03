package roleapp

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
	. "github.com/leyle/ginbase/consolelog"
)

func init() {
	dbandmq.AddIndexKey(IKRoleAndUser)
}

var AuthResultCtxKey = "AUTHRESULT"

const (
	AuthResultInit          = 0 // 内部初始化，无任何验证结果
	AuthResultInternalError = 1 // 内部错误，比如 数据库查询错误
	AuthResultNoPermission  = 2 // role 不对，无对应的操作权限
	AuthResultOK            = 9 // 验证成功
)

// user and roleid
const CollectionNameRoleAndUser = DbPrefix + "roleanduser"

var IKRoleAndUser = &dbandmq.IndexKey{
	Collection: CollectionNameRoleAndUser,
	SingleKey:  []string{"userName"},
	UniqueKey:  []string{"userId"},
}

type RoleAndUser struct {
	Id       string        `json:"id" bson:"_id"`
	UserId   string        `json:"userId" bson:"userId"`
	UserName string        `json:"userName" bson:"userName"` // 非必填，主要是给人看的
	RoleIds  []string      `json:"roleIds" bson:"roleIds"`
	Roles    []*Role       `json:"-" bson:"-"`
	CreateT  *util.CurTime `json:"-" bson:"createT"`
	UpdateT  *util.CurTime `json:"-" bson:"updateT"`
}

type AuthResult struct {
	Result       int          `json:"result"` // 验证结果
	Msg string `json:"msg"`
	UserId       string       `json:"userId"`
	UserName     string       `json:"userName"` // 可能无值
	Roles        []*Role      `json:"roles"`
	ChildrenRole []*ChildRole `json:"childrenRole"`
}

func (ar *AuthResult) Dump() string {
	info, _ := jsoniter.MarshalToString(&ar)
	return info
}

func SetCurUser(c *gin.Context, ar *AuthResult) {
	c.Set(AuthResultCtxKey, ar)
}

func GetCurUser(c *gin.Context) *AuthResult {
	ar, exist := c.Get(AuthResultCtxKey)
	if !exist {
		return nil
	}
	result := ar.(*AuthResult)
	return result
}

func IdInChildrenRole(id string, crs []*ChildRole) bool {
	for _, cr := range crs {
		if cr.Name == SuperChildRole {
			return true
		}
		if cr.Id == id {
			return true
		}
	}
	return false
}

// 验证用户是否有某权限
// 根据 uid 读取用户角色和 api list
// 检查是否可以调用对应的 method/api
func AuthUser(ds *dbandmq.Ds, uid, method, uri string) *AuthResult {
	ar := &AuthResult{
		Result: AuthResultInit,
		Msg: "init",
		UserId: uid,
	}
	roles, err := GetUserRoles(ds, uid)
	if err != nil {
		ar.Result = AuthResultInternalError
		ar.Msg = "Internal error, maybe db execute failed"
		return ar
	}
	ar.Roles = roles

	// 一个用户至少有一个角色，那就是默认用户
	items := UnWrapRoles(roles)

	if !hasPermission(items, method, uri) {
		ar.Result = AuthResultNoPermission
		ar.Msg = "user has no permission to call this api"
		return ar
	}

	childrenRoles := UnWrapChildrenRole(roles)
	ar.ChildrenRole = childrenRoles
	ar.Result = AuthResultOK
	ar.Msg = "OK"

	return ar
}

// 把 roles 的所有 item 全部抽取出来
func UnWrapRoles(roles []*Role) []*Item {
	itemMap := make(map[string]*Item)
	for _, role := range roles {
		for _, p := range role.Permissions {
			for _, item := range p.Items {
				itemMap[item.Id] = item
			}
		}
	}

	var items []*Item
	for _, item := range itemMap {
		items = append(items, item)
	}

	return items
}

// 展开所有的子角色
// 子角色不做扩散继承操作，所以一个用户如果需要包含多个子角色，
// 只能通过直接包含的方法获取，不能通过 A 包含 B，B 包含 C，A 就包含了 C 的方式获取
func UnWrapChildrenRole(roles []*Role) []*ChildRole {
	var childrenRole []*ChildRole
	for _, role := range roles {
		if len(role.ChildrenRoles) > 0 {
			childrenRole = append(childrenRole, role.ChildrenRoles...)
			// 同时追加自身进入
			childrenRole = append(childrenRole, &ChildRole{
				Id:   role.Id,
				Name: role.Name,
			})
		}
	}
	if len(childrenRole) > 0 {
		childrenRole = uniqueChildrenRole(childrenRole)
	}

	return childrenRole
}

func uniqueChildrenRole(childrenRole []*ChildRole) []*ChildRole {
	roleMap := make(map[string]*ChildRole)
	for _, cr := range childrenRole {
		roleMap[cr.Id] = cr
	}

	var ret []*ChildRole
	for _, v := range roleMap {
		ret = append(ret, v)
	}
	return ret
}

// 根据用户id读取其role
func GetUserRoles(ds *dbandmq.Ds, uid string) ([]*Role, error) {
	rau, err := GetRoleAndUserByUserId(ds, uid)
	if err != nil {
		return nil, err
	}

	if rau == nil {
		// 默认用户无 rau，创建一个默认数据
		rau = &RoleAndUser{
			UserId:  uid,
			RoleIds: []string{DefaultRoleId},
		}
	} else {
		rau.RoleIds = append(rau.RoleIds, DefaultRoleId)
	}

	// 读取补充完整的 role 信息 todo
	roles, err := GetRolesByRoleIds(ds, rau.RoleIds, true)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func GetRoleAndUserByUserId(ds *dbandmq.Ds, uid string) (*RoleAndUser, error) {
	f := bson.M{
		"userId": uid,
	}

	var rau *RoleAndUser
	err := ds.C(CollectionNameRoleAndUser).Find(f).One(&rau)
	if err != nil && err != mgo.ErrNotFound {
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	return rau, nil
}

func hasPermission(items []*Item, method, path string) bool {
	if len(items) == 0 {
		return false
	}

	// 按照 method 分组 key 是 method， value 是 uri 的列表
	infos := make(map[string][]string)
	for _, item := range items {
		ps, ok := infos[item.Method]
		if ok {
			ps = append(ps, item.Path)
			infos[item.Method] = ps
		} else {
			infos[item.Method] = []string{item.Path}
		}
	}

	uris, ok := infos[method]
	if !ok {
		// 连方法都不存在，直接就是 false
		return false
	}

	found := false
	for _, uri := range uris {
		// 数据库保存的 uri 支持一个 * 通配符
		if uri == "*" {
			found = true
			break
		}

		// 包含通配符，需要正则校验
		if strings.Contains(uri, "*") {
			uri = strings.ReplaceAll(uri, "*", "\\w+")
			uri := "^" + uri + "$"
			re, err := regexp.Compile(uri)
			if err != nil {
				Logger.Errorf("", "检查用户权限时，系统配置错误，无法 compile 正则表达式, %s", err.Error())
				return false
			}
			match := re.MatchString(path)
			if match {
				found = true
				break
			} else {
				continue
			}
		} else {
			// 否则直接对比
			if uri == path {
				found = true
				break
			}
		}
	}

	if found {
		return true
	}

	return false
}