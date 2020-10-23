package service

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"google.golang.org/api/compute/v1"
)

type GcpPeeringService struct {
	Client             *hazelcastcloud.Client
	CustomerProperties *GcpCustomerPeeringProperties
}

type GcpCustomerPeeringProperties struct {
	ClusterId   string
	ProjectId   string
	NetworkName string
}

func NewGcpPeeringService(client *hazelcastcloud.Client) GcpPeeringService {
	return GcpPeeringService{
		Client: client,
	}
}

func (s GcpPeeringService) Create(customerProperties *GcpCustomerPeeringProperties) error {
	hazelcastProperties := internal.Validate(s.Client.GcpPeering.GetProperties(context.Background(), &models.GetGcpPeeringPropertiesInput{
		ClusterId: customerProperties.ClusterId,
	})).(*models.GcpPeeringProperties)

	computeService, computeServiceErr := compute.NewService(context.Background())
	if computeServiceErr != nil {
		return fmt.Errorf("you need to have GOOGLE_APPLICATION_CREDENTIALS environment variable set in order to perform this action. For more information https://docs.cloud.hazelcast.com/docs/gcp-vpc-peering . GCP Error:%s", computeServiceErr)
	}

	_, addPeeringErr := computeService.Networks.AddPeering(customerProperties.ProjectId, customerProperties.NetworkName, &compute.NetworksAddPeeringRequest{
		Name:             fmt.Sprintf("%s-%s", hazelcastProperties.ProjectId, hazelcastProperties.NetworkName),
		PeerNetwork:      fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", hazelcastProperties.ProjectId, hazelcastProperties.NetworkName),
		AutoCreateRoutes: true,
	}).Do()
	if addPeeringErr != nil {
		return addPeeringErr
	}

	_ = internal.Validate(s.Client.GcpPeering.Accept(context.Background(), &models.AcceptGcpPeeringInput{
		ClusterId:   customerProperties.ClusterId,
		ProjectId:   customerProperties.ProjectId,
		NetworkName: customerProperties.NetworkName,
	}))

	return nil
}
