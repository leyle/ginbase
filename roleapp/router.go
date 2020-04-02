package roleapp

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
)

func RoleRouter(ds *dbandmq.Ds, g *gin.RouterGroup) {
	roleR := g.Group("/role")
	// item manage
	itemR := roleR.Group("/item")
	{
		// 新建 item
		itemR.POST("", func(c *gin.Context) {
			CreateItemHandler(c, ds)
		})

		// 修改 item
		itemR.PUT("/:id", func(c *gin.Context) {
			UpdateItemHandler(c, ds)
		})

		// 删除 item
		itemR.DELETE("/:id", func(c *gin.Context) {
			DeleteItemHandler(c, ds)
		})

		// 读取 item 明细
		itemR.GET("/:id", func(c *gin.Context) {
			GetItemInfoHandler(c, ds)
		})

		// 搜索 item
		roleR.GET("/items", func(c *gin.Context) {
			QueryItemHandler(c, ds)
		})
	}

	// permission manage
	permissionR := roleR.Group("/permission")
	{
		// 新建 permission
		permissionR.POST("", func(c *gin.Context) {
			CreatePermissionHandler(c, ds)
		})

		// 给权限添加 item，可多个
		permissionR.POST("/:id/additems", func(c *gin.Context) {
			AddItemsToPermissionHandler(c, ds)
		})

		// 给权限取消某个或某些 item，可多个
		permissionR.POST("/:id/delitems", func(c *gin.Context) {
			RemoveItemsFromPermissionHandler(c, ds)
		})

		// 修改权限基本信息
		permissionR.PUT("/:id", func(c *gin.Context) {
			UpdatePermissionInfoHandler(c, ds)
		})

		// 删除权限
		permissionR.DELETE("/:id", func(c *gin.Context) {
			DeletePermissionHandler(c, ds)
		})

		// 读取权限明细
		permissionR.GET("/:id", func(c *gin.Context) {
			GetPermissionHandler(c, ds)
		})

		// 搜索权限列表
		roleR.GET("/permissions", func(c *gin.Context) {
			QueryPermissionHandler(c, ds)
		})
	}

	// role manage
	rR := roleR.Group("/role")
	{
		// 新建 role
		rR.POST("", func(c *gin.Context) {
			CreateRoleHandler(c, ds)
		})

		// 给 role 添加 permission
		rR.POST("/:id/addps", func(c *gin.Context) {
			AddPermissionsToRoleHandler(c, ds)
		})

		// 从 role 中移除 permission
		rR.POST("/:id/delps", func(c *gin.Context) {
			RemovePermissionsFromRoleHandler(c, ds)
		})

		// 修改 role 信息
		rR.PUT("/:id", func(c *gin.Context) {
			UpdateRoleInfoHandler(c, ds)
		})

		// 删除role
		rR.DELETE("/:id", func(c *gin.Context) {
			DeleteRoleHandler(c, ds)
		})

		// 给 role 添加 childrole
		rR.POST("/:id/addchildrole", func(c *gin.Context) {
			AddChildRolesToRoleHandler(c, ds)
		})

		// 删除 role 的 childrole
		rR.POST("/:id/delchildrole", func(c *gin.Context) {
			DelChildRolesFromRoleHandler(c, ds)
		})

		// 查看 role 明细
		rR.GET("/:id", func(c *gin.Context) {
			GetRoleInfoHandler(c, ds)
		})

		// 搜索 role
		roleR.GET("/roles", func(c *gin.Context) {
			QueryRoleHandler(c, ds)
		})
	}
}