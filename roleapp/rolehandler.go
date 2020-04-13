package roleapp

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/returnfun"
	"github.com/leyle/ginbase/util"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// item manage handlers
// 新建 item
type CreateItemForm struct {
	Name   string `json:"name" binding:"required"`
	Method string `json:"method" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Group  string `json:"group" binding:"required"` // 属于哪个分组
}

func CreateItemHandler(c *gin.Context, db *dbandmq.Ds) {
	var form CreateItemForm
	var err error
	err = c.BindJSON(&form)
	middleware.StopExec(err)

	ds := db.CopyDs()
	defer ds.Close()

	name := strings.TrimSpace(form.Name)

	dbitem, err := GetItemByName(ds, name)
	middleware.StopExec(err)

	if dbitem != nil {
		returnfun.ReturnErrJson(c, "name已存在")
		return
	}

	if strings.Contains(form.Path, ":id") {
		form.Path = strings.ReplaceAll(form.Path, ":id", "*")
	}

	item := &Item{
		Id:      util.GenerateDataId(),
		Name:    form.Name,
		Method:  strings.ToUpper(form.Method),
		Path:    form.Path,
		Group:   form.Group,
		Deleted: false,
		Source:  RoleDataSourceApi,
		CreateT: util.GetCurTime(),
	}
	item.UpdateT = item.CreateT

	err = ds.C(CollectionNameItem).Insert(item)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, item)
	return
}

// 修改 item
type UpdateItemForm struct {
	Name   string `json:"name" binding:"required"`
	Method string `json:"method" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Group  string `json:"group" binding:"required"` // 属于哪个分组
}

func UpdateItemHandler(c *gin.Context, db *dbandmq.Ds) {
	var form UpdateItemForm
	var err error
	err = c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	dbitem, err := GetItemById(ds, id)
	middleware.StopExec(err)

	if dbitem == nil {
		middleware.StopExec(middleware.ErrNoIdData.Append(id))
	}

	if strings.Contains(form.Path, ":id") {
		form.Path = strings.ReplaceAll(form.Path, ":id", "*")
	}

	dbitem.Name = form.Name
	dbitem.Method = form.Method
	dbitem.Path = form.Path
	dbitem.Group = form.Group
	dbitem.Deleted = false
	dbitem.UpdateT = util.GetCurTime()

	filter := bson.M{
		"_id":    id,
		"source": RoleDataSourceApi,
	}

	err = ds.C(CollectionNameItem).Update(filter, dbitem)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, dbitem)
	return
}

// 删除 item
func DeleteItemHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")
	ds := db.CopyDs()
	defer ds.Close()

	filter := bson.M{
		"_id":    id,
		"source": RoleDataSourceApi,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
			"updateT": util.GetCurTime(),
		},
	}

	err := ds.C(CollectionNameItem).Update(filter, update)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, "")
}

// 根据 id 读取 item 信息
func GetItemInfoHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")
	ds := db.CopyDs()
	defer ds.Close()

	item, err := GetItemById(ds, id)
	middleware.StopExec(err)
	returnfun.ReturnOKJson(c, item)
	return
}

func QueryItemHandler(c *gin.Context, db *dbandmq.Ds) {
	var andCondition []bson.M

	// 过滤掉 admin
	andCondition = append(andCondition, bson.M{"name": bson.M{"$ne": AdminItemName}})

	name := c.Query("name")
	if name != "" {
		andCondition = append(andCondition, bson.M{"name": bson.M{"$regex": name}})
	}

	path := c.Query("path")
	if path != "" {
		andCondition = append(andCondition, bson.M{"path": bson.M{"$regex": path}})
	}

	method := c.Query("method")
	if method != "" {
		method = strings.ToUpper(method)
		andCondition = append(andCondition, bson.M{"method": method})
	}

	group := c.Query("group")
	if group != "" {
		andCondition = append(andCondition, bson.M{"group": group})
	}

	deleted := c.Query("deleted")
	if deleted != "" {
		deleted = strings.ToUpper(deleted)
		if deleted == "TRUE" {
			andCondition = append(andCondition, bson.M{"deleted": true})
		} else {
			andCondition = append(andCondition, bson.M{"deleted": false})
		}
	}

	query := bson.M{}
	if len(andCondition) > 0 {
		query = bson.M{
			"$and": andCondition,
		}
	}

	ds := db.CopyDs()
	defer ds.Close()

	Q := ds.C(CollectionNameItem).Find(query)
	total, err := Q.Count()
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	var items []*Item
	page, size, skip := util.GetPageAndSize(c)
	err = Q.Sort("-_id").Skip(skip).Limit(size).All(&items)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	ret := returnfun.QueryListData{
		Total: total,
		Page:  page,
		Size:  size,
		Data:  items,
	}

	returnfun.ReturnOKJson(c, ret)
	return
}

