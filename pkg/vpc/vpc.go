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

type Vpc struct {
	Azs                              []string
	Cidr                             pulumi.String
	Database                         Subnet
	DhcpOption                       DhcpOption
	EnableDnsHostnames               pulumi.Bool
	EnableDnsSupport                 pulumi.Bool
	EnableNetworkAddressUsageMetrics pulumi.Bool
	EnableIpv6                       pulumi.Bool
	InstanceTenancy                  pulumi.String
	InternetGateway                  InternetGateway
	Ipv4IpamPoolId                   pulumi.String
	Ipv4NetmaskLength                pulumi.String
	Ipv6Cidr                         pulumi.String
	Ipv6CidrBlockNetworkBorderGroup  pulumi.String
	Ipv6IpamPoolId                   pulumi.String
	Ipv6NetmaskLength                pulumi.Int
	Name                             string
	NatGateway                       NatGateway
	NetworkAcl                       NetworkAcl
	PrivateSubnet                    Subnet
	PublicSubnet                     Subnet
	SecondaryCidr                    pulumi.StringArray
	Tags                             pulumi.StringMap
	UseIpamPool                      pulumi.Bool
}

type DhcpOption struct {
	Create             bool
	DomainName         string
	DomainNameServers  pulumi.StringArray
	NetbiosNameServers pulumi.StringArray
	NetbiosNodeType    pulumi.StringPtrInput
	NtpServers         pulumi.StringArray
	Tags               pulumi.StringMap
}

type Subnet struct {
	AssignIpv6AddressOnCreation             pulumi.Bool
	Cidrs                                   []string
	CreateDatabaseInternetGatewayRoute      pulumi.Bool
	CreateDatabaseNatGatewayRoute           pulumi.Bool
	CreateDatabaseSubnetGroup               pulumi.Bool
	CreateDatabaseSubnetRouteTable          pulumi.Bool
	DatabaseSubnetGroupName                 pulumi.String
	DatabaseSubnetGroupTags                 pulumi.StringMap
	EnableDns64                             pulumi.Bool
	EnableResourceNameDnsAaaaRecordOnLaunch pulumi.Bool
	EnableResourceNameDnsARecordOnLaunch    pulumi.Bool
	Ipv6Native                              pulumi.Bool
	Ipv6Prefixes                            pulumi.StringArray
	MapIpOnLaunch                           pulumi.Bool
	PrivateDnsHostnameTypeOnLaunch          pulumi.String
	RouteTableTags                          pulumi.StringMap
	Tags                                    pulumi.StringMap
	TagsPerAz                               pulumi.StringMap
}

type NetworkAcl struct {
	AclTags             pulumi.StringMap
	DedicatedNetworkAcl pulumi.Bool
	InboundAclRules     []pulumi.StringMap
	OutboundAclRules    []pulumi.StringMap
}

type InternetGateway struct {
	CreateEgressOnlyIgw pulumi.Bool
	IgwTags             pulumi.StringMap
}

type NatGateway struct {
	ExternalNatIpIds               pulumi.StringArray
	ExternalNatIps                 pulumi.StringArray
	NatEipTags                     pulumi.StringMap
	NatGatewayDestinationCidrBlock pulumi.String
	NatGatewayTags                 pulumi.StringMap
	OneNatGatewayPerAz             pulumi.Bool
	ReuseNatIps                    pulumi.Bool
	SingleNatGateway               pulumi.Bool
}

type VpcCreateOutput struct {
	Vpc                       *ec2.Vpc
	DhcpOptionID              *ec2.VpcDhcpOptions
	InternetGateway           *ec2.InternetGateway
	PublicRouteTable          *ec2.RouteTable
	PublicSubnets             []*ec2.Subnet
	PrivateRouteTables        []*ec2.RouteTable
	PrivateSubnets            []*ec2.Subnet
	NatGateways               []*ec2.NatGateway
	EgressOnlyInternetGateway *ec2.EgressOnlyInternetGateway
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
	vpcCreateOutput.Vpc = vpc

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
		vpcCreateOutput.DhcpOptionID = dhcpOptionId

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

	vpcCreateOutput.InternetGateway = igw

	// Create egress only internetGateway
	if v.InternetGateway.CreateEgressOnlyIgw {
		egressIgw, err := ec2.NewEgressOnlyInternetGateway(ctx, v.Name, &ec2.EgressOnlyInternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags:  mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name)}, v.InternetGateway.IgwTags, v.Tags),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		vpcCreateOutput.EgressOnlyInternetGateway = egressIgw
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
	vpcCreateOutput.PublicRouteTable = publicRouteTable

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
	var publicSubnets []*ec2.Subnet
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
			Tags:                                    mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name + "-public-" + v.Azs[index])}, v.PublicSubnet.Tags, v.Tags),
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
		publicSubnets = append(publicSubnets, subnet)
	}
	vpcCreateOutput.PublicSubnets = publicSubnets

	// create NatGateway
	var natGatewayCount int
	var natGateways []*ec2.NatGateway
	var privateRouteTables []*ec2.RouteTable

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
			SubnetId:     publicSubnets[i].ID(),
			Tags:         mergeTags(v.NatGateway.NatGatewayTags, v.Tags),
		}, pulumi.DependsOn([]pulumi.Resource{
			igw,
		}))
		if err != nil {
			return vpcCreateOutput, err
		}
		natGateways = append(natGateways, natGw)

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
			GatewayId:            natGw.ID(),
		})
		if err != nil {
			return vpcCreateOutput, err
		}
		privateRouteTables = append(privateRouteTables, privateRouteTable)
	}
	vpcCreateOutput.NatGateways = natGateways
	vpcCreateOutput.PrivateRouteTables = privateRouteTables

	// Create private Subnets
	var privateSubnets []*ec2.Subnet
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
			Tags:                                    mergeTags(pulumi.StringMap{"Name": pulumi.String(v.Name + "-private-" + v.Azs[index])}, v.PrivateSubnet.Tags, v.Tags),
		})
		if err != nil {
			if err != nil {
				return vpcCreateOutput, err
			}
		}
		if natGatewayCount > 1 {
			_, err = ec2.NewRouteTableAssociation(ctx, v.Name+"-private-"+v.Azs[index], &ec2.RouteTableAssociationArgs{
				SubnetId:     subnet.ID(),
				RouteTableId: privateRouteTables[index].ID(),
			})
			if err != nil {
				if err != nil {
					return vpcCreateOutput, err
				}
			}
		} else {
			_, err = ec2.NewRouteTableAssociation(ctx, v.Name+"-private-"+v.Azs[index], &ec2.RouteTableAssociationArgs{
				SubnetId:     subnet.ID(),
				RouteTableId: privateRouteTables[0].ID(),
			})
			if err != nil {
				if err != nil {
					return vpcCreateOutput, err
				}
			}
		}
		privateSubnets = append(privateSubnets, subnet)
	}
	vpcCreateOutput.PrivateRouteTables = privateRouteTables
	vpcCreateOutput.PrivateSubnets = privateSubnets

	return vpcCreateOutput, nil
}
