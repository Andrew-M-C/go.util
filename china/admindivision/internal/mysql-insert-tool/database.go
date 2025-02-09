package main

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	// import mysql driver
	_ "github.com/go-sql-driver/mysql"
)

func connectDB(c config) (*sqlx.DB, error) {
	db := c.Database
	uri := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		db.Username, db.Password, db.Host, db.Port, db.DB,
	)
	// log.Printf("URI: %s", uri)
	return sqlx.Open("mysql", uri)
}

//go:embed table.sql
var createTableTemplate string

func createTable(db *sqlx.DB, c config) error {
	statement := fmt.Sprintf(createTableTemplate, c.Database.TableName)
	_, err := db.Exec(statement)
	return err
}

type insertLine struct {
	Name     string
	Code     string
	Province string
	City     string
	County   string
	Town     string
	Village  string
}

func generateInsertParams(data downloadData) []insertLine {
	var lines []insertLine

	parse := func(in adminNodeFormat) (insertLine, error) {
		province, city, county, town, village := "", "00", "00", "000", "000"
		switch len(in.Code) {
		case 2:
			province = in.Code
		case 4:
			province = in.Province
			city = in.Code[2:4]
		case 6:
			province = in.Province
			city = in.Code[2:4]
			county = in.Code[4:6]
		case 9:
			province = in.Province
			city = in.Code[2:4]
			county = in.Code[4:6]
			town = in.Code[6:9]
		case 12:
			province = in.Province
			city = in.Code[2:4]
			county = in.Code[4:6]
			town = in.Code[6:9]
			village = in.Code[9:12]
		default:
			return insertLine{}, fmt.Errorf("无法识别的代码 '%s', 所属 '%s'", in.Code, in.Name)
		}
		return insertLine{in.Name, in.Code, province, city, county, town, village}, nil
	}

	iterate := func(in []adminNodeFormat) {
		for _, item := range in {
			line, err := parse(item)
			if err != nil {
				log.Println(err)
				continue
			}
			lines = append(lines, line)
		}
	}
	iterate(data.Provinces)
	iterate(data.Cities)
	iterate(data.Counties)
	iterate(data.Towns)
	iterate(data.Villages)
	return lines
}

func doInsert(db *sqlx.DB, tableName string, params []insertLine) error {
	count := 0
	start := time.Now()

	iterate := func(p insertLine) error {
		statement := fmt.Sprintf(
			"INSERT INTO `%s` (name, code, province, city, county, town, village) "+
				"VALUES (?, ?, ?, ?, ?, ?, ?)",
			tableName,
		)
		if _, err := db.Exec(statement,
			p.Name, p.Code, p.Province, p.City, p.County, p.Town, p.Village,
		); err != nil {
			if isDuplicateError(err) {
				return nil
			}
			return err
		}
		count++
		if count%100 == 0 {
			ela := time.Since(start)
			log.Printf("已处理 %d 条数据, 耗时 %v", count, ela)
			start = time.Now()
		}
		return nil
	}
	for _, p := range params {
		if err := iterate(p); err != nil {
			return fmt.Errorf("插入数据失败 (%w), 源数据 %+v", err, p)
		}
	}
	return nil
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "uplicat")
}
