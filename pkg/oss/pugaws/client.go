package pugaws

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/onlythinking/pug-go/internal/config"
	"github.com/onlythinking/pug-go/pkg/help"
	log "github.com/onlythinking/pug-go/pkg/logging"
	"github.com/onlythinking/pug-go/pkg/pugerr"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type PugS3 interface {
	getDefaultBucket() string
	getBucketName(bucket string) string
}

// S3客户端
type S3Client struct {
	*s3.S3
	defaultBucket string
}

// S3 批量下载器
type S3Downloader struct {
	*s3manager.Downloader
	defaultBucket string
}

// S3 批量上传器
type S3Uploader struct {
	*s3manager.Uploader
	defaultBucket string
}

func (ths *S3Client) getDefaultBucket() string {
	return ths.defaultBucket
}

func (ths *S3Client) getBucketName(bucket string) string {
	var bucketName = ths.defaultBucket
	if len(bucket) != 0 {
		bucketName = bucket
	}
	return bucketName
}

func (ths *S3Downloader) getDefaultBucket() string {
	return ths.defaultBucket
}

func (ths *S3Downloader) getBucketName(bucket string) string {
	var bucketName = ths.defaultBucket
	if len(bucket) != 0 {
		bucketName = bucket
	}
	return bucketName
}

func (ths *S3Uploader) getDefaultBucket() string {
	return ths.defaultBucket
}

func (ths *S3Uploader) getBucketName(bucket string) string {
	var bucketName = ths.defaultBucket
	if len(bucket) != 0 {
		bucketName = bucket
	}
	return bucketName
}

// @title    查看所有桶
// @description   查看所有桶
func (ths *S3Client) ShowListBuckets() string {
	resp, err := ths.ListBuckets(nil)
	if err != nil {
		log.Error("Show list buckets err", err)
	}
	return resp.String()
}

// @title    创建桶
func (ths *S3Client) CreateBucketBy(bucketName string) error {
	request := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := ths.CreateBucket(request)
	if err != nil {
		return handleError(err)
	}
	log.Debugf("Create bucket %s return %s", bucketName, resp.String())
	return nil
}

// 上传默认桶单个文件
func (ths *S3Client) PutObjectBody(objectKey string, data []byte) error {
	return ths.doPutObjectBody("", objectKey, data, nil)
}

// 执行上传
func (ths *S3Client) doPutObjectBody(bucket string, objectKey string, data []byte, metadata map[string]string) error {
	var request *s3.PutObjectInput

	if metadata != nil {
		request = &s3.PutObjectInput{
			Bucket:   aws.String(ths.getBucketName(bucket)),
			Key:      aws.String(objectKey),
			Body:     bytes.NewReader(data),
			Metadata: aws.StringMap(metadata),
		}
	} else {
		request = &s3.PutObjectInput{
			Bucket: aws.String(ths.getBucketName(bucket)),
			Key:    aws.String(objectKey),
			Body:   bytes.NewReader(data),
		}
	}

	resp, err := ths.PutObject(request)
	if err != nil {
		return handleError(err)
	}
	log.Debugf("Put object %s  return %s", objectKey, resp.String())
	return nil
}

// 下载默认桶单个文件
func (ths *S3Client) GetObjectBody(objectKey string) ([]byte, error) {
	return ths.GetObjectBodyByBucket("", objectKey)
}

// 下载指定桶文件
func (ths *S3Client) GetObjectBodyByBucket(bucket string, objectKey string) ([]byte, error) {
	result, err := ths.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(ths.getBucketName(bucket)),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, handleError(err)
	}
	defer func() {
		err := result.Body.Close()
		if err != nil {
			log.Error("Close object stream err", err)
		}
	}()

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Error("Read object stream error", err)
		return nil, err
	}

	return data, err
}

// 删除默认桶文件
func (ths *S3Client) DeLObject(objectKey string) error {
	return ths.DeLObjectByBucket("", objectKey)
}

// 删除指定桶文件
func (ths *S3Client) DeLObjectByBucket(bucket string, objectKey string) error {
	resp, err := ths.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(ths.getBucketName(bucket)),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return handleError(err)
	}
	log.Debugf("Put object %s  return %s", objectKey, resp.String())
	return nil
}

// 创建S3客户端
func NewS3Client() *S3Client {
	cfg := config.App()
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.Oss.Region),
		LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			log.Debug(args)
		}),
		Credentials: credentials.NewStaticCredentials(cfg.Oss.AccessKeyId, cfg.Oss.SecretAccessKey, ""),
	})

	if err != nil {
		log.Error("Create s3 client err", err)
		return nil
	}
	svc := s3.New(sess)

	return &S3Client{
		svc,
		cfg.Oss.DefaultBucket,
	}
}

// Context 获取配置和日志
func NewS3ClientWithContext(ctx context.Context) *S3Client {
	logger := log.FromContext(ctx)
	cfg := config.FromContext(ctx)

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.Oss.Region),
		LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			logger.Debug(args)
		}),
		Credentials: credentials.NewStaticCredentials(cfg.Oss.AccessKeyId, cfg.Oss.SecretAccessKey, ""),
	})

	if err != nil {
		logger.Error("Create s3 client err", err)
		return nil
	}
	svc := s3.New(sess)

	return &S3Client{
		svc,
		cfg.Oss.DefaultBucket,
	}
}

// 创建S3下载器
func NewS3Downloader() *S3Downloader {
	cfg := config.App()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Oss.Region),
		Credentials: credentials.NewStaticCredentials(cfg.Oss.AccessKeyId, cfg.Oss.SecretAccessKey, ""),
	})

	if err != nil {
		log.Error("Create s3 downloader err", err)
	}

	// 以下参数根据下载文件大小和CPU内存进行调配
	// PartSize    下载增量大小 5MB (1024 * 1024 * 5)
	// Concurrency 下载启动的goroutine数量 默认 5
	svc := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = 1024 * 1024 * 5
		d.Concurrency = 8
	})
	return &S3Downloader{
		Downloader:    svc,
		defaultBucket: cfg.Oss.DefaultBucket}
}

