module github.com/hazelcast/hazelcast-cloud-cli

go 1.15

require (
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
	github.com/hazelcast/hazelcast-cloud-sdk-go v1.2.0
	github.com/jedib0t/go-pretty/v6 v6.0.5
	github.com/spf13/cobra v1.1.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.33.0
)

//TODO: Remove Replace
replace (
 	github.com/hazelcast/hazelcast-cloud-sdk-go => ../hazelcast-cloud-sdk-go
)
