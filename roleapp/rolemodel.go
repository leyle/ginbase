package roleapp

import (
	. "github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

func init() {
	dbandmq.AddIndexKey(IKItem)
	dbandmq.AddIndexKey(IKPermission)
	dbandmq.AddIndexKey(IKRole)
}

const DbPrefix = "role_"

// 系统内置 api 的 group 名称
const ItemGroupSystem = "system"

// 拥有这个就可以给任何用户赋予任何角色
// 其他 role 就只能给用户赋予自己附属的 childrole
const (
	SuperChildRoleId   = "5e86e466fa080a3ac0956db4"
	SuperChildRoleName = "superchildrole"
)

const (
	RoleDataSourceInternal = "SYSTEM" // 系统内部的数据
	RoleDataSourceApi      = "USER"   // 用户添加的数据
)

const DefaultRoleId = "5e85a88a22b9b93f458de2d8"

var DefaultRoleName = "registereduser" // 可修改

// 系统管理员的id 可修改
var (
	AdminUserId   = "5e86dc88fa080a3ac0956db0"
	AdminUserName = "admin"
)

const (
	AdminItemId   = "5e86dfa8fa080a3ac0956db6"
	AdminItemName = "adminItem"

	AdminPermissionId   = "5e86dfa8fa080a3ac0956db7"
	AdminPermissionName = "adminPermission"

	AdminRoleId   = "5e86dfa8fa080a3ac0956db8"
	AdminRoleName = "adminRole"
)

var AdminChildRole = &ChildRole{
	Id:   SuperChildRoleId,
	Name: SuperChildRoleName,
}

// api 管理员权限
const (
	ApiAdminPermissionId   = "5e86f751fa080a3ac0956db5"
	ApiAdminPermissionName = "sysApiManager"
)

// api 管理员角色
const (
	ApiAdminRoleId   = "5e86f81cfa080a3ac0956db6"
	ApiAdminRoleName = "sysApiRole"
)

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
	Method string `json:"method" bson:"method"`
	Path   string `json:"path" bson:"path"`
	Group  string `json:"group" bson:"group"` // 分组名字，属于哪一个功能模块

	Deleted bool `json:"deleted" bson:"deleted"`

	Source  string        `json:"source" bson:"source"`
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

	Deleted bool `json:"deleted" bson:"deleted"`

	Source  string        `json:"source" bson:"source"`
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

	Deleted bool `json:"deleted" bson:"deleted"`

	Source  string        `json:"source" bson:"source"`
	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// 记录值，归属于某个上层 role
type ChildRole struct {
	Id   string `json:"id" bson:"id"`     // role Id
	Name string `json:"name" bson:"name"` // role name，展示查看用
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
		Logger.Errorf("", "根据name[%s]读取 role item 失败, %s", name, err.Error())
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

// 根据 permissionIds 读取 permission 信息
func GetPermissionsByPermissionIds(db *dbandmq.Ds, pids []string) ([]*Permission, error) {
	f := bson.M{
		"deleted": false,
		"_id": bson.M{
			"$in": pids,
		},
	}

	var ps []*Permission
	err := db.C(CollectionNamePermission).Find(f).All(&ps)
	if err != nil {
		Logger.Errorf("", "根据permissionIds读取permission信息失败, %s", err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}
	if len(ps) == 0 {
		return ps, nil
	}

	wg := sync.WaitGroup{}
	finished := make(chan bool, 1)
	errChan := make(chan error, 1)
	for _, p := range ps {
		wg.Add(1)
		go fullPermission(&wg, db, p, errChan)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err = <-errChan:
		Logger.Errorf("", "查询permissions完整信息失败, %s", err.Error())
		return nil, err
	}

	return ps, nil
}

func fullPermission(wg *sync.WaitGroup, db *dbandmq.Ds, permission *Permission, errChan chan<- error) {
	defer wg.Done()
	ndb := db.CopyDs()
	defer ndb.Close()

	items, err := GetItemsByItemIds(ndb, permission.ItemIds)
	if err != nil {
		errChan <- err
		return
	}
	permission.Items = items
}

// 根据 name 读取 role
func GetRoleByName(db *dbandmq.Ds, name string, more bool) (*Role, error) {
	f := bson.M{
		"name": name,
	}

	var role *Role
	err := db.C(CollectionNameRole).Find(f).One(&role)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据role name[%s]读取role信息失败, %s", name, err.Error())
		return nil, middleware.ErrDbExec.Append(err.Error())
	}

	if role == nil {
		return nil, nil
	}

	if more {
		ps, err := GetPermissionsByPermissionIds(db, role.PermissionIds)
		if err == nil {
			role.Permissions = ps
		}
	}

	return role, nil
}

func GetRoleById(db *dbandmq.Ds, id string, more bool) (*Role, error) {
	var role *Role
	err := db.C(CollectionNameRole).FindId(id).One(&role)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据role id[%s]读取role信息失败, %s", id, err.Error())
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	if more {
		ps, err := GetPermissionsByPermissionIds(db, role.PermissionIds)
		if err == nil {
			role.Permissions = ps
		}
	}

	return role, nil
}

// 根据 roleId 列表读取完整的 roles 信息
func GetRolesByRoleIds(db *dbandmq.Ds, roleIds []string, more bool) ([]*Role, error) {
	if len(roleIds) > 1 {
		roleIds = util.UniqueStringArray(roleIds)
	}

	f := bson.M{
		"deleted": false,
		"_id": bson.M{
			"$in": roleIds,
		},
	}

	var roles []*Role
	err := db.C(CollectionNameRole).Find(f).All(&roles)
	if err != nil {
		Logger.Errorf("", "根据roleIds读取role信息失败, %s", err.Error())
		return nil, err
	}

	if !more {
		return roles, nil
	}

	// 并行的去完善 roles 信息
	wg := sync.WaitGroup{}
	finished := make(chan bool, 1)
	errChan := make(chan error, 1)

	for _, role := range roles {
		wg.Add(1)
		go fullRole(&wg, db, role, errChan)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err = <-errChan:
		Logger.Errorf("", "查询完整roles信息失败, %s", err.Error())
		return nil, err
	}

	return roles, nil
}

func fullRole(wg *sync.WaitGroup, db *dbandmq.Ds, role *Role, errChan chan<- error) {
	defer wg.Done()
	ndb := db.CopyDs()
	defer ndb.Close()

	ps, err := GetPermissionsByPermissionIds(ndb, role.PermissionIds)
	if err != nil {
		errChan <- err
		return
	}
	role.Permissions = ps
}
