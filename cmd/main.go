package main

import (
	"fmt"
	"github.com/yuyitech/db"
	_ "github.com/yuyitech/db/adapter/mongo"
	"os"
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
		URI:     os.Getenv("MONGO_URI"),
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
