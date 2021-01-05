package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/internal/pdl/advance"
	"github.com/onlythinking/pug-go/internal/pdl/loan"
	"github.com/onlythinking/pug-go/pkg/logging"
	"github.com/onlythinking/pug-go/pkg/oss/pugaws"
	"github.com/sethvargo/go-signalcontext"
	"sync"
	"time"
)

func main() {

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	logger := logging.DefaultLogger()

	cfgPath := "config/config.yml"

	appConfig := config.App()

	fmt.Println(appConfig)

	db, err := gorm.Open("mysql", appConfig.Mysql.Url)
	if err != nil {
		fmt.Println(err)
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

	<-ctx.Done()

	logger.Info("Closed.")
}
