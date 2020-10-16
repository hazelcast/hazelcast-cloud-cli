package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"net/http"
	"time"
)

type AzurePeeringService struct {
	Client *hazelcastcloud.Client
}

type AzureCustomerPeeringProperties struct {
	ClusterId         string
	VnetName          string
	SubscriptionId    string
	TenantId          string
	ResourceGroupName string
}

func NewAzurePeeringService(client *hazelcastcloud.Client) AzurePeeringService {
	return AzurePeeringService{
		Client: client,
	}
}

func (s AzurePeeringService) Create(customerProperties *AzureCustomerPeeringProperties, indicator *util.LoadingIndicator) error {
	hazelcastPeeringProperties := internal.Validate(s.Client.AzurePeering.GetProperties(context.Background(), &models.GetAzurePeeringPropertiesInput{
		ClusterId: customerProperties.ClusterId,
	})).(*models.AzurePeeringProperties)

	indicator.SetStep("Authorizers collecting...", 10)
	hazelcastMultiTenantAuthorizer, hazelcastMultiTenantAuthorizerErr := s.getHazelcastMultiTenantAuthorizer(hazelcastPeeringProperties, *customerProperties)
	if hazelcastMultiTenantAuthorizerErr != nil {
		return hazelcastMultiTenantAuthorizerErr
	}

	customerMultiTenantAuthorizer, customerMultiTenantAuthorizerErr := s.getCustomerMultiTenantAuthorizer(hazelcastPeeringProperties, *customerProperties)
	if customerMultiTenantAuthorizerErr != nil {
		return customerMultiTenantAuthorizerErr
	}

	customerAuthorizer, customerAuthorizerErr := s.getCustomerAuthorizer()
	if customerAuthorizerErr != nil {
		return customerAuthorizerErr
	}

	customerGraphAuthorizer, customerGraphAuthorizerErr := s.getCustomerGraphAuthorizer()
	if customerGraphAuthorizerErr != nil {
		return customerGraphAuthorizerErr
	}

	hazelcastVnetPeeringClient := network.NewVirtualNetworkPeeringsClient(hazelcastPeeringProperties.SubscriptionId)
	hazelcastVnetPeeringClient.Authorizer = hazelcastMultiTenantAuthorizer
	customerVnetPeeringClient := network.NewVirtualNetworkPeeringsClient(customerProperties.SubscriptionId)
	customerVnetPeeringClient.Authorizer = customerMultiTenantAuthorizer
	customerServicePrincipalClient := graphrbac.NewServicePrincipalsClient(customerProperties.TenantId)
	customerServicePrincipalClient.Authorizer = customerGraphAuthorizer
	customerRoleAssignmentClient := authorization.NewRoleAssignmentsClient(customerProperties.SubscriptionId)
	customerRoleAssignmentClient.Authorizer = customerAuthorizer

	indicator.SetStep("Service Principal creating...", 30)
	servicePrincipal, createServicePrincipalErr := customerServicePrincipalClient.Create(context.Background(), graphrbac.ServicePrincipalCreateParameters{
		AppID:          &hazelcastPeeringProperties.AppRegistrationId,
		AccountEnabled: to.BoolPtr(true),
	})
	if createServicePrincipalErr != nil {
		if createServicePrincipalErr.(autorest.DetailedError).StatusCode != 409 {
			return createServicePrincipalErr
		}
		list, hazelcastServicePrincipalFindErr := customerServicePrincipalClient.List(context.Background(), fmt.Sprintf("appId eq '%s'", hazelcastPeeringProperties.AppRegistrationId))
		if hazelcastServicePrincipalFindErr != nil {
			return hazelcastServicePrincipalFindErr
		}
		servicePrincipal = list.Values()[0]
	} else {
		indicator.SetStep("Waiting for Service Principal to be ready...", 40)
		time.Sleep(30 * time.Second)
	}

	indicator.SetStep("Role Assignment creating...", 50)
	networkContributorRoleId := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7", customerProperties.SubscriptionId)
	_, roleAssignmentErr := customerRoleAssignmentClient.Create(context.Background(), s.getCustomerVnetId(customerProperties),
		uuid.New().String(), authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &networkContributorRoleId,
				PrincipalID:      servicePrincipal.ObjectID,
			},
		})
	if roleAssignmentErr != nil && roleAssignmentErr.(autorest.DetailedError).StatusCode != 409 {
		return roleAssignmentErr
	}

	indicator.SetStep("Deleting Orphan Peerings...", 60)



	indicator.SetStep("Customer Peering started...", 70)
	hazelcastVnetId := s.getHazelcastVnetId(hazelcastPeeringProperties)
	customerPeerName := uuid.New().String()
	_, customerPeerErr := customerVnetPeeringClient.CreateOrUpdate(context.Background(),
		customerProperties.ResourceGroupName, customerProperties.VnetName, customerPeerName, network.VirtualNetworkPeering{
			VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
				AllowVirtualNetworkAccess: to.BoolPtr(true),
				AllowForwardedTraffic:     to.BoolPtr(true),
				RemoteVirtualNetwork: &network.SubResource{
					ID: &hazelcastVnetId,
				},
			},
			Name: &customerPeerName,
		})
	if customerPeerErr != nil && customerPeerErr.(autorest.DetailedError).Original.(*azure.ServiceError).Code != "AnotherPeeringAlreadyReferencesRemoteVnet" {
		return customerPeerErr
 	}

	if customerPeerErr == nil {
		marshal, _ := json.Marshal(struct {
			PeeringConnectionId string `json:"peeringConnectionId"`
			ClusterId           string `json:"clusterId"`
		}{
			customerPeerName,
			customerProperties.ClusterId,
		})
		peeringEndpoint := fmt.Sprintf("%s/peerings", s.Client.BaseURL)
		req, reqErr := http.NewRequest("POST", peeringEndpoint, bytes.NewBuffer(marshal))
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Client.Token))
		client := &http.Client{}
		resp, respErr := client.Do(req)
		if respErr != nil {
			return respErr
		}
		fmt.Println(resp)
	}

	indicator.SetStep("Hazelcast Peering started...", 90)
	hazelcastPeerName := uuid.New().String()
	customerVnetId := s.getCustomerVnetId(customerProperties)
	_, hazelcastPeeringErr := hazelcastVnetPeeringClient.CreateOrUpdate(context.Background(),
		hazelcastPeeringProperties.ResourceGroupName, hazelcastPeeringProperties.VnetName, hazelcastPeerName, network.VirtualNetworkPeering{
			VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
				AllowVirtualNetworkAccess: to.BoolPtr(true),
				AllowForwardedTraffic:     to.BoolPtr(true),
				RemoteVirtualNetwork: &network.SubResource{
					ID: &customerVnetId,
				},
			},
			Name: &hazelcastPeerName,
		})
	if hazelcastPeeringErr != nil && hazelcastPeeringErr.(autorest.DetailedError).Original.(*azure.ServiceError).Code != "AnotherPeeringAlreadyReferencesRemoteVnet" {
		return hazelcastPeeringErr
	}
	return nil
}

