package domain

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"hash/crc32"
	"time"
)

//***************Model******************************
type PersistentDomain struct {
	Id          string    `json:"id" gorm:"primary_key;type:varchar(40);comment:'ID'"`
	CreatedTime time.Time `json:"createdTime" gorm:"column:created_time;type:datetime;comment:'创建时间'"`
}

type Domain struct {
	PersistentDomain
	LastModifiedTime time.Time `json:"lastModifiedTime" gorm:"column:last_modified_time;type:datetime;comment:'最后更新时间';default null"`
	Remark           string    `json:"remark" gorm:"type:varchar(200);comment:'备注'"`
}

type Record struct {
	PersistentDomain
	Remark string `json:"remark" gorm:"type:varchar(200);comment:'备注'"`
}

//***************Utils******************************
func (ths PersistentDomain) GenerateUUID() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}

func (ths PersistentDomain) GenerateIntUUID() int {
	return int(crc32.ChecksumIEEE(uuid.NewV4().Bytes()))
}

func (ths PersistentDomain) PrePersist() {
	ths.Id = ths.GenerateUUID()
	ths.CreatedTime = time.Now()
}

func (ths Domain) PrePersist() {
	ths.Id = ths.GenerateUUID()
	ths.CreatedTime = time.Now()
}

func (ths Domain) PreUpdate() {
	ths.LastModifiedTime = time.Now()
}

//*************** https://gorm.io/docs ******************************
var dbTp *gorm.DB

func Init(db *gorm.DB) {
	dbTp = db
}

func (ths PersistentDomain) SqlTemplate() *gorm.DB {
	return dbTp
}
