package eks

import (
	"strconv"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (e *Eks) CreateEKS(ctx *pulumi.Context) (*EksCreateOutPut, error) {
	eksCreateOutput := &EksCreateOutPut{}

	//Get partition
	partition, err := aws.GetPartition(ctx, nil, nil)
	if err != nil {
		return eksCreateOutput, err
	}
	ctx.Export("Partition", pulumi.String(partition.Id))

	// Get caller Identity
	caller, err := aws.GetCallerIdentity(ctx, nil, nil)
	if err != nil {
		return eksCreateOutput, err
	}

	// Get session Context
	session, err := iam.GetSessionContext(ctx, &iam.GetSessionContextArgs{
		Arn: caller.Arn,
	}, nil)
	if err != nil {
		return eksCreateOutput, err
	}
	ctx.Export("Partition", pulumi.String(session.Id))

	// Create Cluster and Node Security Group
	clusterSg, err := ec2.NewSecurityGroup(ctx, e.Name+"-cluster", &ec2.SecurityGroupArgs{
		Description: e.ClusterSecurityGroup.Description,
		VpcId:       e.ClusterSecurityGroup.VpcId,
	})
	if err != nil {
		return eksCreateOutput, err
	}

	nodeSg, err := ec2.NewSecurityGroup(ctx, e.Name+"-node", &ec2.SecurityGroupArgs{
		Description: e.NodeSecurityGroup.Description,
		VpcId:       e.NodeSecurityGroup.VpcId,
	})
	if err != nil {
		return eksCreateOutput, err
	}

	// Cluster Security Group Rule

	if e.ClusterSecurityGroup.Create {
		clusterSecurityGroupRules := []securityGroupRule{
			{
				kind:                  "ingress",
				fromPort:              443,
				toPort:                443,
				protocol:              "tcp",
				description:           "Node groups to cluster API",
				sourceSecurityGroupId: nodeSg.ID(),
			},
			// {
			// 	kind:        "egress",
			// 	fromPort:    0,
			// 	toPort:      0,
			// 	protocol:    "-1",
			// 	description: "Allow all egress",
			// 	cidrBlocks:  "0.0.0.0/0",
			// 	// ipv6CidrBlocks: "::/0",
			// },
		}
		clusterSecurityGroupRules = append(clusterSecurityGroupRules, e.ClusterSecurityGroup.AdditionalRules...)
		for index, rule := range clusterSecurityGroupRules {
			securityGroupRuleArgs := createSecurityGroupRule(rule)
			securityGroupRuleArgs.SecurityGroupId = clusterSg.ID()
			_, err := ec2.NewSecurityGroupRule(ctx, e.Name+strconv.Itoa(index), securityGroupRuleArgs,
				pulumi.DependsOn([]pulumi.Resource{
					nodeSg,
					clusterSg,
				}))
			if err != nil {
				return eksCreateOutput, err
			}
		}
	}

	// Create Cluster Iam role
	roleArgs := &iam.RoleArgs{}
	roleArgs.Tags = e.Tags
	roleArgs.AssumeRolePolicy = pulumi.String(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Sid": "EKSClusterAssumeRole",
			"Effect": "Allow",
			"Principal": {
				"Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}]
	}`)
	if e.CloudWatchLogGroup.Create {
		roleArgs.InlinePolicies = &iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name: pulumi.String("cloudwatch"),
				Policy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Deny",
						"Resource": "*",
						"Action": "logs:CreateLogGroup"
					}]
				}`),
			},
		}
	}
	if e.EncryptionKey.Create {
		roleArgs.InlinePolicies = &iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name: pulumi.String("kms"),
				Policy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Resource": "*",
						"Action": [
							"kms:Encrypt",
							"kms:Decrypt",
							"kms:ListGrants",
							"kms:DescribeKey",
						]
					}]
				}`),
			},
		}
	}
	eksRole, err := iam.NewRole(ctx, e.Name, roleArgs)
	if err != nil {
		return eksCreateOutput, err
	}

	// Role attachment
	eksPolicies := append([]string{
		"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}, e.IamRoleAdditionalPolicieArns...)
	for index, eksPolicy := range eksPolicies {
		_, err := iam.NewRolePolicyAttachment(ctx, e.Name+strconv.Itoa(index), &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(eksPolicy),
			Role:      eksRole.Name,
		})
		if err != nil {
			return eksCreateOutput, err
		}
	}

	// Create EKS cluster
	cluster, err := eks.NewCluster(ctx, e.Name, &eks.ClusterArgs{
		RoleArn: eksRole.Arn,
		VpcConfig: &eks.ClusterVpcConfigArgs{
			EndpointPrivateAccess: e.ClusterEndpointPrivateAccess,
			EndpointPublicAccess:  e.ClusterEndpointPublicAccess,
			SubnetIds: e.SubnetIds,
			PublicAccessCidrs:     e.ClusterEndpointPublicAccessCidrs,
			SecurityGroupIds: append(pulumi.StringArray{clusterSg.ID()}, e.AdditionalSecurityGroupIds...),
		},
		Tags: e.Tags,
		// EnabledClusterLogTypes: e.EnabledLogTypes,
		KubernetesNetworkConfig: &eks.ClusterKubernetesNetworkConfigArgs{
			IpFamily:        pulumi.String("ipv4"),
			ServiceIpv4Cidr: e.ClusterServiceIpv4Cidr,
		},
		// EncryptionConfig: &eks.ClusterEncryptionConfigArgs{
		// 	Provider: &eks.ClusterEncryptionConfigProviderArgs{
		// 		// KeyArn: e.EncryptionKey.,
		// 	},
		// 	Resources: pulumi.StringArray{pulumi.String("secret")},
		// },
	}, pulumi.DependsOn([]pulumi.Resource{
		// eksVpcPolicy,
		// eksPolicy,
	}))
	if err != nil {
		return eksCreateOutput, err
	}
	eksCreateOutput.Cluster = cluster
	clusterTag := pulumi.StringMap{}
	clusterName := cluster.Name.ApplyT(func(clusterName string) string {
		clusterTag["kubernetes.io/cluster/"+clusterName] = pulumi.String("owned")
		return "kubernetes.io/cluster/" + clusterName
	}).(pulumi.StringOutput)

	e.Tags = mergeTags(clusterTag, e.Tags)

	for index, subnet := range e.SubnetIds {
		_, err = ec2.NewTag(ctx, e.Name+strconv.Itoa(index), &ec2.TagArgs{
			ResourceId: subnet,
			Key:        clusterName,
			Value:      pulumi.String("owned"),
		})
		if err != nil {
			return eksCreateOutput, err
		}
	}

	// Create Cloudwatch Log Group
	if e.CloudWatchLogGroup.Create {

		_, err := cloudwatch.NewLogGroup(ctx, e.Name, &cloudwatch.LogGroupArgs{
			RetentionInDays: e.CloudWatchLogGroup.RetentionInDays,
			// KmsKeyId: ,
			Tags: e.Tags,
		})
		if err != nil {
			return eksCreateOutput, err
		}
	}

	//################################################################################
	//# IRSA
	//# Note - this is different from EKS identity provider
	//################################################################################
	clusterCert := tls.GetCertificateOutput(ctx, tls.GetCertificateOutputArgs{
		Url: cluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer(),
	})
	_, err = iam.NewOpenIdConnectProvider(ctx, e.Name, &iam.OpenIdConnectProviderArgs{
		ClientIdLists: pulumi.StringArray{
			pulumi.String("sts.amazonaws.com"),
		},
		ThumbprintLists: pulumi.StringArray{
			clusterCert.Certificates().Index(pulumi.Int(0)).Sha1Fingerprint(),
		},
		Url: cluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer().Elem().ToStringOutput(),
	})
	if err != nil {
		return eksCreateOutput, err
	}

	// Create NodeGroup
	nodeGroupOutput, err := e.CreateEksNodeGroups(ctx, nodeSg.ID(), clusterSg.ID(), cluster.Name)
	if err != nil {
		return eksCreateOutput, err
	}
	eksCreateOutput.NodeGroupOutput = nodeGroupOutput
	// create Addons
	for _, addon := range e.ClusterAddons {
		err := e.CreateAddon(ctx, addon, cluster, nodeGroupOutput.NodeGroup)
		if err != nil {
			return eksCreateOutput, err
		}
	}

	return eksCreateOutput, nil
}

// subtent tags
// kubernetes.io/cluster/CLUSTER_NAME
