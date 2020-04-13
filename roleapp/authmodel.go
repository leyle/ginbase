package roleapp

import (
	"errors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	. "github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
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

// 验证结果结构
type SimpleRole struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type AuthResult struct {
	Result   int           `json:"result"` // 验证结果
	Msg      string        `json:"msg"`
	UserId   string        `json:"userId"`
	UserName string        `json:"userName"` // 可能无值
	Roles    []*SimpleRole `json:"roles"`
	SubRoles []*SubRole    `json:"subRoles"`
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

// 这是一个全量检测，要求  subroleids 都在 user 的 subroles 里面
func IdInSubRoles(user *AuthResult, subRoleIds []string) bool {
	// 管理员 ok
	if user.UserId == AdminUserId {
		return true
	}

	// 拥有超级 sub role ok
	if hasSubRoles(SuperSubRoleId, user.SubRoles) {
		return true
	}

	// 对比是否有对应的 sub role
	findS := func(s string) bool {
		for _, sr := range user.SubRoles {
			if sr.Id == s {
				return true
			}
		}
		return false
	}

	for _, sr := range subRoleIds {
		if !findS(sr) {
			return false
		}
	}

	return true
}

func hasSubRoles(src string, dsts []*SubRole) bool {
	for _, dst := range dsts {
		if src == dst.Id {
			return true
		}
	}
	return false
}

// 把 roles 的所有 item 全部抽取出来
func unWrapRoles(roles []*Role) []*Item {
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
func UnWrapSubRoles(roles []*Role) []*SubRole {
	var subRoles []*SubRole
	for _, role := range roles {
		if len(role.SubRoles) > 0 {
			subRoles = append(subRoles, role.SubRoles...)
		}
		// 同时追加自身进入
		if role.Id != DefaultRoleId {
			subRoles = append(subRoles, &SubRole{
				Id:   role.Id,
				Name: role.Name,
			})
		}
	}
	if len(subRoles) > 0 {
		subRoles = uniqueSubRoles(subRoles)
	}

	return subRoles
}

func uniqueSubRoles(subRoles []*SubRole) []*SubRole {
	roleMap := make(map[string]*SubRole)
	for _, cr := range subRoles {
		roleMap[cr.Id] = cr
	}

	var ret []*SubRole
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

	roles, err := GetRolesByRoleIds(ds, rau.RoleIds, true)
	if err != nil {
		return nil, err
	}
	Logger.Debugf("", "CurrentUser[%s] roles: [%s]", uid, DebugPrintRoles(roles))

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

	// method 可能是 *, path 也可能是 *
	// 如果 method 包含了 *,直接看 * 对应的 api 是否满足，如果不满足，才后续处理
	// 否则根据情况直接返回
	asterisk := "*"
	auris, ok := infos[asterisk]
	if ok {
		// 检查 * 是否也存在与 auris 里面
		// 如果存在，就是 method / path 都是 *
		for _, auri := range auris {
			if asterisk == auri {
				return true
			} else {
				if uriMatch(path, auri) {
					return true
				}
			}
		}
	}

	// 检查完毕后，发现没有一个满足，还需要后续检查具体的 method 方法
	uris, ok := infos[method]
	if !ok {
		// 连方法都不存在，直接就是 false
		return false
	}

	for _, uri := range uris {
		// 数据库保存的 uri 支持一个 * 通配符
		if uri == "*" {
			return true
		}

		if uriMatch(path, uri) {
			return true
		}
	}

	return false
}

// path 是目标路径
// uri 是基准
// 对比 path 是否与 uri 一致
func uriMatch(path, uri string) bool {
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
			return true
		} else {
			return false
		}
	} else {
		// 否则直接对比
		if uri == path {
			return true
		}
	}
	return false
}

func AddOrUpdateRoleAndRole(ds *dbandmq.Ds, uid, rid, rname string) error {
	roleId := rid
	if rid == "" {
		if rname == "" {
			return errors.New("rid 与 rname 必须至少一个有值")
		}
		dbrole, err := GetRoleByName(ds, rname, false)
		if err != nil {
			return err
		}
		if dbrole == nil {
			return errors.New("没有指定rname的role")
		}
		roleId = dbrole.Id
	}

	if roleId == "" {
		return errors.New("缺少roleId数据")
	}

	rau, err := GetRoleAndUserByUserId(ds, uid)
	if err != nil {
		return err
	}

	if rau == nil {
		rau = &RoleAndUser{
			Id:      util.GenerateDataId(),
			UserId:  uid,
			RoleIds: []string{roleId},
			CreateT: util.GetCurTime(),
		}
		rau.UpdateT = rau.CreateT
		err = ds.C(CollectionNameRoleAndUser).Insert(rau)
		return err
	}

	rau.RoleIds = append(rau.RoleIds, roleId)
	rau.RoleIds = util.UniqueStringArray(rau.RoleIds)
	update := bson.M{
		"$set": bson.M{
			"roleIds": rau.RoleIds,
			"updateT": util.GetCurTime(),
		},
	}

	err = ds.C(CollectionNameRoleAndUser).UpdateId(rau.Id, update)

	return err
}
