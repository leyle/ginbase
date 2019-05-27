package infiniteclass

import "github.com/gin-gonic/gin"

func CategoryRouter(g *gin.RouterGroup) {
	catR := g.Group("/cat/:domain")
	{
		// 新建分类
		catR.POST("", NewInfiniteClassHandler)

		// 修改指定分类
		catR.PUT("/info/:id", UpdateInfiniteClassHandler)

		// 禁用分类
		catR.POST("/info/:id/disable", DisableInfiniteClassHandler)

		// 启用分类
		catR.POST("/info/:id/enable", EnableInfiniteClassHandler)

		// ?id=xxx   || ?name=xxx&level=1
		catR.GET("/info", GetInfiniteClassInfoHandler)

		// 读取指定 level 的分类列表
		catR.GET("/level/:level", QueryLevelInfiniteClassHandler)
	}
}
