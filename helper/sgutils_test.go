package helper

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func deleteSecurityGroup(client *ec2.Client, groupID string) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(groupID),
	}
	_, err := client.DeleteSecurityGroup(context.Background(), input)
	return err
}

func TestCreateSecurityGroup(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		t.Fatalf("unable to load SDK config, %v", err)
	}

	client := ec2.NewFromConfig(cfg)

	tests := []struct {
		name     string
		client   *ec2.Client
		subnetID string
		wantErr  bool
	}{
		{
			name:     "Valid subnet ID",
			client:   client,
			subnetID: "subnet-08854212983b84d1e", // Replace with a valid subnet ID
			wantErr:  false,
		},
		{
			name:     "Invalid subnet ID",
			client:   client,
			subnetID: "subnet-invalid",
			wantErr:  true,
		},
		{
			name:     "Error Describing Subnet",
			client:   client,
			subnetID: "subnet-invalid",
			wantErr:  true,
		},
		{
			name:     "Error Creating Security Group",
			client:   client,
			subnetID: "subnet-invalid", // Using an invalid subnet ID to trigger an error
			wantErr:  true,
		},
		{
			name:     "Error Authorizing Security Group Ingress",
			client:   client,
			subnetID: "subnet-08854212983b84d1e", // Replace with a valid subnet ID
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Error Describing Subnet" {
				tt.subnetID = "subnet-invalid"
			}
			got, err := CreateSecurityGroup(tt.client, tt.subnetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && got == "" {
				t.Errorf("Expected non-empty security group ID, got %v", got)
			} else if !tt.wantErr {
				defer deleteSecurityGroup(tt.client, got)
			}
		})
	}
}
