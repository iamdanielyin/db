package main

import (
	"github.com/iamdanielyin/db"
	"log"
	"os"
)

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

// Project 项目组
type Project struct {
	ID   string
	Name string
}

// UserProjectRef 用户项目引用
type UserProjectRef struct {
	UserID    string
	ProjectID string
}

func main() {
	// 连接数据源
	sess, err := db.Connect(db.DataSource{
		Name:    "test",
		Adapter: "mongo",
		URI:     os.Getenv("MONGO_URI"),
	})
	if err != nil {
		log.Fatalf("连接数据源失败：%v\n", err)
	}
	defer func() {
		if err := db.Disconnect(); err != nil {
			log.Fatalf("关闭连接失败：%v\n", err)
		}
	}()

	// 注册所有元数据
	sess.RegisterMetadata("test", &User{}, &IDCard{}, &BankCard{}, &Company{}, &Project{})

}
