package datadictionary

import (
	"github.com/globalsign/mgo"
	"github.com/leyle/ginbase"
	"gopkg.in/mgo.v2/bson"
	. "github.com/leyle/gsimplelog"
)

// 数据字典维护功能
type DDOption struct {
	TbName string // mongodb 的 collection 名字
	MgoOption *ginbase.MgoOption
}

var Opt *DDOption

// 数据字典维护分类两部分，一部分是集合/分类的维护
// 另外一部分是对这些集合/分类下的名字的维护


// 默认 key，可以作为所有的其他的 key 的 key
const SET_KEY_NAME = "SET"

type DDInfo struct {
	Id string `json:"-" bson:"_id"`
	SetName string `json:"setName" bson:"setName"` // key
	Value string `json:"value" bson:"value"` // value
	Valid bool `json:"-" bson:"valid"`
}

func GetSetNameByValue(db *ginbase.Ds, value string) (*DDInfo, error) {
	f := bson.M{
		"setName": SET_KEY_NAME,
		"value": value,
		"valid": true,
	}

	var d *DDInfo
	err := db.C(Opt.TbName).Find(f).One(&d)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("根据set[%s]读取其拥有的成员名字[%s]查询字典数据失败, %s", SET_KEY_NAME, value, err.Error())
		return nil, err
	}

	return d, nil
}

func GetNameBySetNameAndValue(db *ginbase.Ds, setName, value string) (*DDInfo, error) {
	f := bson.M{
		"setName": setName,
		"value": value,
		"valid": true,
	}

	var d *DDInfo
	err := db.C(Opt.TbName).Find(f).One(&d)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("根据set[%s]读取其拥有的成员名字[%s]查询字典数据失败, %s", setName, value, err.Error())
		return nil, err
	}
	return d, nil
}

func GetAllSetNames(db *ginbase.Ds) ([]string, error) {
	f := bson.M{
		"setName": SET_KEY_NAME,
		"valid": true,
	}

	var dds []DDInfo
	err := db.C(Opt.TbName).Find(f).All(&dds)
	if err != nil {
		return nil, err
	}

	var sets []string
	for _, dd := range dds {
		sets = append(sets, dd.Value)
	}

	return sets, nil
}
