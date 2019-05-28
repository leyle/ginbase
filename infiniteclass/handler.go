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
	ParentId string `json:"pid" binding:"required"`
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

	domain := c.Param("domain")

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	ic, err := NewInfiniteClass(db, form.ParentId, form.Name, form.Icon, form.Info, domain)
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
// 可选参数是是否筛选状态
func GetInfiniteClassInfoHandler(c *gin.Context) {
	id := c.Query("id")
	if id != "" {
		getInfiniteClassById(c, id)
		return
	}

	name := c.Query("name")
	if name != "" {
		getInfiniteClassByName(c, name)
		return
	}

	if id == "" && name == "" {
		ginbase.ReturnErrJson(c, "缺少id或name")
		return
	}
}

func getInfiniteClassById(c *gin.Context, id string) {
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

	// disable 参数，如果不传，就是选择所有，如果传了，就是指定状态的
	// Y-禁用的 N-非禁用，空，不筛选
	disable := c.Query("disable")
	disable = strings.ToUpper(disable)

	if disable != "" {
		if disable != "Y" && disable != "N" {
			ginbase.ReturnErrJson(c, "错误的 disable 参数值")
			return
		}
	}

	// 如果 child == Y，就需要读取其所有下级分类，如果不是，只读取自己的信息
	child := c.Query("children")
	child = strings.ToUpper(child)
	if child == "Y" {
		// 读取其下级
		err = QueryAllChildrenByParentClass(db, ic, disable)
		if err != nil {
			ginbase.ReturnErrJson(c, err.Error())
			return
		}
	}

	ginbase.ReturnOKJson(c, ic)
	return
}

// 读取指定名字的分类
// 因为不同的level可以有相同的名字，所以这里还要传递一个 level 值
// 默认为 1
func getInfiniteClassByName(c *gin.Context, name string) {
	iLevel := 1
	level := c.Query("level")
	if level != "" {
		il, err := strconv.Atoi(level)
		if err == nil {
			iLevel = il
		}
	}

	// 是否包含子类
	child := c.Query("children")
	child = strings.ToUpper(child)
	more := false
	if child == "Y" {
		more = true
	}

	// disable 参数，如果不传，就是选择所有，如果传了，就是指定状态的
	// Y-禁用的 N-非禁用，空，不筛选
	disable := c.Query("disable")
	disable = strings.ToUpper(disable)

	if disable != "" {
		if disable != "Y" && disable != "N" {
			ginbase.ReturnErrJson(c, "错误的 disable 参数值")
			return
		}
	}

	domain := c.Param("domain")

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	f := bson.M{
		"domain": domain,
		"level": iLevel,
		"name": bson.M{"$regex": name},
	}
	if disable == "Y" {
		f["disable"] = true
	} else if disable == "N" {
		f["disable"] = false
	}

	var ics []*InfiniteClass
	err := db.C(Opt.TbName).Find(f).All(&ics)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if more {
		// 读取子类
		for _, ic := range ics {
			err = QueryAllChildrenByParentClass(db, ic, disable)
			if err != nil {
				// todo
			}
		}
	}

	ginbase.ReturnOKJson(c, ics)
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

	domain := c.Param("domain")

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	child := c.Query("children")
	child = strings.ToUpper(child)
	more := false
	if child == "Y" {
		more = true
	}

	// disable 参数，如果不传，就是选择所有，如果传了，就是指定状态的
	// Y-禁用的 N-非禁用，空，不筛选
	disable := c.Query("disable")
	disable = strings.ToUpper(disable)

	if disable != "" {
		if disable != "Y" && disable != "N" {
			ginbase.ReturnErrJson(c, "错误的 disable 参数值")
			return
		}
	}

	ics, err := QueryInfiniteClassByLevel(db, domain, ilevel, disable, more)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c,  ics)
	return
}

// 读取 parentId 是指定值的分类列表，仅包含当前层级， 不包含下一层
func QueryInfiniteClassUseParentIdHandler(c *gin.Context) {
	pid := c.Param("id")

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	// disable 参数，如果不传，就是选择所有，如果传了，就是指定状态的
	// Y-禁用的 N-非禁用，空，不筛选
	disable := c.Query("disable")
	disable = strings.ToUpper(disable)

	if disable != "" {
		if disable != "Y" && disable != "N" {
			ginbase.ReturnErrJson(c, "错误的 disable 参数值")
			return
		}
	}

	f := bson.M{
		"parentId": pid,
	}
	if disable == "Y" {
		f["disable"] = true
	} else if disable == "N" {
		f["disable"] = false
	}

	var ics []*InfiniteClass
	err := db.C(Opt.TbName).Find(f).All(&ics)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, ics)
	return
}

