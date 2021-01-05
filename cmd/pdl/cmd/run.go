/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/internal/pdl/advance"
	"github.com/onlythinking/pug-go/internal/pdl/loan"
	"github.com/onlythinking/pug-go/pkg/help"
	"github.com/onlythinking/pug-go/pkg/logging"
	"github.com/onlythinking/pug-go/pkg/oss/pugaws"
	"github.com/sethvargo/go-signalcontext"
	"github.com/spf13/cobra"
	"time"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Download pan img or request advance ocr",
	Long:  `Download pan img or request advance ocr`,
	Run: func(cmd *cobra.Command, args []string) {
		step, _ := cmd.Flags().GetString("type")
		excelPath, _ := cmd.Flags().GetString("data")
		cfgPath, _ := cmd.Flags().GetString("config")
		if ok, err := help.PathExists(cfgPath); !ok || err != nil {
			panic("Config file not found.")
		}
		start(excelPath, step, cfgPath)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("config", "c", "config.yml", "Config file path")
	runCmd.Flags().StringP("data", "f", "excel/pan_all.xlsx", "Excel file path")
	runCmd.Flags().StringP("type", "t", "1", "1. Download pan img | 2. Request ocr")
	runCmd.Flags().BoolP("daemon", "d", false, "Daemon")
}

func start(excelPath string, step string, cfgPath string) {

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	logger := logging.DefaultLogger()
	config.InitConfigFile(cfgPath)
	appConfig := config.App()

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

	switch step {
	case "1":
		loan.BatchDownloadImg(excelPath)
	case "2":
		loan.BatchReqAdvIdCardOcr(excelPath)
	}

	<-ctx.Done()

	logger.Info("Closed.")
}
