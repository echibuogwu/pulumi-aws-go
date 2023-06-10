package vpc

import (
	"strconv"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	// "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func mergeTags(tags ...pulumi.StringMap) pulumi.StringMap {
	merged := make(pulumi.StringMap)
	for _, tag := range tags {
		for k, v := range tag {
			merged[k] = v
		}
	}
	return merged
}

func (v *Vpc) CreateVpc(ctx *pulumi.Context) (*VpcCreateOutput, error) {
	vpcCreateOutput := &VpcCreateOutput{}
	// Create VPC
	vpc, err := ec2.NewVpc(ctx, v.Name, &ec2.VpcArgs{
		CidrBlock:                        v.Cidr,
		InstanceTenancy:                  v.InstanceTenancy,
		EnableDnsHostnames:               v.EnableDnsHostnames,
		EnableDnsSupport:                 v.EnableDnsSupport,
		EnableNetworkAddressUsageMetrics: v.EnableNetworkAddressUsageMetrics,
		Tags:                             mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name)}, v.Tags),
	})
	if err != nil {
		return vpcCreateOutput, err
	}
	vpcCreateOutput.VpcId = vpc.ID()

	// Secondary Cidr association
	for index, cidr := range v.SecondaryCidr {
		_, err = ec2.NewVpcIpv4CidrBlockAssociation(ctx, v.Name+v.Azs[index], &ec2.VpcIpv4CidrBlockAssociationArgs{
			VpcId:     vpc.ID(),
			CidrBlock: cidr,
		})
		if err != nil {
			return vpcCreateOutput, err
		}
	}

	// Dhcp Options
	if v.DhcpOption.Create {
		dhcpOptionId, err := ec2.NewVpcDhcpOptions(ctx, v.DhcpOption.DomainName, &ec2.VpcDhcpOptionsArgs{
			DomainNameServers:  v.DhcpOption.DomainNameServers,
			NetbiosNameServers: v.DhcpOption.NetbiosNameServers,
			NetbiosNodeType:    v.DhcpOption.NetbiosNodeType,
			NtpServers:         v.DhcpOption.NtpServers,
			Tags:               mergeTags(pulumi.StringMap{"Name": pulumi.String(v.DhcpOption.DomainName)}, v.DhcpOption.Tags, v.Tags),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		vpcCreateOutput.DhcpOptionId = dhcpOptionId.ID()

		// Dhcp Options Association
		_, err = ec2.NewVpcDhcpOptionsAssociation(ctx, "dnsResolver", &ec2.VpcDhcpOptionsAssociationArgs{
			VpcId:         vpc.ID(),
			DhcpOptionsId: dhcpOptionId.ID(),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
	}

	// Create internet gateway
	igw, err := ec2.NewInternetGateway(ctx, v.Name, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags:  mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name)}, v.InternetGateway.IgwTags, v.Tags),
	})
	if err != nil {
		return vpcCreateOutput, err
	}

	vpcCreateOutput.InternetGatewayId = igw.ID()

	// Create egress only internetGateway
	if v.InternetGateway.CreateEgressOnlyIgw {
		egressIgw, err := ec2.NewEgressOnlyInternetGateway(ctx, v.Name, &ec2.EgressOnlyInternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags:  mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name)}, v.InternetGateway.IgwTags, v.Tags),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		vpcCreateOutput.EgressOnlyInternetGatewayId = egressIgw.ID()
	}

	// create public route table
	publicRouteTable, err := ec2.NewRouteTable(ctx, "public", &ec2.RouteTableArgs{
		VpcId:  vpc.ID(),
		Routes: ec2.RouteTableRouteArray{},
		Tags:   mergeTags(pulumi.StringMap{"Name": pulumi.String("public")}, v.PublicSubnet.Tags, v.PublicSubnet.RouteTableTags),
	})
	if err != nil {
		return vpcCreateOutput, err
	}
	vpcCreateOutput.PublicRouteTableId = publicRouteTable.ID()

	// Internet route
	_, err = ec2.NewRoute(ctx, v.Name+"-internet-route", &ec2.RouteArgs{
		RouteTableId:         publicRouteTable.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
	})
	if err != nil {
		return vpcCreateOutput, err
	}

	// Create public Subnets
	var publicSubnets pulumi.StringArray
	for index, subnet := range v.PublicSubnet.Cidrs {
		subnet, err := ec2.NewSubnet(ctx, v.Name+"-public-"+v.Azs[index], &ec2.SubnetArgs{
			VpcId:                                   vpc.ID(),
			CidrBlock:                               pulumi.String(subnet),
			AssignIpv6AddressOnCreation:             v.PublicSubnet.AssignIpv6AddressOnCreation,
			AvailabilityZone:                        pulumi.String(v.Azs[index]),
			MapPublicIpOnLaunch:                     v.PublicSubnet.MapIpOnLaunch,
			PrivateDnsHostnameTypeOnLaunch:          v.PublicSubnet.PrivateDnsHostnameTypeOnLaunch,
			EnableDns64:                             v.PublicSubnet.EnableDns64,
			EnableResourceNameDnsARecordOnLaunch:    v.PublicSubnet.EnableResourceNameDnsARecordOnLaunch,
			EnableResourceNameDnsAaaaRecordOnLaunch: v.PublicSubnet.EnableResourceNameDnsAaaaRecordOnLaunch,
			Tags:                                    mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name + "-public-" + v.Azs[index])}, pulumi.StringMap{"kubernetes.io/role/elb": pulumi.String("1")}, v.PublicSubnet.Tags, v.Tags),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, v.Name+"-public-"+v.Azs[index], &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		publicSubnets = append(publicSubnets, subnet.ID())
	}
	vpcCreateOutput.PublicSubnetIds = publicSubnets

	// create NatGateway
	var natGatewayCount int
	var natGateways pulumi.StringArray
	var privateRouteTables pulumi.StringArray

	if v.NatGateway.SingleNatGateway {
		natGatewayCount = 1
	} else {
		natGatewayCount = len(v.Azs)
	}

	for i := 0; i < natGatewayCount; i++ {
		eip, err := ec2.NewEip(ctx, v.Name+strconv.Itoa(i), &ec2.EipArgs{
			Tags: v.NatGateway.NatEipTags,
		})
		if err != nil {
			return vpcCreateOutput, err
		}

		// Create Natgateway
		natGw, err := ec2.NewNatGateway(ctx, "natgateway-"+strconv.Itoa(i), &ec2.NatGatewayArgs{
			AllocationId: eip.ID(),
			SubnetId:     publicSubnets[i],
			Tags:         mergeTags(v.NatGateway.NatGatewayTags, v.Tags),
		}, pulumi.DependsOn([]pulumi.Resource{
			igw,
		}))
		if err != nil {
			return vpcCreateOutput, err
		}
		natGateways = append(natGateways, natGw.ID())

		// create private route table
		privateRouteTable, err := ec2.NewRouteTable(ctx, "private-"+v.Azs[i], &ec2.RouteTableArgs{
			VpcId:  vpc.ID(),
			Routes: ec2.RouteTableRouteArray{},
			Tags:   mergeTags(pulumi.StringMap{"Name": pulumi.String("private-" + v.Azs[i])}, v.PrivateSubnet.Tags, v.PrivateSubnet.RouteTableTags),
		})
		if err != nil {
			return vpcCreateOutput, err
		}

		// Create private to natgateway route
		_, err = ec2.NewRoute(ctx, "private-natgateway-"+v.Azs[i], &ec2.RouteArgs{
			RouteTableId:         privateRouteTable.ID(),
			DestinationCidrBlock: v.NatGateway.NatGatewayDestinationCidrBlock,
			NatGatewayId:         natGw.ID(),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		privateRouteTables = append(privateRouteTables, privateRouteTable.ID())
	}
	vpcCreateOutput.NatGatewaysIds = natGateways
	vpcCreateOutput.PrivateRouteTableIds = privateRouteTables

	// Create private Subnets
	var privateSubnets pulumi.StringArray
	for index, subnet := range v.PrivateSubnet.Cidrs {
		subnet, err := ec2.NewSubnet(ctx, v.Name+"-private-"+v.Azs[index], &ec2.SubnetArgs{
			VpcId:                                   vpc.ID(),
			CidrBlock:                               pulumi.String(subnet),
			AssignIpv6AddressOnCreation:             v.PrivateSubnet.AssignIpv6AddressOnCreation,
			AvailabilityZone:                        pulumi.String(v.Azs[index]),
			PrivateDnsHostnameTypeOnLaunch:          v.PrivateSubnet.PrivateDnsHostnameTypeOnLaunch,
			EnableDns64:                             v.PrivateSubnet.EnableDns64,
			EnableResourceNameDnsARecordOnLaunch:    v.PrivateSubnet.EnableResourceNameDnsARecordOnLaunch,
			EnableResourceNameDnsAaaaRecordOnLaunch: v.PrivateSubnet.EnableResourceNameDnsAaaaRecordOnLaunch,
			Tags:                                    mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name + "-private-" + v.Azs[index])}, pulumi.StringMap{"kubernetes.io/role/internal-elb": pulumi.String("1")}, v.PrivateSubnet.Tags, v.Tags),
		})
		if err != nil {
			if err != nil {
				return vpcCreateOutput, err
			}
		}
		if natGatewayCount > 1 {
			_, err = ec2.NewRouteTableAssociation(ctx, v.Name+"-private-"+v.Azs[index], &ec2.RouteTableAssociationArgs{
				SubnetId:     subnet.ID(),
				RouteTableId: privateRouteTables[index],
			})
			if err != nil {
				if err != nil {
					return vpcCreateOutput, err
				}
			}
		} else {
			_, err = ec2.NewRouteTableAssociation(ctx, v.Name+"-private-"+v.Azs[index], &ec2.RouteTableAssociationArgs{
				SubnetId:     subnet.ID(),
				RouteTableId: privateRouteTables[0],
			})
			if err != nil {
				if err != nil {
					return vpcCreateOutput, err
				}
			}
		}
		privateSubnets = append(privateSubnets, subnet.ID())
	}
	vpcCreateOutput.PrivateRouteTableIds = privateRouteTables
	vpcCreateOutput.PrivateSubnetsIds = privateSubnets

	return vpcCreateOutput, nil
}
