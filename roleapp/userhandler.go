package roleapp

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/returnfun"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// 给用户添加 role
// role id 与 role name 必须有一个存在
type AddRoleToUserForm struct {
	UserId    string   `json:"userId" binding:"required"`
	UserName  string   `json:"userName"` // 可选值
	RoleIds   []string `json:"roleIds"`
	RoleNames []string `json:"roleNames"`
}

func AddRoleToUserHandler(c *gin.Context, db *dbandmq.Ds) {
	var form AddRoleToUserForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	ds := db.CopyDs()
	defer ds.Close()
	if len(form.RoleIds) == 0 && len(form.RoleNames) == 0 {
		returnfun.ReturnErrJson(c, "roleId 与 roleName 必须要有一个存在")
		return
	}

	var roleIds []string
	if len(form.RoleIds) > 0 {
		for _, rid := range form.RoleIds {
			rid = strings.TrimSpace(rid)
			if rid != "" {
				roleIds = append(roleIds, rid)
			}
		}
	} else if len(form.RoleNames) > 0 {
		// 此处值一般不会多，所以多读几次数据库即可
		for _, rn := range form.RoleNames {
			rn = strings.TrimSpace(rn)
			if rn != "" {
				dbrole, err := GetRoleByName(ds, rn, false)
				middleware.StopExec(err)
				if dbrole != nil {
					roleIds = append(roleIds, dbrole.Id)
				}
			}
		}
	}

	if len(roleIds) == 0 {
		returnfun.ReturnErrJson(c, "传递的 role 相关数据全部不合法")
		return
	}

	// 检查当前操作用户是否能够给别人分配对应的 role
	curUser := GetCurUser(c)
	if curUser == nil {
		returnfun.ReturnJson(c, 417, 417, "服务器配置错误，未正确配置用户验证", "")
		return
	}

	if !IdInSubRoles(curUser, roleIds) {
		returnfun.Return403Json(c, "当前用户无权给用户赋予某些角色")
		return
	}

	// 检查用户是否已经有权限了，如果有，就添加，无就新增
	rau, err := GetRoleAndUserByUserId(ds, strings.TrimSpace(form.UserId))
	middleware.StopExec(err)

	if rau == nil {
		rau = &RoleAndUser{
			Id:       util.GenerateDataId(),
			UserId:   form.UserId,
			UserName: form.UserName,
			RoleIds:  roleIds,
			CreateT:  util.GetCurTime(),
		}
		rau.UpdateT = rau.CreateT
		err = ds.C(CollectionNameRoleAndUser).Insert(rau)
		middleware.StopExec(err)
		returnfun.ReturnOKJson(c, rau)
		return
	}

	rau.RoleIds = append(rau.RoleIds, roleIds...)
	rau.RoleIds = util.UniqueStringArray(rau.RoleIds)

	update := bson.M{
		"$set": bson.M{
			"roleIds": rau.RoleIds,
			"updateT": util.GetCurTime(),
		},
	}

	err = ds.C(CollectionNameRoleAndUser).UpdateId(rau.Id, update)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, rau)
	return
}

// 移除用户的某些 role
type RemoveUserRoleForm struct {
	UserId    string   `json:"userId" binding:"required"`
	RoleIds   []string `json:"roleIds"`
	RoleNames []string `json:"roleNames"`
}

