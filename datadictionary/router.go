package datadictionary

import "github.com/gin-gonic/gin"

func DRouter(g *gin.RouterGroup) {
	dR := g.Group("/dd")
	{
		// 维护setname
		// 创建 setname
		dR.POST("/setname", CreateSetHandler)

		// 修改 setname
		dR.PUT("/setname", UpdateSetHandler)

		// 移除 setname
		dR.DELETE("/setname", DelSetHandler)

		// 查询所有的 setname
		dR.GET("/setname", GetAllSetNameHandler)

		// 维护 value
		// 创建 name
		dR.POST("/setname/value", CreateNameHandler)

		// 修改 name
		dR.PUT("/setname/value", UpdateNameHandler)

		// 移除 name
		dR.DELETE("/setname/value", DelNameHandler)

		// 读取所有的名字
		dR.GET("/setname/value", GetAllNamesHandler)
	}
}
