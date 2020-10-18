module github.com/hazelcast/hazelcast-cloud-cli

go 1.15

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/azure-sdk-for-go v47.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.10
	github.com/Azure/go-autorest/autorest/adal v0.9.5
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/aws/aws-sdk-go v1.35.9
	github.com/blang/semver/v4 v4.0.0
	github.com/fatih/color v1.9.0
	github.com/google/uuid v1.1.2
	github.com/hazelcast/hazelcast-cloud-sdk-go v1.0.1
	github.com/jedib0t/go-pretty/v6 v6.0.5
	github.com/magiconair/properties v1.8.4
	github.com/spf13/cobra v1.1.0
	github.com/spf13/viper v1.7.1
	google.golang.org/api v0.33.0
)

replace github.com/hazelcast/hazelcast-cloud-sdk-go v1.0.1 => github.com/yunussandikci/hazelcast-cloud-sdk-go v1.1.5
