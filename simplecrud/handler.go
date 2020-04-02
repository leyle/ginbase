package simplecrud

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/returnfun"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// 新建 simple data
type CreateForm struct {
	Value string `json:"value" binding:"required"`
}
func CreateSimpleDataHandler(c *gin.Context, ds *dbandmq.Ds) {
	var form CreateForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	value := strings.TrimSpace(form.Value)
	if value == "" {
		middleware.StopExec(ErrEmptyValue)
	}

	nds := ds.CopyDs()
	defer nds.Close()

	kp := getKeyPointer(c)

	// 需要校验 value 唯一性
	// 不考虑并发情况
	dbSd, err := kp.GetSimpleDataByName(nds, value)
	middleware.StopExec(err)
	if dbSd != nil {
		middleware.StopExec(ErrValueHasExist.Append(value))
	}

	sd := &SimpleData{
		Id:    util.GenerateDataId(),
		Key:   kp.Key,
		Value: value,
	}
	err = nds.C(CollectionNameSimpleData).Insert(sd)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, sd)
	return
}

// 修改 name
// id 和 name 必须有一个存在
// 优先使用 id
type UpdateForm struct {
	Id string `json:"id"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue" binding:"required"`
}
func UpdateSimpleDataHandler(c *gin.Context, db *dbandmq.Ds) {
	var form UpdateForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	kp := getKeyPointer(c)

	ds := db.CopyDs()
	defer ds.Close()

	dbSd := &SimpleData{}

	if form.Id != "" {
		dbSd, err = kp.GetSimpleDataById(ds, form.Id)
		middleware.StopExec(err)
	} else if form.OldValue != "" {
		dbSd, err = kp.GetSimpleDataByName(ds, form.OldValue)
		middleware.StopExec(err)
	}

	if dbSd == nil {
		middleware.StopExec(middleware.ErrNoIdData.Append(form.Id + form.OldValue))
	}
	update := bson.M{
		"$set": bson.M{
			"value": form.NewValue,
		},
	}

	err = ds.C(CollectionNameSimpleData).UpdateId(dbSd.Id, update)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}
	dbSd.Value = form.NewValue
	returnfun.ReturnOKJson(c, dbSd)
	return
}

// 删除指定 id 或 value 的 数据
// 两者取其中一个
type DeleteForm struct {
	Id string `json:"id"`
	Value string `json:"value"`
}
func DeleteSimpleDataHandler(c *gin.Context, db *dbandmq.Ds) {
	var form DeleteForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	kp := getKeyPointer(c)
	ds := db.CopyDs()
	defer ds.Close()

	dbSd := &SimpleData{}

	if form.Id != "" {
		dbSd, err = kp.GetSimpleDataById(ds, form.Id)
		middleware.StopExec(err)
	} else if form.Value != "" {
		dbSd, err = kp.GetSimpleDataByName(ds, form.Value)
		middleware.StopExec(err)
	}
	if dbSd == nil {
		middleware.StopExec(middleware.ErrNoIdData.Append(form.Id + form.Value))
	}

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
		},
	}

	err = ds.C(CollectionNameSimpleData).UpdateId(dbSd.Id, update)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}
	returnfun.ReturnOKJson(c, "")
	return
}

// 根据 id 读取对应的 name 数据
func GetSimpleDataByIdHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")
	kp := getKeyPointer(c)

	ds := db.CopyDs()
	defer ds.Close()

	sd, err := kp.GetSimpleDataById(ds, id)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, sd)
	return
}

// 搜索 simpledata
func QuerySimpleDataHandler(c *gin.Context, db *dbandmq.Ds) {
	kp := getKeyPointer(c)
	var andCondition []bson.M
	andCondition = append(andCondition, bson.M{"key": kp.Key})

	value := c.Query("value")
	if value != "" {
		andCondition = append(andCondition, bson.M{"value": bson.M{"$regex": value}})
	}

	query := bson.M{
		"$and": andCondition,
	}

	ds := db.CopyDs()
	defer ds.Close()

	Q := ds.C(CollectionNameSimpleData).Find(query)
	total, err := Q.Count()
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	var sds []*SimpleData

	page, size, skip := util.GetPageAndSize(c)
	err = Q.Sort("value").Skip(skip).Limit(size).All(&sds)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	ret := &returnfun.QueryListData{
		Total: total,
		Page:  page,
		Size:  size,
		Data:  sds,
	}

	returnfun.ReturnOKJson(c, ret)
	return
}

