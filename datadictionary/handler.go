package datadictionary

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
	"gopkg.in/mgo.v2/bson"
)

// 集合/分类的维护
// 新增集合/分类名字
type AddSetForm struct {
	Value string `json:"value" binding:"required"`
}
func CreateSetHandler(c *gin.Context) {
	var form AddSetForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	// 读取是否已存在同名的数据
	ddinfo, err := GetSetNameByValue(db, form.Value)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}
	if ddinfo != nil {
		ginbase.ReturnErrJson(c, "已存在同名集合")
		return
	}

	ddinfo = &DDInfo{
		Id: ginbase.GenerateDataId(),
		SetName: SET_KEY_NAME,
		Value: form.Value,
		Valid: true,
	}

	err = db.C(Opt.TbName).Insert(ddinfo)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 修改集合名字
// 联动更新
type UpdateSetForm struct {
	OldValue string `json:"oldValue" binding:"required"`
	NewValue string `json:"newValue" binding:"required"`
}
func UpdateSetHandler(c *gin.Context) {
	var form UpdateSetForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if form.OldValue == form.NewValue {
		ginbase.ReturnErrJson(c, "新旧名字一致，无变动")
		return
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	// 非并发安全的
	olddd, err := GetSetNameByValue(db, form.OldValue)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if olddd == nil {
		ginbase.ReturnErrJson(c, "无旧名字")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"value": form.NewValue,
		},
	}

	err = db.C(Opt.TbName).UpdateId(olddd.Id, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	// 同时修改使用了此名字的作为 setname 的所有数据
	f := bson.M{
		"setName": form.OldValue,
	}
	updateAll := bson.M{
		"$set": bson.M{
			"setName": form.NewValue,
		},
	}

	_, err = db.C(Opt.TbName).UpdateAll(f, updateAll)
	if err != nil {
		emsg := fmt.Sprintf("联动更新使用的数据失败, %s", err.Error())
		ginbase.ReturnErrJson(c, emsg)
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 删除集合名字
// 仅仅标记为不可用，不需要联动更新
type DelSetForm struct {
	Value string `json:"value" binding:"required"`
}
func DelSetHandler(c *gin.Context) {
	var form DelSetForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	f := bson.M{
		"setName": SET_KEY_NAME,
		"value": form.Value,
	}
	update := bson.M{
		"$set": bson.M{
			"valid": false,
		},
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	err := db.C(Opt.TbName).Update(f, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}
	ginbase.ReturnOKJson(c, "")
	return
}

// 读取所有的集合名字
func GetAllSetNameHandler(c *gin.Context) {
	f := bson.M{
		"setName": SET_KEY_NAME,
		"valid": true,
	}
	var dds []*DDInfo
	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()
	err := db.C(Opt.TbName).Find(f).All(&dds)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	var names []string
	for _, dd := range dds {
		names = append(names, dd.Value)
	}

	ginbase.ReturnOKJson(c, names)
	return
}

// 新建指定 set 下的名字
type CreateNameForm struct {
	SetName string `json:"setName" binding:"required"`
	Value string `json:"value" binding:"required"`
}
func CreateNameHandler(c *gin.Context) {
	var form CreateNameForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	// 检查 setname 是否存在，不存在报错
	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	setInfo, err := GetSetNameByValue(db, form.SetName)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if setInfo == nil {
		ginbase.ReturnErrJson(c, "不存在输入的分类名字,请先创建")
		return
	}

	// 检查 value 是否存在
	valueInfo, err := GetNameBySetNameAndValue(db, form.SetName, form.Value)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if valueInfo != nil {
		ginbase.ReturnErrJson(c, "数据重复，本分类下已有同名数据")
		return
	}

	ddInfo := &DDInfo{
		Id: ginbase.GenerateDataId(),
		SetName: form.SetName,
		Value: form.Value,
		Valid: true,
	}

	err = db.C(Opt.TbName).Insert(ddInfo)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 更新名字
type UpdateNameForm struct {
	SetName string `json:"setName" binding:"required"`
	OldValue string `json:"oldValue" binding:"required"`
	NewValue string `json:"newValue" binding:"required"`
}
func UpdateNameHandler(c *gin.Context) {
	var form UpdateNameForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if form.OldValue == form.NewValue {
		ginbase.ReturnErrJson(c, "新旧名字一致，无变动")
		return
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	// 检查旧名字是否存在
	oldinfo, err := GetNameBySetNameAndValue(db, form.SetName, form.OldValue)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	if oldinfo == nil {
		ginbase.ReturnErrJson(c, "不存在旧名字数据")
		return
	}

	update := bson.M{
		"$set": bson.M{
			"value": form.NewValue,
		},
	}

	err = db.C(Opt.TbName).UpdateId(oldinfo.Id, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "'")
	return
}

// 移除指定名字
type DelNameForm struct {
	SetName string `json:"setName" binding:"required"`
	Value string `json:"value" binding:"required"`
}
func DelNameHandler(c *gin.Context) {
	var form DelNameForm
	if err := c.BindJSON(&form); err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	f := bson.M{
		"setName": form.SetName,
		"value": form.Value,
	}

	update := bson.M{
		"$set": bson.M{
			"valid": false,
		},
	}

	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	err := db.C(Opt.TbName).Update(f, update)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	ginbase.ReturnOKJson(c, "")
	return
}

// 读取所有的名字
// 如果传递了 setName，读取指定 setName 的数据，如果没有，就读取所有的
func GetAllNamesHandler(c *gin.Context) {
	var allSetNames []string
	db := ginbase.NewDs(Opt.MgoOption)
	defer db.Close()

	setName := c.Query("setname")
	if setName == "" {
		allSetNames = []string{SET_KEY_NAME}
	} else {
		names, err := GetAllSetNames(db)
		if err != nil {
			ginbase.ReturnErrJson(c, err.Error())
			return
		}
		allSetNames = names
	}

	// 根据名字读取所有的数据
	f := bson.M{
		"setName": bson.M{
			"$in": allSetNames,
		},
	}

	var dds []*DDInfo
	err := db.C(Opt.TbName).Find(f).All(&dds)
	if err != nil {
		ginbase.ReturnErrJson(c, err.Error())
		return
	}

	type GroupedData struct {
		SetName string `json:"setName"`
		Values []string `json:"values"`
	}

	setNames := make(map[string][]string)

	for _, dd := range dds {
		values, ok := setNames[dd.SetName]
		if !ok {
			// 不存在新建数据
			setNames[dd.SetName] = []string{dd.Value}
		} else {
			values = append(values, dd.Value)
			setNames[dd.SetName] = values
		}
	}

	var retDatas []*GroupedData
	for k, v := range setNames {
		gd := &GroupedData{
			SetName: k,
			Values: v,
		}
		retDatas = append(retDatas, gd)
	}

	ginbase.ReturnOKJson(c, retDatas)
	return
}