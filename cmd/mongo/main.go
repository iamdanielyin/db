package main

import (
	"fmt"
	_ "github.com/yuyitech/db/adapter/mongo"
	"github.com/yuyitech/db/internal/json"
	"github.com/yuyitech/db/pkg/db"
	"os"
)

type Customer struct {
	ID        uint   `db:"_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int    `db:"age"`
	Roles     []string
}

var settings = &db.DataSource{
	Name:    "test",
	Adapter: "mongo",
	DSN: fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin&connectTimeoutMS=300000&maxPoolSize=50",
		os.Getenv("MONGO_USERNAME"),
		os.Getenv("MONGO_PASSWORD"),
		os.Getenv("MONGO_HOST"),
		os.Getenv("MONGO_DB"),
	),
	IsDefault: true,
}

func main() {
	if err := db.Connect(settings); err != nil {
		panic(err)
	}
	if err := db.RegisterModel(&Customer{}); err != nil {
		panic(err)
	}
	// 增加
	db.Model("Customer").Create(&Customer{
		FirstName: "Eason",
		LastName:  "Chan",
		Age:       47,
		Roles:     []string{"Singer", "Actor"},
	})
	db.Model("Customer").Create(&Customer{
		FirstName: "Edison",
		LastName:  "Chen",
		Age:       41,
		Roles:     []string{"Singer", "Actor"},
	})
	db.Model("Customer").Create(&Customer{
		FirstName: "Andy",
		LastName:  "Lau",
		Age:       60,
		Roles:     []string{"Singer", "Actor"},
	})
	db.Model("Customer").Create(&Customer{
		FirstName: "Stephen",
		LastName:  "Chow",
		Age:       59,
		Roles:     []string{"Director", "Actor"},
	})
	// 查询
	var list []Customer
	res := db.Model("Customer").Find()
	if err := res.All(&list); err != nil {
		panic(err)
	}
	for _, item := range list {
		fmt.Println(json.Stringify(&item, false))
	}
	// 修改
	res = db.Model("Customer").Find(db.Cond{
		"FirstName": "Edison",
		"LastName":  "Chen",
	})
	if n, err := res.Update(&Customer{Age: 40}); err != nil {
		panic(err)
	} else {
		fmt.Println(n)
	}
	// 删除
	res = db.Model("Customer").Find(db.Cond{
		"age >": 40,
	})
	if n, err := res.Delete(); err != nil {
		panic(err)
	} else {
		fmt.Println(n)
	}

	_, _ = db.Model("Customer").Find().Delete()
}
