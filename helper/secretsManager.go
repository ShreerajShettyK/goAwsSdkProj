package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var cfg aws.Config

func init() {
	// Load AWS SDK configuration
	var err error
	cfg, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Error loading AWS SDK config: %v", err)
	}
}

// Fetches the value of a secret from AWS Secrets Manager
func getSecret(secretName string) (string, error) {
	var secretValue string
	client := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := client.GetSecretValue(context.Background(), input)
	if err != nil {
		return secretValue, fmt.Errorf("failed to get secret value: %v", err)
	}

	if result.SecretString == nil {
		return secretValue, fmt.Errorf("secret string is nil")
	}
	secretValue = aws.ToString(result.SecretString)

	return secretValue, nil
}

// Fetches required values from Secrets Manager
func FetchSecrets() (string, string, string, string, string, string, error) {
	var secretData map[string]string
	var amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region string

	secretValue, err := getSecret("task1/InfraProvision")
	if err != nil {
		return "", "", "", "", "", "", err
	}

	// Parsing the JSON format of the secret string
	err = json.Unmarshal([]byte(secretValue), &secretData)
	if err != nil {
		log.Printf("Error parsing secret string: %v\n", err)
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, err
	}

	// Extracting values from the parsed secret data
	amiID, ok := secretData["amiID"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("amiID not found in secret data")
	}

	subnetID, ok = secretData["subnetID"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("subnetID not found in secret data")
	}

	iamRoleName, ok = secretData["iamRoleName"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("iamRoleName not found in secret data")
	}

	instanceType, ok = secretData["instanceType"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("instanceType not found in secret data")
	}

	mongoDbConnectionString, ok = secretData["mongoDbConnectionString"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("mongoDbConnectionString not found in secret data")
	}

	region, ok = secretData["region"]
	if !ok {
		return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, fmt.Errorf("region not found in secret data")
	}

	return amiID, subnetID, iamRoleName, instanceType, mongoDbConnectionString, region, nil
}
