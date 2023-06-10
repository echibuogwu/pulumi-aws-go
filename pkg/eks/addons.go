package eks

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (e *Eks) CreateAddon(ctx *pulumi.Context, addon Addon, cluster *eks.Cluster) error {
	_addon := &eks.AddonArgs{}
	_addon.Tags = e.Tags
	_addon.ClusterName = cluster.Name
	_addon.AddonName = pulumi.String(addon.Name)

	if addon.Version != "" {
		_addon.AddonVersion = addon.Version
	}
	if addon.ConfigurationValues != "" {
		_addon.ConfigurationValues = addon.ConfigurationValues
	}
	if addon.Preserve {
		_addon.Preserve = addon.Preserve
	}
	if addon.ServiceAccountRoleArn != "" {
		_addon.ServiceAccountRoleArn = addon.ServiceAccountRoleArn
	}

	_, err := eks.NewAddon(ctx, addon.Name, _addon, pulumi.DependsOn([]pulumi.Resource{
		cluster,
	}))
	if err != nil {
		return err
	}
	return nil
}
