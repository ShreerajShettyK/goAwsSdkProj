package helper

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type ec2InstanceInterface interface {
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceStatus(ctx context.Context, params *ec2.DescribeInstanceStatusInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceStatusOutput, error)
}

func createUserDataScript() string {
	return `#!/bin/bash
        sudo apt update
        sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
        sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
        sudo apt update
        sudo apt install -y docker-ce
        sudo systemctl start docker
        sudo systemctl enable docker
        sudo usermod -aG docker ubuntu
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        docker --version
        docker-compose --version
        sudo snap install amazon-ssm-agent --classic
        sudo systemctl start snap.amazon-ssm-agent.amazon-ssm-agent
        sudo systemctl enable snap.amazon-ssm-agent.amazon-ssm-agent
    `
}

func createInstanceInput(securityGroupID, instanceType, amiID, iamRoleName, userData string) *ec2.RunInstancesInput {
	return &ec2.RunInstancesInput{
		ImageId:          aws.String(amiID),
		InstanceType:     types.InstanceType(instanceType),
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		SecurityGroupIds: []string{securityGroupID},
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Name: aws.String(iamRoleName),
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(userData))),
	}
}

func waitForInstanceRunning(client ec2InstanceInterface, instanceID string) (string, error) {
	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	log.Printf("Waiting for instance to be in running state...")
	waiter := ec2.NewInstanceRunningWaiter(client)
	if err := waiter.Wait(context.Background(), describeInstancesInput, 5*time.Minute); err != nil {
		return "", fmt.Errorf("instance did not reach running state in time: %v", err)
	}
	log.Printf("Instance is now running")

	describeInstancesResult, err := client.DescribeInstances(context.Background(), describeInstancesInput)
	if err != nil {
		return "", fmt.Errorf("failed to describe instances: %v", err)
	}
	publicDNS := aws.ToString(describeInstancesResult.Reservations[0].Instances[0].PublicDnsName)

	return publicDNS, nil
}

func waitForInstanceStatusChecks(client ec2InstanceInterface, instanceID string) error {
	describeInstanceStatusInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{instanceID},
	}
	log.Printf("Waiting for instance status checks to complete...")
	waiter := ec2.NewInstanceStatusOkWaiter(client)
	if err := waiter.Wait(context.Background(), describeInstanceStatusInput, 10*time.Minute); err != nil {
		return fmt.Errorf("instance did not pass status checks in time: %v", err)
	}
	log.Printf("Instance has passed status checks")
	return nil
}

func CreateEC2Instance(client ec2InstanceInterface, securityGroupID, instanceType, amiID, iamRoleName string) (string, string, error) {
	userData := createUserDataScript()
	instanceInput := createInstanceInput(securityGroupID, instanceType, amiID, iamRoleName, userData)

	runResult, err := client.RunInstances(context.Background(), instanceInput)
	if err != nil {
		return "", "", fmt.Errorf("failed to run instances: %v", err)
	}

	instanceID := aws.ToString(runResult.Instances[0].InstanceId)

	publicDNS, err := waitForInstanceRunning(client, instanceID)
	if err != nil {
		return "", "", err
	}

	if err := waitForInstanceStatusChecks(client, instanceID); err != nil {
		return "", "", err
	}

	return instanceID, publicDNS, nil
}
