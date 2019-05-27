package infiniteclass

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

// 增删查改
// 新建分类
type NewInfiniteClassForm struct {
	ParentId string `json:"parentId" binding:"required"`
	Name string `json:"name" binding:"required"`
	Icon string `json:"icon"`
	Info string `json:"info"`
}
func NewInfiniteClassHandler(c *gin.Context) {
	var form NewInfiniteClassForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := NewInfiniteClass(db, form.ParentId, form.Name, form.Icon, form.Info)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, ic)
	return
}

// 修改分类名字、图标、描述信息
type UpdateInfiniteClassForm struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	Info string `json:"info"`
}
func UpdateInfiniteClassHandler(c *gin.Context) {
	var form UpdateInfiniteClassForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	id := c.Param("id")

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := GetInfiniteClassById(db, id)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if ic == nil {
		ginbase.ReturnErrJson(c, "无指定id的分类信息")
		return
	}

	ic.Name = form.Name
	ic.Icon = form.Icon
	ic.Info = form.Info
	err = db.C(Opt.TbName).UpdateId(ic.Id, ic)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 禁用分类
func DisableInfiniteClassHandler(c *gin.Context) {
	id := c.Param("id")
	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := GetInfiniteClassById(db, id)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if ic == nil {
		ginbase.ReturnErrJson(c, "无指定id的分类信息")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"disable": true,
		},
	}

	err = db.C(Opt.TbName).UpdateId(ic.Id, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 启用分类
func EnableInfiniteClassHandler(c *gin.Context) {
	id := c.Param("id")
	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := GetInfiniteClassById(db, id)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if ic == nil {
		ginbase.ReturnErrJson(c, "无指定id的分类信息")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"disable": false,
		},
	}

	err = db.C(Opt.TbName).UpdateId(ic.Id, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 读取指定id的分类明细
// 包含了它所有的下一级分类
func GetInfiniteClassInfoHandler(c *gin.Context) {
	id := c.Param("id")
	// 如果 child == Y，就需要读取其所有下级分类，如果不是，只读取自己的信息

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := GetInfiniteClassById(db, id)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if ic == nil {
		ginbase.ReturnErrJson(c, "无指定id的分类数据")
		return
	}

	child := c.Query("children")
	child = strings.ToUpper(child)
	if child == "Y" {
		// 读取其下级
		err = QueryAllChildrenByParentClass(db, ic)
		if err != nil {
			ginbase.ReturnErrJson(c, err.Error())
			return
		}
	}

	ginbase.ReturnOKJson(c, ic)
	return
}

// 读取指定level的分类列表
func QueryLevelInfiniteClassHandler(c *gin.Context) {
	level := c.Param("level")
	ilevel, err := strconv.Atoi(level)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	child := c.Query("children")
	child = strings.ToUpper(child)
	more := false
	if child == "Y" {
		more = true
	}

	ics, err := QueryInfiniteClassByLevel(db, ilevel, more)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c,  ics)
	return
}