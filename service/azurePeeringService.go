package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"strings"
	"time"
)

type AzurePeeringService struct {
	client                         *hazelcastcloud.Client
	customerPeeringProperties      *AzureCustomerPeeringProperties
	hazelcastPeeringProperties     *models.AzurePeeringProperties
	hazelcastVnetPeeringClient     network.VirtualNetworkPeeringsClient
	customerVnetPeeringClient      network.VirtualNetworkPeeringsClient
	customerServicePrincipalClient graphrbac.ServicePrincipalsClient
	customerRoleAssignmentClient   authorization.RoleAssignmentsClient
	servicePrincipal               graphrbac.ServicePrincipal
	customerVnetPeering            network.VirtualNetworkPeering
	hazelcastVnetPeering           network.VirtualNetworkPeering
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
		client:                    client,
		customerPeeringProperties: customerProperties,
	}
}

func (s *AzurePeeringService) Create(indicator *util.LoadingIndicator) error {
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
	deleteOrphanPeeringsErr := s.deleteOrphanPeerings()
	if deleteOrphanPeeringsErr != nil {
		return deleteOrphanPeeringsErr
	}
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

func (s *AzurePeeringService) notifyPeering() error {
	jsonObject, _ := json.Marshal(struct {
		ClusterId           string `json:"clusterId"`
		PeeringConnectionId string `json:"peeringConnectionId"`
		VnetId              string `json:"vpcId"`
		VnetCidr            string `json:"vpcCidr"`
	}{
		s.customerPeeringProperties.ClusterId,
		*s.hazelcastVnetPeering.Name,
		s.customerPeeringProperties.VnetName,
		(*s.hazelcastVnetPeering.RemoteAddressSpace.AddressPrefixes)[0],
	})
	request, requestErr := http.NewRequest("POST", fmt.Sprintf("https://%s/peerings", s.client.BaseURL.Host), bytes.NewBuffer(jsonObject))
	if requestErr != nil {
		return requestErr
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	request.Header.Add("User-Agent", "hazelcast-cloud-sdk-go")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.client.Token))
	client := &http.Client{}
	_, respErr := client.Do(request)
	if respErr != nil {
		return respErr
	}
	return nil
}

func (s *AzurePeeringService) createHazelcastPeering() error {
	hazelcastPeeringName :=  s.generatePeeringName()
	createHazelcastPeering, createHazelcastPeeringErr := s.hazelcastVnetPeeringClient.CreateOrUpdate(context.Background(),
		s.hazelcastPeeringProperties.ResourceGroupName, s.hazelcastPeeringProperties.VnetName, hazelcastPeeringName,
		network.VirtualNetworkPeering{VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
			AllowVirtualNetworkAccess: to.BoolPtr(true),
			AllowForwardedTraffic:     to.BoolPtr(true),
			RemoteVirtualNetwork: &network.SubResource{
				ID: to.StringPtr(s.getCustomerVnetId()),
			},
		}})
	if createHazelcastPeeringErr != nil {
		return createHazelcastPeeringErr
	}
	_ = createHazelcastPeering.WaitForCompletionRef(context.Background(), s.hazelcastVnetPeeringClient.Client)
	hazelcastPeering, hazelcastPeeringErr := s.hazelcastVnetPeeringClient.Get(context.Background(),
		s.hazelcastPeeringProperties.ResourceGroupName, s.hazelcastPeeringProperties.VnetName, hazelcastPeeringName)
	if hazelcastPeeringErr != nil {
		return hazelcastPeeringErr
	}
	s.hazelcastVnetPeering = hazelcastPeering
	return nil
}

func (s *AzurePeeringService) createCustomerPeering() error {
	customerPeeringName := s.generatePeeringName()
	createCustomerPeering, createCustomerPeeringErr := s.customerVnetPeeringClient.CreateOrUpdate(context.Background(),
		s.customerPeeringProperties.ResourceGroupName, s.customerPeeringProperties.VnetName, customerPeeringName,
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
	_ = createCustomerPeering.WaitForCompletionRef(context.Background(), s.customerVnetPeeringClient.Client)
	customerPeering, customerPeeringErr := s.customerVnetPeeringClient.Get(context.Background(),
		s.customerPeeringProperties.ResourceGroupName, s.customerPeeringProperties.VnetName, customerPeeringName)
	if customerPeeringErr != nil {
		return customerPeeringErr
	}
	s.customerVnetPeering = customerPeering
	return nil
}

func (s *AzurePeeringService) deleteOrphanPeerings() error {
	customerPeeringList, customerPeeringListErr := s.customerVnetPeeringClient.List(context.Background(),
		s.customerPeeringProperties.ResourceGroupName, s.customerPeeringProperties.VnetName)
	if customerPeeringListErr != nil {
		return customerPeeringListErr
	}
	for _,customerPeer := range customerPeeringList.Values() {
		if  *customerPeer.RemoteVirtualNetwork.ID == s.getHazelcastVnetId() {
			if customerPeer.PeeringState != network.VirtualNetworkPeeringStateConnected {
				deleteCustomerPeering, deleteCustomerPeeringErr := s.customerVnetPeeringClient.Delete(context.Background(),
					s.customerPeeringProperties.ResourceGroupName, s.customerPeeringProperties.VnetName, *customerPeer.Name)
				if deleteCustomerPeeringErr != nil {
					return deleteCustomerPeeringErr
				}
				_ = deleteCustomerPeering.WaitForCompletionRef(context.Background(), s.customerVnetPeeringClient.Client)
			} else {
				return errors.New(fmt.Sprintf("You already have one connected peering connection named %s.", *customerPeer.Name))
			}
		}
	}
	return nil
}

func (s *AzurePeeringService) createRoleAssignment() error {
	networkContributorRoleId := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7",
		s.customerPeeringProperties.SubscriptionId)
	_, roleAssignmentErr := s.customerRoleAssignmentClient.Create(context.Background(), s.getCustomerVnetId(),
		uuid.New().String(), authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &networkContributorRoleId,
				PrincipalID:      s.servicePrincipal.ObjectID,
			},
		})
	if roleAssignmentErr != nil {
		if roleAssignmentErr.(autorest.DetailedError).StatusCode != 409 {
			return roleAssignmentErr
		} else {
			return nil
		}
	}
	time.Sleep(20 * time.Second)
	return nil
}