// 创建S3下载器
func NewS3Uploader() *S3Uploader {
	cfg := config.App()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Oss.Region),
		Credentials: credentials.NewStaticCredentials(cfg.Oss.AccessKeyId, cfg.Oss.SecretAccessKey, ""),
	})

	if err != nil {
		log.Error("Create s3 uploader err", err)
	}

	// 以下参数根据上传文件大小和CPU内存进行调配
	// PartSize    		上传增量大小可以将有效负载的块缓冲 5MB (1024 * 1024 * 5)
	// MaxUploadParts   上传增量块最大限制  默认 10000
	// Concurrency 		上传启动的goroutine数量 默认 5
	svc := s3manager.NewUploader(sess, func(d *s3manager.Uploader) {
		d.PartSize = 1024 * 1024 * 5
		d.MaxUploadParts = 10000
		d.Concurrency = 8
	})
	return &S3Uploader{svc,
		cfg.Oss.DefaultBucket}
}

// baseDir 本地根路径文件夹
// s3 objectKey 集合
var lock = sync.Mutex{}

func (ths *S3Downloader) BatchDownload(baseDir string, keys []string) error {

	lock.Lock()
	if ok, _ := help.PathExists(baseDir); !ok {
		// 创建目录
		err := os.Mkdir(baseDir, os.ModePerm)
		if err != nil {
			log.Errorf("Create dir %s fail on batch download ", err)
			return err
		}
	}
	lock.Unlock()

	// 清空目录
	//err := help.CleanDir(baseDir)
	//if err != nil {
	//	log.Errorf("Clean dir %s fail on batch download ", err)
	//	return err
	//}

	var objects []s3manager.BatchDownloadObject

	var bucketName = ths.getDefaultBucket()
	var keyCount = len(keys)
	for _, key := range keys {
		// 创建Key文件
		fileName := filepath.Join(baseDir, key)
		err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
		if nil != err {
			log.Errorf("Create key %s dir err", err)
			continue
		}

		tmpFile, err := os.Create(fileName)
		if nil != err {
			log.Errorf("Create key file %s  err", err)
			continue
		}

		objects = append(objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			},
			Writer: tmpFile,
			After: func() error {
				keyCount--
				log.Debugf("Remaining %s", strconv.Itoa(keyCount))
				if keyCount <= 0 {
					log.Debug("----------------------Download done--------------------------")
				}
				return nil
			},
		})
	}

	log.Debugf("----------------------Download total: %d--------------------------", len(keys))

	iter := &s3manager.DownloadObjectsIterator{Objects: objects}
	if err := ths.Downloader.DownloadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return err
	}
	return nil
}

// 上传单个大文件
func (ths *S3Uploader) UploadFile(filename string) error {
	return ths.UploadFileByBucket("", filename)
}

// 上传单个大文件指定的桶
func (ths *S3Uploader) UploadFileByBucket(bucket string, filename string) error {

	if ok, _ := help.PathExists(filename); !ok {
		return pugerr.ViolationError(filename + " not found .")
	}
	uploadFile, err := os.Open(filename)
	if err != nil {
		return pugerr.ViolationErrorWithErr("open "+filename+" not found .", err)
	}

	_, err = ths.Upload(&s3manager.UploadInput{
		Bucket: aws.String(ths.getBucketName(bucket)),
		Key:    aws.String(filename),
		Body:   uploadFile,
	},
		func(u *s3manager.Uploader) {
			u.PartSize = 10 * 1024 * 1024 // 缓存块
			u.LeavePartsOnError = true    // 上传失败时，不删除零时文件
		})

	log.Debugf("Upload ok, return %s")
	return nil
}

// 批量上传文件
func (ths *S3Uploader) BatchUpload(uploadDir string, extension map[string]int) error {

	if ok, _ := help.PathExists(uploadDir); !ok {
		return pugerr.ViolationError(uploadDir + " not found .")
	}

	relFileMap, err := help.WalkDir(uploadDir, extension)

	if err != nil {
		return pugerr.UndefinedError(err)
	}

	var objects []s3manager.BatchUploadObject
	bucketName := ths.getDefaultBucket()
	var keyCount = len(relFileMap)

	for key, filePath := range relFileMap {
		tmpFile, err := os.Open(filePath)
		if nil != err {
			log.Errorf("Open file %s  err %s", filePath, err)
			continue
		}

		objects = append(objects, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
				Body:   tmpFile,
			},
			After: func() error {
				keyCount--
				log.Debugf("Remaining %s", strconv.Itoa(keyCount))
				if keyCount <= 0 {
					log.Debug("----------------------Upload done--------------------------")
				}
				return nil
			},
		})
	}

	log.Debugf("----------------------Upload total: %d--------------------------", len(relFileMap))

	iter := &s3manager.UploadObjectsIterator{Objects: objects}
	if err := ths.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return err
	}

	return nil
}

func handleError(err error) error {
	aerr, ok := err.(awserr.Error)
	if ok {
		if aerr.Code() == s3.ErrCodeNoSuchKey {
			return pugerr.ViolationError("Path does not exist.")
		} else if aerr.Code() == s3.ErrCodeNoSuchBucket {
			return pugerr.ViolationError("Bucket does not exist.")
		} else if aerr.Code() == s3.ErrCodeBucketAlreadyExists {
			return pugerr.ViolationError("Bucket already existed.")
		} else {
			return pugerr.UndefinedError(err)
		}
	}
	return err
}
