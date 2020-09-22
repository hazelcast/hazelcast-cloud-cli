package internal

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestConfigService_CreateConfig(t *testing.T) {
	//given
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}
	defer os.RemoveAll(path)
	newConfigService := NewConfigServiceWith(path)

	//when
	newConfigService.CreateConfig()

	//then
	_, err = os.Stat(fmt.Sprintf("%s/.hazelcastcloud/config.yaml", path))
	assert.Equal(t, err, nil)
}

func TestConfigService_Set_And_Get(t *testing.T) {
	//given
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}
	defer os.RemoveAll(path)
	newConfigService := NewConfigServiceWith(path)

	//when
	newConfigService.Set("key-1", "value-1")

	//then
	key := newConfigService.GetString("key-1")
	assert.Equal(t, key, "value-1")
}

func TestNewConfigService_Set_And_Get_Int64(t *testing.T) {
	//given
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}
	defer os.RemoveAll(path)
	newConfigService := NewConfigServiceWith(path)

	//when
	newConfigService.Set("key-1", 1600251122)
	key := newConfigService.GetInt64("key-1")

	//then
	assert.Equal(t, uint64(key), uint64(1600251122))
}

func TestConfigService_GetFullConfigPath(t *testing.T) {
	//given
	dir, _ := os.UserHomeDir()
	path := fmt.Sprintf("%s/.hazelcastcloud/config.yaml", dir)
	newConfigService := NewConfigService()

	//when
	fullPathConfig := newConfigService.GetFullConfigPath()

	//then
	assert.Equal(t, fullPathConfig, path)
}
