package infiniteclass

import "github.com/gin-gonic/gin"

func CategoryRouter(g *gin.RouterGroup) {
	catR := g.Group("/cat")
	{
		// 新建分类
		catR.POST("", NewInfiniteClassHandler)

		// 修改指定分类
		catR.PUT("/info/:id", UpdateInfiniteClassHandler)

		// 禁用分类
		catR.POST("/info/:id/disable", DisableInfiniteClassHandler)

		// 启用分类
		catR.POST("/info/:id/enable", EnableInfiniteClassHandler)

		// 读取指定id的分类，可选参数 child=Y，读取所有下级
		catR.GET("/info/:id", GetInfiniteClassInfoHandler)

		// 读取指定 level 的分类列表
		catR.GET("/level/:level", QueryLevelInfiniteClassHandler)
	}
}
