package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	DB *gorm.DB
)

func InitDB() {
	// 设置 gorm 相关配置
	gormConfig := &gorm.Config{
		PrepareStmt:            true, // 执行任一 SQL 语句时, 都会创建 Prepare Statement 并缓存, 以提高后续执行效率
		SkipDefaultTransaction: true, // 禁止在事务中进行写入操作, 性能提升约 30%
		// 覆盖默认命名策略
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 表名映射不加复数, 仅仅是驼峰转为蛇形
		},
	}
	db, err := gorm.Open(mysql.Open("go_chatery_tester:123456@tcp(localhost:3306)/go_chatery?charset=utf8mb4&parseTime=True&loc=Local"), gormConfig)
	if err != nil {
		fmt.Println("数据库初始化失败")
		panic(err)
	}
	DB = db
}
