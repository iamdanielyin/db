package schema_test

import (
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID                   string        `db:"native:_id;primary"`
	Username             string        `db:"native:username;title:账号;trim;required"`
	Password             string        `db:"native:password;title:密码;type:password;trim;required"`
	Nickname             string        `db:"native:nickname;title:昵称;trim;required"`
	Status               int           `db:"native:status;title:状态;enum:0=正常,1=禁用,2=审核中,3=审核拒绝;default:0;required"`
	Gender               null.String   `db:"native:gender;title:性别;enum:male,female,unknown"`
	PhoneNumber          null.String   `db:"native:phone_number;title:手机号"`
	CountryCode          null.String   `db:"native:country_code;title:城市/地区代码"`
	PhoneNumberConfirmed null.Time     `db:"native:phone_number_confirmed;title:手机号验证时间"`
	Score                null.Int      `db:"native:score;title:用户积分"`
	CreditCard           *CreditCard   `db:"title:关联信用卡;ref:CreditCard;owner;foreignKey:UserID"`
	RoleBindings         []RoleBinding `db:"title:关联角色;ref:RoleBinding;owner;localKey:UserID;foreignKey:RoleID"`
	RegisteredAt         int64         `db:"native:registered_at;title:注册时间;type:timestamp;default:$now;required"`
}

type RoleBinding struct {
	ID     string `db:"native:_id;primary"`
	UserID string `db:"native:user_id;title:关联用户ID;required"`
	User   *User  `db:"native:ref:User;title:关联用户;localKey:UserID"`
	RoleID string `db:"native:role_id;title:关联角色ID;required"`
	Role   *Role  `db:"title:关联角色;ref:Role;localKey:RoleID"`
}

type Role struct {
	ID                 string              `db:"native:_id;primary"`
	Name               string              `db:"native:name;title:角色名称;required"`
	Description        null.String         `db:"native:description;title:详细描述"`
	PermissionBindings []PermissionBinding `db:"title:关联权限;ref:PermissionBinding;owner;localKey:RoleID;foreignKey:PermissionID"`
}

type PermissionBinding struct {
	ID           string      `db:"native:_id;primary"`
	RoleID       string      `db:"native:role_id;title:关联角色ID;required"`
	Role         *Role       `db:"native:ref:Role;title:关联角色;localKey:RoleID"`
	PermissionID string      `db:"native:permission_id;title:关联权限ID;required"`
	Permission   *Permission `db:"title:关联权限;ref:Permission;localKey:PermissionID"`
}

type Permission struct {
	ID          string      `db:"native:_id;primary"`
	Name        string      `db:"native:name;title:权限名称"`
	Description null.String `db:"native:description;title:详细描述"`
}

type CreditCard struct {
	ID     string `db:"native:_id;primary"`
	Number string `db:"native:number;title:卡号"`
	UserID string `db:"native:user_id;title:关联用户ID;required"`
	User   *User  `db:"title:关联用户;ref:User;localKey:UserID"`
}
