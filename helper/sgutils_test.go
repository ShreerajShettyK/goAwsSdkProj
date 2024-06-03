package helper

import (
	"context"
	"fmt"

	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

// Mock implementation of securitygroupInterface for testing
type MockSecurityGroupClient struct {
	DescribeSubnetsErr               error
	DescribeSecurityGroupsErr        error
	CreateSecurityGroupErr           error
	AuthorizeSecurityGroupIngressErr error
}

func TestCreateSecurityGroup(t *testing.T) {
	t.Run("DescribeSubnetsError", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr: fmt.Errorf("describe subnets error"),
		}
		_, err := CreateSecurityGroup(&client, "subnet-123456", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe subnet")
	})

	t.Run("UseDefaultNoDefaultGroup", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr:        nil,
			DescribeSecurityGroupsErr: fmt.Errorf("describe security groups error"),
		}
		_, err := CreateSecurityGroup(&client, "subnet-123456", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe security groups")
	})

	t.Run("UseDefaultSuccess", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr:        nil,
			DescribeSecurityGroupsErr: nil,
		}
		groupID, err := CreateSecurityGroup(&client, "subnet-123456", true)
		assert.NoError(t, err)
		assert.NotNil(t, groupID)
	})

	t.Run("CreateSecurityGroupError", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr:     nil,
			CreateSecurityGroupErr: fmt.Errorf("create security group error"),
		}
		_, err := CreateSecurityGroup(&client, "subnet-123456", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create security group")
	})

	t.Run("CreateSecurityGroupDuplicateName", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr:     nil,
			CreateSecurityGroupErr: fmt.Errorf("InvalidGroup.Duplicate: duplicate group name"),
		}
		_, err := CreateSecurityGroup(&client, "subnet-123456", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security group with name")
	})

	t.Run("AuthorizeSecurityGroupIngressError", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr:               nil,
			AuthorizeSecurityGroupIngressErr: fmt.Errorf("authorize ingress error"),
		}
		_, err := CreateSecurityGroup(&client, "subnet-123456", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authorize security group ingress")
	})

	t.Run("Success", func(t *testing.T) {
		client := MockSecurityGroupClient{
			DescribeSubnetsErr: nil,
		}
		groupID, err := CreateSecurityGroup(&client, "subnet-123456", false)
		assert.NoError(t, err)
		assert.NotNil(t, groupID)
	})
}

// Implementing the securitygroupInterface for MockSecurityGroupClient

func (client *MockSecurityGroupClient) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	if client.DescribeSubnetsErr != nil {
		return nil, client.DescribeSubnetsErr
	}
	return &ec2.DescribeSubnetsOutput{
		Subnets: []types.Subnet{
			{
				VpcId: aws.String("vpc-123456"),
			},
		},
	}, nil
}

func (client *MockSecurityGroupClient) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	if client.DescribeSecurityGroupsErr != nil {
		return nil, client.DescribeSecurityGroupsErr
	}

	var vpcID *string
	for _, filter := range params.Filters {
		if aws.ToString(filter.Name) == "vpc-id" {
			vpcID = aws.String(filter.Values[0])
			break
		}
	}

	if vpcID != nil && *vpcID == "vpc-123456" {
		return &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId: aws.String("sg-123456"),
				},
			},
		}, nil
	}
	return &ec2.DescribeSecurityGroupsOutput{}, nil
}

func (client *MockSecurityGroupClient) CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	if client.CreateSecurityGroupErr != nil {
		return nil, client.CreateSecurityGroupErr
	}
	return &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("sg-123456"),
	}, nil
}

func (client *MockSecurityGroupClient) AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	if client.AuthorizeSecurityGroupIngressErr != nil {
		return nil, client.AuthorizeSecurityGroupIngressErr
	}
	return &ec2.AuthorizeSecurityGroupIngressOutput{}, nil
}
