package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	region            = "us-east-1"
	instanceType      = types.InstanceTypeT2Micro
	amiID             = "ami-0e001c9271cf7f3b9" // Example AMI ID for Ubuntu 22.04
	securityGroupName = "auto-generated-security-group8"
	subnetID          = "subnet-08854212983b84d1e" // Replace with your subnet ID
	iamRoleName       = "SSMManagedInstanceRole"   // Ensure this matches the IAM role name you created
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	securityGroupID, err := createSecurityGroup(ec2Client, subnetID)
	if err != nil {
		log.Fatalf("unable to create security group: %v", err)
	}
	fmt.Printf("Created security group %s\n", securityGroupID)

	instanceID, publicDNS, err := createEC2Instance(ec2Client, securityGroupID)
	if err != nil {
		log.Fatalf("unable to create instance: %v", err)
	}
	fmt.Printf("Created instance %s with public DNS %s\n", instanceID, publicDNS)

	ssmClient := ssm.NewFromConfig(cfg)
	commands := []string{
		"sudo apt update",
		"sudo apt install -y apt-transport-https ca-certificates curl software-properties-common",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
		"sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
		"sudo apt update",
		"sudo apt install -y docker-ce",
		"sudo systemctl start docker",
		"sudo systemctl enable docker",
		"sudo usermod -aG docker ubuntu",
		"sudo curl -L \"https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)\" -o /usr/local/bin/docker-compose",
		"sudo chmod +x /usr/local/bin/docker-compose",
		"docker --version",
		"docker-compose --version",
	}

	commandInput := &ssm.SendCommandInput{
		InstanceIds:  []string{instanceID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": commands,
		},
	}

	output, err := ssmClient.SendCommand(context.TODO(), commandInput)
	if err != nil {
		log.Fatalf("failed to send SSM command, %v", err)
	}

	fmt.Println("Successfully sent SSM command to install Docker and SSM Agent")
	fmt.Printf("SSM Command ID: %s\n", *output.Command.CommandId)

	// Wait for the command to complete
	time.Sleep(30 * time.Second)
	describeCommandOutput, err := ssmClient.GetCommandInvocation(context.TODO(), &ssm.GetCommandInvocationInput{
		CommandId:  output.Command.CommandId,
		InstanceId: aws.String(instanceID),
	})
	if err != nil {
		log.Fatalf("failed to describe command invocation, %v", err)
	}
	fmt.Printf("Command Status: %s\n", describeCommandOutput.Status)
	fmt.Printf("Command Output: %s\n", describeCommandOutput.StandardOutputContent)
}

func createSecurityGroup(client *ec2.Client, subnetID string) (string, error) {
	// Retrieve VPC ID from the subnet
	subnetInput := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}
	subnetResult, err := client.DescribeSubnets(context.TODO(), subnetInput)
	if err != nil {
		return "", fmt.Errorf("failed to describe subnet: %v", err)
	}
	vpcID := *subnetResult.Subnets[0].VpcId

	sgInput := &ec2.CreateSecurityGroupInput{
		Description: aws.String("Security group for SSH access"),
		GroupName:   aws.String(securityGroupName),
		VpcId:       aws.String(vpcID),
	}
	sgResult, err := client.CreateSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return "", fmt.Errorf("failed to create security group: %v", err)
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
	_, err = client.AuthorizeSecurityGroupIngress(context.TODO(), authInput)
	if err != nil {
		return "", fmt.Errorf("failed to authorize security group ingress: %v", err)
	}

	return *sgResult.GroupId, nil
}

func createEC2Instance(client *ec2.Client, securityGroupID string) (string, string, error) {
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
		InstanceType:     instanceType,
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

	fmt.Println("Waiting for instance to be in running state...")
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
	fmt.Println("Instance is now running")

	fmt.Println("Waiting for instance status checks to complete...")
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
	fmt.Println("Instance has passed status checks")

	describeInstancesResult, err := client.DescribeInstances(context.TODO(), describeInstancesInput)
	if err != nil {
		return "", "", fmt.Errorf("failed to describe instances: %v", err)
	}
	publicDNS := *describeInstancesResult.Reservations[0].Instances[0].PublicDnsName

	return instanceID, publicDNS, nil
}
