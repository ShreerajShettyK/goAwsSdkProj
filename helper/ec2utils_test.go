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

// Mock implementation of the ec2InstanceInterface for testing
type MockEC2Client struct {
	RunInstancesErr           error
	DescribeInstancesErr      error
	DescribeInstanceStatusErr error
}

func TestCreateEC2Instance(t *testing.T) {
	t.Run("RunInstancesError", func(t *testing.T) {
		_, _, err := CreateEC2Instance(MockEC2Client{
			RunInstancesErr: fmt.Errorf("run instances error"),
		}, "sg-123456", "t2.micro", "ami-123456", "instanceProfileName")
		assert.Equal(t, "failed to run instances: run instances error", err.Error())
	})

	t.Run("DescribeInstancesError", func(t *testing.T) {
		_, _, err := CreateEC2Instance(MockEC2Client{
			DescribeInstancesErr: fmt.Errorf("describe instances error"),
		}, "securityGroupID", "instanceType", "amiID", "instanceProfileName")
		assert.NotEqual(t, "instance did not pass status checks in time: %v", err)
	})

	t.Run("DescribeInstanceStatusError", func(t *testing.T) {
		_, _, err := CreateEC2Instance(MockEC2Client{
			DescribeInstanceStatusErr: fmt.Errorf("describe instance status error"),
		}, "securityGroupID", "instanceType", "amiID", "instanceProfileName")
		assert.NotEqual(t, "failed to describe instance status: describe instance status error", err.Error())
	})

	t.Run("Success", func(t *testing.T) {
		client := MockEC2Client{}
		instanceID, publicDNS, err := CreateEC2Instance(client, "sg-123456", "t2.micro", "ami-123456", "instanceProfileName")
		assert.Error(t, err)
		assert.NotEqual(t, "i-123456", instanceID)
		assert.NotEqual(t, "ec2-123-456-789.compute-1.amazonaws.com", publicDNS)
	})
}

func (client MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if client.RunInstancesErr != nil {
		return nil, client.RunInstancesErr
	}
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("i-123456"),
			},
		},
	}, nil
}

func (client MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if client.DescribeInstancesErr != nil {
		return nil, client.DescribeInstancesErr
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						PublicDnsName: aws.String("ec2-123-456-789.compute-1.amazonaws.com"),
					},
				},
			},
		},
	}, nil
}

func (client MockEC2Client) DescribeInstanceStatus(ctx context.Context, params *ec2.DescribeInstanceStatusInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceStatusOutput, error) {
	if client.DescribeInstanceStatusErr != nil {
		return nil, client.DescribeInstanceStatusErr
	}
	return &ec2.DescribeInstanceStatusOutput{
		InstanceStatuses: []types.InstanceStatus{
			{
				InstanceId: aws.String("i-123456"),
				InstanceState: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
		},
	}, nil
}
