<a name="zTrhs"></a>
# 目录
- [快速开始](#ZH38i)
- [数据源](#RiEEy)
- [元数据](#cm6b7)
   - [传入Metadata](#n4teo)
   - [传入结构体](#p5do1)
- [新增](#S75Ro)
   - [单个新增](#qn1U7)
   - [批量新增](#BW1TR)
- [查询](#MDT4w)
   - [单个查询](#aucZe)
   - [列表查询](#LXaih)
   - [迭代查询](#oKG1I)
   - [条件查询](#pXhaf)
      - [字符串运算符](#KbWQr)
      - [数值运算符](#yAytk)
      - [范围运算符](#thJKJ)
      - [存在运算符](#RPGBz)
      - [逻辑运算符](#w1OQt)
   - [数量查询](#jGzY2)
   - [分页查询](#viDrv)
   - [排序查询](#AhP9Z)
- [修改](#usHdi)
   - [单个修改](#s8Ylo)
   - [批量修改](#cBYBK)
- [删除](#nWBSo)
   - [单个删除](#lXp0m)
   - [批量删除](#MtLe7)
   - [逻辑删除](#Rnlna)
   - [物理删除](#SKYEm)
- [事务](#yGnyc)
   - [StartTransaction](#cixqD)
   - [WithTransaction](#FnqFR)
- [本地化脚本](#OMeK7)
   - [查询类脚本](#ks9it)
   - [执行类脚本](#FJJWr)
- [中间件](#PRphn)
   - [CRUD中间件](#L8OPY)
   - [字段中间件](#iEvuC)
- [元数据引用](#htZqq)
   - [关系定义](#inKDt)
   - [元数据定义](#edQl8)
   - [拥有一个](#hn3OE)
   - [拥有多个](#B0tvR)
   - [关联一个](#QVMkd)
   - [关联多个](#xx3Gi)
<a name="ZH38i"></a>
# 快速开始
```go
package main

import (
	"fmt"
	"github.com/yuyitech/db"
)

type Member struct {
	FirstName string
	LastName  string
}

func main() {
	// 连接数据源
	sess, err := db.Connect(db.DataSource{
		Name:    "test",
		Adapter: "mongo",
		URI:     "mongodb://admin:123456@localhost/test",
	})
	if err != nil {
		panic("连接失败！")
	}

	// 注册元数据
	sess.RegisterMetadata(&Member{})

	// 新增数据
	res, _ := db.Model("Member").InsertMany([]Member{
		{
			FirstName: "Eason",
			LastName:  "Chan",
		},
		{
			FirstName: "Daniel",
			LastName:  "Wu",
		},
		{
			FirstName: "Steve",
			LastName:  "Jobs",
		},
	})
    fmt.Println(res.StringIDs())

	// 查询数据
	var member Member
	db.Model("Member").Find(db.Cond{"FirstName": "Eason"}).One(&member)

	var members []Member
	db.Model("Member").Find().All(&members)

	// 修改数据
	db.Model("Member").Find(db.Cond{"FirstName": "Daniel"}).UpdateOne(&Member{
		LastName: "Yin",
	})

	// 删除数据
	db.Model("Member").Find(db.Cond{"FirstName": "Daniel"}).DeleteOne()

	// 关闭连接
	_ = sess.Disconnect()
}
```
<a name="RiEEy"></a>
# 数据源
连接数据源：
```go
db.Connect(db.DataSource{
    Name: 'test',
    Adapter: 'mongo',
    URI: 'mongodb://admin:123456@localhost/test'
})
```
获取缓存连接：
```go
sess := db.Session('test')
```
关闭单个连接：
```go
sess.Disconnect()
```
关闭所有连接：
```go
db.Disconnect()
```
<a name="cm6b7"></a>
# 元数据
注册元数据：
```go
// 通过session实例注册元数据
sess.RegisterMetadata(...)

// 通过全局函数注册元数据
db.RegisterMetadata('test', ...)
```
注意：结构体名称需要**全局唯一**，`RegisterMetadata`支持传入**Metadata**和**结构体指针**两种格式。
<a name="n4teo"></a>
## 传入Metadata
```go
db.RegisterMetadata("test", db.Metadata{
    Name:        "User",
    DisplayName: "用户",
    Properties: db.Fields{
        "ID": db.Field{
            Type:        db.String,
            PrimaryKey:  "true",
            DisplayName: "数据唯一ID",
            Description: "系统自动生成",
        },
        "Username": db.Field{
            Type:        db.String,
            DisplayName: "用户名",
            Description: "不允许重复",
            Trim:        "both",
            Required:    "true",
            Unique:      "true",
        },
        "Password": db.Field{
            Type:        db.Password,
            DisplayName: "密码",
            Description: "加密存储",
            Trim:        "both",
        },
        "Nickname": db.Field{
            Type:        db.String,
            DisplayName: "昵称",
            Trim:        "both",
        },
        "Avatar": db.Field{
            Type:        db.String,
            DisplayName: "头像",
        },
        "Gender": db.Field{
            Type:        db.String,
            DisplayName: "性别",
            Enum: db.Enum{
                db.EnumItem{
                    Label: "男",
                    Value: "male",
                },
                db.EnumItem{
                    Label: "女",
                    Value: "female",
                },
                db.EnumItem{
                    Label: "未知",
                    Value: "unknown",
                },
            },
            DefaultValue: "unknown",
        },
        "Status": db.Field{
            Type:        db.Int,
            DisplayName: "用户状态",
            Enum: db.Enum{
                db.EnumItem{
                    Label: "正常",
                    Value: 1,
                },
                db.EnumItem{
                    Label: "已禁用",
                    Value: -1,
                },
                db.EnumItem{
                    Label: "审核中",
                    Value: -2,
                },
            },
            DefaultValue: 1,
        },
        "DenyLogin": db.Field{
            Type:        db.Bool,
            DisplayName: "禁止登录",
        },
        "CountryCode": db.Field{
            Type:        db.String,
            DisplayName: "国家/地区代码",
            Required:    "+PhoneNumber",
        },
        "PhoneNumber": db.Field{
            Type:        db.String,
            DisplayName: "手机号码",
            Description: "不含国家/地区代码",
            Required:    "contact",
        },
        "EmailAddress": db.Field{
            Type:        db.String,
            DisplayName: "邮箱地址",
            Required:    "contact",
        },
        "CreatedAt": db.Field{
            Type:         db.Timestamp,
            DisplayName:  "注册时间",
            DefaultValue: "$now",
        },
    },
})
```

- `Trim` - 表示对字符串进行空格裁剪，支持传入`both`、`start`、`end`三个值；
- `Required` - 表示该字段不允许为空，传入`true`时，表示该字段必填；当传入`+字段名`格式时，表示当指定字段不为空时必填；传入其他值时表示分组名称，组内所有字段为多选一必填（即设置相同值的所有字段不能全部为空）；
- `Unique` - 表示该字段唯一，传入`true`时，表示该字段唯一。
- `DefaultValue` - 表示该字段默认值，当传入`$now`时，表示取当前时间的Unix时间戳。
<a name="p5do1"></a>
## 传入结构体
为简化配置，以下属性推荐使用简写配置：

- `primaryKey` - 简写为`pk`；
- `displayName` - 简写为`name`；
- `nativeName` - 简写为`native`；
- `description` - 简写为`desc`；
- `required` - 简写为`rqd`；
- `unique` - 简写为`uniq`；
- `defaultValue` - 简写为`default`。
```go
type User struct {
	ID           string `db:"name=数据唯一ID;desc=系统自动生成;pk=true"`
	Username     string `db:"name=用户名;desc=不允许重复;trim=both;rqd=true;uniq=true"`
	Password     string `db:"name=密码;desc=加密存储;trim=both;type=password"`
	Nickname     string `db:"name=昵称;trim=both"`
	Avatar       string `db:"name=头像;trim=both"`
	Gender       string `db:"name=性别;enum=[male:男,female:女,unknown:未知];default=unknown"`
	Status       int    `db:"name=用户状态;enum=[1:正常,-1:已禁用,-2:审核中];default=1"`
	DenyLogin    bool   `db:"name=禁止登录"`
	CountryCode  string `db:"name=国家/地区代码;rqd=+PhoneNumber"`
	PhoneNumber  string `db:"name=手机号码;desc=不含国家/地区代码;rqd=contact"`
	EmailAddress string `db:"name=邮箱地址;rqd=contact"`
	CreatedAt    int    `db:"name=注册时间;default=$now"`
}

func (u *User) Metadata() db.Metadata {
	return db.Metadata{
		DisplayName: "用户",
	}
}

db.RegisterMetadata('test', &User{})
```
注意：默认使用结构体名称（以上为`User`）作为元数据名称，也可以通过实现`Metadata`函数来覆盖自动解析生成的任意属性。
<a name="S75Ro"></a>
# 新增
支持单个新增和批量新增两种。
<a name="qn1U7"></a>
## 单个新增
```go
res, _ := db.Model('User').InsertOne(User{
    Username:     "foo",
    EmailAddress: "foo@example.com",
})

fmt.Println(res.StringID()) // 5326bfc0e6f780b21635248f
```
<a name="BW1TR"></a>
## 批量新增
```go
res, _ := db.Model('User').InsertMany([]User{
    {
        Username:     "foo",
        EmailAddress: "foo@example.com",
    },
    {
        Username:     "bar",
        EmailAddress: "bar@example.com",
    },
    {
        Username:    "foobar",
        CountryCode: "86",
        PhoneNumber: "13800138000",
    },
})

fmt.Println(res.StringIDs()) // [5326bfc0e6f780b21635248f, 2094bfc0e6f780b21635938d, 9287bfc0e6f780b216350129t]
```
<a name="MDT4w"></a>
# 查询
<a name="aucZe"></a>
## 单个查询
```go
var user User
_ = db.Model("User").Find().One(&user)
```
<a name="LXaih"></a>
## 列表查询
```go
var users []User
_ = db.Model("User").Find().All(&users)
```
<a name="oKG1I"></a>
## 迭代查询
大数据量场景下，推荐使用迭代查询：
```go
var users []User
cur, _ := db.Model("User").Find().Cursor()
for cur.HasNext() {
    var user User
    _ = cur.Next(&user)
    users = append(users, user)
}
cur.Close()
```
<a name="pXhaf"></a>
## 条件查询
`Find`函数支持传入`db.Cond`进行条件查询：
```go
db.Model("User").Find(db.Cond{
    "Status":       1,
    "DenyLogin !=": false,
    "Username":     "foo",
    "EmailAddress": "foo@example.com",
})
```
`db.Cond`为`Map`结构，其中`Key`支持传入`字段名 运算符`格式，支持的运算符列表如下

- `=` - 默认值（`Key`仅为字段名时等价于传入等号）
- `!=` - 不等于
- `*=` - 模糊匹配任意开头，类似于SQL的`LIKE '%aaa'`
- `=*` - 模糊匹配任意结尾，类似于SQL的`LIKE 'aaa%'`
- `*` - 模糊任意子字符串，类似于SQL的`LIKE '%aaa%'`
- `>` - 大于
- `<` - 小于
- `>=` - 大于等于
- `<=` - 小于等于
- `~=` - 正则匹配
- `$in` - 包含在指定数组内
- `$nin` - 不包含在指定数组内
- `$exists` - 是否存在
<a name="KbWQr"></a>
### 字符串运算符
```go
// 表示查询 EmailAddress 以“foo”开头的数据
db.Cond{"EmailAddress *=": "foo"}

// 表示查询 EmailAddress 以“example.com”结尾的数据
db.Cond{"EmailAddress =*": "example.com"}

// 表示查询 EmailAddress 包含“example”的数据
db.Cond{"EmailAddress *": "example"}

// 表示查询 EmailAddress 包含“example”的数据
db.Cond{"EmailAddress ~=": "/example/igm"}
```
<a name="yAytk"></a>
### 数值运算符
```go
// 表示查询 Status 为“1”的数据
db.Cond{"Status": 1} // 等价于db.Cond{ "Status =": 1 }

// 表示查询 Status 不为“1”的数据
db.Cond{"Status !=": 1}

// 表示查询 Status 大于“0”的数据
db.Cond{"Status >": 0}

// 表示查询 Status 大于等于“1”的数据
db.Cond{"Status >=": 1}

// 表示查询 Status 小于“0”的数据
db.Cond{"Status <": 0}

// 表示查询 Status 小于等于“0”的数据
db.Cond{"Status <=": 0}
```
<a name="thJKJ"></a>
### 范围运算符
```go
// 表示查询 “Status IN (1, -1, -2)” 的数据
db.Cond{"Status $in": []int{1, -1, -2}}

// 表示查询 “Status NOT IN (-1, -2)” 的数据
db.Cond{"Status $nin": []int{-1, -2}}

// 表示查询 “CreatedAt >= 1633536000 AND CreatedAt <= 1633622399” 的数据
db.And(
    db.Cond{"CreatedAt >=": 1633536000},
    db.Cond{"CreatedAt <=": 1633622399},
)

// 表示查询 “(Username = "foo" OR Username = "bar") AND Status = 1” 的数据
db.And(
    db.Or(
        db.Cond{"Username": "foo"},
        db.Cond{"Username": "bar"},
    ),
    db.Cond{"Status": 1},
)
```
<a name="RPGBz"></a>
### 存在运算符
```go
// 表示查询 PhoneNumber 字段存在且不为空的数据（类似于PhoneNumber IS NOT NULL）
db.Cond{"PhoneNumber $exists": true}

// 表示查询 PhoneNumber 字段不存在或值为空的数据（类似于PhoneNumber IS NULL）
db.Cond{"PhoneNumber $exists": false}
```
<a name="w1OQt"></a>
### 逻辑运算符
```go
// 表示查询 “Username = "foo" OR Username = "bar"” 的数据
db.And(
    db.Cond{"Username": "foo"},
    db.Cond{"Username": "bar"},
)

// 表示查询 “(CountryCode = "86" AND PhoneNumber = "13800138000") OR (EmailAddress IS NOT NULL)” 的数据
db.Or(
    db.And(
        db.Cond{"CountryCode": "86"},
        db.Cond{"PhoneNumber": "13800138000"},
    ),
    db.Cond{"EmailAddress $exists": true},
)
```
<a name="jGzY2"></a>
## 数量查询
```go
count, _ := db.Model("User").Find().Count()
```
<a name="viDrv"></a>
## 分页查询
```go
// 按10条每页分页
p := db.Model("User").Find().Paginate(10)

// 查询第1页数据
p.All()

// 查询第2页数据
p.Page(2).All()

// 查询所有记录数
recordCount, _ := p.TotalRecords()

// 查询所有页数
pageCount, _ := p.TotalPages()
```
<a name="AhP9Z"></a>
## 排序查询
```go
// 单个字段排序（“-”开头表示逆序）
db.Model("User".Find().OrderBy("-Status").All()

// 多个字段排序：方式一
db.Model("User".Find().OrderBy("-CreatedAt").OrderBy("Status").All()

// 多个字段排序：方式二
db.Model("User".Find().OrderBy("-CreatedAt", "Status").All()
```
<a name="usHdi"></a>
# 修改
支持单个修改和批量修改两种，修改语法如下：
```go
n, err := db.Model("User").Find(...).UpdateXxx(...)
if err != nil {
    panic("修改失败")
}

n // 受影响记录数：1
```
<a name="s8Ylo"></a>
## 单个修改
```go
// 传入结构体进行查询和修改
db.Model("User").Find(&User{ID: "1"}).UpdateOne(&User{
	Nickname: "Foo",
	Status:   1,
})

// 传入Cond进行查询、传入结构体进行修改   
db.Model('User'.Find(db.Cond{"ID": "1"}).UpdateOne(&User{
	Nickname: "Foo",
	Status:   1,
})

// 传入Cond进行查询、传入Map进行修改
db.Model('User'.Find(db.Cond{"ID": "1"}).UpdateOne(db.D{
	"Nickname": "Foo",
	"Status":   1,
})
```
注意：虽然`Find`和`Update`均支持传入结构体，但由于 Go 语言本身无法区分空值和零值的情况（在 Go 中，每个基础类型在未初始化时都对应一个零值：布尔类型是 `false `，整型和浮点型都是 `0` ，字符串是`""`），框架底层将自动忽略所有值为`nil`、`0`、`false`、`""`的字段，所以使用结构体注册元数据时，可允许为空的字段，推荐使用[https://github.com/guregu/null](https://github.com/guregu/null)包进行类型声明。
```go
import "gopkg.in/guregu/null.v4"

type User struct {
	ID           string      `db:"name=数据唯一ID;desc=系统自动生成"`
	Username     string      `db:"name=用户名;desc=不允许重复;trim=both;rqd=true;uniq=true"`
	Password     null.String `db:"name=密码;desc=加密存储;trim=both;type=password"`
	Nickname     null.String `db:"name=昵称;trim=both"`
	Avatar       null.String `db:"name=头像;trim=both"`
	Gender       string      `db:"name=性别;enum=[male:男,female:女,unknown:未知];default=unknown"`
	Status       int         `db:"name=用户状态;enum=[1:正常,-1:已禁用,-2:审核中];default=1"`
	DenyLogin    null.Bool   `db:"name=禁止登录"`
	CountryCode  null.String `db:"name=国家/地区代码;rqd=+PhoneNumber"`
	PhoneNumber  null.String `db:"name=手机号码;desc=不含国家/地区代码;rqd=contact"`
	EmailAddress null.String `db:"name=邮箱地址;rqd=contact"`
}
```
<a name="cBYBK"></a>
## 批量修改
```go
// 更新所有记录的 Status 的值为“1”
db.Model("User").Find().UpdateMany(&User{
    Status: 1,
})
```
<a name="nWBSo"></a>
# 删除
支持单个删除和批量删除两种，删除语法如下：
```go
n, err := db.Model("User").Find(...).DeleteXxx()
if err != nil {
    panic("修改失败")
}

n // 受影响记录数：1
```
<a name="lXp0m"></a>
## 单个删除
```go
// 删除 ID 为“1”的数据
db.Model("User").Find(&User{ID: "1"}).DeleteOne()

// 删除 ID 为“1”的数据
db.Model("User".Find(db.Cond{"ID": "1"}).DeleteOne()
```
<a name="MtLe7"></a>
## 批量删除
```go
// 删除 “Username = "foo" OR Username = "bar"” 的数据
db.Model("User".Find(db.Or(
    db.Cond{"Username": "foo"},
    db.Cond{"Username": "bar"},
)).DeleteMany()
```
<a name="Rnlna"></a>
## 逻辑删除
框架自带对逻辑删除的支持，但使用前需要注册逻辑删除规则。
```go
// 针对所有元数据设置全局规则（全局优先级最低）
db.RegisterLogicDeleteRule("*", &db.LogicDeleteRule{
    SetValue: map[string]string{
        "DeletedAt": "$now",
    },
    GetValue: db.Cond{"DeletedAt $exists": false},
})

// 针对某一组元数据设置规则（组规则优于全局规则）
db.RegisterLogicDeleteRule("Acc*", &db.LogicDeleteRule{
	SetValue: map[string]string{
        "IsDeleted": "$int(1)",
    },
    GetValue: db.Cond{"DeletedAt !=": 1},
})

// 针对单个元数据设置规则（元数据规则高于组规则）
db.RegisterLogicDeleteRule("User", &db.LogicDeleteRule{
	SetValue: map[string]string{
        "DeletedAt": "$now",
    },
    GetValue: db.Cond{"DeletedAt $exists": false},
})
```
注意：

- `RegisterLogicDeleteRule`第一个参数为Glob语法（具体用法请参考[https://github.com/gobwas/glob](https://github.com/gobwas/glob)）；
- 每个元数据只会有**一条**规则生效，规则优先级为`元数据规则 > 组规则 > 全局规则`；
- `SetValue`的可选值如下：
   - `$now` - 当前时间的Unix时间戳；
   - `$now_iso` - 当前时间的ISO格式字符串；
   - `$int(v)` - 格式化v为整型类型；
   - `$float(v)` - 格式化v为浮点类型；
   - `$bool(v)` - 格式化v为布尔类型；
   - 其他非空值为原样设置。
- `GetValue`可接收`db.Cond`、`db.And`或`db.Or`类型数据。
<a name="SKYEm"></a>
## 物理删除
支持在`Find`以后链式调用`Unscoped`函数忽略**所有逻辑删除规则**。
```go
// 单个物理删除
db.Model("User").Find().Unscoped().DeleteOne()

// 批量物理删除
db.Model("User").Find().Unscoped().DeleteMany()
```
<a name="yGnyc"></a>
# 事务
支持`StartTransaction`和`WithTransaction`两种方式。
<a name="cixqD"></a>
## StartTransaction
需手动调用`Commit`或`Rollback`：
```go
// 指定数据源创建事务实例
tx, _ := db.StartTransaction("test")
// 等价于
tx, _ := db.Session("test").StartTransaction()

// 事务内的各种数据库操作
tx.Model("User").Find()
tx.Model("User").Find(&User{ID: "1"}).Update(&User{Status: 1})
tx.Model("User").Find(&User{Username: "foo"}).Update(&User{Nickname: "Foo"})

// 提交或回滚事务
if err := tx.Commit(); err != nil {
    tx.Rollback()
}
```
<a name="FnqFR"></a>
## WithTransaction
框架根据返回的`error`判断自动`Commit`还是`Rollback`：
```go
db.WithTransaction("test", func(tx) error {
    tx.Model("User").Find()
    if _, err := tx.Model("User").Find(&User{ID: "1"}).Update(&User{Status: 1}); err != nil {
        return err
    }
    if _, err := tx.Model("User").Find(&User{Username: "foo"}).Update(&User{Nickname: "Foo"}); err != nil {
        return err
    }
    return nil
})

// 等价于
db.Session("test").WithTransaction(func(tx) error {
    ...
})
```
<a name="OMeK7"></a>
# 本地化脚本
支持查询类脚本和执行类脚本两种，查询类脚本返回查询对象，执行类脚本返回执行结果、受影响记录数等。
<a name="ks9it"></a>
## 查询类脚本
SQL类脚本：
```go
q, _ := db.Raw("test", `SELECT * FROM users WHERE id = ?`, "1").Query()
// 等价于
q, _ := db.Session("test").Raw(...).Query()

// 可进行反序列化操作
q.All()
q.One()
q.Cursor()
```
MongoDB类脚本：
```go
q, _ := db.Raw("test", `
    {
      "collection": "User",
      "action": "aggregate",
      "options": [
        {
          "$project": {
            "_id": 0,
            "username": 1,
            "status": 1
          }
        },
        {
          "$group": {
            "_id": "$status",
            "valid_count": {
              "$sum": 1
            }
          }
        }
      ]
    }
`).Query()
```
<a name="FJJWr"></a>
## 执行类脚本
```go
n, _ := db.Raw("test", ...).Exec()
// 等价于
n, _ := db.Session("test").Raw(...).Exec()
```
<a name="PRphn"></a>
# 中间件
包含元数据中间件（CRUD中间件）和字段中间件两种。
<a name="L8OPY"></a>
## CRUD中间件
默认支持的中间件列表如下：

- `beforeCreate` - 新增前调用
- `afterCreate` - 新增后调用
- `beforeUpdate` - 修改前调用
- `afterUpdate` - 修改后调用
- `beforeSave` - 新增、修改前皆调用
- `afterSave` - 新增、修改后皆调用
- `beforeQuery` - 查询前调用
- `afterQuery` - 查询后调用
- `beforeDelete` - 删除前调用
- `afterDelete` - 删除后调用

注册语法如下：
```go
// 针对所有元数据注册全局中间件（全局）
db.RegisterMiddleware("*:beforeCreate", func(scope *db.Scope) error {
    ...
})

// 针对部分元数据注册分组中间件（分组后于全局执行）
db.RegisterMiddleware("Acc*:beforeCreate", func(scope *db.Scope) error {
    ...
})

// 针对指定元数据注册元数据中间件（元数据中间件最后执行）
db.RegisterMiddleware("User:beforeCreate", func(scope *db.Scope) error {
    ...
})
```
注意：

- `RegisterMiddleware`的第一个参数为Glob语法（具体用法请参考[https://github.com/gobwas/glob](https://github.com/gobwas/glob)），配置格式固定为`元数据名称:中间件名称`；
- 所有中间件均会执行，执行顺序从先到后排列为`全局中间件>分组中间件>元数据中间件`；
- 传入参数`scope`的重要属性或方法的含义如下：
   - `Session` - 当前操作使用的连接会话；
   - `Metadata` - 当前元数据；
   - `Conditions` - 当前操作关联的所有查询条件；
   - `Action` - 当前数据库操作；
      - `INSERT_ONE`
      - `INSERT_MANY`
      - `UPDATE_ONE`
      - `UPDATE_MANY`
      - `DELETE_ONE`
      - `DELETE_MANY`
      - `QUERY_ONE`
      - `QUERY_ALL`
      - `QUERY_CURSOR`
      - `QUERY_COUNT`
      - `QUERY_PAGE`
   - `OrderBys` - 排序参数；
   - `PageSize` - 分页条数；
   - `PageNum` - 当前页码；
   - `InsertDocs` - 新增的数据，Map数组结构；
   - `InsertOneResult` - 单个新增结果；
   - `InsertManyResult` - 批量新增结果；
   - `UpdateDoc` - 修改的数据，Map结构；
   - `UpdateOneResult` - 单个修改执行结果；
   - `UpdateManyResult` - 批量修改执行结果；
   - `DeleteOneResult` - 单个删除执行结果；
   - `DeleteManyResult` - 批量删除执行结果；
   - `Skip()` - 跳过后续所有中间件的执行；
   - `HasError()` - 当前调用链中是否包含错误。
<a name="iEvuC"></a>
## 字段中间件
字段中间件和元数据中间件的注册语法很像，只需要多添加一个`:`符号传入字段名即可，其余用法与元数据中间件完全一致：
```go
// 该中间件仅在创建User时Username有值时被触发
db.RegisterMiddleware("User:beforeCreate:Username", func(scope *db.Scope) error {
    ...
})

// 该中间件仅在创建User时CountryCode和PhoneNumber均有值时被触发
db.RegisterMiddleware("User:beforeCreate:CountryCode,PhoneNumber", func(scope *db.Scope) error {
    ...
})

// 该中间件仅在有修改User的PhoneNumber或EmailAddress时被触发
db.RegisterMiddleware("User:beforeUpdate:PhoneNumber|EmailAddress", func(scope *db.Scope) error {
    ...
})
```
注意：

- `RegisterMiddleware`的第一个参数为Glob语法（具体用法请参考[https://github.com/gobwas/glob](https://github.com/gobwas/glob)），配置格式固定为`元数据名称:中间件名称:字段规则`；
- 字段规则支持三种配置：
   - 仅传入单个字段名：表示仅指定字段发生变化时触发；
   - 传入多个字段名，以英文逗号`,`分隔：表示仅指定的所有字段均发生变化时触发；
   - 传入多个字段名，以英文竖线`|`分隔：表示指定的所有字段中某一个发生变化时触发。
- 字段规则在不同CRUD操作时匹配的位置不同：
   - `INSERT_ONE` - 匹配`InsertOneDoc`；
   - `INSERT_MANY` - 匹配`InsertManyDocs`；
   - `UPDATE_XXX` - 匹配`UpdateDoc`；
   - `DELETE_XXX` - 匹配`Conditions`。
  <a name="htZqq"></a>
# 元数据关联
在常见的ORM框架中，我们常常听到以下几种关联关系：

- 一对一
- 一对多
- 多对一
- 多对多

这几种关联关系对于新手来说往往难以理解，建模时经常会不知所措。但不论判断哪种关联关系，一定要先确定参照物，就像你和你爸的关系，站在你爸的角度，他叫你儿子，站在你的角度，你叫他爸爸，不同的角度下的叫法也是不同的。
<a name="inKDt"></a>
## 关系定义

1. **拥有一个**（Has One）：用户拥有一张身份证，每个用户对自己的身份证信息拥有**读写权**（一对一）；
1. **拥有多个**（Has Many）：用户拥有多张银行卡，每个用户对自己的银行卡信息拥有**读写权**（一对多）；
1. **关联一个**（Associate One）：用户属于一个公司，每个用户对自己所属的公司拥有**只读权**，即每个用户不能修改公司资料（一对一）；
1. **关联多个**（Associate Many）：用户参与多个项目，每个用户对自己参与的项目拥有**只读权**（多对多）。

综上：

- **拥有关系**中，**主模型**对**子模型**拥有完整**读写权**，即我们可以在操作主模型的同时对其子模型进行增删改查。
- **关联关系**中，**主模型**对**被关联模型**仅拥有**只读权**，即我们只能联查被关联模型、与其建立关联关系以及解除关联关系。
<a name="edQl8"></a>
## 元数据定义
我们通过ref属性来定义以上四种关联关系，参数如下：

- `type` - 关系类型（必填）；
   - `HAS_ONE` - 拥有一个
   - `HAS_MANY` - 拥有多个
   - `ASSC_ONE` - 关联一个
   - `ASSC_MANY` - 关联多个
- `meta` - 被引用元数据名称（可选，默认解析结构体名称）；
- `src` - 当前元数据中的关系字段名称（可选，默认为主键名称）；
- `dst` - 被引用元数据中的关系字段名称（可选，默认为主键名称）；
- `int_meta` - 类似多对多种的中间表，指定元数据名称（当`type=ASSC_MANY`时必填）
- `int_meta_src` - 中间表中当前元数据中的关系字段名称（当`type=ASSC_MANY`时必填）
- `int_meta_dst` - 中间表中被引用元数据中的关系字段名称（当`type=ASSC_MANY`时必填）

以下是配置示例：
```go
// User 用户
type User struct {
	ID        string
	RealName  string
	IDCard    *IDCard    `db:"ref=type:HAS_ONE,dst:UserID"`
	BankCards []BankCard `db:"ref=type:HAS_MANY,dst:UserID"`
	CompanyID string
	Company   *Company  `db:"ref=type:ASSC_ONE,src:CompanyID"`
	Projects  []Project `db:"ref=type:ASSC_MANY,int_meta:UserProjectRef,int_src:UserID,int_dst:ProjectID"`
}

// IDCard 身份证
type IDCard struct {
	ID      string
	CardNum string
	UserID  string
	User    *User `db:"ref=type:ASSC_ONE,src:UserID"`
}

// BankCard 银行卡
type BankCard struct {
	ID      string
	CardNum string
	UserID  string
}

// Company 公司
type Company struct {
	ID   string
	Name string
}

// Project 项目
type Project struct {
	ID   string
	Name string
}

// UserProjectRef 用户项目关联
type UserProjectRef struct {
	UserID    string
	ProjectID string
}
```

<a name="hn3OE"></a>
## 拥有一个

### 引用联查

#### 单个查询

#### 多个查询

#### 迭代查询

### 引用变更

我们整理了三种场景

#### 引用替换

#### 引用合并

#### 引用删除

#### 建立关系

#### 解除关系

#### 引用新增

#### 引用修改

#### 引用删除



<a name="B0tvR"></a>
## 拥有多个

### 引用联查

#### 单个查询

#### 多个查询

#### 迭代查询

### 引用变更

以下三种常见场景，几乎覆盖子档案的联动变更的所有场景：

- 引用替换：使用传入列表覆盖数据库列表；
- 引用合并：将传入列表合并至数据库列表；
- 引用删除：仅删除传入列表的数据。

注意：默认情况下，引用删除时都只会删除与主档案之间的引用关系，并不会删除引用档案本身，如需在删除引用关系的同时也一并删掉引用档案，请通过指定`DeleteObjects`参数。

#### 引用新增

```go
db.Model("User").Find().DeleteOne()
```

#### 引用修改

#### 引用删除

#### 建立关系

#### 解除关系

#### 引用新增

#### 引用修改

#### 引用删除


<a name="QVMkd"></a>
## 关联一个


<a name="xx3Gi"></a>
## 关联多个