// permission manage handlers
// 新建一个 permission 容器
type CreatePermissionForm struct {
	Name    string   `json:"name" binding:"required"`
	ItemIds []string `json:"itemIds"` // 不是必选的
}

func CreatePermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	var form CreatePermissionForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	// 检查名字是否存在，不加锁
	ds := db.CopyDs()
	defer ds.Close()

	name := strings.TrimSpace(form.Name)
	dbp, err := GetPermissionByName(ds, name, false)
	middleware.StopExec(err)

	if dbp != nil {
		returnfun.ReturnErrJson(c, "name已存在")
		return
	}

	permission := &Permission{
		Id:      util.GenerateDataId(),
		Name:    form.Name,
		ItemIds: form.ItemIds,
		Deleted: false,
		Source:  RoleDataSourceApi,
		CreateT: util.GetCurTime(),
	}
	permission.UpdateT = permission.CreateT

	err = ds.C(CollectionNamePermission).Insert(permission)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, permission)
	return
}

// 给 permission 添加 items
type AddItemsToPermissionForm struct {
	ItemIds []string `json:"itemIds" binding:"required"`
}

func AddItemsToPermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	var form AddItemsToPermissionForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	dbp, err := GetPermissionById(ds, id, false)
	middleware.StopExec(err)

	if dbp == nil || dbp.Deleted {
		middleware.StopExec(middleware.ErrNoIdData.Append(id))
	}

	dbp.ItemIds = append(dbp.ItemIds, form.ItemIds...)
	dbp.ItemIds = util.UniqueStringArray(dbp.ItemIds)
	dbp.UpdateT = util.GetCurTime()

	err = ds.C(CollectionNamePermission).UpdateId(dbp.Id, dbp)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, dbp)
	return
}

// 把 items 从 permission 中移除
type RemoveItemFromPermissionForm struct {
	ItemIds []string `json:"itemIds" binding:"required"`
}

func RemoveItemsFromPermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	var form RemoveItemFromPermissionForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	dbp, err := GetPermissionById(ds, id, false)
	middleware.StopExec(err)
	if dbp == nil || dbp.Deleted {
		middleware.StopExec(middleware.ErrNoIdData.Append(err.Error()))
	}

	var remainIds []string
	for _, dbpId := range dbp.ItemIds {
		remain := true
		for _, rid := range form.ItemIds {
			if dbpId == rid {
				remain = false
				break
			}
		}

		if remain {
			remainIds = append(remainIds, dbpId)
		}
	}

	dbp.ItemIds = remainIds
	dbp.UpdateT = util.GetCurTime()

	err = ds.C(CollectionNamePermission).UpdateId(dbp.Id, dbp)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, dbp)
	return
}

// 修改 permission 基本信息
type UpdatePermissionForm struct {
	Name string `json:"name" binding:"required"`
}

func UpdatePermissionInfoHandler(c *gin.Context, db *dbandmq.Ds) {
	var form UpdatePermissionForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")
	name := strings.TrimSpace(form.Name)

	ds := db.CopyDs()
	defer ds.Close()

	update := bson.M{
		"$set": bson.M{
			"name":    name,
			"deleted": false,
			"updateT": util.GetCurTime(),
		},
	}

	err = ds.C(CollectionNamePermission).UpdateId(id, update)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}
	returnfun.ReturnOKJson(c, "")
	return
}