func (s AzurePeeringService) getCustomerPeeringName(customerPeeringProperties *AzureCustomerPeeringProperties, pee) string {
	return fmt.Sprintf("%s-to-%s", peeringProperties.VnetName, )
}

func (s AzurePeeringService) getCustomerVnetId(peeringProperties *AzureCustomerPeeringProperties) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/VirtualNetworks/%s",
		peeringProperties.SubscriptionId, peeringProperties.ResourceGroupName, peeringProperties.VnetName)
}

func (s AzurePeeringService) getHazelcastVnetId(peeringProperties *models.AzurePeeringProperties) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/VirtualNetworks/%s",
		peeringProperties.SubscriptionId, peeringProperties.ResourceGroupName, peeringProperties.VnetName)
}

func (s AzurePeeringService) getCustomerGraphAuthorizer() (autorest.Authorizer, error) {
	azureEnv, azureEnvErr := azure.EnvironmentFromName("AzurePublicCloud")
	if azureEnvErr != nil {
		return nil, azureEnvErr
	}
	return auth.NewAuthorizerFromCLIWithResource(azureEnv.GraphEndpoint)
}

func (s AzurePeeringService) getCustomerAuthorizer() (autorest.Authorizer, error) {
	return auth.NewAuthorizerFromCLI()
}

func (s AzurePeeringService) getHazelcastMultiTenantAuthorizer(peeringProperties *models.AzurePeeringProperties, customerPeeringProperties AzureCustomerPeeringProperties) (autorest.Authorizer, error) {
	env, envErr := azure.EnvironmentFromName("AzurePublicCloud")
	if envErr != nil {
		return nil, envErr
	}

	oAuthConfig, oAuthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint, peeringProperties.TenantId, []string{customerPeeringProperties.TenantId}, adal.OAuthOptions{})
	if oAuthConfigErr != nil {
		return nil, oAuthConfigErr
	}

	token, envErr := adal.NewMultiTenantServicePrincipalToken(oAuthConfig, peeringProperties.AppRegistrationId, peeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
	return autorest.NewMultiTenantBearerAuthorizer(token), nil
}

func (s AzurePeeringService) getCustomerMultiTenantAuthorizer(peeringProperties *models.AzurePeeringProperties, customerPeeringProperties AzureCustomerPeeringProperties) (autorest.Authorizer, error) {
	env, envErr := azure.EnvironmentFromName("AzurePublicCloud")
	if envErr != nil {
		return nil, envErr
	}

	oAuthConfig, oAuthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint, customerPeeringProperties.TenantId, []string{peeringProperties.TenantId}, adal.OAuthOptions{})
	if oAuthConfigErr != nil {
		return nil, oAuthConfigErr
	}

	token, envErr := adal.NewMultiTenantServicePrincipalToken(oAuthConfig, peeringProperties.AppRegistrationId, peeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
	return autorest.NewMultiTenantBearerAuthorizer(token), nil
}