func (s *AzurePeeringService) createServicePrincipal() error {
	servicePrincipal, createServicePrincipalErr := s.customerServicePrincipalClient.Create(context.Background(),
		graphrbac.ServicePrincipalCreateParameters{
			AppID:          &s.hazelcastPeeringProperties.AppRegistrationId,
			AccountEnabled: to.BoolPtr(true),
		})
	if createServicePrincipalErr != nil {
		if createServicePrincipalErr.(autorest.DetailedError).StatusCode != 409 {
			return createServicePrincipalErr
		}
		servicePrincipalList, servicePrincipalListErr := s.customerServicePrincipalClient.List(
			context.Background(), fmt.Sprintf("appId eq '%s'", s.hazelcastPeeringProperties.AppRegistrationId))
		if servicePrincipalListErr != nil {
			return servicePrincipalListErr
		}
		s.servicePrincipal = servicePrincipalList.Values()[0]
		return nil
	}
	time.Sleep(60 * time.Second)
	s.servicePrincipal = servicePrincipal
	return nil
}

func (s *AzurePeeringService) initHazelcastPeeringProperties() error {
	hazelcastPeeringProperties, _, hazelcastPeeringPropertiesErr := s.client.AzurePeering.GetProperties(context.Background(), &models.GetAzurePeeringPropertiesInput{
		ClusterId: s.customerPeeringProperties.ClusterId,
	})
	if hazelcastPeeringPropertiesErr != nil {
		return hazelcastPeeringPropertiesErr
	}
	s.hazelcastPeeringProperties = hazelcastPeeringProperties
	return nil
}

func (s *AzurePeeringService) initClients() error {
	env, envErr := azure.EnvironmentFromName("AzurePublicCloud")
	if envErr != nil {
		return envErr
	}

	hazelcastOauthConfig, hazelcastOauthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint,
		s.hazelcastPeeringProperties.TenantId, []string{s.customerPeeringProperties.TenantId}, adal.OAuthOptions{})
	if hazelcastOauthConfigErr != nil {
		return hazelcastOauthConfigErr
	}

	hazelcastToken, hazelcastTokenErr := adal.NewMultiTenantServicePrincipalToken(hazelcastOauthConfig,
		s.hazelcastPeeringProperties.AppRegistrationId, s.hazelcastPeeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
	if hazelcastTokenErr != nil {
		return hazelcastOauthConfigErr
	}

	customerOauthConfig, customerOauthConfigErr := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint,
		s.customerPeeringProperties.TenantId, []string{s.hazelcastPeeringProperties.TenantId}, adal.OAuthOptions{})
	if customerOauthConfigErr != nil {
		return customerOauthConfigErr
	}

	customerToken, customerTokenErr := adal.NewMultiTenantServicePrincipalToken(customerOauthConfig,
		s.hazelcastPeeringProperties.AppRegistrationId, s.hazelcastPeeringProperties.AppRegistrationKey, env.ResourceManagerEndpoint)
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

	s.hazelcastVnetPeeringClient = network.NewVirtualNetworkPeeringsClient(s.hazelcastPeeringProperties.SubscriptionId)
	s.hazelcastVnetPeeringClient.Authorizer = autorest.NewMultiTenantBearerAuthorizer(hazelcastToken)
	s.customerVnetPeeringClient = network.NewVirtualNetworkPeeringsClient(s.customerPeeringProperties.SubscriptionId)
	s.customerVnetPeeringClient.Authorizer = autorest.NewMultiTenantBearerAuthorizer(customerToken)
	s.customerServicePrincipalClient = graphrbac.NewServicePrincipalsClient(s.customerPeeringProperties.TenantId)
	s.customerServicePrincipalClient.Authorizer = customerGraphAuthorizer
	s.customerRoleAssignmentClient = authorization.NewRoleAssignmentsClient(s.customerPeeringProperties.SubscriptionId)
	s.customerRoleAssignmentClient.Authorizer = customerAuthorizer

	return nil
}

func (s *AzurePeeringService) getCustomerVnetId() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s",
		s.customerPeeringProperties.SubscriptionId, s.customerPeeringProperties.ResourceGroupName, s.customerPeeringProperties.VnetName)
}

func (s *AzurePeeringService) getHazelcastVnetId() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s",
		s.hazelcastPeeringProperties.SubscriptionId, s.hazelcastPeeringProperties.ResourceGroupName, s.hazelcastPeeringProperties.VnetName)
}

func (s *AzurePeeringService) generatePeeringName() string {
	return strings.Replace(strings.ToLower(uuid.New().String()), "-", "", -1)[0:8]
}