// 删除 permission
func DeletePermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")
	ds := db.CopyDs()
	defer ds.Close()

	filter := bson.M{
		"_id":    id,
		"source": RoleDataSourceApi,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
			"updateT": util.GetCurTime(),
		},
	}

	err := ds.C(CollectionNamePermission).Update(filter, update)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}
	returnfun.ReturnOKJson(c, "")
	return
}

// 读取 permission 信息，包含 items
func GetPermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")
	ds := db.CopyDs()
	defer ds.Close()

	p, err := GetPermissionById(ds, id, true)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, p)
	return
}

// 搜索 permission 列表
func QueryPermissionHandler(c *gin.Context, db *dbandmq.Ds) {
	var andCondition []bson.M

	// 过滤掉 admin
	andCondition = append(andCondition, bson.M{"name": bson.M{"$ne": AdminPermissionName}})

	name := c.Query("name")
	if name != "" {
		andCondition = append(andCondition, bson.M{"name": bson.M{"$regex": name}})
	}

	deleted := c.Query("deleted")
	if deleted != "" {
		deleted = strings.ToUpper(deleted)
		if deleted == "TRUE" {
			andCondition = append(andCondition, bson.M{"deleted": true})
		} else {
			andCondition = append(andCondition, bson.M{"deleted": false})
		}
	}

	query := bson.M{}
	if len(andCondition) > 0 {
		query = bson.M{
			"$and": andCondition,
		}
	}

	ds := db.CopyDs()
	defer ds.Close()

	Q := ds.C(CollectionNamePermission).Find(query)
	total, err := Q.Count()
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	var ps []*Permission
	page, size, skip := util.GetPageAndSize(c)
	err = Q.Sort("-_id").Skip(skip).Limit(size).All(&ps)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	retData := returnfun.QueryListData{
		Total: total,
		Page:  page,
		Size:  size,
		Data:  ps,
	}

	returnfun.ReturnOKJson(c, retData)
	return
}

// 新建 role
type CreateRoleForm struct {
	Name     string     `json:"name" binding:"required"`
	Pids     []string   `json:"pids"`     // 可以没有值
	SubRoles []*SubRole `json:"subRoles"` // 可以无值
}

func CreateRoleHandler(c *gin.Context, db *dbandmq.Ds) {
	var form CreateRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	ds := db.CopyDs()
	defer ds.Close()

	name := strings.TrimSpace(form.Name)
	dbrole, err := GetRoleByName(ds, name, false)
	middleware.StopExec(err)

	if dbrole != nil {
		returnfun.ReturnErrJson(c, "name已存在")
		return
	}

	role := &Role{
		Id:            util.GenerateDataId(),
		Name:          name,
		PermissionIds: form.Pids,
		SubRoles:      form.SubRoles,
		Deleted:       false,
		Source:        RoleDataSourceApi,
		CreateT:       util.GetCurTime(),
	}
	role.UpdateT = role.CreateT

	err = ds.C(CollectionNameRole).Insert(role)
	if err != nil {
		middleware.StopExec(middleware.ErrDbExec.Append(err.Error()))
	}

	returnfun.ReturnOKJson(c, role)
	return
}

// 给 role 添加 permission
type AddPToRoleForm struct {
	Pids []string `json:"pids" binding:"required"`
}

func AddPermissionsToRoleHandler(c *gin.Context, db *dbandmq.Ds) {
	var form AddPToRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	dbrole, err := GetRoleById(ds, id, false)
	middleware.StopExec(err)

	if dbrole == nil || dbrole.Deleted {
		middleware.StopExec(middleware.ErrNoIdData.Append(id))
	}

	// 检查 pids 的合法性 todo
	dbrole.PermissionIds = append(dbrole.PermissionIds, form.Pids...)
	dbrole.PermissionIds = util.UniqueStringArray(dbrole.PermissionIds)
	dbrole.UpdateT = util.GetCurTime()

	err = ds.C(CollectionNameRole).UpdateId(id, dbrole)
	middleware.StopExec(err)
	returnfun.ReturnOKJson(c, dbrole)
	return
}

