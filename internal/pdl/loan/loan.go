package loan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/internal/pdl/advance"
	log "github.com/onlythinking/pug-go/pkg/logging"
	"github.com/onlythinking/pug-go/pkg/model"
	"github.com/onlythinking/pug-go/pkg/oss/pugaws"
	uuid "github.com/satori/go.uuid"
	"github.com/tealeg/xlsx/v3"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LoanFile struct {
	CustNo   string `json:"custNo"`
	BusiType string `json:"busiType"`
	InPath   string `json:"inPath"`
	AppNo    string `json:"appNo"`
	PhoneNo  string `json:"phoneNo"`
}

type CuCustOcrResultDtl struct {
	Id         string    `json:"id" gorm:"primary_key;type:varchar(40);comment:'ID'"`
	InstTime   time.Time `json:"instTime" gorm:"column:INST_TIME;type:datetime;comment:'插入时间'"`
	UpdtTime   time.Time `json:"updtTime" gorm:"column:UPDT_TIME;type:datetime;comment:'修改时间'"`
	InstUserNo string    `json:"instUserNo" gorm:"column:INST_USER_NO;type:varchar(40);comment:'插入用户编码'"`
	UpdtUserNo string    `json:"updtUserNo" gorm:"column:UPDT_USER_NO;type:varchar(40);comment:'修改用户编码'"`
	Remark     string    `json:"REMARK" gorm:"column:REMARK;type:varchar(400);comment:'备注（修改记录）'"`
	CustNo     string    `json:"custNo" gorm:"column:CUST_NO;type:varchar(40);comment:'客户唯一编码'"`
	BusiType   string    `json:"busiType" gorm:"column:BUSI_TYPE;type:varchar(8);comment:'业务类型（码类：1007）'"`
	AdvCode    string    `json:"advCode" gorm:"column:ADV_CODE;type:varchar(200);comment:'ADV返回code'"`
	Message    string    `json:"message" gorm:"column:MESSAGE;type:varchar(200);comment:'ADV返回message'"`
	PanNo      string    `json:"panNo" gorm:"column:PAN_NO;type:varchar(40);comment:'Pan卡编号'"`
	CustName   string    `json:"custName" gorm:"column:CUST_NAME;type:varchar(500);comment:'客户姓名'"`
	Birthday   string    `json:"birthday" gorm:"column:BIRTHDAY;type:varchar(40);comment:'生日'"`
	FatherName string    `json:"fatherName" gorm:"column:FATHER_NAME;type:varchar(500);comment:'父亲姓名'"`
}

type PointThirdServiceRecord struct {
	//Id              string    `json:"id" gorm:"primary_key;type:varchar(40);comment:'ID'"`
	//InstTime        model.JsonTime `json:"instTime" gorm:"column:INST_TIME;type:datetime;comment:'插入时间'"`
	//UpdtTime        model.JsonTime `json:"updtTime" gorm:"column:UPDT_TIME;type:datetime;comment:'修改时间'"`
	RequestTime  model.JsonTime `json:"requestTime" gorm:"column:REQUEST_TIME;type:datetime;comment:'调用时间'"`
	ResponseTime model.JsonTime `json:"responseTime" gorm:"column:RESPONSE_TIME;type:datetime;comment:'响应时间'"`
	InstUserNo   string         `json:"instUserNo" gorm:"column:INST_USER_NO;type:varchar(40);comment:'插入用户编码'"`
	//UpdtUserNo      string    `json:"updtUserNo" gorm:"column:UPDT_USER_NO;type:varchar(40);comment:'修改用户编码'"`
	Remark          string `json:"remark" gorm:"column:REMARK;type:varchar(400);comment:'备注（修改记录）'"`
	AppNo           string `json:"appNo" gorm:"column:APP_NO;type:varchar(8);comment:'APP编号'"`
	RegistNo        string `json:"registNo" gorm:"column:REGIST_NO;type:varchar(40);comment:'客户注册手机号'"`
	TransactionId   string `json:"transactionId" gorm:"column:TRANSACTION_ID;type:varchar(80);comment:'三方响应的transactionId'"`
	ServiceName     string `json:"serviceName" gorm:"column:SERVICE_NAME;type:varchar(40);comment:'三方服务名称'"`
	ResponseMessage string `json:"responseMessage" gorm:"column:RESPONSE_MESSAGE;type:varchar(400);comment:'三方响应的message'"`
	ResponseStatus  string `json:"responseStatus" gorm:"column:RESPONSE_STATUS;type:varchar(40);comment:'三方是否有响应（码类：1000）'"`
	ResponseCode    string `json:"responseCode" gorm:"column:RESPONSE_CODE;type:varchar(40);comment:'三方响应的code'"`
	IsPay           string `json:"isPay" gorm:"column:IS_PAY;type:varchar(8);comment:'是否收费（码类：1000）'"`
}

