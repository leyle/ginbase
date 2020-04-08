package roleapp

import "github.com/leyle/ginbase/dbandmq"

// 引用这个包的功能，需要调用这里的一些方法，来进行初始化
func InitRoleApp(ds *dbandmq.Ds, dfName, adminId, adminName, uriPrefix string) error {
	if dfName != "" {
		DefaultRoleName = dfName
	}

	if adminId != "" {
		AdminUserId = adminId
	}

	if adminName != "" {
		AdminUserName = adminName
	}

	var err error
	// 初始化 defautl role
	err = insureDefaultRole(ds)
	if err != nil {
		return err
	}

	// 初始化管理员
	err = insureAdmin(ds)
	if err != nil {
		return err
	}

	// 初始化一堆系统内置 role 相关的 api
	err = insureRoleAppItems(ds, uriPrefix)
	if err != nil {
		return err
	}
	return nil
}

// Auth 方法
// 验证用户是否有某权限
// 根据 uid 读取用户角色和 api list
// 检查是否可以调用对应的 method/api
func AuthUser(ds *dbandmq.Ds, uid, method, uri string) *AuthResult {
	ar := &AuthResult{
		Result: AuthResultInit,
		Msg:    "init",
		UserId: uid,
	}
	roles, err := GetUserRoles(ds, uid)
	if err != nil {
		ar.Result = AuthResultInternalError
		ar.Msg = "Internal error, maybe db execute failed"
		return ar
	}

	// 展开用户的 roles
	var simpleRoles []*SimpleRole
	for _, role := range roles {
		sr := &SimpleRole{
			Id:   role.Id,
			Name: role.Name,
		}
		simpleRoles = append(simpleRoles, sr)
	}
	ar.Roles = simpleRoles

	childrenRoles := UnWrapChildrenRole(roles)
	ar.ChildrenRole = childrenRoles

	// 一个用户至少有一个角色，那就是默认用户
	items := unWrapRoles(roles)
	if !hasPermission(items, method, uri) {
		ar.Result = AuthResultNoPermission
		ar.Msg = "No permission to call this api"
		return ar
	}

	ar.Result = AuthResultOK
	ar.Msg = "OK"

	return ar
}
