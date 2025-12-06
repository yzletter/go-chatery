package mysql_test

import (
	"fmt"
	"testing"

	"github.com/yzletter/go-chatery/im_rabbitmq/mysql"
)

func TestInit(t *testing.T) {
	mysql.InitDB()

	if mysql.DB != nil {
		sqlDB, _ := mysql.DB.DB()
		err := sqlDB.Ping()
		if err != nil {
			fmt.Println("Ping MySQL 失败 ...")
			return
		}
		fmt.Println("Ping MySQL 成功 ...")
		return
	}
}
