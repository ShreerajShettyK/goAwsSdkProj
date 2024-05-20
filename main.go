package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}

	// Create an EC2 service client
	svc := ec2.New(sess)

	userData := `#!/bin/bash
	sudo apt update
	sudo apt upgrade
	sudo apt install openjdk-11-jdk
	sudo wget -O /usr/share/keyrings/jenkins-keyring.asc \
		https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key
	echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]" \
		https://pkg.jenkins.io/debian-stable binary/ | sudo tee \
		/etc/apt/sources.list.d/jenkins.list > /dev/null
	sudo apt-get update
	sudo apt-get install fontconfig openjdk-11-jre
	sudo apt-get install jenkins
	sudo systemctl status jenkins`

	// Encode user data to base64
	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))

	// Create a security group
	createSGOutput, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String("MySecurityGroup"),
		Description: aws.String("Security group for SSH and HTTP access"),
		VpcId:       aws.String("vpc-07e3c6cc53a279937"), // replace with your VPC ID
	})
	if err != nil {
		log.Fatalf("Failed to create security group: %s", err)
	}
	sgID := createSGOutput.GroupId

	_, err = svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: sgID,
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(22),
				ToPort:     aws.Int64(22),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(80),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(8080),
				ToPort:     aws.Int64(8080),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to set security group ingress rules: %s", err)
	}

	// Specify the instance details
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String("ami-04b70fa74e45c3917"), // replace with your Ubuntu AMI ID
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		UserData:     aws.String(userDataEncoded),
		// SecurityGroupIds: []*string{
		// 	sgID,
		// },
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{
				AssociatePublicIpAddress: aws.Bool(true),
				DeviceIndex:              aws.Int64(0),
				SubnetId:                 aws.String("subnet-00537b3921f471b4b"),
				Groups: []*string{
					sgID,
				},
			},
		},
		KeyName: aws.String("goTask"), // replace with your key pair name
	})
	if err != nil {
		log.Fatalf("Could not create instance: %s", err)
	}

	// Get the instance ID of the created instance
	instanceID := runResult.Instances[0].InstanceId
	fmt.Printf("Successfully created instance %s\n", *instanceID)

	// Tag the instance
	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{instanceID},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("Task3Instance"),
			},
		},
	})
	if err != nil {
		log.Fatalf("Could not create tags for instance %s, %s\n", *instanceID, err)
	}

	fmt.Printf("Successfully tagged instance %s\n", *instanceID)

	// Wait until the instance is running
	fmt.Println("Waiting for instance to be in running state...")
	err = svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{instanceID},
	})
	if err != nil {
		log.Fatalf("Error waiting for instance to be in running state: %s", err)
	}

	fmt.Printf("Instance %s is running\n", *instanceID)
}