// 从 role 中移除 permission
type RemovePFromRoleForm struct {
	Pids []string `json:"pids" binding:"required"`
}

func RemovePermissionsFromRoleHandler(c *gin.Context, db *dbandmq.Ds) {
	var form RemovePFromRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")

	ds := db.CopyDs()
	defer ds.Close()

	dbrole, err := GetRoleById(ds, id, false)
	middleware.StopExec(err)

	if dbrole == nil || dbrole.Deleted {
		middleware.StopExec(middleware.ErrNoIdData.Append(id))
	}

	// 检查 pids 的合法性
	var remainPids []string
	for _, dbpId := range dbrole.PermissionIds {
		remain := true
		for _, pid := range form.Pids {
			if dbpId == pid {
				remain = false
				break
			}
		}

		if remain {
			remainPids = append(remainPids, dbpId)
		}
	}

	dbrole.PermissionIds = remainPids
	dbrole.UpdateT = util.GetCurTime()

	err = ds.C(CollectionNameRole).UpdateId(id, dbrole)
	middleware.StopExec(err)
	returnfun.ReturnOKJson(c, dbrole)
	return
}

// 修改 role 信息
type UpdateRoleForm struct {
	Name string `json:"name" binding:"required"`
}

func UpdateRoleInfoHandler(c *gin.Context, db *dbandmq.Ds) {
	var form UpdateRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	id := c.Param("id")
	name := strings.TrimSpace(form.Name)

	ds := db.CopyDs()
	defer ds.Close()

	update := bson.M{
		"$set": bson.M{
			"name":    name,
			"deleted": false,
			"updateT": util.GetCurTime(),
		},
	}

	err = ds.C(CollectionNameRole).UpdateId(id, update)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, "")
	return
}

// 删除 role
func DeleteRoleHandler(c *gin.Context, db *dbandmq.Ds) {
	id := c.Param("id")

	if id == DefaultRoleId {
		returnfun.Return403Json(c, "cannot delete this data")
		return
	}

	filter := bson.M{
		"_id":    id,
		"source": RoleDataSourceApi,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
			"updateT": util.GetCurTime(),
		},
	}

	ds := db.CopyDs()
	defer ds.Close()

	err := ds.C(CollectionNameRole).Update(filter, update)
	middleware.StopExec(err)
	returnfun.ReturnOKJson(c, "")
	return
}

// 给 role 添加 subrole
type SubRoleForm struct {
	Roles []*SubRole `json:"subRoles" binding:"required"`
}

func AddSubRolesToRoleHandler(c *gin.Context, ds *dbandmq.Ds) {
	var form SubRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	roleId := c.Param("id")
	db := ds.CopyDs()
	defer db.Close()

	dbRole, err := GetRoleById(db, roleId, false)
	middleware.StopExec(err)
	if dbRole == nil {
		returnfun.ReturnErrJson(c, "无指定id的role信息")
		return
	}
	if dbRole.Deleted {
		returnfun.ReturnErrJson(c, "角色已被删除，要修改请先恢复此角色")
		return
	}

	// 检查要添加的 roleId 的有效性
	var roleIds []string
	for _, r := range form.Roles {
		roleIds = append(roleIds, r.Id)
	}
	roleIds = util.UniqueStringArray(roleIds)

	addRoles, err := GetRolesByRoleIds(db, roleIds, false)
	middleware.StopExec(err)
	findR := func(rid string) *Role {
		for _, ar := range addRoles {
			if ar.Id == rid {
				return ar
			}
		}
		return nil
	}
	var validRoles []*SubRole
	var invalidRoles []*SubRole
	for _, addr := range form.Roles {
		vr := findR(addr.Id)
		if vr != nil {
			validRoles = append(validRoles, addr)
		} else {
			invalidRoles = append(invalidRoles, addr)
		}
	}

	if len(validRoles) == 0 {
		returnfun.ReturnErrJson(c, "要添加的子角色全部无效")
		return
	}

	// 当前角色已有的 role 和新的 role 要去重
	var allRoles []*SubRole
	allRoles = append(allRoles, validRoles...)
	if len(dbRole.SubRoles) > 0 {
		allRoles = append(allRoles, dbRole.SubRoles...)
	}

	roleMap := make(map[string]*SubRole)
	for _, r := range allRoles {
		roleMap[r.Id] = r
	}
	allRoles = []*SubRole{}
	for _, v := range roleMap {
		allRoles = append(allRoles, v)
	}

	update := bson.M{
		"$set": bson.M{
			"subRoles": allRoles,
			"updateT":  util.GetCurTime(),
		},
	}

	err = db.C(CollectionNameRole).UpdateId(dbRole.Id, update)
	middleware.StopExec(err)

	retData := gin.H{
		"validRoles":   validRoles,
		"invalidRoles": invalidRoles,
	}

	returnfun.ReturnOKJson(c, retData)
	return
}

