package roleapp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/returnfun"
	"testing"
)

func getDs() *dbandmq.Ds {
	host := "192.168.2.136"
	port := "27018"
	user := "test"
	passwd := "test"
	dbname := "test"
	ds := dbandmq.NewDs(host, port, user, passwd, dbname)
	return ds
}

func TestRoleRouter(t *testing.T) {
	ds := getDs()
	defer ds.Close()
	ds.InsureCollectionKeys()

	r := middleware.SetupGin()

	apiR := r.Group("/api")

	// 初始化
	DefaultRoleName = "patient"
	err := InitRoleApp(ds, DefaultRoleName, "", "", "/api")
	if err != nil {
		t.Error(err)
		return
	}

	RoleRouter(apiR.Group("", func(c *gin.Context) {
		auth(c, ds)
	}), ds)
	UserAndRoleRouter(apiR.Group("", func(c *gin.Context) {
		auth(c, ds)
	}), ds)
	NoNeedAuthRouter(apiR.Group(""), ds)

	addr := "127.0.0.1:8000"
	err = r.Run(addr)
	if err != nil {
		fmt.Println(err)
	}
}

func auth(c *gin.Context, db *dbandmq.Ds) {
	ds := db.CopyDs()
	defer ds.Close()

	ar := AuthUser(ds, "e1cc4bc1-0222-4a4e-a2d2-2e8cd6cfa604", c.Request.Method, c.Request.RequestURI)
	if ar.Result == AuthResultOK {
		SetCurUser(c, ar)
		c.Next()
	} else {
		returnfun.Return403Json(c, ar.Dump())
	}
}

func TestInsureroleitem(t *testing.T) {
	insureRoleAppItems(nil, "/api")
}
