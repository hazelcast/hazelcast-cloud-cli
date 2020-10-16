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
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"net/http"
	"time"
)

type AzurePeeringService struct {
	Client                         *hazelcastcloud.Client
	CustomerPeeringProperties      *AzureCustomerPeeringProperties
	HazelcastPeeringProperties     *models.AzurePeeringProperties
	HazelcastVnetPeeringClient     network.VirtualNetworkPeeringsClient
	CustomerVnetPeeringClient      network.VirtualNetworkPeeringsClient
	CustomerServicePrincipalClient graphrbac.ServicePrincipalsClient
	CustomerRoleAssignmentClient   authorization.RoleAssignmentsClient
	ServicePrincipalObjectId       string
}

type AzureCustomerPeeringProperties struct {
	ClusterId         string
	VnetName          string
	SubscriptionId    string
	TenantId          string
	ResourceGroupName string
}

func NewAzurePeeringService(client *hazelcastcloud.Client, customerProperties *AzureCustomerPeeringProperties) AzurePeeringService {
	return AzurePeeringService{
		Client:                    client,
		CustomerPeeringProperties: customerProperties,
	}
}

func (s AzurePeeringService) Create(indicator *util.LoadingIndicator) error {
	indicator.SetStep("Peering Properties collecting...", 10)
	initHazelcastPeeringPropertiesErr := s.initHazelcastPeeringProperties()
	if initHazelcastPeeringPropertiesErr != nil {
		return initHazelcastPeeringPropertiesErr
	}
	indicator.SetStep("Clients initializing...", 20)
	initClientErr := s.initClients()
	if initClientErr != nil {
		return initClientErr
	}
	indicator.SetStep("Service Principal creating...", 30)
	initServicePrincipalErr := s.createServicePrincipal()
	if initServicePrincipalErr != nil {
		return initServicePrincipalErr
	}
	indicator.SetStep("Role Assignment creating...", 45)
	initRoleAssignmentsErr := s.createRoleAssignment()
	if initRoleAssignmentsErr != nil {
		return initRoleAssignmentsErr
	}
	indicator.SetStep("Orphan Peerings deleting...", 55)
	s.deleteOrphanPeerings()
	indicator.SetStep("Customer Peering creating...", 65)
	initCustomerPeeringErr := s.createCustomerPeering()
	if initCustomerPeeringErr != nil {
		return initCustomerPeeringErr
	}
	indicator.SetStep("Hazelcast Peering creating...", 80)
	initHazelcastPeeringErr := s.createHazelcastPeering()
	if initHazelcastPeeringErr != nil {
		return initHazelcastPeeringErr
	}
	indicator.SetStep("Peering notifying...", 95)
	notifyPeeringErr := s.notifyPeering()
	if notifyPeeringErr != nil {
		return notifyPeeringErr
	}
	return nil
}

func (s AzurePeeringService) notifyPeering() error {
	marshal, _ := json.Marshal(struct {
		PeeringConnectionId string `json:"peeringConnectionId"`
		ClusterId           string `json:"clusterId"`
	}{
		s.getHazelcastPeeringName(),
		s.CustomerPeeringProperties.ClusterId,
	})
	request, requestErr := http.NewRequest("POST", fmt.Sprintf("%s/peerings", s.Client.BaseURL), bytes.NewBuffer(marshal))
	if requestErr != nil {
		return requestErr
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Client.Token))
	client := &http.Client{}
	_, respErr := client.Do(request)
	if respErr != nil {
		return respErr
	}
	return nil
}

func (s AzurePeeringService) createHazelcastPeering() error {
	_, createCustomerPeeringErr := s.HazelcastVnetPeeringClient.CreateOrUpdate(context.Background(),
		s.HazelcastPeeringProperties.ResourceGroupName, s.HazelcastPeeringProperties.VnetName, s.getHazelcastPeeringName(),
		network.VirtualNetworkPeering{VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
			AllowVirtualNetworkAccess: to.BoolPtr(true),
			AllowForwardedTraffic:     to.BoolPtr(true),
			RemoteVirtualNetwork: &network.SubResource{
				ID: to.StringPtr(s.getCustomerVnetId()),
			},
		}})
	if createCustomerPeeringErr != nil {
		return createCustomerPeeringErr
	}
	return nil
}

func (s AzurePeeringService) createCustomerPeering() error {
	_, createCustomerPeeringErr := s.CustomerVnetPeeringClient.CreateOrUpdate(context.Background(),
		s.CustomerPeeringProperties.ResourceGroupName, s.CustomerPeeringProperties.VnetName, s.getCustomerPeeringName(),
		network.VirtualNetworkPeering{VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
			AllowVirtualNetworkAccess: to.BoolPtr(true),
			AllowForwardedTraffic:     to.BoolPtr(true),
			RemoteVirtualNetwork: &network.SubResource{
				ID: to.StringPtr(s.getHazelcastVnetId()),
			},
		}})
	if createCustomerPeeringErr != nil {
		return createCustomerPeeringErr
	}
	return nil
}

func (s AzurePeeringService) deleteOrphanPeerings() {
	_, _ = s.HazelcastVnetPeeringClient.Delete(context.Background(),
		s.HazelcastPeeringProperties.ResourceGroupName, s.HazelcastPeeringProperties.VnetName, s.getHazelcastPeeringName())
	_, _ = s.CustomerVnetPeeringClient.Delete(context.Background(),
		s.CustomerPeeringProperties.ResourceGroupName, s.CustomerPeeringProperties.VnetName, s.getCustomerPeeringName())
}

