package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Eks struct {
	AdditionalSecurityGroupIds            pulumi.StringArray
	AttachClusterEncryptionPolicy         pulumi.Bool
	CloudWatchLogGroup                    cloudWatchLogGroup
	ClusterAddons                         []Addon
	ClusterEncryptionConfig               pulumi.StringArrayMap
	ClusterEndpointPrivateAccess          pulumi.Bool
	ClusterEndpointPublicAccess           pulumi.Bool
	ClusterEndpointPublicAccessCidrs      pulumi.StringArray
	ClusterSecurityGroup                  securityGroup
	ClusterServiceIpv4Cidr                pulumi.String
	ClusterServiceIpv6Cidr                pulumi.String
	ClusterTimeouts                       pulumi.StringMap
	ControlPlaneSubnetIds                 pulumi.StringArray
	CreateClusterPrimarySecurityGroupTags pulumi.Bool
	EnabledLogTypes                       pulumi.StringArray
	EncryptionKey                         Kms
	IamRoleAdditionalPolicieArns          []string
	IdentityProvider                      identityProvider
	Irsa                                  irsa
	Name                                  string
	NodeSecurityGroup                     securityGroup
	ManagedNodeGroups                     NodeGroup
	SubnetIds                             pulumi.StringArray
	Tags                                  pulumi.StringMap
	Version                               pulumi.String
}

type Kms struct {
	Create                  bool
	Description             pulumi.String
	DeletionWindowInDays    pulumi.Int
	EnableRotation          pulumi.Bool
	EnableDefaultPolicy     pulumi.Bool
	Owners                  pulumi.StringArray
	Administrators          pulumi.StringArray
	Users                   pulumi.StringArray
	ServiceUsers            pulumi.StringArray
	SourcePolicyDocuments   pulumi.StringArray
	OverridePolicyDocuments pulumi.StringArray
	Aliases                 pulumi.StringArray
}

type cloudWatchLogGroup struct {
	Create          bool
	KmsKeyId        pulumi.String
	RetentionInDays pulumi.Int
}

type securityGroup struct {
	AdditionalRules         []securityGroupRule
	Create                  bool
	Description             pulumi.String
	EnableDefaultRules      bool
	ExistingSecurityGroupId pulumi.String
	Name                    string
	Tags                    pulumi.StringMap
	VpcId                   pulumi.StringInput
}

type securityGroupRule struct {
	cidrBlocks            string
	description           string
	fromPort              int
	ipv6CidrBlocks        string
	kind                  string
	protocol              string
	self                  bool
	sourceSecurityGroupId pulumi.IDOutput
	toPort                int
}

type irsa struct {
	CustomOidcThumbprints   pulumi.StringArray
	Enabled                 bool
	OpendIdConnectAudiences pulumi.StringArray
}

type Addon struct {
	Name                  string
	Version               pulumi.String
	ResolveConflicts      pulumi.String
	ServiceAccountRoleArn pulumi.String
	Preserve              pulumi.Bool
	ConfigurationValues   pulumi.String
}

type identityProvider struct {
	Name pulumi.StringMap
}

type NodeGroup struct {
	AmiId                              pulumi.String
	AmiReleaseVersion                  pulumi.String
	AmiType                            pulumi.String
	BlockDeviceMappings                pulumi.String
	CapacityReservationSpecification   pulumi.String
	CapacityType                       pulumi.String
	ClusterIpFamily                    pulumi.String
	ClusterName                        pulumi.String
	ClusterVersion                     pulumi.String
	CpuOptions                         pulumi.String
	CreateIamRole                      pulumi.String
	CreateLaunchTemplate               pulumi.Bool
	CreditSpecification                pulumi.String
	DesiredSize                        pulumi.Int
	DisableApiTermination              pulumi.String
	DiskSize                           pulumi.Int
	EbsOptimized                       pulumi.String
	ElasticGpuSpecifications           pulumi.String
	ElasticInferenceAccelerator        pulumi.String
	EnableMonitoring                   pulumi.String
	EnableRemoteAccess                 pulumi.Bool
	EnclaveOptions                     pulumi.String
	ForceUpdateVersion                 pulumi.Bool
	IamRoleAdditionalPolicies          []string
	IamRoleArn                         pulumi.String
	IamRoleAttachCniPolicy             pulumi.String
	IamRoleDescription                 pulumi.String
	IamRoleName                        pulumi.String
	IamRolePath                        pulumi.String
	IamRolePermissionsBoundary         pulumi.String
	IamRoleTags                        pulumi.String
	IamRoleUseNamePrefix               pulumi.String
	InstanceMarketOptions              pulumi.String
	InstanceTypes                      pulumi.StringArray
	KernelId                           pulumi.String
	KeyName                            pulumi.String
	Labels                             pulumi.StringMap
	ExistingLaunchTemplateId           pulumi.String
	ExistingLaunchTemplateVersion      pulumi.String
	LaunchTemplate                     LaunchTemplate
	LicenseSpecifications              pulumi.String
	MaintenanceOptions                 pulumi.String
	MaxSize                            pulumi.Int
	MetadataOptions                    pulumi.String
	MinSize                            pulumi.Int
	Name                               string
	NetworkInterfaces                  pulumi.String
	ExistingNodeRoleArn                pulumi.String
	Placement                          pulumi.String
	PrivateDnsNameOptions              pulumi.String
	RamDiskId                          pulumi.String
	SubnetIds                          pulumi.StringArray
	Taints                             eks.NodeGroupTaintArrayInput
	Timeouts                           pulumi.String
	UpdateConfig                       pulumi.String
	UpdateLaunchTemplateDefaultVersion pulumi.String
	UseCustomLaunchTemplate            pulumi.String
	UseExistingLaunchTemplate          bool
}

type LaunchTemplate struct {
	BlockDeviceMappings               ec2.LaunchTemplateBlockDeviceMappingArray
	CapacityReservation               ec2.LaunchTemplateCapacityReservationSpecificationArgs
	CpuCores                          pulumi.Int
	DisableApiStop                    pulumi.Bool
	DisableApiTermination             pulumi.Bool
	DiskSize                          pulumi.Int
	EbsOptimized                      pulumi.String
	CreditSpecification               ec2.LaunchTemplateCreditSpecificationArgs
	ElasticGpuSpecifications          ec2.LaunchTemplateElasticGpuSpecificationArray
	ElasticInferenceAccelerator       pulumi.String
	IamInstanceProfileName            pulumi.String
	ImageId                           pulumi.String
	InstanceType                      pulumi.String
	KernelId                          pulumi.String
	KeyName                           pulumi.String
	InstanceInitiatedShutdownBehavior pulumi.String
	RamDiskId                         pulumi.String
	InstanceMarketOptions             ec2.LaunchTemplateInstanceMarketOptionsArgs
	LicenseSpecifications             ec2.LaunchTemplateLicenseSpecificationArray
	MetadataOptions                   ec2.LaunchTemplateMetadataOptionsArgs
	Monitoring                        ec2.LaunchTemplateMonitoringArgs
	VpcSecurityGroupIds               pulumi.StringArray
	Placement                         ec2.LaunchTemplatePlacementArgs
	NetworkInterfaces                 ec2.LaunchTemplateNetworkInterfaceArray
	TagSpecifications                 ec2.LaunchTemplateTagSpecificationArray
	UserData                          pulumi.String
	UpdateDefaultVersion  			  pulumi.Bool
}

type NodeGroupCreateOutPut struct {
	NodeGroup *eks.NodeGroup
}

type EksCreateOutPut struct {
	Cluster *eks.Cluster
	NodeGroupOutput *NodeGroupCreateOutPut
}
