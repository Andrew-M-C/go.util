package main

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func main() {
	var (
		conf   config
		db     *sqlx.DB
		err    error
		data   downloadData
		params []insertLine
	)
	procedures := []func(){
		func() { conf, err = parseConfig() },                           // 获得配置
		func() { data, err = downloadAdminDistricts() },                // 下载远程数据
		func() { db, err = connectDB(conf) },                           // 连接 DB
		func() { err = createTable(db, conf) },                         // 建表
		func() { statisticNameLength(data) },                           // 统计一下最长名称
		func() { params = generateInsertParams(data) },                 // 解析和准备插入参数
		func() { err = doInsert(db, conf.Database.TableName, params) }, // 插入
	}
	for i, f := range procedures {
		start := time.Now()
		f()
		ela := time.Since(start)
		if err != nil {
			log.Printf("执行第 %d 步失败, 错误信息: %v", i+1, err)
			break
		}
		log.Printf("步骤 %d, 耗时 %v", i+1, ela)
	}

	// 结束
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("关闭 DB 失败: %v", err)
		} else {
			log.Println("关闭 DB")
		}
	}
	if err == nil {
		log.Println("运行顺利结束")
	}
}