func (CuCustOcrResultDtl) TableName() string {
	return "cu_cust_ocr_result_dtl"
}

func (PointThirdServiceRecord) TableName() string {
	return "point_third_service_record"
}

//*************** https://gorm.io/docs ******************************
var dbTp *gorm.DB
var downloader *pugaws.S3Downloader
var advClient *advance.AdvClient

func Init(db *gorm.DB, d *pugaws.S3Downloader, client *advance.AdvClient) {
	dbTp = db
	downloader = d
	advClient = client
}

func BatchDownloadImg(excelPath string) {

	allItems, err := ParseExcel(excelPath)
	if err != nil {
		log.Error("ParseExcel err : ", err)
		return
	}

	items := allItems[1:]

	baseDir := config.App().Pdl.BaseDir
	chunkSize := config.App().Pdl.ChunkSize
	asyncSize := config.App().Pdl.AsyncSize

	var chunks [][]LoanFile
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	chunks = append(chunks, items)

	for index, chunk := range chunks {
		log.Infof("--------开始处理【%s】---------", strconv.Itoa(index))

		start := time.Now()

		itemChunk := chunk

		var asyncChunks [][]LoanFile
		for asyncSize < len(itemChunk) {
			itemChunk, asyncChunks = itemChunk[asyncSize:], append(asyncChunks, items[0:asyncSize:asyncSize])
		}
		asyncChunks = append(asyncChunks, itemChunk)

		// 等待批次完成
		countDown := make(chan int, len(asyncChunks))

		for _, asyncItem := range asyncChunks {
			asyncItem := asyncItem
			go func() {
				DownloadImg(baseDir, &asyncItem)
				countDown <- 1
			}()
		}

		<-countDown

		log.Infof("下载完成 【%s】", strconv.Itoa(index))
		log.Infof("耗时 %d ms", time.Since(start).Milliseconds())
		log.Info("--------处理结束---------")
	}

}

func BatchReqAdvIdCardOcr(excelPath string) {
	all, err := ParseExcel(excelPath)
	if err != nil {
		log.Error("ParseExcel err : ", err)
		return
	}

	loans := all[1:]

	baseDir := config.App().Pdl.BaseDir
	chunkSize := config.App().Pdl.ChunkSize
	asyncSize := config.App().Pdl.AsyncSize

	//已处理
	processedMap := GetAllOcrResult()
	// 需要处理的客户编号
	var items []LoanFile
	for _, loan := range loans {
		_, exist := processedMap[loan.CustNo]
		if exist {
			continue
		}
		items = append(items, loan)
	}

	log.Info("----------------")
	log.Info("数据整理完毕: ")
	log.Infof("已处理数 %d ", len(processedMap))
	log.Infof("待处理数 %d", len(items))
	log.Infof("总数 %d", len(loans))
	log.Info("----------------")

	var chunks [][]LoanFile
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	chunks = append(chunks, items)

	for index, chunk := range chunks {
		log.Infof("--------开始处理【%s】---------", strconv.Itoa(index))

		start := time.Now()

		itemChunk := chunk

		var asyncChunks [][]LoanFile
		for asyncSize < len(itemChunk) {
			itemChunk, asyncChunks = itemChunk[asyncSize:], append(asyncChunks, items[0:asyncSize:asyncSize])
		}
		asyncChunks = append(asyncChunks, itemChunk)

		// 等待批次完成
		countDown := make(chan int, len(asyncChunks))

		for _, asyncItem := range asyncChunks {
			asyncItem := asyncItem

			go func() {
				for _, aItem := range asyncItem {
					time.Sleep(time.Millisecond * 20)
					ReqAdvIdCardOcr(baseDir, &aItem)
				}
				countDown <- 1
			}()
		}

		<-countDown

		log.Infof("下载完成 【%s】", strconv.Itoa(index))
		log.Infof("耗时 %d ms", time.Since(start).Milliseconds())
		log.Info("--------处理结束---------")
	}

}

func DownloadImg(baseDir string, files *[]LoanFile) {
	var keys []string
	for _, v := range *files {
		keys = append(keys, v.InPath)
	}
	err := downloader.BatchDownload(baseDir, keys)
	if err != nil {
		log.Error("download loan img err ", err)
	}
}

