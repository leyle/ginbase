package simplecrud

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
)

func SimpleDataRouter(g *gin.RouterGroup, ds *dbandmq.Ds) {
	sdR := g.Group("/simpledata/:key")
	{
		// 新建value
		sdR.POST("/add", func(c *gin.Context) {
			CreateSimpleDataHandler(c, ds)
		})

		// 修改
		sdR.POST("/update", func(c *gin.Context) {
			UpdateSimpleDataHandler(c, ds)
		})

		// 删除
		sdR.POST("/del", func(c *gin.Context) {
			DeleteSimpleDataHandler(c, ds)
		})

		// 根据 id 查看名字
		sdR.GET("/info/:id", func(c *gin.Context) {
			GetSimpleDataByIdHandler(c, ds)
		})

		// 搜索
		sdR.GET("/list", func(c *gin.Context) {
			QuerySimpleDataHandler(c, ds)
		})
	}
}
