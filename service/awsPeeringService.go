package service

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	hazelcastcloud "github.com/hazelcast/hazelcast-cloud-sdk-go"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
)

type AwsPeeringService struct {
	client                     *hazelcastcloud.Client
	customerPeeringProperties  *AwsCustomerPeeringProperties
	hazelcastPeeringProperties *models.AwsPeeringProperties
	ec2                        *ec2.EC2
}

type AwsCustomerPeeringProperties struct {
	SubAccountId string
	ClusterId    string
	Region       string
	VpcId        string
	SubnetIds    []string
}

func NewAwsPeeringService(client *hazelcastcloud.Client, customerProperties *AwsCustomerPeeringProperties) AwsPeeringService {
	return AwsPeeringService{
		client:                    client,
		customerPeeringProperties: customerProperties,
	}
}

func (s *AwsPeeringService) Create(indicator *util.LoadingIndicator) error {
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

	indicator.SetStep("Creating vpc peering connection...", 30)
	peeringConnectionId, peeringErr := s.createPeeringConnection()
	if peeringErr != nil {
		return peeringErr
	}
	indicator.SetStep("Creating routes...", 40)
	createRouteErr := s.createRoute(peeringConnectionId)
	if createRouteErr != nil {
		return createRouteErr
	}
	indicator.SetStep("Verifying vpc peering connection...", 50)
	subnets, subnetsErr := s.getSubnets()
	if subnetsErr != nil {
		return subnetsErr
	}
	vpcCidr, vpcCidrErr := s.getVpcCidr()
	if vpcCidrErr != nil {
		return vpcCidrErr
	}
	_, _, acceptErr := s.client.AwsPeering.Accept(context.Background(), &models.AcceptAwsPeeringInput{
		VpcId:               s.customerPeeringProperties.VpcId,
		VpcCidr:             vpcCidr,
		PeeringConnectionId: peeringConnectionId,
		Subnets:             subnets,
	})
	if acceptErr != nil {
		return acceptErr
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
	sess, sessionErr := session.NewSession(&aws.Config{
		Region: aws.String(s.customerPeeringProperties.Region)},
	)
	if sessionErr != nil {
		return sessionErr
	}
	s.ec2 = ec2.New(sess)
	return nil
}

func (s *AwsPeeringService) createPeeringConnection() (string, error) {
	peering, peeringErr := s.ec2.CreateVpcPeeringConnection(&ec2.CreateVpcPeeringConnectionInput{
		PeerOwnerId: aws.String(s.hazelcastPeeringProperties.OwnerId),
		PeerRegion:  aws.String(s.hazelcastPeeringProperties.Region),
		PeerVpcId:   aws.String(s.hazelcastPeeringProperties.VpcId),
		VpcId:       aws.String(s.customerPeeringProperties.VpcId),
	})
	if peeringErr != nil {
		return "", peeringErr
	}
	return aws.StringValue(peering.VpcPeeringConnection.VpcPeeringConnectionId), nil
}

func (s *AwsPeeringService) createRoute(peeringConnectionId string) error {
	var routeTableIds []string
	var routeTables *ec2.DescribeRouteTablesOutput
	for _, subnet := range s.customerPeeringProperties.SubnetIds {
		tables, routeTablesErr := s.ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("association.subnet-id"),
					Values: aws.StringSlice([]string{subnet}),
				},
			},
		})
		routeTables = tables

		if routeTablesErr != nil {
			return routeTablesErr
		}

		if len(tables.RouteTables) == 0 {
			defaultTables, tablesErr := s.ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
				Filters: []*ec2.Filter{
					{
						Name:   aws.String("association.main"),
						Values: aws.StringSlice([]string{"true"}),
					},
					{
						Name:   aws.String("vpc-id"),
						Values: aws.StringSlice([]string{s.customerPeeringProperties.VpcId}),
					},
				},
			})
			if tablesErr != nil {
				return tablesErr
			}
			routeTables = defaultTables
		}
		routeTableIds = append(routeTableIds, aws.StringValue(routeTables.RouteTables[0].RouteTableId))
	}

	for _, routeTableId := range routeTableIds {
		_, deleteRouteErr := s.ec2.DeleteRoute(&ec2.DeleteRouteInput{
			DestinationCidrBlock: aws.String(s.hazelcastPeeringProperties.VpcCidr),
			RouteTableId:         aws.String(routeTableId),
		})

		if deleteRouteErr != nil {
			return deleteRouteErr
		}

		_, createRouteErr := s.ec2.CreateRoute(&ec2.CreateRouteInput{
			DestinationCidrBlock:   aws.String(s.hazelcastPeeringProperties.VpcCidr),
			RouteTableId:           aws.String(routeTableId),
			VpcPeeringConnectionId: aws.String(peeringConnectionId),
		})

		if createRouteErr != nil {
			return createRouteErr
		}
	}

	return nil
}

func (s *AwsPeeringService) getVpcCidr() (string, error) {
	vpcs, vpcsErr := s.ec2.DescribeVpcs(&ec2.DescribeVpcsInput{VpcIds: aws.StringSlice([]string{s.customerPeeringProperties.VpcId})})
	if vpcsErr != nil {
		return "", vpcsErr
	}
	if len(vpcs.Vpcs) == 0 {
		return "", fmt.Errorf("vpc id %s not found", s.customerPeeringProperties.VpcId)
	}
	return aws.StringValue(vpcs.Vpcs[0].CidrBlock), nil
}

func (s *AwsPeeringService) getSubnets() ([]models.AcceptAwsVpcPeeringInputSubnets, error) {
	subnets, subnetsErr := s.ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice(s.customerPeeringProperties.SubnetIds),
	})
	if subnetsErr != nil {
		return nil, subnetsErr
	}
	var subnetInfos []models.AcceptAwsVpcPeeringInputSubnets
	for _, sub := range subnets.Subnets {
		subnetInfos = append(subnetInfos, models.AcceptAwsVpcPeeringInputSubnets{
			SubnetId:   aws.StringValue(sub.SubnetId),
			SubnetCidr: aws.StringValue(sub.CidrBlock),
		})
	}
	return subnetInfos, nil
}
