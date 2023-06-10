package main

import (
	"github.com/echibuogwu/pulumi-aws-go/pkg/eks"
	"github.com/echibuogwu/pulumi-aws-go/pkg/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		region := "eu-central-1"
		tags := make(pulumi.StringMap)
		tags["team"] = pulumi.String("devops")

		publicSubnet := vpc.Subnet{
			Cidrs:                          []string{"10.40.1.0/24", "10.40.2.0/24", "10.40.3.0/24"},
			Tags:                           pulumi.StringMap{"type": pulumi.String("private")},
			PrivateDnsHostnameTypeOnLaunch: "resource-name",
		}
		privateSubnet := vpc.Subnet{
			Cidrs:                          []string{"10.40.101.0/24", "10.40.102.0/24", "10.40.103.0/24"},
			Tags:                           pulumi.StringMap{"type": pulumi.String("public")},
			PrivateDnsHostnameTypeOnLaunch: "resource-name",
		}
		natGateway := vpc.NatGateway{
			NatGatewayDestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			SingleNatGateway:               true,
		}
		vpc := &vpc.Vpc{
			Name:               ctx.Stack(),
			InstanceTenancy:    "default",
			Azs:                []string{region + "a", region + "b", region + "c"},
			Cidr:               "10.40.0.0/16",
			Tags:               tags,
			PublicSubnet:       publicSubnet,
			PrivateSubnet:      privateSubnet,
			NatGateway:         natGateway,
			EnableDnsHostnames: pulumi.Bool(true),
			EnableDnsSupport:   pulumi.Bool(true),
		}
		output, err := vpc.CreateVpc(ctx)
		if err != nil {
			return err
		}

		// EKS Cluster Configs
		cluster := &eks.Eks{}
		cluster.Name = ctx.Stack()
		cluster.ClusterServiceIpv4Cidr = pulumi.String("172.16.0.0/12")
		cluster.ClusterAddons = []eks.Addon{
			{
				Name: "coredns",
			},
			{
				Name: "vpc-cni",
			},
		}
		clusterSg := &cluster.ClusterSecurityGroup
		nodeSg := &cluster.NodeSecurityGroup
		clusterSg.VpcId = output.VpcId
		clusterSg.Create = true
		nodeSg.Create = true
		nodeSg.VpcId = output.VpcId
		clusterSg.Description = "This is the EKS cluster security group"
		nodeSg.Description = "This is the EKS nodes security group"
		cluster.SubnetIds = output.PrivateSubnetsIds
		cluster.ManagedNodeGroups.Name = "workernode"
		cluster.ManagedNodeGroups.AmiId = ""
		cluster.ManagedNodeGroups.CapacityType = "ON_DEMAND"
		cluster.ManagedNodeGroups.MinSize = 1
		cluster.ManagedNodeGroups.DesiredSize = 1
		cluster.ManagedNodeGroups.MaxSize = 2
		cluster.ManagedNodeGroups.SubnetIds = output.PrivateSubnetsIds
		cluster.ManagedNodeGroups.LauchTemplate.DiskSize = 200
		cluster.ManagedNodeGroups.AmiType = "AL2_ARM_64"
		cluster.ManagedNodeGroups.InstanceTypes = pulumi.StringArray{pulumi.String("t4g.medium")}

		_, err = cluster.CreateEKS(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}
