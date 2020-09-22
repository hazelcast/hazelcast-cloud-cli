package internal

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

const (
	ConfigType string = "yaml"
	ConfigName string = "config"
)

type ConfigKey string

const (
	ApiKey               ConfigKey = "api-key"
	ApiSecret            ConfigKey = "api-secret"
	LastVersionCheckTime ConfigKey = "last-version-check-time"
)

type ConfigService interface {
	CreateConfig()
	Set(key ConfigKey, value interface{})
	GetString(key ConfigKey) string
	GetInt64(key ConfigKey) int64
	GetFullConfigPath() string
}

type configService struct {
	ConfigPath     string
	ConfigFile     string
	FullConfigPath string
}

func NewConfigService() ConfigService {
	homeDir, _ := os.UserHomeDir()
	return NewConfigServiceWith(homeDir)
}

func NewConfigServiceWith(homeDir string) ConfigService {
	configPath := fmt.Sprintf("%s/.hazelcastcloud", homeDir)
	configFile := "config.yaml"
	fullConfigPath := fmt.Sprintf("%s/%s", configPath, configFile)

	return &configService{
		ConfigPath:     configPath,
		ConfigFile:     configFile,
		FullConfigPath: fullConfigPath,
	}
}

func (c configService) CreateConfig() {
	v := viper.New()
	v.SetConfigType(ConfigType)
	v.SetConfigName(ConfigName)
	v.AddConfigPath(c.ConfigPath)

	if readErr := v.ReadInConfig(); readErr != nil {
		if _, ok := readErr.(viper.ConfigFileNotFoundError); ok {
			_, statErr := os.Stat(c.FullConfigPath)
			if !os.IsExist(statErr) {
				_ = os.Mkdir(c.ConfigPath, 0755)
				_, _ = os.Create(c.FullConfigPath)
			}
		}
	}
}

func (c configService) Set(key ConfigKey, value interface{}) {
	viper.AddConfigPath(c.ConfigPath)
	viper.Set(string(key), value)
	_ = viper.WriteConfig()
}

func (c configService) GetString(key ConfigKey) string {
	viper.AddConfigPath(c.ConfigPath)
	_ = viper.ReadInConfig()
	return viper.GetString(string(key))
}

func (c configService) GetInt64(key ConfigKey) int64 {
	viper.AddConfigPath(c.ConfigPath)
	_ = viper.ReadInConfig()
	return viper.GetInt64(string(key))
}

func (c configService) GetFullConfigPath() string {
	return c.FullConfigPath
}
