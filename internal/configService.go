package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ConfigKey string

const (
	ApiKey               ConfigKey = "api-key"
	ApiSecret            ConfigKey = "api-secret"
	LastVersionCheckTime ConfigKey = "last-version-check-time"
)

type ConfigService interface {
	Set(key ConfigKey, value string)
	Get(key ConfigKey) string
}

type configService struct {
	ConfigPath     string
	FullConfigPath string
}

func NewConfigService() ConfigService {
	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		panic(homeDirErr)
	}
	configFile := "config.json"
	configPath := fmt.Sprintf("%s/.hazelcastcloud", homeDir)
	fullConfigPath := fmt.Sprintf("%s/%s", configPath, configFile)

	return &configService{
		ConfigPath:     configPath,
		FullConfigPath: fullConfigPath,
	}
}

func (c configService) getConfig() map[string]string {
	_ = os.Mkdir(c.ConfigPath, 0777)
	readFile, readFileErr := os.OpenFile(c.FullConfigPath, os.O_RDONLY|os.O_CREATE, 0644)
	if readFileErr != nil {
		panic(readFileErr)
	}

	readData, readDataErr := ioutil.ReadAll(readFile)
	if readDataErr != nil {
		panic(readDataErr)
	}

	jsonMap := make(map[string]string)
	unmarshallErr := json.Unmarshal(readData, &jsonMap)
	if unmarshallErr != nil && len(readData) != 0 {
		panic(unmarshallErr)
	}

	return jsonMap
}

func (c configService) Set(key ConfigKey, value string) {
	config := c.getConfig()
	config[string(key)] = value
	configJson, marshallErr := json.Marshal(config)
	if marshallErr != nil {
		panic(marshallErr)
	}
	writeFileErr := ioutil.WriteFile(c.FullConfigPath, configJson, 0644)
	if writeFileErr != nil {
		panic(writeFileErr)
	}
}

func (c configService) Get(key ConfigKey) string {
	return c.getConfig()[string(key)]
}