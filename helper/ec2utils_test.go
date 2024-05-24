package helper

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func TestCreateEC2Instance(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	type args struct {
		client          *ec2.Client
		securityGroupID string
		instanceType    string
		amiID           string
		iamRoleName     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "Successful Instance Creation",
			args: args{
				client:          ec2Client,
				securityGroupID: "sg-0b771161b3d1e1849", // Replace with a valid security group ID
				instanceType:    "t2.micro",
				amiID:           "ami-0e001c9271cf7f3b9", // Example AMI ID for Ubuntu 22.04
				iamRoleName:     "SSMManagedInstanceRole",
			},
			wantErr: false,
		},
		{
			name: "Invalid Security Group ID",
			args: args{
				client:          ec2Client,
				securityGroupID: "invalid-security-group-id",
				instanceType:    "t2.micro",
				amiID:           "ami-0e001c9271cf7f3b9",
				iamRoleName:     "SSMManagedInstanceRole",
			},
			wantErr: true,
		},
		{
			name: "Invalid Instance Type",
			args: args{
				client:          ec2Client,
				securityGroupID: "sg-0b771161b3d1e1849",
				instanceType:    "invalid-instance-type",
				amiID:           "ami-0e001c9271cf7f3b9",
				iamRoleName:     "SSMManagedInstanceRole",
			},
			wantErr: true,
		},
		{
			name: "Invalid AMI ID",
			args: args{
				client:          ec2Client,
				securityGroupID: "sg-0b771161b3d1e1849",
				instanceType:    "t2.micro",
				amiID:           "invalid-ami-id",
				iamRoleName:     "SSMManagedInstanceRole",
			},
			wantErr: true,
		},
		{
			name: "Invalid IAM Role Name",
			args: args{
				client:          ec2Client,
				securityGroupID: "sg-0b771161b3d1e1849",
				instanceType:    "t2.micro",
				amiID:           "ami-0e001c9271cf7f3b9",
				iamRoleName:     "invalid-iam-role-name",
			},
			wantErr: true,
		},
		{
			name: "Timeout during Instance Creation",
			args: args{
				client:          ec2Client,
				securityGroupID: "sg-0b771161b3d1e1849",
				instanceType:    "t2.micro",
				amiID:           "ami-0e001c9271cf7f3b9",
				iamRoleName:     "SSMManagedInstanceRole",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateEC2Instance(tt.args.client, tt.args.securityGroupID, tt.args.instanceType, tt.args.amiID, tt.args.iamRoleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEC2Instance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && (got == "" || got1 == "") {
				t.Errorf("Expected non-empty instance ID and public DNS, got %v and %v", got, got1)
			}
		})
	}
}
