package helper
 
import (
    "context"
    "fmt"
    "testing"
 
    "github.com/aws/aws-sdk-go-v2/aws"
    // "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/stretchr/testify/assert"
)
 
// Mock implementation of the sgInterface for testing
type Client struct {
    DescribeSubnetsError               error
    CreateSecurityGroupError           error
    AuthorizeSecurityGroupIngressError error
}
 
func TestCreateSecurityGroup(t *testing.T) {
    t.Run("DescribeSubnetsError", func(t *testing.T) {
        _, err := CreateSecurityGroup(Client{
            DescribeSubnetsError: fmt.Errorf("couldn't create"),
        }, "SubnetID")
        assert.Equal(t, "failed to describe subnet: couldn't create", err.Error())
    })
 
    t.Run("CreateSecurityGroupError", func(t *testing.T) {
        _, err := CreateSecurityGroup(Client{
            CreateSecurityGroupError: fmt.Errorf("create security group error"),
        },"subnetID")
        assert.Equal(t, "failed to create security group: create security group error", err.Error())
    })
 
    t.Run("AuthorizeSecurityGroupIngressError", func(t *testing.T) {
        _, err := CreateSecurityGroup(Client{
            AuthorizeSecurityGroupIngressError: fmt.Errorf("authorize security group ingress error"),
        }, "SubnetID")
        assert.Equal(t, "failed to authorize security group ingress: authorize security group ingress error", err.Error())
    })
 
    t.Run("Success", func(t *testing.T) {
        groupID, err := CreateSecurityGroup(Client{}, "SubnetID")
        assert.NoError(t, err)
        assert.Equal(t, "sg-123456", groupID)
    })
}
 
func (client Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
    if client.DescribeSubnetsError != nil {
        return nil, client.DescribeSubnetsError
    }
    return &ec2.DescribeSubnetsOutput{
        Subnets: []types.Subnet{
            {
                SubnetId: aws.String("subnetID"),
                VpcId:    aws.String("vpc-123456"),
            },
        },
    }, nil
}
 
 
func (client Client) CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
    if client.CreateSecurityGroupError != nil {
        return nil, client.CreateSecurityGroupError
    }
    return &ec2.CreateSecurityGroupOutput{
        GroupId: aws.String("sg-123456"),
    }, nil
}
 
func (client Client) AuthorizeSecurityGroupIngress(ctx context.Context, params *ec2.AuthorizeSecurityGroupIngressInput, optFns ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
    if client.AuthorizeSecurityGroupIngressError != nil {
        return nil, client.AuthorizeSecurityGroupIngressError
    }
    return &ec2.AuthorizeSecurityGroupIngressOutput{}, nil
}
 