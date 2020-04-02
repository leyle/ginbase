package simplecrud

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	. "github.com/leyle/ginbase/consolelog"
)

func init() {
	dbandmq.AddIndexKey(IKSimpleData)
}

var DbPrefix = "simplecrud_"

// error code
var ErrEmptyValue = &middleware.CustomErrStruct{
	Code: 30001,
	Msg:  "Value is null",
}

// name exist
var ErrValueHasExist = &middleware.CustomErrStruct{
	Code: 30002,
	Msg:  "Name has exist: ",
}

var CollectionNameSimpleData = DbPrefix + "simpledata"
var IKSimpleData = &dbandmq.IndexKey{
	Collection:    CollectionNameSimpleData,
	SingleKey:     []string{"key", "value", "deleted"},
	CompositeKeys: [][]string{[]string{"key", "name", "deleted"}},
}
type SimpleData struct {
	Id string `json:"id" bson:"_id"`
	Key string `json:"-" bson:"key"`
	Value string `json:"value" bson:"value"`
	Deleted bool `json:"-" bson:"deleted"`
}

type KeyPointer struct {
	Key string
}

func getKeyPointer(c *gin.Context) *KeyPointer {
	key := c.Param("key")
	return &KeyPointer{Key: key}
}

func (k *KeyPointer) GetSimpleDataByName(ds *dbandmq.Ds, name string) (*SimpleData, error) {
	var sd *SimpleData
	f := bson.M{
		"key": k.Key,
		"value": name,
		"deleted": false,
	}

	err := ds.C(CollectionNameSimpleData).Find(f).One(&sd)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "按name[%s]查询simpledata失败, %s", name, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}
	fmt.Println(sd)

	return sd, nil
}

func (k *KeyPointer) GetSimpleDataById(ds *dbandmq.Ds, id string) (*SimpleData, error) {
	var sd *SimpleData
	f := bson.M{
		"_id": id,
		"deleted": false,
	}
	err := ds.C(CollectionNameSimpleData).Find(f).One(&sd)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "按id[%s]查询simpledata失败, %s", id, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	return sd, nil
}