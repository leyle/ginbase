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
	Name     string `json:"name" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Group string `json:"group" binding:"required"` // 属于哪个分组
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
		Id:       util.GenerateDataId(),
		Name:     form.Name,
		Method:   strings.ToUpper(form.Method),
		Path:     form.Path,
		Group: form.Group,
		Deleted:  false,
		Source: RoleDataSourceApi,
		CreateT:  util.GetCurTime(),
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
	Name     string `json:"name" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Group string `json:"group" binding:"required"` // 属于哪个分组
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

	err = ds.C(CollectionNameItem).UpdateId(dbitem.Id, dbitem)
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

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
			"updateT": util.GetCurTime(),
		},
	}

	err := ds.C(CollectionNameItem).UpdateId(id, update)
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
	andCondition = append(andCondition, bson.M{"name": bson.M{"$not": bson.M{"$in": AdminItemNames}}})

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
func RemoveItemsFromPermissionHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 修改 permission 基本信息
func UpdatePermissionInfoHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 删除 permission
func DeletePermissionHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 读取 permission 信息，包含 items
func GetPermissionHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 搜索 permission 列表
func QueryPermissionHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 新建 role
func CreateRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 给 role 添加 permission
func AddPermissionsToRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 从 role 中移除 permission
func RemovePermissionsFromRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 修改 role 信息
func UpdateRoleInfoHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 删除 role
func DeleteRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 给 role 添加 childrole
func AddChildRolesToRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 删除 role 的 childroles
func DelChildRolesFromRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 查看 role 明细
func GetRoleInfoHandler(c *gin.Context, ds *dbandmq.Ds) {

}

// 搜索 role
func QueryRoleHandler(c *gin.Context, ds *dbandmq.Ds) {

}