package service

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
)

type AwsPeeringService struct {
	client                     *hazelcastcloud.Client
	customerPeeringProperties  *AwsCustomerPeeringProperties
	hazelcastPeeringProperties *models.AwsPeeringProperties
}

type AwsCustomerPeeringProperties struct {
	ClusterId string
	VpcId     string
	SubnetIds []string
}

func NewAwsPeeringService(client *hazelcastcloud.Client, customerProperties *AwsCustomerPeeringProperties) AwsPeeringService {
	return AwsPeeringService{
		client:                    client,
		customerPeeringProperties: customerProperties,
	}
}

func (s *AwsPeeringService) Create(indicator *util.LoadingIndicator) error {
	initHazelcastPeeringPropertiesErr := s.initHazelcastPeeringProperties()
	if initHazelcastPeeringPropertiesErr != nil {
		return initHazelcastPeeringPropertiesErr
	}
	initClientErr := s.initClients()
	if initClientErr != nil {
		return initClientErr
	}
	return nil
}

func (s *AwsPeeringService) initHazelcastPeeringProperties() error {
	hazelcastPeeringProperties, _, hazelcastPeeringPropertiesErr := s.client.AwsPeering.GetProperties(context.Background(), &models.GetAwsPeeringPropertiesInput{
		ClusterId: s.customerPeeringProperties.ClusterId,
	})
	if hazelcastPeeringPropertiesErr != nil {
		return hazelcastPeeringPropertiesErr
	}
	s.hazelcastPeeringProperties = hazelcastPeeringProperties
	return nil
}

func (s *AwsPeeringService) initClients() error {
	return nil
}