func RemoveRoleFromUserHandler(c *gin.Context, db *dbandmq.Ds) {
	var form RemoveUserRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	ds := db.CopyDs()
	defer ds.Close()
	if len(form.RoleIds) == 0 && len(form.RoleNames) == 0 {
		returnfun.ReturnErrJson(c, "roleId 与 roleName 必须要有一个存在")
		return
	}

	var roleIds []string
	if len(form.RoleIds) > 0 {
		for _, rid := range form.RoleIds {
			rid = strings.TrimSpace(rid)
			if rid != "" {
				roleIds = append(roleIds, rid)
			}
		}
	} else if len(form.RoleNames) > 0 {
		// 此处值一般不会多，所以多读几次数据库即可
		for _, rn := range form.RoleNames {
			rn = strings.TrimSpace(rn)
			if rn != "" {
				dbrole, err := GetRoleByName(ds, rn, false)
				middleware.StopExec(err)
				if dbrole != nil {
					roleIds = append(roleIds, dbrole.Id)
				}
			}
		}
	}

	if len(roleIds) == 0 {
		returnfun.ReturnErrJson(c, "传递的 role 相关数据全部不合法")
		return
	}

	// 检查当前操作用户是否能够给别人分配对应的 role
	curUser := GetCurUser(c)
	if curUser == nil {
		returnfun.ReturnJson(c, 417, 417, "服务器配置错误，未正确配置用户验证", "")
		return
	}

	if !IdInSubRoles(curUser, roleIds) {
		returnfun.Return403Json(c, "当前用户无权给用户赋予某些角色")
		return
	}

	rau, err := GetRoleAndUserByUserId(ds, strings.TrimSpace(form.UserId))
	middleware.StopExec(err)
	if rau == nil {
		returnfun.ReturnErrJson(c, "用户无赋予权限记录")
		return
	}

	// 处理剩下的
	var remainIds []string
	for _, dbr := range rau.RoleIds {
		remain := true
		for _, rid := range roleIds {
			if dbr == rid {
				remain = false
				break
			}
		}
		if remain {
			remainIds = append(remainIds, dbr)
		}
	}

	update := bson.M{
		"$set": bson.M{
			"roleIds": remainIds,
			"updateT": util.GetCurTime(),
		},
	}

	err = ds.C(CollectionNameRoleAndUser).UpdateId(rau.Id, update)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, "")
	return
}

// 读取 userid 与 role 列表
func QueryRoleAndUserHandler(c *gin.Context, db *dbandmq.Ds) {
	var andCondition []bson.M
	uid := c.Query("uid")
	if uid != "" {
		andCondition = append(andCondition, bson.M{"userId": bson.M{"$regex": uid}})
	}

	uname := c.Query("uname")
	if uname != "" {
		andCondition = append(andCondition, bson.M{"userName": bson.M{"$regex": uname}})
	}

	rid := c.Query("rid")
	if rid != "" {
		andCondition = append(andCondition, bson.M{"roleIds": rid})
	}

	query := bson.M{}
	if len(andCondition) > 0 {
		query = bson.M{
			"$and": andCondition,
		}
	}

	ds := db.CopyDs()
	defer ds.Close()

	Q := ds.C(CollectionNameRoleAndUser).Find(query)
	total, err := Q.Count()
	middleware.StopExec(err)

	var raus []*RoleAndUser
	page, size, skip := util.GetPageAndSize(c)
	err = Q.Sort("-_id").Skip(skip).Limit(size).All(&raus)
	middleware.StopExec(err)

	// 组装 role ids 为 simplerole
	var roleIds []string
	for _, rau := range raus {
		roleIds = append(roleIds, rau.RoleIds...)
	}
	roleIds = util.UniqueStringArray(roleIds)

	dbRoles, err := GetRolesByRoleIds(ds, roleIds, false)
	middleware.StopExec(err)
	findR := func(rid string) *SimpleRole {
		for _, dbr := range dbRoles {
			if dbr.Id == rid {
				return &SimpleRole{
					Id:   dbr.Id,
					Name: dbr.Name,
				}
			}
		}
		return nil
	}

	for _, rau := range raus {
		var srs []*SimpleRole
		for _, rid := range rau.RoleIds {
			sr := findR(rid)
			if sr != nil {
				srs = append(srs, sr)
			}
		}
		rau.Roles = srs
	}

	ret := returnfun.QueryListData{
		Total: total,
		Page:  page,
		Size:  size,
		Data:  raus,
	}

	returnfun.ReturnOKJson(c, ret)
	return
}

// 读取用户的 role
// 本接口无需权限
func GetUserRoleHandler(c *gin.Context, db *dbandmq.Ds) {
	uid := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	rau, err := GetRoleAndUserByUserId(ds, uid)
	middleware.StopExec(err)

	if rau == nil {
		// 返回默认用户
		dfr := GetDefaultRole()
		retR := []*SimpleRole{dfr}
		returnfun.ReturnOKJson(c, retR)
		return
	}

	roleIds := rau.RoleIds
	roles, err := GetRolesByRoleIds(ds, roleIds, false)
	middleware.StopExec(err)

	var crs []*SimpleRole
	for _, role := range roles {
		cr := &SimpleRole{
			Id:   role.Id,
			Name: role.Name,
		}
		crs = append(crs, cr)
	}
	rau.Roles = crs

	returnfun.ReturnOKJson(c, rau)
	return
}
