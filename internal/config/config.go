package config

import (
	"context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
)

type AppConfig struct {
	Mysql struct {
		Url      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`

	Oss struct {
		AccessKeyId     string `yaml:"accessKeyId"`
		SecretAccessKey string `yaml:"secretAccessKey"`
		Region          string `yaml:"region"`
		Endpoint        string `yaml:"endpoint"`
		DomainUrl       string `yaml:"domainUrl"`
		DefaultBucket   string `yaml:"defaultBucket"`
	} `yaml:"oss"`

	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Pdl struct {
		BaseDir   string `yaml:"baseDir"`
		ChunkSize int    `yaml:"chunkSize"`
		AsyncSize int    `yaml:"asyncSize"`
		AdvanceAI struct {
			AdvanceAiKey string `yaml:"advanceAiKey"`
			IdCardOcrUrl string `yaml:"idCardOcrUrl"`
		} `yaml:"advanceAI"`
		EventServer struct {
			ThirdUrl string `yaml:"thirdUrl"`
		} `yaml:"eventServer"`
	} `yaml:"pdl"`
}

var (
	instance *AppConfig
	once     sync.Once
)

var cfgFile string

type contextKey string

const appConfigKey = contextKey("appConfig")

// 从上下文加载
func FromContext(ctx context.Context) *AppConfig {
	if appConfig, ok := ctx.Value(appConfigKey).(*AppConfig); ok {
		return appConfig
	}
	return NewAppConfig()
}

// 绑定配置到Context
func WithAppConfig(ctx context.Context, appConfig *AppConfig) context.Context {
	return context.WithValue(ctx, appConfigKey, appConfig)
}

func NewAppConfig() *AppConfig {

	cfg, err := NewConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func NewConfig(configPath string) (*AppConfig, error) {
	config := &AppConfig{}
	filename, _ := filepath.Abs(configPath)
	yamlFile, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	return config, nil
}

// 单例
func InitConfigFile(cfgPath string) {
	cfgFile = cfgPath
}

// 单例
func App() *AppConfig {
	once.Do(func() {
		if cfgFile == "" {
			cfgFile = "configs/secret/config.yml"
		}
		instance = NewAppConfig()
	})
	return instance
}
