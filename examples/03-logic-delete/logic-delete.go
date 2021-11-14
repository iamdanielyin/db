package main

import (
	"github.com/yuyitech/db"
	_ "github.com/yuyitech/db/adapter/mongo"
	"log"
	"os"
)

type Member struct {
	FirstName string
	LastName  string
	DeletedAt int
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

	// 注册元数据
	if err := sess.RegisterMetadata(&Member{}); err != nil {
		log.Fatalf("元数据注册失败：%v\n", err)
	}

	// 注册逻辑删除规则
	db.RegisterLogicDeleteRule("*", &db.LogicDeleteRule{
		SetValue: map[string]string{
			"DeletedAt": "$now",
		},
		GetValue: db.Cond{"DeletedAt $exists": false},
	})

	// 新增数据
	if res, err := db.Model("Member").InsertMany([]Member{
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
	}); err != nil {
		log.Fatalf("新增数据失败：%v\n", err)
	} else {
		log.Printf("新增数据成功：%v\n", res.StringIDs())
	}

	if n, err := db.Model("Member").Find(db.Cond{"FirstName": "Eason"}).Count(); err != nil {
		log.Fatalf("删除前查询数据失败：%v\n", err)
	} else {
		log.Printf("删除前查询数据总数：%d\n", n)
	}

	// 删除数据
	if n, err := db.Model("Member").Find(db.Cond{"FirstName": "Eason"}).DeleteMany(); err != nil {
		log.Fatalf("删除数据失败：%v\n", err)
	} else {
		log.Printf("删除记录条数：%d\n", n)
	}

	if n, err := db.Model("Member").Find(db.Cond{"FirstName": "Eason"}).Count(); err != nil {
		log.Fatalf("删除后查询数据失败：%v\n", err)
	} else {
		log.Printf("删除后查询数据总数：%d\n", n)
	}
}
