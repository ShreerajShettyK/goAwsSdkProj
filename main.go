package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"main.go/helper"
)

const (
	region       = "us-east-1"
	instanceType = "t2.micro"
	amiID        = "ami-0e001c9271cf7f3b9"    // Example AMI ID for Ubuntu 22.04
	subnetID     = "subnet-08854212983b84d1e" // Replace with your subnet ID
	iamRoleName  = "SSMManagedInstanceRole"   // Ensure this matches the IAM role name you created
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	securityGroupID, err := helper.CreateSecurityGroup(ec2Client, subnetID)
	if err != nil {
		log.Fatalf("unable to create security group: %v", err)
	}
	log.Printf("Created security group %s\n", securityGroupID)

	instanceID, publicDNS, err := helper.CreateEC2Instance(ec2Client, securityGroupID, instanceType, amiID, iamRoleName)
	if err != nil {
		log.Fatalf("unable to create instance: %v", err)
	}
	log.Printf("Created instance %s with public DNS %s\n", instanceID, publicDNS)

	err = helper.ExecuteSSMCommands(cfg, instanceID)
	if err != nil {
		log.Fatalf("failed to execute SSM commands: %v", err)
	}
}
