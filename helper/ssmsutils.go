package helper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Wait for the SSM command to reach a terminal state (from inprogress to Success)
func waitForSSMCommandCompletion(ssmClient *ssm.Client, commandID, instanceID string) error {
	waiter := ssm.NewCommandExecutedWaiter(ssmClient)
	describeCommandInput := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandID),
		InstanceId: aws.String(instanceID),
	}
	if err := waiter.Wait(context.Background(), describeCommandInput, 10*time.Minute); err != nil {
		return fmt.Errorf("SSM command did not complete in time: %v", err)
	}
	return nil
}

func ExecuteSSMCommands(cfg aws.Config, instanceID string) error {
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
		"sudo apt update",
		"sudo apt upgrade",
		"sudo apt install openjdk-17-jdk",
		"sudo wget -O /usr/share/keyrings/jenkins-keyring.asc https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key",
		"echo \"deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc] https://pkg.jenkins.io/debian-stable binary/\" | sudo tee /etc/apt/sources.list.d/jenkins.list > /dev/null",
		"sudo apt-get update",
		"sudo apt-get install -y fontconfig openjdk-11-jre",
		"sudo apt-get install -y jenkins",
		"sudo systemctl status jenkins",
	}

	commandInput := &ssm.SendCommandInput{
		InstanceIds:  []string{instanceID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": commands,
		},
	}

	output, err := ssmClient.SendCommand(context.Background(), commandInput)
	if err != nil {
		return fmt.Errorf("failed to send SSM command: %v", err)
	}

	log.Printf("Successfully sent SSM command to install SSM Agent, Docker and Jenkins")
	log.Printf("SSM Command ID: %s\n", *output.Command.CommandId)

	// Wait for the command to complete using waiter
	if err := waitForSSMCommandCompletion(ssmClient, aws.ToString(output.Command.CommandId), instanceID); err != nil {
		return err
	}

	describeCommandOutput, err := ssmClient.GetCommandInvocation(context.Background(), &ssm.GetCommandInvocationInput{
		CommandId:  output.Command.CommandId,
		InstanceId: aws.String(instanceID),
	})
	if err != nil {
		return fmt.Errorf("failed to describe command invocation: %v", err)
	}
	log.Printf("Command Status: %s\n", describeCommandOutput.Status)
	log.Printf("Command Output: %v\n", describeCommandOutput.StandardOutputContent)
	return nil
}
