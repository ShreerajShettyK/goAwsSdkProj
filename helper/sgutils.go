package helper

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
type sgInterface interface {
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
}

func CreateSecurityGroup(client sgInterface, subnetID string) (string, error) {
	// Retrieve VPC ID from the subnet
	subnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}
	subnetResult, err := client.DescribeSubnets(context.Background(), subnetInput)
	if err != nil {
		return "", fmt.Errorf("failed to describe subnet: %v", err)
	}
	vpcID := *subnetResult.Subnets[0].VpcId

	var sgResult *ec2.CreateSecurityGroupOutput
	maxRetries := 5
	for retries := 0; retries < maxRetries; retries++ {
		securityGroupName := randString(20)

		sgInput := &ec2.CreateSecurityGroupInput{
			Description: aws.String("Security group for SSH access"),
			GroupName:   aws.String(securityGroupName),
			VpcId:       aws.String(vpcID),
		}

		sgResult, err = client.CreateSecurityGroup(context.Background(), sgInput)
		if err != nil {
			if strings.Contains(err.Error(), "InvalidGroup.Duplicate") {
				continue
			}
			return "", fmt.Errorf("failed to create security group: %v", err)
		}
		break
	}

	if sgResult == nil {
		return "", fmt.Errorf("failed to create a unique security group after %d attempts", maxRetries)
	}

	authInput := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: sgResult.GroupId,
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	}
	_, err = client.AuthorizeSecurityGroupIngress(context.Background(), authInput)
	if err != nil {
		return "", fmt.Errorf("failed to authorize security group ingress: %v", err)
	}

	return *sgResult.GroupId, nil
}
