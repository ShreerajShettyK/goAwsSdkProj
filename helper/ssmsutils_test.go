package helper

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWaitForSSMCommandCompletion(t *testing.T){
	_,_,err:= waitForSSMCommandCompletion(MockSSMClient{
		SendCommandError: fmt.Errorf("Wait for ssm completion"),
	},  "commandID", "instanceID","ssmInterface")
	assert.Equal(t,err.Error(),"SSM command did not complete in time: Wait for ssm completion")
}
               
type MockSSMClient struct {
	SendCommandErr        error
	GetCommandInvocationErr error
}
    
func (client MockSSMClient) SendCommand(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
	if client.SendCommandErr !=nil{
		return nil,client.SendCommandErr
	}
	return &ssm.SendCommandOutput{
		Command: []types.Command{

			{

			},
		},

		
	},nil
}             
    


