package main

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/internal/pdl/advance"
	"github.com/onlythinking/pug-go/internal/pdl/loan"
	"github.com/onlythinking/pug-go/pkg/logging"
	"github.com/onlythinking/pug-go/pkg/oss/pugaws"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"
)

func main() {

	logger := logging.DefaultLogger()

	if pid := syscall.Getpid(); pid != 1 {
		_ = ioutil.WriteFile("pdl.pid", []byte(strconv.Itoa(pid)), 0777)
		defer func() {
			err := os.Remove("pdl.pid")
			if err != nil {
				logger.Error("Remove pdl.pid err ", err)
			}
		}()
	}

	appConfig := config.App()
	db, err := gorm.Open("mysql", appConfig.Mysql.Url)
	if err != nil {
		panic("Failed to connect to the database, please check the configuration")
	}
	// 获取通用数据库对象 sql.DB
	sqlDB := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(64)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dbStats := sqlDB.Stats()
	dbStatsContent, err := json.Marshal(dbStats)
	logger.Infof("Mysql db pool: %s", string(dbStatsContent))

	db.LogMode(true)
	downloader := pugaws.NewS3Downloader()
	advOcrClient := advance.NewAdvOcrClient("PAN_FRONT")

	loan.Init(db, downloader, advOcrClient)

	//loan.BatchDownloadImg()
	//loan.BatchReqAdvIdCardOcr()
}