func (s AzurePeeringService) createRoleAssignment() error {
	networkContributorRoleId := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7",
		s.CustomerPeeringProperties.SubscriptionId)
	_, roleAssignmentErr := s.CustomerRoleAssignmentClient.Create(context.Background(), s.getCustomerVnetId(),
		uuid.New().String(), authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &networkContributorRoleId,
				PrincipalID:      &s.ServicePrincipalObjectId,
			},
		})
	if roleAssignmentErr != nil && roleAssignmentErr.(autorest.DetailedError).StatusCode != 409 {
		return roleAssignmentErr
	}
	return nil
}

func (s AzurePeeringService) createServicePrincipal() error {
	servicePrincipal, createServicePrincipalErr := s.CustomerServicePrincipalClient.Create(context.Background(),
		graphrbac.ServicePrincipalCreateParameters{
			AppID:          &s.HazelcastPeeringProperties.AppRegistrationId,
			AccountEnabled: to.BoolPtr(true),
		})
	if createServicePrincipalErr != nil {
		if createServicePrincipalErr.(autorest.DetailedError).StatusCode != 409 {
			return createServicePrincipalErr
		}
		servicePrincipalList, servicePrincipalListErr := s.CustomerServicePrincipalClient.List(
			context.Background(), fmt.Sprintf("appId eq '%s'", s.HazelcastPeeringProperties.AppRegistrationId))
		if servicePrincipalListErr != nil {
			return servicePrincipalListErr
		}
		s.ServicePrincipalObjectId = *servicePrincipalList.Values()[0].ObjectID
		return nil
	}
	time.Sleep(30 * time.Second)
	s.ServicePrincipalObjectId = *servicePrincipal.ObjectID
	return nil
}

func (s AzurePeeringService) initHazelcastPeeringProperties() error {
	hazelcastPeeringProperties, _, hazelcastPeeringPropertiesErr := s.Client.AzurePeering.GetProperties(context.Background(), &models.GetAzurePeeringPropertiesInput{
		ClusterId: s.CustomerPeeringProperties.ClusterId,
	})
	if hazelcastPeeringPropertiesErr != nil {
		return hazelcastPeeringPropertiesErr
	}
	s.HazelcastPeeringProperties = hazelcastPeeringProperties
	return nil
}

func (s AzurePeeringService) initClients() error {
	env, envErr := azure.EnvironmentFromName("AzurePublicCloud")
	if envErr != nil {
		return envErr
	}

	hazelcastOauthConfig, hazelcastOauthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint,
		s.HazelcastPeeringProperties.TenantId, []string{s.CustomerPeeringProperties.TenantId}, adal.OAuthOptions{})
	if hazelcastOauthConfigErr != nil {
		return hazelcastOauthConfigErr
	}

	hazelcastToken, hazelcastTokenErr := adal.NewMultiTenantServicePrincipalToken(hazelcastOauthConfig,
		s.HazelcastPeeringProperties.AppRegistrationId, s.HazelcastPeeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
	if hazelcastTokenErr != nil {
		return hazelcastOauthConfigErr
	}

	customerOauthConfig, customerOauthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint,
		s.CustomerPeeringProperties.TenantId, []string{s.HazelcastPeeringProperties.TenantId}, adal.OAuthOptions{})
	if customerOauthConfigErr != nil {
		return customerOauthConfigErr
	}

	customerToken, customerTokenErr := adal.NewMultiTenantServicePrincipalToken(customerOauthConfig,
		s.HazelcastPeeringProperties.AppRegistrationId, s.HazelcastPeeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
	if customerTokenErr != nil {
		return customerTokenErr
	}

	customerAuthorizer, customerAuthorizerErr := auth.NewAuthorizerFromCLI()
	if customerAuthorizerErr != nil {
		return customerAuthorizerErr
	}

	customerGraphAuthorizer, customerGraphAuthorizerErr := auth.NewAuthorizerFromCLIWithResource(env.GraphEndpoint)
	if customerGraphAuthorizerErr != nil {
		return customerGraphAuthorizerErr
	}

	s.HazelcastVnetPeeringClient = network.NewVirtualNetworkPeeringsClient(s.HazelcastPeeringProperties.SubscriptionId)
	s.HazelcastVnetPeeringClient.Authorizer = autorest.NewMultiTenantBearerAuthorizer(hazelcastToken)
	s.CustomerVnetPeeringClient = network.NewVirtualNetworkPeeringsClient(s.CustomerPeeringProperties.SubscriptionId)
	s.CustomerVnetPeeringClient.Authorizer = autorest.NewMultiTenantBearerAuthorizer(customerToken)
	s.CustomerServicePrincipalClient = graphrbac.NewServicePrincipalsClient(s.CustomerPeeringProperties.TenantId)
	s.CustomerServicePrincipalClient.Authorizer = customerGraphAuthorizer
	s.CustomerRoleAssignmentClient = authorization.NewRoleAssignmentsClient(s.CustomerPeeringProperties.SubscriptionId)
	s.CustomerRoleAssignmentClient.Authorizer = customerAuthorizer

	return nil
}

func (s AzurePeeringService) getCustomerPeeringName() string {
	return fmt.Sprintf("peering-to-%s", s.HazelcastPeeringProperties.VnetName)
}

func (s AzurePeeringService) getHazelcastPeeringName() string {
	return fmt.Sprintf("peering-to-%s", s.CustomerPeeringProperties.VnetName)
}

func (s AzurePeeringService) getCustomerVnetId() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/VirtualNetworks/%s",
		s.CustomerPeeringProperties.SubscriptionId, s.CustomerPeeringProperties.ResourceGroupName, s.CustomerPeeringProperties.VnetName)
}

func (s AzurePeeringService) getHazelcastVnetId() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/VirtualNetworks/%s",
		s.HazelcastPeeringProperties.SubscriptionId, s.HazelcastPeeringProperties.ResourceGroupName, s.HazelcastPeeringProperties.VnetName)
}
