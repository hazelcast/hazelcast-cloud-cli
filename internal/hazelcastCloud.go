package internal

import (
	"github.com/fatih/color"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"os"
	"reflect"
	"strings"
)

var hazelcastCloudClient *hazelcastcloud.Client

func NewClient() *hazelcastcloud.Client {
	var apiKey = os.Getenv("HZ_CLOUD_API_KEY")
	var apiSecret = os.Getenv("HZ_CLOUD_API_SECRET")
	if len(strings.TrimSpace(apiKey)) == 0 || len(strings.TrimSpace(apiSecret)) == 0 {
		configService := NewConfigService()
		apiKey = configService.GetString("api-key")
		apiSecret = configService.GetString("api-secret")
	}

	if len(strings.TrimSpace(apiKey)) == 0 || len(strings.TrimSpace(apiSecret)) == 0 {
		println("Authentication Error: hzcloud CLI tool  is  not configured correctly, please see for more details https://github.com/hazelcast/hazelcast-cloud-cli/configuration")
		os.Exit(1)
	}

	client := Validate(hazelcastcloud.NewFromCredentials(apiKey, apiSecret))
	hazelcastCloudClient = client.(*hazelcastcloud.Client)
	return hazelcastCloudClient
}

func Validate(a interface{}, b *hazelcastcloud.Response, c error) interface{} {
	if c != nil {
		if reflect.TypeOf(c) == reflect.TypeOf(&hazelcastcloud.ErrorResponse{}) {
			color.Red("Message:%s CorrelationId:%s", c.(*hazelcastcloud.ErrorResponse).Message,
				c.(*hazelcastcloud.ErrorResponse).CorrelationId)
		} else {
			color.Red(c.Error())
		}
		os.Exit(1)
	}
	return a
}
