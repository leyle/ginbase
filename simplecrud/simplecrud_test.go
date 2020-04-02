package simplecrud

import (
	"fmt"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
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

func TestSimpleDataRouter(t *testing.T) {
	ds := getDs()
	defer ds.Close()
	r := middleware.SetupGin()

	apiR := r.Group("/api")

	SimpleDataRouter(apiR.Group(""), ds)

	addr := "127.0.0.1:8000"
	err := r.Run(addr)
	if err != nil {
		fmt.Println(err)
	}
}
