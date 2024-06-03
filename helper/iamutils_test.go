package helper

import (
	"context"

	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
)

type MockIAMClient struct {
	GetRoleErr                  error
	CreatePolicyErr             error
	CreateRoleErr               error
	AttachRolePolicyErr         error
	CreateInstanceProfileErr    error
	AddRoleToInstanceProfileErr error
}

func TestEnsureIAMRole(t *testing.T) {
	roleName := "test-role"

	t.Run("GetRoleError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr: fmt.Errorf("get role error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "get role error", err.Error())
	})

	t.Run("CreatePolicyError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr:      &types.NoSuchEntityException{},
			CreatePolicyErr: fmt.Errorf("create policy error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "failed to create IAM policy: create policy error", err.Error())
	})

	t.Run("CreateRoleError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr:    &types.NoSuchEntityException{},
			CreateRoleErr: fmt.Errorf("create role error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "failed to create IAM role: create role error", err.Error())
	})

	t.Run("AttachRolePolicyError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr:          &types.NoSuchEntityException{},
			AttachRolePolicyErr: fmt.Errorf("attach role policy error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "failed to attach IAM policy to role: attach role policy error", err.Error())
	})

	t.Run("CreateInstanceProfileError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr:               &types.NoSuchEntityException{},
			CreateInstanceProfileErr: fmt.Errorf("create instance profile error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "failed to create instance profile: create instance profile error", err.Error())
	})

	t.Run("AddRoleToInstanceProfileError", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr:                  &types.NoSuchEntityException{},
			AddRoleToInstanceProfileErr: fmt.Errorf("add role to instance profile error"),
		}
		_, err := EnsureIAMRole(client, roleName)
		assert.Error(t, err)
		assert.Equal(t, "failed to add role to instance profile: add role to instance profile error", err.Error())
	})

	t.Run("Success", func(t *testing.T) {
		client := MockIAMClient{
			GetRoleErr: &types.NoSuchEntityException{},
		}
		result, err := EnsureIAMRole(client, roleName)
		assert.NoError(t, err)
		assert.Equal(t, roleName, result)
	})
}

func (m MockIAMClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	if m.GetRoleErr != nil {
		return nil, m.GetRoleErr
	}
	return &iam.GetRoleOutput{
		Role: &types.Role{
			RoleName: params.RoleName,
		},
	}, nil
}

func (m MockIAMClient) CreatePolicy(ctx context.Context, params *iam.CreatePolicyInput, optFns ...func(*iam.Options)) (*iam.CreatePolicyOutput, error) {
	if m.CreatePolicyErr != nil {
		return nil, m.CreatePolicyErr
	}
	return &iam.CreatePolicyOutput{
		Policy: &types.Policy{
			Arn: aws.String("arn:aws:iam::123456789012:policy/SSM-SessionManager-Policy"),
		},
	}, nil
}

func (m MockIAMClient) CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	if m.CreateRoleErr != nil {
		return nil, m.CreateRoleErr
	}
	return &iam.CreateRoleOutput{}, nil
}

func (m MockIAMClient) AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	if m.AttachRolePolicyErr != nil {
		return nil, m.AttachRolePolicyErr
	}
	return &iam.AttachRolePolicyOutput{}, nil
}

func (m MockIAMClient) CreateInstanceProfile(ctx context.Context, params *iam.CreateInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.CreateInstanceProfileOutput, error) {
	if m.CreateInstanceProfileErr != nil {
		return nil, m.CreateInstanceProfileErr
	}
	return &iam.CreateInstanceProfileOutput{}, nil
}

func (m MockIAMClient) AddRoleToInstanceProfile(ctx context.Context, params *iam.AddRoleToInstanceProfileInput, optFns ...func(*iam.Options)) (*iam.AddRoleToInstanceProfileOutput, error) {
	if m.AddRoleToInstanceProfileErr != nil {
		return nil, m.AddRoleToInstanceProfileErr
	}
	return &iam.AddRoleToInstanceProfileOutput{}, nil
}
