package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/joho/godotenv"
	"main.go/helper"
)

// Constants initialized from environment variables
var (
	Region        string
	InstanceType  string
	AmiID         string
	SubnetID      string
	IAMRoleName   string
	PolicyEC2Role string
	PolicySSMCore string
)

// init initializes the constants from environment variables
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	Region = os.Getenv("region")
	InstanceType = os.Getenv("instanceType")
	AmiID = os.Getenv("amiID")
	SubnetID = os.Getenv("subnetID")
	IAMRoleName = os.Getenv("iamRoleName")
	PolicyEC2Role = os.Getenv("policyEC2Role")
	PolicySSMCore = os.Getenv("policySSMCore")
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(Region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)

	// Check if IAM role exists and create if it does not
	instanceProfileName, err := helper.EnsureIAMRole(iamClient, IAMRoleName, PolicyEC2Role, PolicySSMCore)
	if err != nil {
		log.Fatalf("unable to ensure IAM role: %v", err)
	}
	log.Printf("Successfully created or ensured IAM role %s\n", IAMRoleName)

	securityGroupID, err := helper.CreateSecurityGroup(ec2Client, SubnetID)
	if err != nil {
		log.Fatalf("unable to create security group: %v", err)
	}
	log.Printf("Created security group %s\n", securityGroupID)

	instanceID, publicDNS, err := helper.CreateEC2Instance(ec2Client, securityGroupID, InstanceType, AmiID, instanceProfileName)
	if err != nil {
		log.Fatalf("unable to create instance: %v", err)
	}
	log.Printf("Created instance %s with public DNS %s\n", instanceID, publicDNS)

	err = helper.ExecuteSSMCommands(cfg, instanceID)
	if err != nil {
		log.Fatalf("failed to execute SSM commands: %v", err)
	}
}
