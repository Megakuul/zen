package zen

import (
	"fmt"
	"net/url"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Operator struct {
	region string
	account string
	domains []string
	deleteProtection bool
}

func (o *Operator) Deploy(ctx *pulumi.Context) error {
	tableOutputs, err := o.deployTable(ctx)
	if err!=nil {
		return fmt.Errorf("failed to deploy table: %v", err)
	}
	schedulerOutputs, err := o.deployScheduler(ctx)
	if err!=nil {
		return fmt.Errorf("failed to deploy scheduler: %v", err)
	}
	proxyOutputs, err := o.deployProxy(ctx, &proxyInput{
		SchedulerDomain: schedulerOutputs.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err!=nil {
				return fmt.Sprintf("invalid.domain")
			}
			return url.Host
		}).(pulumi.StringOutput),
		ManagerDomain: schedulerOutputs.PublicUrl.ApplyT(func(input string) string {
			url, err := url.Parse(input)
			if err!=nil {
				return fmt.Sprintf("invalid.domain")
			}
			return url.Host
		}).(pulumi.StringOutput),
	})
	return nil
}