func ReqAdvIdCardOcr(baseDir string, file *LoanFile) {
	data, err := advClient.ReqIdCardOcr(filepath.Join(baseDir, file.InPath))
	advResp := advance.AdvResp{}
	err = json.Unmarshal(data, &advResp)

	if err != nil {
		log.Errorf("ReqAdvIdCardOcr to json err: %s ", string(data), err)
		return
	}

	ocrResult := CuCustOcrResultDtl{
		BusiType:   file.BusiType,
		AdvCode:    advResp.Code,
		Message:    advResp.Message,
		CustNo:     file.CustNo,
		PanNo:      advResp.Data.Values.IdNumber,
		CustName:   advResp.Data.Values.Name,
		Birthday:   advResp.Data.Values.Birthday,
		FatherName: advResp.Data.Values.FatherName,
	}

	ocrResult.InsertOcrResult()

	var isPay = "10000000"
	if "PAY" == advResp.PricingStrategy {
		isPay = "10000001"
	}
	reqPoint := PointThirdServiceRecord{
		AppNo:           file.CustNo[1:4],
		TransactionId:   advResp.TransactionId,
		ServiceName:     "PAN OCR",
		InstUserNo:      "sys",
		ResponseStatus:  "10000001",
		ResponseCode:    advResp.Code,
		ResponseMessage: advResp.Message,
		RequestTime:     model.JsonTime(time.Now()),
		ResponseTime:    model.JsonTime(time.Now()),
		IsPay:           isPay,
		Remark:          string(data),
	}

	reqPointData, err := json.Marshal(reqPoint)
	if err != nil {
		log.Errorf("ReqOrcRecord to json err %s", err)
		return
	}
	go WriteReqOcrRecord(string(reqPointData))
}

func (ths CuCustOcrResultDtl) SqlTemplate() *gorm.DB {
	return dbTp
}

func (ths CuCustOcrResultDtl) GenerateUUID() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

func (ths CuCustOcrResultDtl) InsertOcrResult() {
	dbTp := ths.SqlTemplate()
	if "SUCCESS" != ths.AdvCode {
		return
	}
	ths.Id = ths.GenerateUUID()
	ths.InstTime = time.Now()
	ths.UpdtTime = time.Now()
	dbTp.Create(ths)
}

func GetAllOcrResult() map[string]string {
	var ocrResult []CuCustOcrResultDtl
	dbTp.Find(&ocrResult)
	dbTp.Model(&ocrResult)
	if ocrResult != nil {
		var processedMap = make(map[string]string, len(ocrResult))
		for _, v := range ocrResult {
			processedMap[v.CustNo] = v.AdvCode
		}
		return processedMap
	}
	return nil
}

// 解析Excel
func ParseExcel(filename string) ([]LoanFile, error) {
	wb, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s file not exist.", filename))
	}
	sh := wb.Sheets[0]
	log.Info("XLS Read num total:", sh.MaxRow)

	var loanFiles []LoanFile

	for i := 0; i < sh.MaxRow; i++ {
		row, err := sh.Row(i)
		if err != nil {
			log.Errorf("XLS row %s err: %s", i, err)
			continue
		}
		custNoCell := row.GetCell(0)
		custNo, _ := custNoCell.FormattedValue()
		busiTypeCell := row.GetCell(1)
		busiType, _ := busiTypeCell.FormattedValue()
		inPathCell := row.GetCell(2)
		inPath, _ := inPathCell.FormattedValue()
		appNoCell := row.GetCell(3)
		appNo, _ := appNoCell.FormattedValue()
		phoneNoCell := row.GetCell(4)
		phoneNo, _ := phoneNoCell.FormattedValue()
		if inPath == "" {
			continue
		}
		loanFiles = append(loanFiles, LoanFile{CustNo: custNo,
			BusiType: busiType,
			InPath:   strings.ReplaceAll(inPath, "https://qt-fpdl-app.s3.ap-south-1.amazonaws.com/", ""),
			AppNo:    appNo,
			PhoneNo:  phoneNo,
		})
	}
	return loanFiles, nil
}

var pointUrl = config.App().Pdl.EventServer.ThirdUrl

// 调用埋点
func WriteReqOcrRecord(reqBody string) {
	data := []byte(reqBody)
	req, err := http.NewRequest("POST", pointUrl, bytes.NewBuffer(data))
	if err != nil {
		log.Errorf("ReqOrcRecord new request err: %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("ReqOrcRecord req err: %s", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Errorf("ReqOrcRecord return err: %s", string(body))
	}

}
