package main

import (
	"context"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"golang.org/x/crypto/ssh"
)

const (
	region            = "us-east-1"
	instanceType      = types.InstanceTypeT2Micro
	amiID             = "ami-04b70fa74e45c3917" // Example AMI ID for Ubuntu 20.04
	keyPairName       = "my-new-key-pair"
	securityGroupName = "auto-generated-security-group"
	subnetID          = "subnet-093b5b9b6b1da8e29" // Replace with your subnet ID
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	keyPair, err := createKeyPair(ec2Client)
	if err != nil {
		log.Fatalf("unable to create key pair: %v", err)
	}
	fmt.Printf("Created key pair %s\n", *keyPair.KeyName)

	securityGroupID, err := createSecurityGroup(ec2Client, subnetID)
	if err != nil {
		log.Fatalf("unable to create security group: %v", err)
	}
	fmt.Printf("Created security group %s\n", securityGroupID)

	instanceID, publicDNS, err := createEC2Instance(ec2Client, keyPairName, securityGroupID)
	if err != nil {
		log.Fatalf("unable to create instance: %v", err)
	}
	fmt.Printf("Created instance %s with public DNS %s\n", instanceID, publicDNS)

	err = runShellCommands(publicDNS, keyPair.KeyMaterial)
	if err != nil {
		log.Fatalf("unable to run shell commands: %v", err)
	}
	fmt.Println("Shell script executed successfully")
}

func createKeyPair(client *ec2.Client) (*ec2.CreateKeyPairOutput, error) {
	input := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyPairName),
	}
	result, err := client.CreateKeyPair(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create key pair: %v", err)
	}
	return result, nil
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

func createEC2Instance(client *ec2.Client, keyName, securityGroupID string) (string, string, error) {
	instanceInput := &ec2.RunInstancesInput{
		ImageId:          aws.String(amiID),
		InstanceType:     instanceType,
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		KeyName:          aws.String(keyName),
		SecurityGroupIds: []string{securityGroupID},
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

	describeInstancesResult, err := client.DescribeInstances(context.TODO(), describeInstancesInput)
	if err != nil {
		return "", "", fmt.Errorf("failed to describe instances: %v", err)
	}
	publicDNS := *describeInstancesResult.Reservations[0].Instances[0].PublicDnsName

	return instanceID, publicDNS, nil
}

func runShellCommands(publicDNS string, keyMaterial *string) error {
	block, _ := pem.Decode([]byte(*keyMaterial))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("failed to decode PEM block containing private key")
	}
	privateKey, err := ssh.ParsePrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: "ubuntu", // Replace with appropriate SSH username for your AMI
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	fmt.Println("Waiting for SSH to be available...")
	startTime := time.Now()
	for {
		conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", publicDNS), sshConfig)
		if err == nil {
			conn.Close()
			break
		}
		if time.Since(startTime) > 2*time.Minute {
			return fmt.Errorf("SSH connection timeout")
		}
		time.Sleep(5 * time.Second)
	}
	fmt.Println("SSH is now available")

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", publicDNS), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	err = session.Run("sudo apt-get update -y && sudo apt-get upgrade -y")
	if err != nil {
		return fmt.Errorf("failed to run: %v", err)
	}

	return nil
}
