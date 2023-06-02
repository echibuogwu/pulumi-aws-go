package main

import (
	
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/echibuogwu/pulumi-aws-go/pkg/vpc"
	
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
			NatGatewayDestinationCidrBlock: "0.0.0.0/0",
			SingleNatGateway:               true,
		}
		vpc := &vpc.Vpc{
			Name:            ctx.Stack(),
			InstanceTenancy: "default",
			Azs:             []string{region + "a", region + "b", region + "c"},
			Cidr:            "10.40.0.0/16",
			Tags:            tags,
			PublicSubnet:    publicSubnet,
			PrivateSubnet:   privateSubnet,
			NatGateway:      natGateway,
		}
		output, err := vpc.CreateVpc(ctx)
		if err != nil {
			return err
		}
		ctx.Export("VPC_ID", output.Vpc.ID())
		return nil
	})
}
