package database

import (
	"fmt"
	"ginbase"
	"gopkg.in/mgo.v2"
)

type Ds struct {
	se *mgo.Session
	DbName string
}

var dbSession *Ds

func InitMongodbConn(conf ginbase.IMongoDbConf) error {
	url := conf.Url()
	session, err := mgo.Dial(url)
	if err != nil {
		e := fmt.Errorf("连接 mongodb 失败, %s", err.Error())
		fmt.Println(e.Error())
		return e
	}

	dbSession = &Ds{
		se: session,
		DbName: conf.DbName(),
	}

	return nil
}

func NewDs() *Ds {
	copySession := dbSession.se.Copy()
	ds := &Ds{
		se: copySession,
		DbName: dbSession.DbName,
	}

	return ds
}

func (ds *Ds) C(tbname string) *mgo.Collection {
	return ds.se.DB(ds.DbName).C(tbname)
}

func (ds *Ds) Close() {
	ds.se.Close()
}
