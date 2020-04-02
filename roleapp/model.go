package roleapp

import (
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2"
	. "github.com/leyle/ginbase/consolelog"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	dbandmq.AddIndexKey(IKItem)
	dbandmq.AddIndexKey(IKPermission)
	dbandmq.AddIndexKey(IKRole)
	dbandmq.AddIndexKey(IKApiGroup)
}

const DbPrefix = "role_"

// id 和 name 都叫这个，拥有这个就可以给任何用户赋予任何角色
// 其他 role 就只能给用户赋予自己附属的 childrole
const SuperChildRole = "superchildrole"

const (
	RoleDataSourceInternal = "SYSTEM" // 系统内部的数据
	RoleDataSourceApi = "USER" // 用户添加的数据
)

const (
	AdminRoleName       = "admin"
	AdminPermissionName = "admin"
	AdminItemName       = "admin:"
)

var DefaultRoleName = "registereduser" // 可修改
var DefaultRoleId = "5e85a88a22b9b93f458de2d8"

var AdminItemNames = []string{
	AdminItemName + "GET",
	AdminItemName + "POST",
	AdminItemName + "PUT",
	AdminItemName + "DELETE",
	AdminItemName + "PATCH",
	AdminItemName + "OPTION",
	AdminItemName + "HEAD",
}

// item
const CollectionNameItem = DbPrefix + "item"
var IKItem = &dbandmq.IndexKey{
	Collection: CollectionNameItem,
	SingleKey:  []string{"method", "path", "group", "source", "deleted"},
	UniqueKey:  []string{"name"},
}
type Item struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"` // 名字要求唯一，目的是避免在系统内造成脏数据

	// api
	Method   string `json:"method" bson:"method"`
	Path     string `json:"path" bson:"path"`
	Group string `json:"group" bson:"group"` // 分组名字，属于哪一个功能模块

	Deleted bool `json:"deleted" bson:"deleted"`

	Source string `json:"source" bson:"source"`
	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// permission
const CollectionNamePermission = DbPrefix + "permission"
var IKPermission = &dbandmq.IndexKey{
	Collection: CollectionNamePermission,
	SingleKey:  []string{"itemIds", "deleted", "source"},
	UniqueKey:  []string{"name"},
}
type Permission struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`

	ItemIds []string `json:"-" bson:"itemIds"`
	Items   []*Item  `json:"items" bson:"-"`

	Deleted  bool   `json:"deleted" bson:"deleted"`

	Source string `json:"source" bson:"source"`
	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// role
const CollectionNameRole = DbPrefix + "role"
var IKRole = &dbandmq.IndexKey{
	Collection: CollectionNameRole,
	SingleKey:  []string{"permissionIds", "deleted", "source"},
	UniqueKey:  []string{"name"},
}
type Role struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`

	PermissionIds []string      `json:"-" bson:"permissionIds"`
	Permissions   []*Permission `json:"permissions" bson:"-"`

	// 包含的下属 role 列表，当前 role 所属用户可以给自己的下属用户赋予的权限
	ChildrenRoles []*ChildRole `json:"childrenRole" bson:"childrenRole"`

	Deleted  bool   `json:"deleted" bson:"deleted"`

	Source string `json:"source" bson:"source"`
	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// 记录值，归属于某个上层 role
type ChildRole struct {
	Id   string `json:"id" bson:"id"`     // role Id
	Name string `json:"name" bson:"name"` // role name，展示查看用
}

// api group 管理
const CollectionNameApiGroup = DbPrefix + "apigroup"
var IKApiGroup = &dbandmq.IndexKey{
	Collection:    CollectionNameApiGroup,
	UniqueKey:     []string{"name"},
}
type ApiGroup struct {
	Id string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

// 根据 id 读取 item
func GetItemById(ds *dbandmq.Ds, id string) (*Item, error) {
	var item *Item
	err := ds.C(CollectionNameItem).FindId(id).One(&item)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据id[%s]读取 item 信息失败, %s", id, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}
	return item, nil
}

// 根据 name 读取 item
func GetItemByName(ds *dbandmq.Ds, name string) (*Item, error) {
	f := bson.M{
		"name": name,
	}

	var item *Item
	err := ds.C(CollectionNameItem).Find(f).One(&item)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据name[%s]读取 role item 失败, %s", err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	return item, nil
}

// 根据 itemIds 读取 items 信息
func GetItemsByItemIds(db *dbandmq.Ds, itemIds []string) ([]*Item, error) {
	if len(itemIds) == 0 {
		return nil, nil
	}

	f := bson.M{
		"deleted": false,
		"_id": bson.M{
			"$in": itemIds,
		},
	}

	var items []*Item
	err := db.C(CollectionNameItem).Find(f).All(&items)
	if err != nil {
		Logger.Errorf("", "根据itemIds读取item信息失败, %s", err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	return items, nil
}

// 根据 name 读取 permission
func GetPermissionByName(db *dbandmq.Ds, name string, more bool) (*Permission, error) {
	f := bson.M{
		"name": name,
	}

	var p *Permission
	err := db.C(CollectionNamePermission).Find(f).One(&p)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据permission name[%s]读取permission信息失败, %s", name, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	if p == nil {
		return nil, nil
	}

	if more {
		items, err := GetItemsByItemIds(db, p.ItemIds)
		if err == nil {
			p.Items = items
		}
	}

	return p, nil
}

func GetPermissionById(db *dbandmq.Ds, id string, more bool) (*Permission, error) {
	var p *Permission
	err := db.C(CollectionNamePermission).FindId(id).One(&p)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据 permission id[%s]读取permission信息失败, %s", id, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	if p == nil {
		return nil, nil
	}

	if more {
		items, err := GetItemsByItemIds(db, p.ItemIds)
		if err == nil {
			p.Items = items
		}
	}

	return p, nil
}