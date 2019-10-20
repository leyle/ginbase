package dbandmq

import (
	"fmt"
	. "github.com/leyle/ginbase/consolelog"
	"gopkg.in/mgo.v2"
)

type MgoOption struct {
	Host string
	Port string
	User string
	Passwd string
	Database string
}

func (opt *MgoOption) String() string {
	return fmt.Sprintf("[%s:%s][%s:%s]%s", opt.Host, opt.Port, opt.User, opt.Passwd, opt.Database)
}

type Ds struct {
	Se *mgo.Session
	Opt *MgoOption
}

func (opt *MgoOption) ConnectUrl() string {
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?connect=direct", opt.User, opt.Passwd, opt.Host, opt.Port, opt.Database)
	return url
}

type IndexKey struct {
	Collection string
	SingleKey []string
	CompositeKeys [][]string
	UniqueKey []string
}

var IndexKeys = []*IndexKey{}

var ds *Ds = nil

func InitMongodbSession(opt *MgoOption) error {
	if ds == nil {
		url := opt.ConnectUrl()
		session, err := mgo.Dial(url)
		if err != nil {
			Logger.Errorf("", "初始化连接 mgo 失败,[%s], %s", url, err.Error())
			return err
		}

		ds = &Ds{
			Se: session,
			Opt: opt,
		}
	}

	Logger.Debugf("", "初始化连接 mongodb[%s]成功", opt.String())
	return nil
}

func NewDs(opt *MgoOption) *Ds {
	if ds == nil {
		// 就不用先初始化了
		err := InitMongodbSession(opt)
		if err != nil {
			panic(err)
		}
	}

	se := ds.Se.Copy()

	newDs := &Ds{
		Se: se,
		Opt: opt,
	}

	return newDs
}

func (d *Ds) Close() {
	d.Se.Close()
}

// 为什么不直接叫 Copy，为了避免自动补全时，看错了，把 Copy Close 搞混
func (d *Ds) CopyDs() *Ds {
	se := ds.Se.Copy()
	newDs := &Ds{
		Se:  se,
		Opt: ds.Opt,
	}
	return newDs
}

func (d *Ds) C(collection string) *mgo.Collection {
	return d.Se.DB(d.Opt.Database).C(collection)
}

// 创建单键索引
// 传入的列表中的每一个字段创建一个索引
func (d *Ds)InsureSingleIndex(collection string, keys []string) error {
	Logger.Debugf("", "Insure mongodb [%s] index, %s, starting...", collection, keys)
	var err error
	for _, key := range keys {
		err = d.C(collection).EnsureIndexKey(key)
		if err != nil {
			Logger.Errorf("", "create [%s] index [%s] failed, %s", collection, key, err.Error())
			return err
		}
	}
	Logger.Debugf("", "Insure mongodb [%s] index, %s, done", collection, keys)
	return nil
}

// 创建复合索引
func (d *Ds)InsureCompositeIndex(collection string, keys []string) error {
	Logger.Debugf("", "Insure mongodb [%s] index, %s, starting...", collection, keys)
	err := d.C(collection).EnsureIndexKey(keys...)
	if err != nil {
		Logger.Errorf("", "create [%s] index [%s] failed, %s", collection, keys, err.Error())
		return err
	}
	Logger.Debugf("", "Insure mongodb [%s] index, %s, done", collection, keys)
	return nil
}

func (d *Ds)InsureUniqueIndex(collection string, keys []string) error {
	Logger.Debugf("", "Insure mongodb [%s] unique index, %s, starting...", collection, keys)
	var err error
	for _, key := range keys {
		err = d.C(collection).EnsureIndex(mgo.Index{
			Key: []string{key},
			Unique: true,
		})

		if err != nil {
			Logger.Errorf("", "create [%s] unqiue index [%s] failed, %s", collection, key, err.Error())
			return err
		}
	}
	Logger.Debugf("", "Insure mongodb [%s] unique index, %s, done", collection, keys)
	return nil
}

func (d *Ds)InsureCollectionKeys() error {
	var err error
	for _, ik := range IndexKeys {
		name := ik.Collection
		if len(ik.SingleKey) > 0 {
			err = d.InsureSingleIndex(name, ik.SingleKey)
			if err != nil {
				return err
			}
		}

		if len(ik.CompositeKeys) > 0 {
			for _, ckey := range ik.CompositeKeys {
				err = d.InsureCompositeIndex(name, ckey)
				if err != nil {
					return err
				}
			}
		}

		if len(ik.UniqueKey) > 0 {
			err = d.InsureUniqueIndex(name, ik.UniqueKey)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
