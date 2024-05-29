package helper

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// EnsureIAMRole checks if the IAM role exists and creates it if it doesn't, attaching the necessary policies.
func EnsureIAMRole(client *iam.Client, roleName string) (string, error) {
	_, err := client.GetRole(context.Background(), &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err == nil {
		log.Printf("IAM role %s already exists\n", roleName)
		return roleName, nil
	} else {
		var notFound *iamTypes.NoSuchEntityException
		if !errors.As(err, &notFound) {
			return "", err
		}

		// Define the policy document allowing access to Session Manager
		policyDocument := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Action": [
						"ssm:GetDocument",
						"ssm:DescribeDocument",
						"ssm:GetManifest",
						"ssm:GetParameters",
						"ssm:UpdateAssociationStatus",
						"ssm:UpdateInstanceAssociationStatus",
						"ssm:UpdateInstanceInformation",
						"ssm:SendCommand",
						"ssm:StartSession",
						"ssm:TerminateSession"
					],
					"Resource": "*"
				},
				{
					"Effect": "Allow",
					"Action": [
						"ssmmessages:CreateControlChannel",
						"ssmmessages:CreateDataChannel",
						"ssmmessages:OpenControlChannel",
						"ssmmessages:OpenDataChannel"
					],
					"Resource": "*"
				},
				{
					"Effect": "Allow",
					"Action": [
						"ec2messages:AcknowledgeMessage",
						"ec2messages:DeleteMessage",
						"ec2messages:FailMessage",
						"ec2messages:GetEndpoint",
						"ec2messages:GetMessages",
						"ec2messages:SendReply"
					],
					"Resource": "*"
				},
				{
					"Effect": "Allow",
					"Action": [
						"ec2:DescribeInstanceStatus"
					],
					"Resource": "*"
				}
			]
}`

		// Create the IAM policy
		createPolicyOutput, err := client.CreatePolicy(context.Background(), &iam.CreatePolicyInput{
			PolicyDocument: aws.String(policyDocument),
			PolicyName:     aws.String("SSMSessionManagerPolicy"),
			Description:    aws.String("Allows access to Session Manager for EC2 instances"),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create IAM policy: %v", err)
		}
		log.Printf("Created IAM policy SSMSessionManagerPolicy")

		// Create the IAM role
		_, err = client.CreateRole(context.Background(), &iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(`{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "ec2.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create IAM role: %v", err)
		}
		log.Printf("Created IAM role %s\n", roleName)

		// Attach the IAM policy to the role
		_, err = client.AttachRolePolicy(context.Background(), &iam.AttachRolePolicyInput{
			PolicyArn: aws.String(*createPolicyOutput.Policy.Arn),
			RoleName:  aws.String(roleName),
		})
		if err != nil {
			return "", fmt.Errorf("failed to attach IAM policy to role: %v", err)
		}
		log.Printf("Attached IAM policy SSMSessionManagerPolicy to role %s\n", roleName)

		// Create an instance profile
		_, err = client.CreateInstanceProfile(context.Background(), &iam.CreateInstanceProfileInput{
			InstanceProfileName: aws.String(roleName),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create instance profile: %v", err)
		}
		log.Printf("Created instance profile %s\n", roleName)

		// Add the role to the instance profile
		_, err = client.AddRoleToInstanceProfile(context.Background(), &iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(roleName),
			RoleName:            aws.String(roleName),
		})
		if err != nil {
			return "", fmt.Errorf("failed to add role to instance profile: %v", err)
		}
		log.Printf("Added role %s to instance profile %s\n", roleName, roleName)

		// Return the instance profile name
		return roleName, nil
	}

}
