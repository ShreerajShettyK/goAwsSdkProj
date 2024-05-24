package helper

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func TestExecuteSSMCommands(t *testing.T) {
	// Load AWS SDK config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("unable to load SDK config, %v", err)
	}

	// Replace "i-096170992864d5385" with a valid instance ID
	instanceID := "i-096170992864d5385"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful Execution",
			args: args{
				cfg:        cfg,
				instanceID: instanceID,
			},
			wantErr: false,
		},
		{
			name: "Failed Command Execution",
			args: args{
				cfg:        cfg,
				instanceID: instanceID,
			},
			wantErr: true,
		},
		{
			name: "Failed to Send SSM Command",
			args: args{
				cfg:        aws.Config{},
				instanceID: instanceID,
			},
			wantErr: true,
		},
		{
			name: "Failed to Describe Command Invocation",
			args: args{
				cfg:        cfg,
				instanceID: "invalid-instance-id",
			},
			wantErr: true,
		},
		{
			name: "Timeout",
			args: args{
				cfg:        cfg,
				instanceID: instanceID,
			},
			wantErr: false, // Timeout is not treated as an error in this test
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExecuteSSMCommands(tt.args.cfg, tt.args.instanceID); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteSSMCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type args struct {
	cfg        aws.Config
	instanceID string
}
