[TOC]

# infinite class api

本接口支持无限级的联级分类。所有级的最终级 parentId 值是字符串 0.

---

## 接口路径

此处列出来的是本包提供的本身的 api 路径。即实际使用时，可能会在这个路径之外，再此扩展更多路径。

比如新建分类的路径是`/cat/:domain`,实际使用可能会追加`/api/`上去，联合起来就是 `/api/cat/:domain`

### 路径中 :domain 的用处

路径中的 domain 可以有两个用处：

1. 当程序为多个不同的应用提供服务时，使用 domain 可以用来区分数据归属于哪个应用；
2. 当只为一个应用提供服务时，使用 domain 可以作为顶级分类使用。

下面会介绍这种使用方法中，对数据的管理逻辑的问题。

---

## 管理接口

为了使接口通用，即同一套系统，可以用在很多个 app 上。路径中配置了 :domain 参数。这个参数会写入到分类的属性中。不同的 app 通过 domain 的值来区分。

同一个层级，name + domain 联合约束唯一值，也就是 name + level + domain  约束了唯一的值。

提交的消息体都是 application/json 格式的数据。

返回的数据也是 json 格式的数据。

---

### 新建分类

```json
// POST /cat/:domain
// 输入数据
// parentId 指的是当前数据的父级是谁，如果是一级，那么父级值就是 0，如果是其他层级，那么就是上一级的真实 id
// name 指的是本分类的名字
// icon 是分类的图标
// info 是额外的一些补充信息
// parentId 和 name 值必输
{
    "parentId": "0",
    "name": "ShangHai",
    "icon": "https://example.com/icon.jpg",
    "info": "some extra info"
}
```



---

### 修改分类信息

```json
// PUT /cat/:domain/info/:id
// 路径中的 :id 指的是要修改的分类的 id
// name 值必输
// 另外两个可选输入
// 程序实现时，使用提交上来的数据覆盖原有的数据，所以即使以下三个值中的某些值没有改变，也需要原样提交上来。
{
    "name": "some name",
    "icon": "https://example.com/icon.jpg",
    "info": "some extra info"
}
```



---

### 禁用分类

```json
// POST /cat/:domain/:id/disable
```



---

### 启用分类

```json
// POST /cat/:domain/:id/enable
```



---

### 读取指定 id/name 的分类信息（可包含子类）

此接口的作用主要是读取当前及子类的信息。

```json
// GET /cat/:domain/info
// 支持两种传递参数，一种是传递 id，另外一种是传递 name
// 1.  传递 id 查询
// 1.1 支持的参数如下
// 1.1.1 children=y/n 指的是是否查询返回子类，不传递不返回子类数据
// 1.1.2 disable=y/n y 指的是仅返回 disable 的数据，n 仅返回有效数据，不传递返回所有
// 1.2 例子
// GET /cat/:domain/info?id=xxx // 仅返回 id 所属层级数据
// GET /cat/:domain/info?id=xxx&children=y&disable=n // 返回 id 所有层级及下层的 disable=n 的数据
// ...


// 2. 传递 name 查询
// 因为不同的 level 可以有相同的 name，所以传递 name 时，需要同步传递一个 level 参数指明这是哪个层级的数据
// 2.1 支持的参数如下
// 2.1.0 level=2 // 必传参数
// 2.1.1 children=y/n 指的是是否查询返回子类，不传递不返回子类数据
// 2.1.2 disable=y/n y 指的是仅返回 disable 的数据，n 仅返回有效数据，不传递返回所有
// 2.2 例子
// GET /cat/:domain/info?name=address&level=1
// GET /cat/:domain/info?name=address&level=1&children=y&disable=n
```



---

### 读取指定 level 的信息（可包含子类）

本接口从 level 的角度提供了查询分类及子类的方法

根据传递的 url 参数，可以查询所有的 level 数据，也可以限定某个父类的子类的 level 数据。

比如有一个父类名为 address 的分类，下面分为了 ACity / BCity / CCity 三个子类，每一个子类又有各自的行政区域分类。

address 作为 level 1 的数据 ， ACity 等作为 level 2 的数据，Acity 等下属的子类就是 level 3 的数据。

现在可以使用 /cat/:domain/level/1?children=y 把 address 和所有的子类读取回来。

也可以使用 /cat/:domain/level/2?children=y&pname=address 把 address 下属的所有子类读取回来，但是不包含 address 这一层级。

```json
// GET /cat/:domain/level/:level
// 路径中的 :level 指的是层级，是一个数字。
// 额外支持三个参数
// children=y/n 指的是是否查询返回子类，不传递不返回子类数据
// disable=y/n y 指的是仅返回 disable 的数据，n 仅返回有效数据，不传递返回所有
// pid/pname 可以限定这个 :level 的数据是属于哪一个父类的数据
// 例子
// GET /cat/:domain/level/1 // 仅返回所有一级分类的数据
// GET /cat/:domain/level/1?children=y&disable=n // 返回所有的 1 级分类及子类数据
// GET /cat/:domain/level/2?children=y&pname=address // 返回父类名为 address 的 2 级分类数据及其子类
```

