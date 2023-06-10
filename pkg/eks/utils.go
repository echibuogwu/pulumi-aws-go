package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createSecurityGroupRule(securityGroupRule securityGroupRule) *ec2.SecurityGroupRuleArgs {
	securityGroupRuleArgs := &ec2.SecurityGroupRuleArgs{}
	securityGroupRuleArgs.Type = pulumi.String(securityGroupRule.kind)
	securityGroupRuleArgs.ToPort = pulumi.Int(securityGroupRule.toPort)
	securityGroupRuleArgs.FromPort = pulumi.Int(securityGroupRule.fromPort)
	securityGroupRuleArgs.Protocol = pulumi.String(securityGroupRule.protocol)
	if securityGroupRule.description != "" {
		securityGroupRuleArgs.Description = pulumi.String(securityGroupRule.description)
	}
	if securityGroupRule.cidrBlocks != "" {
		securityGroupRuleArgs.CidrBlocks = pulumi.StringArray{
			pulumi.String(securityGroupRule.cidrBlocks),
		}
	} else if securityGroupRule.self {
		securityGroupRuleArgs.Self = pulumi.BoolPtr(securityGroupRule.self)
	} else {
		securityGroupRuleArgs.SourceSecurityGroupId = securityGroupRule.sourceSecurityGroupId
	}

	if securityGroupRule.ipv6CidrBlocks != "" {
		securityGroupRuleArgs.Ipv6CidrBlocks = pulumi.StringArray{
			pulumi.String(securityGroupRule.ipv6CidrBlocks),
		}
	}
	return securityGroupRuleArgs
}

func mergeTags(tags ...pulumi.StringMap) pulumi.StringMap {
	merged := make(pulumi.StringMap)
	for _, tag := range tags {
		for k, v := range tag {
			merged[k] = v
		}
	}
	return merged
}
