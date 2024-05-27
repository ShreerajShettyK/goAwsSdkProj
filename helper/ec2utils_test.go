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

func TestEc2Instance(t *testing.T) {
	_, _, err := CreateEC2Instance(MockEc2Creation{
		RunInstancesError: fmt.Errorf("Couldnt create ec2 instance"),
	}, "securityGroupId", "instanceType", "amiId", "iamRoleName")
	assert.Equal(t, err.Error(), "failed to run instances: Couldnt create ec2 instance")
}

type MockEc2Creation struct {
	RunInstancesError           error
	DescribeInstancesError      error
	DescribeInstanceStatusError error
}

func (m MockEc2Creation) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesError != nil {
		return nil, m.RunInstancesError
	}
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("ec2InstanceId"),
			},
		},
	}, nil
}
func (m MockEc2Creation) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return nil, nil
}
func (m MockEc2Creation) DescribeInstanceStatus(ctx context.Context, params *ec2.DescribeInstanceStatusInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceStatusOutput, error) {
	return nil, nil
}
