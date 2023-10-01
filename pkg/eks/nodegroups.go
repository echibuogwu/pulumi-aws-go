package eks

import (
	"strconv"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (e *Eks) CreateEksNodeGroups(ctx *pulumi.Context, nodeSecurityGroupId, clusterSecurityGroupId pulumi.IDOutput, clusterName pulumi.StringOutput) (*NodeGroupCreateOutPut, error) {
	nodeGroupCreateOutput := &NodeGroupCreateOutPut{}

	// Create Node SecurityGroupRule
	nodeSecurityGroupRules := []securityGroupRule{
		{
			kind:                  "ingress",
			fromPort:              443,
			toPort:                443,
			protocol:              "tcp",
			description:           "Cluster API to node groups",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:                  "ingress",
			fromPort:              10250,
			toPort:                10250,
			protocol:              "tcp",
			description:           "Cluster API to node kubelets",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:        "ingress",
			fromPort:    53,
			toPort:      53,
			protocol:    "tcp",
			description: "Node to node CoreDNS",
			self:        true,
		},
		{
			kind:        "ingress",
			fromPort:    53,
			toPort:      53,
			protocol:    "udp",
			description: "Node to node CoreDNS UDP",
			self:        true,
		},
		{
			kind:        "ingress",
			fromPort:    1025,
			toPort:      65535,
			protocol:    "tcp",
			description: "Node to node ingress on ephemeral ports",
			self:        true,
		},
		{
			kind:                  "ingress",
			fromPort:              4443,
			toPort:                4443,
			protocol:              "tcp",
			description:           "Cluster API to node 4443/tcp webhook",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:                  "ingress",
			fromPort:              6443,
			toPort:                6443,
			protocol:              "tcp",
			description:           "Cluster API to node 6443/tcp webhook",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:                  "ingress",
			fromPort:              8443,
			toPort:                8443,
			protocol:              "tcp",
			description:           "Cluster API to node 8443/tcp webhook",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:                  "ingress",
			fromPort:              9443,
			toPort:                9443,
			protocol:              "tcp",
			description:           "Cluster API to node 9443/tcp webhook",
			sourceSecurityGroupId: clusterSecurityGroupId,
		},
		{
			kind:        "egress",
			fromPort:    0,
			toPort:      0,
			protocol:    "-1",
			description: "Allow all egress",
			cidrBlocks:  "0.0.0.0/0",
		},
	}

	nodeSecurityGroupRules = append(nodeSecurityGroupRules, e.NodeSecurityGroup.AdditionalRules...)
	for index, rule := range nodeSecurityGroupRules {
		securityGroupRuleArgs := createSecurityGroupRule(rule)
		securityGroupRuleArgs.SecurityGroupId = nodeSecurityGroupId
		_, err := ec2.NewSecurityGroupRule(ctx, e.ManagedNodeGroups.Name+strconv.Itoa(index), securityGroupRuleArgs)
		if err != nil {
			return nodeGroupCreateOutput, nil
		}
	}

	ngArgs := &eks.NodeGroupArgs{}
	remoteAccess := &eks.NodeGroupRemoteAccessArgs{}

	if e.ManagedNodeGroups.ExistingNodeRoleArn == "" {
		nodeEksRole, err := iam.NewRole(ctx, e.ManagedNodeGroups.Name+"-node", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
						"Service": "ec2.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
				}]
			}`),
		})
		if err != nil {
			return nodeGroupCreateOutput, nil
		}

		ngArgs.NodeRoleArn = nodeEksRole.Arn

		nodeGroupPolicies := append([]string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		}, e.ManagedNodeGroups.IamRoleAdditionalPolicies...)

		for index, nodeGroupPolicy := range nodeGroupPolicies {
			_, err := iam.NewRolePolicyAttachment(ctx, e.ManagedNodeGroups.Name+strconv.Itoa(index), &iam.RolePolicyAttachmentArgs{
				Role:      nodeEksRole.Name,
				PolicyArn: pulumi.String(nodeGroupPolicy),
			})
			if err != nil {
				return nodeGroupCreateOutput, nil
			}
		}
	} else {
		ngArgs.NodeRoleArn = e.ManagedNodeGroups.ExistingNodeRoleArn
	}

	if e.ManagedNodeGroups.EnableRemoteAccess {
		// Generate an AWS Key Pair for SSH access
		privateKey, err := tls.NewPrivateKey(ctx, e.ManagedNodeGroups.Name, &tls.PrivateKeyArgs{
			Algorithm: pulumi.String("ED25519"),
		})
		if err != nil {
			return nodeGroupCreateOutput, nil
		}
		ctx.Export("WorkerNodes-ssh", privateKey.PrivateKeyOpenssh)

		keyPair, err := ec2.NewKeyPair(ctx, e.ManagedNodeGroups.Name, &ec2.KeyPairArgs{
			KeyName:   pulumi.String(e.ManagedNodeGroups.Name),
			PublicKey: privateKey.PublicKeyOpenssh,
		})
		if err != nil {
			return nodeGroupCreateOutput, nil
		}
		remoteAccess.Ec2SshKey = keyPair.KeyName
		// remoteAccess.SourceSecurityGroupIds     // TBD
	}

	if e.ManagedNodeGroups.AmiType != "" {
		ngArgs.AmiType = e.ManagedNodeGroups.AmiType
	}
	if e.ManagedNodeGroups.CapacityType != "" {
		ngArgs.CapacityType = e.ManagedNodeGroups.CapacityType
	}

	if e.ManagedNodeGroups.DiskSize != 0 {
		ngArgs.DiskSize = e.ManagedNodeGroups.DiskSize
	}

	if e.ManagedNodeGroups.ForceUpdateVersion {
		ngArgs.ForceUpdateVersion = e.ManagedNodeGroups.ForceUpdateVersion
	}

	if len(e.ManagedNodeGroups.InstanceTypes) < 1 {
		ngArgs.InstanceTypes = e.ManagedNodeGroups.InstanceTypes
	}

	if e.ManagedNodeGroups.AmiReleaseVersion != "" {
		ngArgs.ReleaseVersion = e.ManagedNodeGroups.AmiReleaseVersion
	}

	if e.ManagedNodeGroups.ClusterVersion != "" {
		ngArgs.Version = e.ManagedNodeGroups.ClusterVersion
	}

	ngArgs.Labels = e.ManagedNodeGroups.Labels
	ngArgs.NodeGroupName = pulumi.String(e.ManagedNodeGroups.Name)
	ngArgs.ClusterName = clusterName
	ngArgs.SubnetIds = e.ManagedNodeGroups.SubnetIds
	ngArgs.Taints = e.ManagedNodeGroups.Taints
	ngArgs.RemoteAccess = remoteAccess
	ngArgs.Tags = e.Tags
	ngArgs.ScalingConfig = &eks.NodeGroupScalingConfigArgs{
		MaxSize:     e.ManagedNodeGroups.MaxSize,
		MinSize:     e.ManagedNodeGroups.MinSize,
		DesiredSize: e.ManagedNodeGroups.DesiredSize,
	}

	if e.ManagedNodeGroups.UseExistingLaunchTemplate {
		ngArgs.LaunchTemplate = &eks.NodeGroupLaunchTemplateArgs{
			Id:      e.ManagedNodeGroups.ExistingLaunchTemplateId,
			Version: e.ManagedNodeGroups.ExistingLaunchTemplateVersion,
		}
	} else if e.ManagedNodeGroups.CreateLaunchTemplate {
		template, err := e.CreateLaunchTemplate(ctx, nodeSecurityGroupId)
		ngArgs.LaunchTemplate = &eks.NodeGroupLaunchTemplateArgs{
			Id:      template.ID(),
			Version: pulumi.Sprintf("%d", template.LatestVersion.ToIntOutput()),
		}
		if err != nil {
			return nodeGroupCreateOutput, err
		}
	} else {
		ngArgs.DiskSize = e.ManagedNodeGroups.DiskSize
	}
	//  ###############  Create EKS Node Group  ###############
	nodeGroup, err := eks.NewNodeGroup(ctx, e.ManagedNodeGroups.Name, ngArgs)
	if err != nil {
		return nodeGroupCreateOutput, nil
	}
	nodeGroupCreateOutput.NodeGroup = nodeGroup

	return nodeGroupCreateOutput, nil
}
