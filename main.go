package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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
	SecretName    string
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
	log.Printf("Successfully received .env variables %s\n", IAMRoleName)
	SecretName = "dev/GoAwsSdkProj/AwsAccessKeys" // Set the name of your secret here
}

func main() {
	// Create a context
	ctx := context.Background()

	// Retrieve the secret value from Secrets Manager
	secretValue, err := retrieveSecret(ctx, SecretName)
	if err != nil {
		log.Fatalf("unable to retrieve secret value: %v", err)
	}

	// Parse the secret value
	var secretMap map[string]string
	err = json.Unmarshal([]byte(secretValue), &secretMap)
	if err != nil {
		log.Fatalf("failed to unmarshal secret value: %v", err)
	}

	// Extract the AWS credentials from the secret
	accessKeyID := secretMap["AccesskeyID"]
	secretAccessKey := secretMap["SecretAccessKey"]

	log.Printf("Using AccessKeyID: %s", accessKeyID)

	// Configure AWS SDK with the retrieved credentials
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""))
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config with retrieved credentials: %v", err)
	}

	// Create EC2 and IAM clients using the updated configuration
	ec2Client := ec2.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)

	// Ensure IAM role
	roleName, err := helper.EnsureIAMRole(iamClient, IAMRoleName)
	if err != nil {
		log.Fatalf("unable to ensure IAM role: %v", err)
	}
	log.Printf("Successfully created or ensured IAM role %s\n", roleName)
	log.Println("Waiting for IAM role to be available...")
	time.Sleep(10 * time.Second)

	// Create security group
	securityGroupID, err := helper.CreateSecurityGroup(ec2Client, SubnetID, true)
	if err != nil {
		log.Fatalf("unable to create security group: %v", err)
	}
	log.Printf("Security group: %s\n", securityGroupID)

	// Create EC2 instance
	instanceID, publicDNS, err := helper.CreateEC2Instance(ec2Client, securityGroupID, InstanceType, AmiID, roleName)
	if err != nil {
		log.Fatalf("unable to create instance: %v", err)
	}
	log.Printf("Created instance %s with public DNS %s\n", instanceID, publicDNS)

	// Execute SSM commands
	err = helper.ExecuteSSMCommands(cfg, instanceID)
	if err != nil {
		log.Fatalf("failed to execute SSM commands: %v", err)
	}
}

// retrieveSecret retrieves the secret value from Secrets Manager
func retrieveSecret(ctx context.Context, secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	// Retrieve the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	return *result.SecretString, nil
}
