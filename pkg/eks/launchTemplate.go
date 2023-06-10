package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (e *Eks) CreateLaunchTemplate(ctx *pulumi.Context, nodeSecurityGroupId pulumi.IDOutput) (*ec2.LaunchTemplate, error) {
	lauchTemplateArgs := &ec2.LaunchTemplateArgs{}
	// lauchTemplateArgs.CapacityReservationSpecification = e.ManagedNodeGroups.LauchTemplate.CapacityReservation
	lauchTemplateArgs.VpcSecurityGroupIds = append(e.ManagedNodeGroups.LauchTemplate.VpcSecurityGroupIds, nodeSecurityGroupId)
	// lauchTemplateArgs.CreditSpecification = e.ManagedNodeGroups.LauchTemplate.CreditSpecification
	// lauchTemplateArgs.ElasticGpuSpecifications = e.ManagedNodeGroups.LauchTemplate.ElasticGpuSpecifications
	// lauchTemplateArgs.InstanceMarketOptions = e.ManagedNodeGroups.LauchTemplate.InstanceMarketOptions
	// lauchTemplateArgs.LicenseSpecifications = e.ManagedNodeGroups.LauchTemplate.LicenseSpecifications
	// lauchTemplateArgs.MetadataOptions = e.ManagedNodeGroups.LauchTemplate.MetadataOptions
	// lauchTemplateArgs.Monitoring = e.ManagedNodeGroups.LauchTemplate.Monitoring

	// lauchTemplateArgs.Placement = e.ManagedNodeGroups.LauchTemplate.Placement
	// lauchTemplateArgs.NetworkInterfaces = e.ManagedNodeGroups.LauchTemplate.NetworkInterfaces
	// lauchTemplateArgs.TagSpecifications = e.ManagedNodeGroups.LauchTemplate.TagSpecifications
	if e.ManagedNodeGroups.LauchTemplate.DiskSize > 0 {
		lauchTemplateArgs.BlockDeviceMappings = ec2.LaunchTemplateBlockDeviceMappingArray{
			&ec2.LaunchTemplateBlockDeviceMappingArgs{
				DeviceName: pulumi.String("/dev/xvda"),
				Ebs: &ec2.LaunchTemplateBlockDeviceMappingEbsArgs{
					VolumeSize:          e.ManagedNodeGroups.LauchTemplate.DiskSize,
					VolumeType:          pulumi.String("gp3"),
					Iops:                pulumi.Int(10000),
					Throughput:          pulumi.Int(1000),
					DeleteOnTermination: pulumi.String("true"),
				},
			},
		}
	}
	if e.ManagedNodeGroups.LauchTemplate.CpuCores > 0 {
		lauchTemplateArgs.CpuOptions = &ec2.LaunchTemplateCpuOptionsArgs{
			CoreCount: e.ManagedNodeGroups.LauchTemplate.CpuCores,
		}
	}

	if e.ManagedNodeGroups.LauchTemplate.IamInstanceProfileName != "" {
		lauchTemplateArgs.IamInstanceProfile = &ec2.LaunchTemplateIamInstanceProfileArgs{
			Name: e.ManagedNodeGroups.LauchTemplate.IamInstanceProfileName,
		}
	}

	if e.ManagedNodeGroups.LauchTemplate.ElasticInferenceAccelerator != "" {
		lauchTemplateArgs.ElasticInferenceAccelerator = &ec2.LaunchTemplateElasticInferenceAcceleratorArgs{
			Type: e.ManagedNodeGroups.LauchTemplate.ElasticInferenceAccelerator,
		}
	}

	if e.ManagedNodeGroups.LauchTemplate.DisableApiStop {
		lauchTemplateArgs.DisableApiStop = e.ManagedNodeGroups.LauchTemplate.DisableApiStop
	}

	if e.ManagedNodeGroups.LauchTemplate.DisableApiTermination {
		lauchTemplateArgs.DisableApiTermination = e.ManagedNodeGroups.LauchTemplate.DisableApiTermination
	}

	if e.ManagedNodeGroups.LauchTemplate.EbsOptimized != "" {
		lauchTemplateArgs.EbsOptimized = e.ManagedNodeGroups.LauchTemplate.EbsOptimized
	}

	if e.ManagedNodeGroups.LauchTemplate.ImageId != "" {
		lauchTemplateArgs.ImageId = e.ManagedNodeGroups.LauchTemplate.ImageId
	}

	if e.ManagedNodeGroups.LauchTemplate.InstanceType != "" {
		lauchTemplateArgs.InstanceType = e.ManagedNodeGroups.LauchTemplate.InstanceType
	}

	if e.ManagedNodeGroups.LauchTemplate.KernelId != "" {
		lauchTemplateArgs.KernelId = e.ManagedNodeGroups.LauchTemplate.KernelId
	}

	if e.ManagedNodeGroups.LauchTemplate.KeyName != "" {
		lauchTemplateArgs.KeyName = e.ManagedNodeGroups.LauchTemplate.KeyName
	}

	if e.ManagedNodeGroups.LauchTemplate.InstanceInitiatedShutdownBehavior != "" {
		lauchTemplateArgs.InstanceInitiatedShutdownBehavior = e.ManagedNodeGroups.LauchTemplate.InstanceInitiatedShutdownBehavior
	}

	if e.ManagedNodeGroups.LauchTemplate.RamDiskId != "" {
		lauchTemplateArgs.RamDiskId = e.ManagedNodeGroups.LauchTemplate.RamDiskId
	}

	if e.ManagedNodeGroups.LauchTemplate.UserData != "" {
		lauchTemplateArgs.UserData = e.ManagedNodeGroups.LauchTemplate.UserData
	}

	template, err := ec2.NewLaunchTemplate(ctx, e.ManagedNodeGroups.Name, lauchTemplateArgs)
	if err != nil {
		return &ec2.LaunchTemplate{}, err
	}
	return template, nil
}
