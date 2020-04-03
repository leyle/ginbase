package roleapp

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/returnfun"
)

// 默认用户角色
func GetDefaultRole() *ChildRole {
	r := &ChildRole{
		Id:   DefaultRoleId,
		Name: DefaultRoleName,
	}
	return r
}

func isUnchangeable(g string) bool {
	if g == SysApiGroup {
		return true
	}
	return false
}

type SystemConfig struct {
	Items []*Item
	Ps []*Permission
	Roles []*Role
}

func (s *SystemConfig) addItem(item *Item) {

}

func (s *SystemConfig) addPermission(p *Permission) {

}

func (s *SystemConfig) addRole(p *Role) {

}

func PreCheckAuth(c *gin.Context) {
	user := GetCurUser(c)
	if user == nil {
		returnfun.ReturnJson(c, 417, 417, "服务器配置错误，调用本接口需要配置用户授权", "")
		return
	}
	c.Next()
}