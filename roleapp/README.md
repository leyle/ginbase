[TOC]

# RBAC 验证中的 Role 管理接口

本处提供一个快速集成 rbac 中的 role 方向的管理接口，主要实现的功能是管理 role 方面的内容。

项目的模型逻辑包含三个部分内容：

（下述的**多个**指的是零个或多个，如果是零个，说明是空的集合）

1. item - 最基础的 api 的信息结构，包含路径、http method 等内容；
2. permission - 由多个 item 构成的集合，呈现了某一个方面的操作逻辑；
3. role - 由多个 permission 构成的集合，呈现了某个用户角色能够完成的功能的；

验证逻辑包含两部分，一部分是赋予用户对应的 role 信息，另外一部分是验证用户是否能够调用指定的 api。

赋予了用户指定 role 后，用户包含的 role 最终会归结到用户对一堆 items 的可使用上。

验证功能拿到了用户的 userId 和要调用的 api。根据 userId 最终得到了用户能够调用的 items 列表，与目标 api 进行比较，如果存在，即表明当前用户可以调用此接口，否则用户无权限调用此接口。



若无特殊说明，提交的 http request body 均为 json 格式的数据。

---

## 接口路径

因为是为了方便其他项目快速使用上 rbac 中的 r 的管理，所以此处的接口仅描述了自身的路径，实际开发时，一般还会在前面再次追加更多的路径，以构成完整的 api 路径。

比如，新建 item 的路径是 **/role/m/item**，在引用了本库后，其他程序，可能会在前面再追加一些内容，比如 **/api/rbac**，这样的情况下，完整的路径就是 **/api/rbac/role/m/item** 了。

在本文档后续的描述中，api 路径指定是 **/role/m/item** 这样的，而 api prefix 指的是就是 **/api/rbac** 这样的

---



## Admin 与 Default Role

### admin user id 与 admin name

整个系统，有一个初始化的 root 用户，此用户拥有最高权限。需要在引用库包时，调用相应的方法传递过来。需要让哪个用户成为 root 用户，传递此用户的 userId 与 name 即可。

### default role

系统中有些接口仅需要验证用户是否登录（比如读取用户自己的基本信息）。这样的通用接口，归属于系统内的所有用户都可以调用，就可以配置到默认的 role 下。

系统提供了默认 role 的初始值，在引用集成本库时，可以提供另外的 default role name  值。

配置 admin 与 default role 的方法

```go
// 调用如下方法，传递对应的值即可
func InitRoleApp(ds *dbandmq.Ds, dfName, adminId, adminName, uriPrefix string) error {
}
```

---



## role 相关管理接口

role 包含三部分，item、permission、role。下面依次阐述此情况。



### item 管理

#### 新建 item

```json
// POST /role/m/item
// name - 必填，具有唯一性
// method - 必填，http method 名字，比如 GET / POST / PUT 等
// path - 必填，具体的路径，如果路径中有参数，使用 * 代替
// group - 必填，分组名。建议 group name 从我提供的另外一个包 simpledata 中来管理

// 例1. 一般输入
{
    "name": "create vsp",
    "method": "POST",
    "path": "/api/vsp/vsp",
    "group": "vaccine"
}

// 例2. path 中有参数的输入
// 此种情况下，路径参数要么使用 :id，要么使用 *，提交的 :id 在系统内部也会被转换成 *
{
    "name": "get vsp info",
    "method": "GET",
    "path": "/api/vsp/vsp/:id",
    "group": "vaccine"
}
```

---



#### 修改 item

```json
// PUT /role/m/item/:id
// 路径中的 :id 是要被修改的 item 的 id 值
// 修改传递的 body 参数与新建一致，所以即使只修改了 item 中某一项，没有更改的需要原样提交上来。
{
    "name": "create self vsp",
    "method": "POST",
    "path": "/api/vsp/vsp",
    "group": "vaccine"
}
```

---



#### 删除 item

```json
// DELETE /role/m/item/:id
// 路径中的 :id 是要被删除的 item 的 id 值
// 删除是软删除，会标记此 item 为 deleted。
// 后续如果创建了一个与已删除的 item 同名的数据，会提示数据已存在。不会主动去更新原被删除的 item。
// 如果要启用一个已删除的 item，调用 修改 item 接口即可恢复回来。
```

---



#### 读取指定 id 的 item

```json
// GET /role/m/item/:id
// 路径中 :id 即为要读取的 item 的数据 id
// 即使一个 item 被删除了，此处也会返回数据，只是 deleted 值为 true.

// 返回例子
{
    "code": 200,
    "msg": "OK",
    "data": {
        "id": "5e9428c9c9d95708a25dff2b",
        "name": "get vsp detail by id",
        "method": "GET",
        "path": "/api/vsp/vsp/*",
        "group": "vaccine",
        "deleted": true,
        "source": "USER"
    }
}
```

---



#### 搜索 item

