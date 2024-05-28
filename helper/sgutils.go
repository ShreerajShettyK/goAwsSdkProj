package helper

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

// CreateSecurityGroup creates a security group or returns the default security group if specified.
func CreateSecurityGroup(client *ec2.Client, subnetID string, useDefault bool) (string, error) {
	// Retrieve VPC ID from the subnet
	subnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}
	subnetResult, err := client.DescribeSubnets(context.Background(), subnetInput)
	if err != nil {
		return "", fmt.Errorf("failed to describe subnet: %v", err)
	}
	vpcID := aws.ToString(subnetResult.Subnets[0].VpcId)

	if useDefault {
		// Retrieve the default security group
		vpcInput := &ec2.DescribeSecurityGroupsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
				{
					Name:   aws.String("group-name"),
					Values: []string{"default"},
				},
			},
		}
		vpcResult, err := client.DescribeSecurityGroups(context.Background(), vpcInput)
		if err != nil {
			return "", fmt.Errorf("failed to describe security groups: %v", err)
		}
		if len(vpcResult.SecurityGroups) == 0 {
			return "", fmt.Errorf("no default security group found in VPC %s", vpcID)
		}
		return *vpcResult.SecurityGroups[0].GroupId, nil
	}

	// Create a new security group
	rand.Seed(time.Now().UnixNano())
	securityGroupName := "SSH-Access-" + randString(6)

	sgInput := &ec2.CreateSecurityGroupInput{
		Description: aws.String("Security group for SSH access"),
		GroupName:   aws.String(securityGroupName),
		VpcId:       aws.String(vpcID),
	}

	sgResult, err := client.CreateSecurityGroup(context.Background(), sgInput)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidGroup.Duplicate") {
			return "", fmt.Errorf("security group with name %s already exists", securityGroupName)
		}
		return "", fmt.Errorf("failed to create security group: %v", err)
	}

	// Authorize security group ingress for SSH
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
