package main

import (
	"github.com/yuyitech/db"
	_ "github.com/yuyitech/db/adapter/mongo"
	"log"
	"os"
)

type Member struct {
	FirstName string
	LastName  string `bson:"last_name"`
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

	// 查询数据
	var member Member
	if err := db.Model("Member").Find(db.Cond{"FirstName": "Eason"}).One(&member); err != nil {
		log.Fatalf("单个查询失败：%v\n", err)
	} else {
		log.Printf("单个查询成功：%v\n", member)
	}

	var members []Member
	if err := db.Model("Member").Find().All(&members); err != nil {
		log.Fatalf("列表查询失败：%v\n", err)
	} else {
		log.Printf("列表查询成功：%v\n", members)
	}

	// 修改数据
	if n, err := db.Model("Member").Find(db.Cond{"FirstName": "Daniel"}).UpdateOne(&Member{
		LastName: "Yin",
	}); err != nil {
		log.Fatalf("修改数据失败：%v\n", err)
	} else {
		log.Printf("修改记录条数：%d", n)
	}

	// 删除数据
	if n, err := db.Model("Member").Find(db.Cond{"FirstName": "Daniel"}).DeleteOne(); err != nil {
		log.Fatalf("删除数据失败：%v\n", err)
	} else {
		log.Printf("删除记录条数：%d\n", n)
	}

	// 清空所有
	//_, _ = db.Model("Member").Find().DeleteMany()
}