```json
// GET /role/m/items
// 支持的 url query parameters 参数如下
// name - 支持部分匹配
// path - 支持部分匹配
// method - 精确匹配，不区分大小写
// group - 精确匹配
// deleted - true / false，筛选是否是删除的数据，如果不传递此值，返回的是所有数据
// page - 默认值 1
// size - 默认值 10

// 例 GET /role/m/items?group=vaccine&deleted=true
{
    "code": 200,
    "msg": "OK",
    "data": {
        "total": 1,
        "page": 1,
        "size": 10,
        "data": [
            {
                "id": "5e9428c9c9d95708a25dff2b",
                "name": "get vsp detail by id",
                "method": "GET",
                "path": "/api/vsp/vsp/*",
                "group": "vaccine",
                "deleted": true,
                "source": "USER"
            }
        ]
    }
}
```

---



### permission 管理

#### 新建 permission 集合

```json
// POST /role/m/permission
// name 参数必填
// 此时还有一个可选参数 itemIds，指的是此 可以包含到此 permission 下的 items

// 例1，仅包含 name 参数
{
    "name": "manage vsp"
}

// 例2. 同时包含了 itemIds 参数
{
    "name": "manage vsp",
    "itemIds": ["5e9428c9c9d95708a25dff2b", "5e9428c0c9d95708a25dff29"]
}
```

---



#### 给 permission 添加 items

一个 permission 可以包含多个 items，可以重复提交。程序会去重。

```json
// POST /role/m/permission/:id/additems
// 路径中的 :id 指的是 permission id
{
    "itemIds": ["5e9428c9c9d95708a25dff2b", "5e9428c0c9d95708a25dff29"]
}
```

---



#### 从 permission 中移除 items

如果要去掉 permission 对某接口的权限，调用此接口即可。

```json
// POST /role/m/permission/:id/delitems
// 路径中的 :id 指的是 permission id
{
    "itemIds": ["5e9428c9c9d95708a25dff2b", "5e9428c0c9d95708a25dff29"]
}
```

---



#### 修改 permission 的 name

permission 是一个容器，所以本身的信息，就只有 name 一个属性。

```json
// PUT /role/m/permission/:id
// 路径中的 :id 指的是 permission id
{
    "name": "new permission name"
}
```

---



#### 删除 permission

```json
// DELETE /role/m/permission/:id
// 路径中的 :id 指的是 permission id
// 软删除，标记为 deleted = true
```

---



#### 读取指定 id 的 permission 明细

此处读取的 permission 信息会包含其拥有的 items 列表信息

```json
// GET /role/m/permission/:id
// 路径中的 :id 指的是 permission id
// 读取回来的包含的 items 仅包含未被删除的 item。
{
    "code": 200,
    "msg": "OK",
    "data": {
        "id": "5e942e1ac9d95708a25dff38",
        "name": "new permission name",
        "items": [
            {
                "id": "5e9428c0c9d95708a25dff29",
                "name": "create vsp",
                "method": "POST",
                "path": "/api/vsp/vsp",
                "group": "vaccine",
                "deleted": false,
                "source": "USER"
            },
            {
                "id": "5e9428c9c9d95708a25dff2b",
                "name": "get vsp detail by id",
                "method": "GET",
                "path": "/api/vsp/vsp/*",
                "group": "vaccine",
                "deleted": false,
                "source": "USER"
            }
        ],
        "deleted": false,
        "source": "USER"
    }
}
```

---



#### 搜索 permission

```json
// GET /role/m/permissions
// 支持的 url query parameters 如下
// name - 部分匹配
// deleted - 可选值 true / false，如果没有此参数，返回所有
// page - 从 1 开始
// size - 默认 10
```

---



### role 管理

role 也是个容器，包含了 permission 列表，为了简单处理， role 之间不能继承。但是一个 role 可以拥有 sub roles。sub role 不参数用户的接口调用权限验证，仅在一个用户A给另外一个用户B赋予 role 时，检查用户A的 sub roles 中是否包含此 role，如果存在，就允许，否则拒绝。

#### 新建 role

```json
// POST /role/m/role
// name - 必输
// pids - permission id 列表，可选输入，后续有接口可以单独维护
// subRoles - 可赋予给其他用户的 role 列表，可选输，后续有接口可以单独维护
// 例1. 仅包含 name
{
    "name": "vmadmin"
}

// 例2. 包含了 pids
{
    "name": "haspids",
    "pids": ["5e942e1ac9d95708a25dff38", "5e942e0cc9d95708a25dff36"]
}

// 例3. 包含了 subRoles
// subrole 中的 id 指的是 role id，name 指的是 role name
{
    "name": "hasSubRoles",
    "pids": ["5e942e1ac9d95708a25dff38", "5e942e0cc9d95708a25dff36"],
    "subRoles": [
        {
            "id": "roleId1",
            "name": "rolename1"
        },
        {
            "id": "roleid2",
            "name": "rolename2"
        }
    ]
}
```

---



#### 给 role 添加 permission

```json
// POST /role/m/role/:id/addps
// 路径中的 :id 指的是 role id
{
    "pids": ["5e942e1ac9d95708a25dff38", "5e942e0cc9d95708a25dff36"]
}
```

---



####  从 role 中移除 permission

