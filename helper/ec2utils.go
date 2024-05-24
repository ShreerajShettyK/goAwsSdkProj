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

func CreateEC2Instance(client *ec2.Client, securityGroupID, instanceType, amiID, iamRoleName string) (string, string, error) {
	userData := `#!/bin/bash
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

	instanceInput := &ec2.RunInstancesInput{
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

	runResult, err := client.RunInstances(context.TODO(), instanceInput)
	if err != nil {
		return "", "", fmt.Errorf("failed to run instances: %v", err)
	}

	instanceID := *runResult.Instances[0].InstanceId

	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	log.Printf("Waiting for instance to be in running state...")
	startTime := time.Now()
	for {
		describeInstancesResult, err := client.DescribeInstances(context.TODO(), describeInstancesInput)
		if err != nil {
			return "", "", fmt.Errorf("failed to describe instances: %v", err)
		}
		instanceState := describeInstancesResult.Reservations[0].Instances[0].State.Name
		if instanceState == types.InstanceStateNameRunning {
			break
		}
		if time.Since(startTime) > 5*time.Minute {
			return "", "", fmt.Errorf("instance did not reach running state in time")
		}
		time.Sleep(5 * time.Second)
	}
	log.Printf("Instance is now running")

	log.Printf("Waiting for instance status checks to complete...")
	for {
		describeInstanceStatusInput := &ec2.DescribeInstanceStatusInput{
			InstanceIds: []string{instanceID},
		}

		describeInstanceStatusResult, err := client.DescribeInstanceStatus(context.TODO(), describeInstanceStatusInput)
		if err != nil {
			return "", "", fmt.Errorf("failed to describe instance status: %v", err)
		}

		if len(describeInstanceStatusResult.InstanceStatuses) > 0 {
			instanceStatus := describeInstanceStatusResult.InstanceStatuses[0]
			if instanceStatus.InstanceStatus.Status == "ok" &&
				instanceStatus.SystemStatus.Status == "ok" {
				break
			}
		}

		if time.Since(startTime) > 10*time.Minute {
			return "", "", fmt.Errorf("instance did not pass status checks in time")
		}
		time.Sleep(10 * time.Second)
	}
	log.Printf("Instance has passed status checks")

	describeInstancesResult, err := client.DescribeInstances(context.TODO(), describeInstancesInput)
	if err != nil {
		return "", "", fmt.Errorf("failed to describe instances: %v", err)
	}
	publicDNS := *describeInstancesResult.Reservations[0].Instances[0].PublicDnsName

	return instanceID, publicDNS, nil
}
