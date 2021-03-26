package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yuyitech/db/internal/json"
	"github.com/yuyitech/db/pkg/db"
	"gopkg.in/guregu/null.v4"
	"os"
	"testing"
	"time"
)

func init() {
	_ = db.Connect(&db.DataSource{
		Name:    "test",
		Adapter: "mysql",
		DSN:     os.Getenv("MYSQL_DSN"),
	})
}

func TestAdapter_NativeCollectionMetadata(t *testing.T) {
	meta, err := db.DB("test").NativeCollectionMetadata()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(meta)
}

func TestExecResult_All(t *testing.T) {
	var all []struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Mobile    string    `json:"mobile"`
	}
	query := db.DB("test").Query("SELECT * FROM hyg_Freelancer")
	if err := query.All(&all); err != nil {
		t.Error(err)
		return
	}
	t.Log(all)
}

func TestExecResult_One(t *testing.T) {
	var one struct {
		ID           uint      `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		Mobile       string    `json:"mobile"`
		SignNotified null.Bool `json:"sign_notified"`
	}
	query := db.DB("test").Query("SELECT * FROM hyg_Freelancer")
	if err := query.One(&one); err != nil {
		t.Error(err)
		return
	}
	t.Log(one.Mobile)
}

func TestFindResult_All(t *testing.T) {
	var list []struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret_data"`
		Token     string    `json:"token"`
		Iat       int64     `json:"iat"`
		Exp       int64     `json:"exp"`
		Method    string    `json:"method"`
	}
	res := db.Model("TESTAccSignSecret").Find(
		db.Cond{
			"iat >":  1587870150,
			"method": "LOGIN",
		},
	).Order("method", "-exp").Page(2).Size(3)
	totalPages, _ := res.TotalPages()
	t.Log(totalPages)
	_ = res.All(&list)
	t.Log(json.Stringify(&list, true))
}

func TestFindResult_One(t *testing.T) {
	var secret struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret"`
		Token     string    `json:"token"`
		Iat       int64     `json:"iat"`
		Exp       int64     `json:"exp"`
		Method    string    `json:"method"`
	}
	_ = db.Model("TESTAccSignSecret").Find(db.Cond{"method": "LOGIN"}).One(&secret)
	t.Log(secret)
}

func TestModel_Create(t *testing.T) {
	secret := struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret"`
		Token     string    `json:"token"`
		Iat       int64     `json:"iat"`
		Exp       int64     `json:"exp"`
		Method    string    `json:"method"`
	}{
		CreatedAt: time.Now(),
		UID:       10000,
		Secret:    "hello",
		Token:     "test-_hello",
		Iat:       19012891230,
	}
	id, n, err := db.Model("TESTAccSignSecret").Create(&secret)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id)
	t.Log(n)
}

func TestModel_BatchCreate(t *testing.T) {
	type secret struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret"`
		Token     string    `json:"token"`
		Iat       int64     `json:"iat"`
		Exp       int64     `json:"exp"`
		Method    string    `json:"method"`
	}
	var list []secret
	for i := 0; i < 9000; i++ {
		item := secret{
			CreatedAt: time.Now(),
			UID:       10000,
			Secret:    "hello2",
			Token:     "test-_hello",
			Iat:       19012891230,
		}
		if i%3 == 0 {
			item.Method = fmt.Sprintf("%d", i+1)
			item.Exp = time.Now().Unix()
		}
		list = append(list, item)
	}
	id, n, err := db.Model("TESTAccSignSecret").Create(&list)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id, n)
}

func TestFindResult_Update(t *testing.T) {
	type secret struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret"`
		Token     string    `json:"token"`
		Iat       int64     `json:"iat"`
		Exp       int64     `json:"exp"`
		Method    string    `json:"method"`
	}
	res := db.Model("TESTAccSignSecret").Find(db.Cond{"id": 75568})
	n, err := res.Update(&secret{UpdatedAt: time.Now()})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(n)
}

func TestFindResult_Delete(t *testing.T) {
	res := db.Model("TESTAccSignSecret").Find(db.Cond{"secret": "hello2"})
	n, err := res.Delete()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(n)
}

func TestFindResult_Populate(t *testing.T) {
	type secret struct {
		ID        uint      `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UID       uint      `json:"uid"`
		Secret    string    `json:"secret"`
	}
	type user struct {
		ID            uint      `json:"id"`
		CreatedAt     time.Time `json:"created_at"`
		Username      string    `json:"username"`
		Name          string    `json:"name"`
		Secrets       []secret  `json:"secrets"`
		DefaultSecret secret    `json:"defaultSecret"`
	}
	meta, _ := db.Meta("TESTSysUser")
	meta.RegisterFields(map[string]db.Field{
		"secrets": {
			MetadataName:             "TESTSysUser",
			Name:                     "secrets",
			Type:                     db.TypeArray,
			IsScalarType:             false,
			DisplayName:              "Secrets",
			RelationshipKind:         db.RelationshipHasMany,
			RelationshipModel:        "TESTAccSignSecret",
			RelationshipLocalField:   "id",
			RelationshipForeignField: "uid",
		},
		"default_secret": {
			MetadataName:             "TESTSysUser",
			Name:                     "default_secret",
			Type:                     db.TypeObject,
			IsScalarType:             false,
			DisplayName:              "Default Secret",
			RelationshipKind:         db.RelationshipHasOne,
			RelationshipModel:        "TESTAccSignSecret",
			RelationshipLocalField:   "id",
			RelationshipForeignField: "uid",
		},
	})
	var err error
	var one user
	err = db.Model("TESTSysUser").Find(db.Cond{"id": 171}).Populate("defaultSecret").Populate("secrets").One(&one)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(one.ID, one.DefaultSecret, len(one.Secrets), one.Secrets)

	var list []user
	err = db.Model("TESTSysUser").Find(db.Cond{"id in": []uint{170, 171}}).Populate("secrets").Populate("defaultSecret").All(&list)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range list {
		t.Log(item.ID, item.DefaultSecret, len(item.Secrets), item.Secrets)
	}
}