// 删除 role 的 subroles
func DelSubRolesFromRoleHandler(c *gin.Context, ds *dbandmq.Ds) {
	var form SubRoleForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	roleId := c.Param("id")
	db := ds.CopyDs()
	defer db.Close()

	dbRole, err := GetRoleById(db, roleId, false)
	middleware.StopExec(err)
	if dbRole == nil {
		returnfun.ReturnErrJson(c, "无指定id的role信息")
		return
	}
	if dbRole.Deleted {
		returnfun.ReturnErrJson(c, "角色已被删除，要修改请先恢复此角色")
		return
	}

	// 删除的时候，就直接循环删除即可
	if len(dbRole.SubRoles) == 0 {
		returnfun.ReturnOKJson(c, "")
		return
	}

	findR := func(rid string) *SubRole {
		for _, r := range form.Roles {
			if rid == r.Id {
				return r
			}
		}
		return nil
	}

	var remainRoles []*SubRole
	for _, dbr := range dbRole.SubRoles {
		cr := findR(dbr.Id)
		if cr == nil {
			remainRoles = append(remainRoles, dbr)
		}
	}

	update := bson.M{
		"$set": bson.M{
			"subRoles": remainRoles,
			"updateT":  util.GetCurTime(),
		},
	}

	err = db.C(CollectionNameRole).UpdateId(dbRole.Id, update)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, "")
	return
}

// 查看 role 明细
func GetRoleInfoHandler(c *gin.Context, ds *dbandmq.Ds) {
	id := c.Param("id")
	db := ds.CopyDs()
	defer db.Close()

	role, err := GetRoleById(db, id, true)
	middleware.StopExec(err)
	returnfun.ReturnOKJson(c, role)
	return

}

// 搜索 role
func QueryRoleHandler(c *gin.Context, ds *dbandmq.Ds) {
	var andCondition []bson.M

	// 过滤掉 admin
	andCondition = append(andCondition, bson.M{"name": bson.M{"$ne": AdminRoleName}})

	name := c.Query("name")
	if name != "" {
		andCondition = append(andCondition, bson.M{"name": bson.M{"$regex": name}})
	}

	deleted := c.Query("deleted")
	if deleted != "" {
		deleted = strings.ToUpper(deleted)
		if deleted == "TRUE" {
			andCondition = append(andCondition, bson.M{"deleted": true})
		} else {
			andCondition = append(andCondition, bson.M{"deleted": false})
		}
	}

	query := bson.M{}
	if len(andCondition) > 0 {
		query = bson.M{
			"$and": andCondition,
		}
	}

	db := ds.CopyDs()
	defer db.Close()

	Q := db.C(CollectionNameRole).Find(query)
	total, err := Q.Count()
	middleware.StopExec(err)

	var roles []*Role
	page, size, skip := util.GetPageAndSize(c)
	err = Q.Sort("-_id").Skip(skip).Limit(size).All(&roles)
	middleware.StopExec(err)

	retData := gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"data":  roles,
	}

	returnfun.ReturnOKJson(c, retData)
	return
}
