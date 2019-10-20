package dbandmq

import (
	"github.com/leyle/ginbase/util"
	"testing"
)

func TestIndexKey_Add(t *testing.T) {
	var err error
	opt := &MgoOption{
		Host:     "192.168.2.136",
		Port:     "27018",
		User:     "test",
		Passwd:   "test",
		Database: "testrole",
	}

	err = InitMongodbSession(opt)
	if err != nil {
		t.Error(err)
	}

	AddIndexKey(ik)

	db := NewDs(opt)
	defer db.Close()

	err = db.C(CollectionNameSomeData).Insert(&SomeData{
		Id: util.GenerateDataId(),
		KeyA: "keyadata",
		UKeyB: "unique key b",
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = ds.InsureCollectionKeys()
	if err != nil {
		t.Error(err)
	}


}

const CollectionNameSomeData = "someData"
var ik = &IndexKey{
	Collection:    CollectionNameSomeData,
	SingleKey:     []string{"keyA"},
	UniqueKey:     []string{"uKeyB"},
}



type SomeData struct {
	Id string `json:"id" bson:"_id"`
	KeyA string `json:"keyA" bson:"keyA"`
	UKeyB string `json:"uKeyB" bson:"uKeyB"`
}