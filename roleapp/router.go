package roleapp

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
)

// role 自身数据管理
func RoleRouter(g *gin.RouterGroup, ds *dbandmq.Ds) {
	roleR := g.Group("/role/m/", func(c *gin.Context) {
		PreCheckAuth(c)
	})
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

		// 给 role 添加 subrole
		rR.POST("/:id/addsubroles", func(c *gin.Context) {
			AddSubRolesToRoleHandler(c, ds)
		})

		// 删除 role 的 subrole
		rR.POST("/:id/delsubroles", func(c *gin.Context) {
			DelSubRolesFromRoleHandler(c, ds)
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

// 管理用户与 role 的关系
// 本处不实现接口验证，但是在外部调用此接口的地方，需要实现 auth 接口
// 实现 auth 接口时，需要使用 get/set cur user 的方法
// 这样，接口中的数据才能读取到当前用户
func UserAndRoleRouter(g *gin.RouterGroup, ds *dbandmq.Ds) {
	authR := g.Group("/rau", func(c *gin.Context) {
		PreCheckAuth(c)
	})
	{
		// 给 userid 添加 roles
		authR.POST("/addroles", func(c *gin.Context) {
			AddRoleToUserHandler(c, ds)
		})

		// 取消 userid 的 role
		authR.POST("/delroles", func(c *gin.Context) {
			RemoveRoleFromUserHandler(c, ds)
		})

		// 查询 user and role list
		authR.GET("/users", func(c *gin.Context) {
			QueryRoleAndUserHandler(c, ds)
		})
	}
}

// 无需权限的 api
func NoNeedAuthRouter(g *gin.RouterGroup, ds *dbandmq.Ds) {
	nR := g.Group("")
	{
		// 读取用户的 role list
		nR.GET("/rau/user/:id", func(c *gin.Context) {
			GetUserRoleHandler(c, ds)
		})
	}
}
