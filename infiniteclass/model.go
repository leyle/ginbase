package infiniteclass

import (
	"fmt"
	"github.com/leyle/ginbase"
	. "github.com/leyle/gsimplelog"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const INFINITE_CLASS_ROOT_ID = "0"

type ClassOption struct {
	TbName string
	MgoOption *ginbase.MgoOption
}

var Opt *ClassOption

// 无限极分类
type InfiniteClass struct {
	Id string `json:"id" bson:"_id"`
	ParentId string `json:"parentId" bson:"parentId"`
	Name string `json:"name" bson:"name"`
	Icon string `json:"icon" bson:"icon"` // 图标
	Info string `json:"info" bson:"info"` // 描述信息
	Level int `json:"level" bson:"level"`
	Disable bool `json:"disable" bson:"disable"` // 是否禁用
	Children []*InfiniteClass `json:"children" bson:"-"` // 下一级
}

func (i *InfiniteClass) Desc() string {
	return fmt.Sprintf("Class id[%s],name[%s],icon[%s],info[%s], parentId[%s]", i.Id, i.Name, i.Icon, i.Info, i.ParentId)
}

// 根据 parentId 计算当前分类的 level 值
func CalcLevelByParentId(db *ginbase.Ds, pid string) (int, error) {
	if pid == INFINITE_CLASS_ROOT_ID {
		return 1, nil
	}
	ic, err := GetInfiniteClassById(db, pid)
	if err != nil {
		return -1, err
	}

	if ic == nil {
		return -1, errors.New("无指定id的父级分类信息")
	}

	pLevel := ic.Level
	return pLevel + 1, nil
}

// 新建分类
func NewInfiniteClass(db *ginbase.Ds, pid, name, icon, info string) (*InfiniteClass, error) {
	dbc, err := GetInfiniteClassByParentIdAndName(db, pid, name)
	if err != nil {
		return nil, err
	}

	if dbc != nil {
		e := fmt.Errorf("新建分类时，已存在父级为[%s]的子分类名[%s]", pid, name)
		Logger.Error(e.Error())
		return nil, e
	}

	if pid != INFINITE_CLASS_ROOT_ID {
		// 检查 pid 是否存在
		dbpc, err := GetInfiniteClassById(db, pid)
		if err != nil {
			return nil, err
		}

		if dbpc == nil {
			e := fmt.Errorf("新建分类时，不存在指定的父级[%s]信息", pid)
			Logger.Error(e.Error())
			return nil, e
		}

		if dbpc.Disable {
			e := fmt.Errorf("新建分类时，父级[%s][%s]分类已被禁用", pid, dbpc.Name)
			Logger.Error(e.Error())
			return nil, e
		}
	}

	curLevel, err := CalcLevelByParentId(db, pid)
	if err != nil {
		return nil, err
	}

	c := &InfiniteClass{
		Id: ginbase.GenerateDataId(),
		ParentId: pid,
		Name: name,
		Icon: icon,
		Info: info,
		Level: curLevel,
		Disable: false,
	}

	err = db.C(Opt.TbName).Insert(c)
	if err != nil {
		Logger.Errorf("新建%s失败,%s", c.Desc(), err.Error())
		return nil, err
	}

	return c, nil
}

// 根据 name 和 parentId 读取分类信息
func GetInfiniteClassByParentIdAndName(db *ginbase.Ds, pid, name string) (*InfiniteClass, error) {
	f := bson.M{
		"parentId": pid,
		"name": name,
	}

	var c *InfiniteClass
	err := db.C(Opt.TbName).Find(f).One(&c)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("根据ParentId[%s]和name[%s]查询分类信息失败, %s", pid, name, err.Error())
		return nil, err
	}

	return c, nil
}

// 根据 id 读取分类信息
func GetInfiniteClassById(db *ginbase.Ds, id string) (*InfiniteClass, error) {
	var ic *InfiniteClass
	err := db.C(Opt.TbName).FindId(id).One(&ic)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("根据id[%s]读取分类信息失败, %s", id, err.Error())
		return nil, err
	}

	return ic, nil
}

// 根据 parentId，递归读取其所有的下级
func QueryAllChildrenByParentClass(db *ginbase.Ds, pic *InfiniteClass) (err error) {
	var ics []*InfiniteClass
	f := bson.M{
		"parentId": pic.Id,
	}
	err = db.C(Opt.TbName).Find(f).All(&ics)
	if err != nil {
		Logger.Errorf("根据parent[%s][%s]读取其子分类失败, %s", pic.Id, pic.Name, err.Error())
		return err
	}

	for _, ic := range ics {
		pic.Children = append(pic.Children, ic)
		err = QueryAllChildrenByParentClass(db, ic)
		if err != nil {
			// 不暂停？还是跳出循环？todo
		}
	}

	return nil
}

// 读取指定 level 的分类
func QueryInfiniteClassByLevel(db *ginbase.Ds, level int, more bool) ([]*InfiniteClass, error) {
	var ics []*InfiniteClass
	var err error
	f := bson.M{
		"level": level,
	}

	err = db.C(Opt.TbName).Find(f).All(&ics)
	if err != nil {
		Logger.Errorf("根据level[%d]读取分类列表失败, %s", level, err.Error())
		return nil, err
	}

	if more {
		// 对于每一个分类，读取其所有的子类
		for _, ic := range ics {
			err = QueryAllChildrenByParentClass(db, ic)
			if err != nil {
				// todo
			}
		}
	}

	return ics, nil
}