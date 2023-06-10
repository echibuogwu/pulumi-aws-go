package vpc

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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
	VpcId                       pulumi.IDOutput
	DhcpOptionId                pulumi.IDOutput
	InternetGatewayId           pulumi.IDOutput
	PublicRouteTableId          pulumi.IDOutput
	PublicSubnetIds             pulumi.StringArray
	PrivateRouteTableIds        pulumi.StringArray
	PrivateSubnetsIds           pulumi.StringArray
	NatGatewaysIds              pulumi.StringArray
	EgressOnlyInternetGatewayId pulumi.IDOutput
}
