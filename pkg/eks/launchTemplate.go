package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (e *Eks) CreateLaunchTemplate(ctx *pulumi.Context, nodeSecurityGroupId pulumi.IDOutput) (*ec2.LaunchTemplate, error) {
	launchTemplateArgs := &ec2.LaunchTemplateArgs{}
	// launchTemplateArgs.CapacityReservationSpecification = e.ManagedNodeGroups.LaunchTemplate.CapacityReservation
	launchTemplateArgs.VpcSecurityGroupIds = append(e.ManagedNodeGroups.LaunchTemplate.VpcSecurityGroupIds, nodeSecurityGroupId)
	// launchTemplateArgs.CreditSpecification = e.ManagedNodeGroups.LaunchTemplate.CreditSpecification
	// launchTemplateArgs.ElasticGpuSpecifications = e.ManagedNodeGroups.LaunchTemplate.ElasticGpuSpecifications
	// launchTemplateArgs.InstanceMarketOptions = e.ManagedNodeGroups.LaunchTemplate.InstanceMarketOptions
	// launchTemplateArgs.LicenseSpecifications = e.ManagedNodeGroups.LaunchTemplate.LicenseSpecifications
	// launchTemplateArgs.MetadataOptions = e.ManagedNodeGroups.LaunchTemplate.MetadataOptions
	// launchTemplateArgs.Monitoring = e.ManagedNodeGroups.LaunchTemplate.Monitoring

	// launchTemplateArgs.Placement = e.ManagedNodeGroups.LaunchTemplate.Placement
	// launchTemplateArgs.NetworkInterfaces = e.ManagedNodeGroups.LaunchTemplate.NetworkInterfaces
	// launchTemplateArgs.TagSpecifications = e.ManagedNodeGroups.LaunchTemplate.TagSpecifications
	if e.ManagedNodeGroups.LaunchTemplate.DiskSize > 0 {
		launchTemplateArgs.BlockDeviceMappings = ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvda"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeSize:          e.ManagedNodeGroups.LaunchTemplate.DiskSize,
					VolumeType:          pulumi.String("gp3"),
					Iops:                pulumi.Int(10000),
					Throughput:          pulumi.Int(1000),
					DeleteOnTermination: pulumi.String("true"),
				},
			},
		}
	}
	if e.ManagedNodeGroups.LaunchTemplate.CpuCores > 0 {
		launchTemplateArgs.CpuOptions = &ec2.LaunchTemplateCpuOptionsArgs{
			CoreCount: e.ManagedNodeGroups.LaunchTemplate.CpuCores,
		}
	}

	if e.ManagedNodeGroups.LaunchTemplate.IamInstanceProfileName != "" {
		launchTemplateArgs.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileArgs{
			Name: e.ManagedNodeGroups.LaunchTemplate.IamInstanceProfileName,
		}
	}

	if e.ManagedNodeGroups.LaunchTemplate.ElasticInferenceAccelerator != "" {
		launchTemplateArgs.ElasticInferenceAccelerator = &ec2.LaunchTemplateElasticInferenceAcceleratorArgs{
			Type: e.ManagedNodeGroups.LaunchTemplate.ElasticInferenceAccelerator,
		}
	}

	if e.ManagedNodeGroups.LaunchTemplate.DisableApiStop {
		launchTemplateArgs.DisableApiStop = e.ManagedNodeGroups.LaunchTemplate.DisableApiStop
	}

	if e.ManagedNodeGroups.LaunchTemplate.DisableApiTermination {
		launchTemplateArgs.DisableApiTermination = e.ManagedNodeGroups.LaunchTemplate.DisableApiTermination
	}

	if e.ManagedNodeGroups.LaunchTemplate.EbsOptimized != "" {
		launchTemplateArgs.EbsOptimized = e.ManagedNodeGroups.LaunchTemplate.EbsOptimized
	}

	if e.ManagedNodeGroups.LaunchTemplate.ImageId != "" {
		launchTemplateArgs.ImageId = e.ManagedNodeGroups.LaunchTemplate.ImageId
	}

	if e.ManagedNodeGroups.LaunchTemplate.InstanceType != "" {
		launchTemplateArgs.InstanceType = e.ManagedNodeGroups.LaunchTemplate.InstanceType
	}

	if e.ManagedNodeGroups.LaunchTemplate.KernelId != "" {
		launchTemplateArgs.KernelId = e.ManagedNodeGroups.LaunchTemplate.KernelId
	}

	if e.ManagedNodeGroups.LaunchTemplate.KeyName != "" {
		launchTemplateArgs.KeyName = e.ManagedNodeGroups.LaunchTemplate.KeyName
	}

	if e.ManagedNodeGroups.LaunchTemplate.InstanceInitiatedShutdownBehavior != "" {
		launchTemplateArgs.InstanceInitiatedShutdownBehavior = e.ManagedNodeGroups.LaunchTemplate.InstanceInitiatedShutdownBehavior
	}

	if e.ManagedNodeGroups.LaunchTemplate.RamDiskId != "" {
		launchTemplateArgs.RamDiskId = e.ManagedNodeGroups.LaunchTemplate.RamDiskId
	}

	if e.ManagedNodeGroups.LaunchTemplate.UserData != "" {
		launchTemplateArgs.UserData = e.ManagedNodeGroups.LaunchTemplate.UserData
	}

	if e.ManagedNodeGroups.LaunchTemplate.UpdateDefaultVersion {
		launchTemplateArgs.UpdateDefaultVersion = pulumi.Bool(true)
	}
	
	template, err := ec2.NewLaunchTemplate(ctx, e.ManagedNodeGroups.Name, launchTemplateArgs)
	if err != nil {
		return &ec2.LaunchTemplate{}, err
	}
	return template, nil
}
