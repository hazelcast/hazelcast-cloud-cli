package internal

import (
	"github.com/fatih/color"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"os"
	"reflect"
	"strings"
)

func NewClient() *hazelcastcloud.Client {
	var apiKey = os.Getenv("HZ_CLOUD_API_KEY")
	var apiSecret = os.Getenv("HZ_CLOUD_API_SECRET")

	if len(strings.TrimSpace(apiKey)) == 0 || len(strings.TrimSpace(apiSecret)) == 0 {
		configService := NewConfigService()
		apiKey = configService.Get("api-key")
		apiSecret = configService.Get("api-secret")
	}

	if len(strings.TrimSpace(apiKey)) == 0 || len(strings.TrimSpace(apiSecret)) == 0 {
		println("Authentication Error: hzcloud CLI tool is not configured correctly. " +
			"You need to login via `hzcloud login` or set `HZ_CLOUD_API_KEY` and `HZ_CLOUD_API_SECRET` environment" +
			" variables. For more details https://github.com/hazelcast/hazelcast-cloud-cli#authentication-with-hazelcast-cloud")
		os.Exit(1)
	}

	return Validate(Login(apiKey, apiSecret)).(*hazelcastcloud.Client)
}

func Login(apiKey string, apiSecret string) (*hazelcastcloud.Client, *hazelcastcloud.Response, error) {
	var apiUrl = os.Getenv("HZ_CLOUD_API_URL")
	if len(strings.TrimSpace(apiUrl)) != 0 {
		return hazelcastcloud.NewFromCredentials(apiKey, apiSecret, hazelcastcloud.OptionEndpoint(apiUrl))
	}
	return hazelcastcloud.NewFromCredentials(apiKey, apiSecret)
}

func Validate(a interface{}, b *hazelcastcloud.Response, c error) interface{} {
	if c != nil {
		if reflect.TypeOf(c) == reflect.TypeOf(&hazelcastcloud.ErrorResponse{}) {
			color.Red("Message:%s\nCorrelationId:%s", c.(*hazelcastcloud.ErrorResponse).Message,
				c.(*hazelcastcloud.ErrorResponse).CorrelationId)
		} else {
			color.Red(c.Error())
		}
		os.Exit(1)
	}
	return a
}
