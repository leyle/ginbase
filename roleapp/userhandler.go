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
	UserId string `json:"userId" binding:"required"`
	UserName string `json:"userName"` // 可选值
	RoleId string `json:"roleId"`
	RoleName string `json:"roleName"`
}
func AddRoleToUserHandler(c *gin.Context, db *dbandmq.Ds) {
	var form AddRoleToUserForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	ds := db.CopyDs()
	defer ds.Close()

	roleId := strings.TrimSpace(form.RoleId)
	if form.RoleId == "" {
		if form.RoleName == "" {
			returnfun.ReturnErrJson(c, "roleId 与 roleName 必须要有一个存在")
			return
		} else {
			dbrole, err := GetRoleByName(ds, strings.TrimSpace(form.RoleName), false)
			middleware.StopExec(err)
			if dbrole == nil {
				returnfun.ReturnErrJson(c, "没有指定 roleName 的数据")
				return
			}
			roleId = dbrole.Id
		}
	}

	// 检查当前操作用户是否能够给别人分配对应的 role
	curUser := GetCurUser(c)
	if curUser == nil {
		returnfun.ReturnJson(c, 417, 417, "服务器配置错误，未正确配置用户验证", "")
		return
	}

	if !IdInChildrenRole(roleId, curUser.ChildrenRole) {
		returnfun.Return403Json(c, "当前用户无权给用户赋予指定角色")
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
			RoleIds:  []string{roleId},
			CreateT:  util.GetCurTime(),
		}
		rau.UpdateT = rau.CreateT
		err = ds.C(CollectionNameRoleAndUser).Insert(rau)
		middleware.StopExec(err)
		returnfun.ReturnOKJson(c, rau)
		return
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
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, rau)
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
		retR := []*ChildRole{dfr}
		returnfun.ReturnOKJson(c, retR)
		return
	}

	roleIds := rau.RoleIds
	roles, err := GetRolesByRoleIds(ds, roleIds, false)
	middleware.StopExec(err)

	var crs []*ChildRole
	for _, role := range roles {
		cr := &ChildRole{
			Id:   role.Id,
			Name: role.Name,
		}
		crs = append(crs, cr)
	}


	returnfun.ReturnOKJson(c, crs)
	return
}