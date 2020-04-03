package roleapp

import (
	"github.com/gin-gonic/gin"
	. "github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/returnfun"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// 默认用户角色
func GetDefaultRole() *ChildRole {
	r := &ChildRole{
		Id:   DefaultRoleId,
		Name: DefaultRoleName,
	}
	return r
}

func isUnchangeable(g string) bool {
	if g == ItemGroupSystem {
		return true
	}
	return false
}

type SystemConfig struct {
	Items []*Item
	Ps []*Permission
	Roles []*Role
}

func (s *SystemConfig) addItem(item *Item) {

}

func (s *SystemConfig) addPermission(p *Permission) {

}

func (s *SystemConfig) addRole(p *Role) {

}

func PreCheckAuth(c *gin.Context) {
	user := GetCurUser(c)
	if user == nil {
		returnfun.ReturnJson(c, 417, 417, "服务器配置错误，调用本接口需要配置用户授权", "")
		return
	}
	c.Next()
}

// 生成默认用户的 role 存储到数据库
func insureDefaultRole(ds *dbandmq.Ds) error {
	Logger.Debug("", "初始化系统默认role")
	dbrole, err := GetRoleById(ds, DefaultRoleId, false)
	if err != nil {
		return err
	}
	if dbrole == nil {
		role := &Role{
			Id:            DefaultRoleId,
			Name:          DefaultRoleName,
			Deleted:       false,
			Source:        RoleDataSourceInternal,
			CreateT:       util.GetCurTime(),
		}
		role.UpdateT = role.CreateT
		err = ds.C(CollectionNameRole).Insert(role)
		return err
	}

	if dbrole.Name != DefaultRoleName {
		update := bson.M{
			"$set": bson.M{
				"name": DefaultRoleName,
				"updateT": util.GetCurTime(),
			},
		}

		err = ds.C(CollectionNameRole).UpdateId(dbrole.Id, update)
		return err
	}
	return nil
}

func saveItem(ds *dbandmq.Ds, item *Item) error {
	err := ds.C(CollectionNameItem).Insert(item)
	return err
}

func savePermission(ds *dbandmq.Ds, p *Permission) error {
	return ds.C(CollectionNamePermission).Insert(p)
}

func saveRole(ds *dbandmq.Ds, role *Role) error {
	return ds.C(CollectionNameRole).Insert(role)
}

func saveRoleAndUser(ds *dbandmq.Ds, r *RoleAndUser) error {
	return ds.C(CollectionNameRoleAndUser).Insert(r)
}

// 管理员那一套
func insureAdmin(ds *dbandmq.Ds) error {
	// 包括 item permission  role roleanduser
	Logger.Debug("", "初始化系统admin相关数据")
	var err error
	curT := util.GetCurTime()

	// item
	item := &Item{
		Id:      AdminItemId,
		Name:    AdminItemName,
		Method:  "*",
		Path:    "*",
		Group:   ItemGroupSystem,
		Deleted: false,
		Source:  RoleDataSourceInternal,
		CreateT: curT,
		UpdateT: curT,
	}
	_ , err = addItem(ds, item, KeyQueryId)
	if err != nil {
		return err
	}

	// permission
	p := &Permission{
		Id:      AdminPermissionId,
		Name:    AdminPermissionName,
		ItemIds: []string{AdminItemId},
		Deleted: false,
		Source:  RoleDataSourceInternal,
		CreateT: curT,
		UpdateT: curT,
	}

	err = addPermission(ds, p, KeyQueryId)
	if err != nil {
		return err
	}

	// role
	role := &Role{
		Id:            AdminRoleId,
		Name:          AdminRoleName,
		PermissionIds: []string{AdminPermissionId},
		ChildrenRoles: []*ChildRole{AdminChildRole},
		Deleted:       false,
		Source:        RoleDataSourceInternal,
		CreateT:       curT,
		UpdateT:       curT,
	}

	err = addRole(ds, role, KeyQueryId)
	if err != nil {
		return err
	}

	// superchild role
	err = insureSuperChildRole(ds)
	if err != nil {
		return err
	}

	// 关联 admin role 和 admin userid
	rau := &RoleAndUser{
		Id:       util.GenerateDataId(),
		UserId:   AdminUserId,
		UserName: AdminUserName,
		RoleIds:  []string{AdminRoleId},
		CreateT:  curT,
		UpdateT:  curT,
	}

	err = addRoleAndUser(ds, rau)
	if err != nil {
		return err
	}
	return nil
}

// 把 superchild role 存储到 role 里面
func insureSuperChildRole(ds *dbandmq.Ds) error {
	role := &Role{
		Id:            SuperChildRoleId,
		Name:          SuperChildRoleName,
		Deleted:       false,
		Source:        RoleDataSourceInternal,
		CreateT:       util.GetCurTime(),
	}
	role.UpdateT = role.CreateT
	err := addRole(ds, role, KeyQueryId)
	return err
}

// key 标记按照 id 还是 name 来查询唯一性
const (
	KeyQueryId = "id"
	KeyQueryName = "name"
)
func addItem(ds *dbandmq.Ds, item *Item, key string) (*Item, error) {
	// 检查数据库是否存在，存在就不管了
	var err error
	dbitem := &Item{}
	if key == KeyQueryId {
		dbitem, err = GetItemById(ds, item.Id)
	} else {
		dbitem, err = GetItemByName(ds, item.Name)
	}

	if err != nil {
		return nil, err
	}

	if dbitem != nil {
		return dbitem, nil
	}

	err = saveItem(ds, item)
	return item, err
}

func addPermission(ds *dbandmq.Ds, p *Permission, key string) error {
	var err error
	dbp := &Permission{}
	if key == KeyQueryId {
		dbp, err = GetPermissionById(ds, p.Id, false)
	} else {
		dbp, err = GetPermissionByName(ds, p.Name, false)
	}
	if err != nil {
		return err
	}

	if dbp != nil {
		return nil
	}

	return savePermission(ds, p)
}

func addRole(ds *dbandmq.Ds, r *Role, key string) error {
	var err error
	dbr := &Role{}
	if key == KeyQueryId {
		dbr, err = GetRoleById(ds, r.Id, false)
	} else {
		dbr, err = GetRoleByName(ds, r.Name, false)
	}
	if err != nil {
		return err
	}
	if dbr != nil {
		return nil
	}

	return saveRole(ds, r)
}

func addRoleAndUser(ds *dbandmq.Ds, rau *RoleAndUser) error {
	var err error
	dbr, err := GetRoleAndUserByUserId(ds, rau.UserId)
	if err != nil {
		return err
	}
	if dbr != nil {
		return nil
	}

	return saveRoleAndUser(ds, rau)
}

// 系统内部 item 接口
func insureRoleAppItems(ds *dbandmq.Ds, uriPrefix string) error {
	var items []*Item
	var err error
	curT := util.GetCurTime()

	// 新建 item
	item := generateItem(curT, "roleapp:createitem", "POST", uriPrefix + "/role/m/item")
	items = append(items, item)

	// 修改 item
	item = generateItem(curT, "roleapp:updateitem", "PUT", uriPrefix + "/role/m/item/:id")
	items = append(items, item)

	// 删除 item
	item = generateItem(curT, "roleapp:deleteitem", "DELETE", uriPrefix + "/role/m/item/:id")
	items = append(items, item)

	// 读取 item 明细
	item = generateItem(curT, "roleapp:getitem", "GET", uriPrefix + "/role/m/item/:id")
	items = append(items, item)

	// 搜索 items
	item = generateItem(curT, "roleapp:queryitem", "GET", uriPrefix + "/role/m/items")
	items = append(items, item)

	// 新建 permission
	item = generateItem(curT, "roleapp:createpermission", "POST", uriPrefix + "/role/m/permission")
	items = append(items, item)

	// 给 permission 增加 items
	item = generateItem(curT, "roleapp:additemstopermission", "POST", uriPrefix + "/role/m/permission/:id/additems")
	items = append(items, item)

	// 从 permission 移除 items
	item = generateItem(curT, "roleapp:delitemsfrompermission", "POST", uriPrefix + "/role/m/permission/:id/delitems")
	items = append(items, item)

	// 修改 permission 基本信息
	item = generateItem(curT, "roleapp:updatepermission", "PUT", uriPrefix + "/role/m/permission/:id")
	items = append(items, item)

	// 删除权限
	item = generateItem(curT, "roleapp:deletepermission", "DELETE", uriPrefix + "/role/m/permission/:id")
	items = append(items, item)

	// 读取权限明细
	item = generateItem(curT, "roleapp:getpermission", "GET", uriPrefix + "/role/m/permission/:id")
	items = append(items, item)

	// 搜索权限
	item = generateItem(curT, "roleapp:querypermission", "GET", uriPrefix + "/role/m/permissions")
	items = append(items, item)

	// 新建 role
	item = generateItem(curT, "roleapp:createrole", "POST", uriPrefix + "/role/m/role")
	items = append(items, item)

	// 给 role 添加 permission
	item = generateItem(curT, "roleapp:addpstorole", "POST", uriPrefix + "/role/m/role/:id/addps")
	items = append(items, item)

	// 从 role 中移除 permission
	item = generateItem(curT, "roleapp:delpsfromrole", "POST", uriPrefix + "/role/m/role/:id/delps")
	items = append(items, item)

	// 修改 role 信息
	item = generateItem(curT, "roleapp:updaterole", "PUT", uriPrefix + "/role/m/role/:id")
	items = append(items, item)

	// 删除 role
	item = generateItem(curT, "roleapp:deleterole", "DELETE", uriPrefix + "/role/m/role/:id")
	items = append(items, item)

	// 给 role 添加 childrole
	item = generateItem(curT, "roleapp:addchildroletorole", "POST", uriPrefix + "/role/m/role/:id/addchildrole")
	items = append(items, item)

	// 从 role 中移除 childrole
	item = generateItem(curT, "roleapp:delchildrolefromrole", "POST", uriPrefix + "/role/m/role/:id/delchildrole")
	items = append(items, item)

	// 查看 role 明细
	item = generateItem(curT, "roleapp:getrole", "GET", uriPrefix + "/role/m/role/:id")
	items = append(items, item)

	// 搜索 role
	item = generateItem(curT, "roleapp:queryrole", "GET", uriPrefix + "/role/m/roles")
	items = append(items, item)

	// 给 userid 添加 role
	item = generateItem(curT, "roleapp:addroletouser", "POST", uriPrefix + "/rau/addrole")
	items = append(items, item)

	var itemIds []string
	for _, item := range items {
		dbitem, err :=  addItem(ds, item, KeyQueryName)
		if err != nil {
			return err
		}
		itemIds = append(itemIds, dbitem.Id)
	}

	// api 管理员权限
	per := &Permission{
		Id:      ApiAdminPermissionId,
		Name:    ApiAdminPermissionName,
		ItemIds: itemIds,
		Deleted: false,
		Source:  RoleDataSourceInternal,
		CreateT: curT,
		UpdateT: curT,
	}
	err = addPermission(ds, per, KeyQueryId)
	if err != nil {
		return err
	}

	// api role
	role := &Role{
		Id:            ApiAdminRoleId,
		Name:          ApiAdminRoleName,
		PermissionIds: []string{ApiAdminPermissionId},
		ChildrenRoles: []*ChildRole{AdminChildRole},
		Deleted:       false,
		Source:        RoleDataSourceInternal,
		CreateT:       curT,
		UpdateT:       curT,
	}

	err = addRole(ds, role, KeyQueryId)
	if err != nil {
		return err
	}

	Logger.Debug("", "初始化系统 api item permission role 完成")
	return nil
}

func generateItem(t *util.CurTime, name, method, path string) *Item {
	if strings.Contains(path, ":id") {
		path = strings.ReplaceAll(path, ":id", "*")
	}
	method = strings.ToUpper(method)
	item := &Item{
		Id:      util.GenerateDataId(),
		Name:    name,
		Method:  method,
		Path:    path,
		Group:   ItemGroupSystem,
		Deleted: false,
		Source:  RoleDataSourceInternal,
		CreateT: t,
		UpdateT: t,
	}
	return item
}