```json
// POST /role/m/role/:id/delps
// 路径中的 :id 指的是 role id
{
    "pids": ["5e942e1ac9d95708a25dff38", "5e942e0cc9d95708a25dff36"]
}
```

---



#### 修改 role 的 name

```json
// PUT /role/m/role/:id
// 路径中的 :id 指的是 role id
{
    "name": "new role name"
}
```

---



#### 删除 role

```json
// DELETE /role/m/role/:id
// 路径中的 :id 指的是 role id
// 删除仅仅是标记，如果需要重新启用，调用修改 name 接口即可。
```

---



#### 给 role 添加 sub roles

role 是系统中存在的权限集合。

sub roles 的数据来源也是系统中已存在的 role。

sub role 存在的目的是解决因为 role 没有继承关系带来的给用户赋予 role 的问题。

一个拥有 role a 的用户 a 想要给用户 b 赋予一个 role b，这个时候，无法从 role 自身判断出用户 a 是否能够执行此操作。所以引入了一个 sub role 概念。sub role 存在的意义就是辅助验证用户 a 能否给用户 b 赋予某 role。

如果要赋予的 role 在用户 a 的 sub role 里面，则允许，否则禁止。

用户 a 的 sub role 不一定存在于用户 a 的 role 里面。

比如一个部门经理 role，他可以给自己部门员工赋予 员工 role，会计 role，小组长 role 等。这些他赋予给其他用户的 role，不是他本身的 role，只需要存于他的 sub role 即可。

sub role 也是实际存在于系统中的 role，如果自己生造了一个不存在于系统中的 sub role 数据，那么这个赋予给对应用户后，对应用户实际是无期望的权限的。

```json
// POST /role/m/role/:id/addsubrole
{
    "subRoles": [
        {
            "id": "5e9436f5c9d95709ae02a9b6",
            "name": "new role name"
        },
        {
            "id": "5e943655c9d95709ae02a9b1",
            "name": "vmadmin"
        }
    ]
}
```

---



#### 从 role 中移除 sub roles

```json
// POST /role/m/role/:id/delsubroles
{
    "subRoles": [
        {
            "id": "5e9436f5c9d95709ae02a9b6",
            "name": "new role name"
        },
        {
            "id": "5e943655c9d95709ae02a9b1",
            "name": "vmadmin"
        }
    ]
}
```

---



#### 查看指定 id 的 role 明细

此处返回的 role 明细，包含了 subroles，包含了 permissions，同时还包含了各个 permission 包含的 items。

```json
// GET /role/m/role/:id
// 返回例子
{
    "code": 200,
    "msg": "OK",
    "data": {
        "id": "5e9436f5c9d95709ae02a9b6",
        "name": "new role name",
        "permissions": [
            {
                "id": "5e942e0cc9d95708a25dff36",
                "name": "manage vsp",
                "items": null,
                "deleted": false,
                "source": "USER"
            }
        ],
        "subRoles": [
            {
                "id": "5e9436f5c9d95709ae02a9b6",
                "name": "new role name"
            },
            {
                "id": "roleId1",
                "name": "rolename1"
            },
            {
                "id": "roleid2",
                "name": "rolename2"
            }
        ],
        "deleted": false,
        "source": "USER"
    }
}
```

---



#### 搜索 roles

```json
// GET /role/m/roles
// 支持的 url query parameters 如下
// name - 部分匹配
// deleted - 可选值 true / false，不传返回所有
// page - 从 1 开始
// size - 默认值 10
```

---



## role 与 user 关联相关接口

上述的接口都是管理 role 本身的接口，此处描述的是与用户关联起来，给用户 role 权限。

### role 与用户关联

#### 给 user id 赋予 role

```json
// POST /rau/addroles
// 可以选择按 role id 添加，或者按 role name 添加，role id 与 name 必须有一个存在
// 如果 role id 存在，就以 role id 为准，不会再检查 role name 是否有值
// 添加的 role id 或 name 必须是系统中存在的数据。
// userName 是个可选值，建议调用时还是传递此值，方便系统维护时，能够更直观的看到用户是谁。

// 例1. 使用 roleIds 传递数据
{
    "userId": "someuseridvalue",
    "userName": "Jack Ma",
    "roleIds": ["5e9436eec9d95709ae02a9b4", "5e943655c9d95709ae02a9b1"]
}

// 例2. 使用 roleNames 传递数据
{
    "userId": "someuseridvalue",
    "userName": "Jack Ma",
    "roleName": ["vspadmin", "haspids"]
}
```

---



#### 移除指定 uesr id 的 role

```json
// POST /rau/delroles
// 可以选择按 role id 删除，或者按 role name 删除，role id 与 name 必须有一个存在
// 如果 role id 存在，就以 role id 为准，不会再检查 role name 是否有值
// 删除的 role id 或 name 必须是系统中存在的数据。

// 例1. 使用 roleIds 传递数据
{
    "userId": "someuseridvalue",
    "roleIds": ["5e9436eec9d95709ae02a9b4", "5e943655c9d95709ae02a9b1"]
}

// 例2. 使用 roleNames 传递数据
{
    "userId": "someuseridvalue",
    "roleName": ["vspadmin", "haspids"]
}
```

