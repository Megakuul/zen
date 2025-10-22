package email

import (
	"fmt"

	"github.com/megakuul/zen/internal/deploy/util"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sesv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployInput struct {
	Region  string
	Domains []string
	AutoDns bool
}

type DeployOutput struct {
	EmailName      pulumi.StringOutput
	EmailPolicyArn pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	if len(input.Domains) < 1 {
		return nil, fmt.Errorf("expected at least one domain")
	}
	emailConfig, err := sesv2.NewConfigurationSet(ctx, "email", &sesv2.ConfigurationSetArgs{
		ConfigurationSetName: pulumi.String("zen-email"),
		Region:               pulumi.String(input.Region),
		SuppressionOptions: &sesv2.ConfigurationSetSuppressionOptionsArgs{
			SuppressedReasons: pulumi.ToStringArray([]string{"BOUNCE", "COMPLAINT"}),
		},
		SendingOptions: &sesv2.ConfigurationSetSendingOptionsArgs{
			SendingEnabled: pulumi.Bool(true),
		},
	})
	if err != nil {
		return nil, err
	}
	emailIdentity, err := sesv2.NewEmailIdentity(ctx, "email", &sesv2.EmailIdentityArgs{
		EmailIdentity: pulumi.String(input.Domains[0]),
		Region:        pulumi.String(input.Region),
		DkimSigningAttributes: &sesv2.EmailIdentityDkimSigningAttributesArgs{
			NextSigningKeyLength: pulumi.String("RSA_2048_BIT"),
		},
		ConfigurationSetName: emailConfig.ConfigurationSetName,
	})
	if err != nil {
		return nil, err
	}
	envelopeDomain := fmt.Sprintf("%s.%s", "bounce", input.Domains[0])
	_, err = sesv2.NewEmailIdentityMailFromAttributes(ctx, "email", &sesv2.EmailIdentityMailFromAttributesArgs{
		EmailIdentity:       emailIdentity.EmailIdentity,
		BehaviorOnMxFailure: pulumi.String("REJECT_MESSAGE"),
		MailFromDomain:      pulumi.String(envelopeDomain),
		Region:              pulumi.String(input.Region),
	})
	if err != nil {
		return nil, err
	}

	if input.AutoDns {
		envelopeZone, err := util.LookupZone(ctx, envelopeDomain)
		if err != nil {
			return nil, err
		}
		_, err = route53.NewRecord(ctx, "email-bounce", &route53.RecordArgs{
			ZoneId: pulumi.String(envelopeZone.Id),
			Name:   pulumi.String(envelopeDomain),
			Type:   pulumi.String("MX"),
			Records: pulumi.ToStringArray([]string{
				fmt.Sprintf("10 feedback-smtp.%s.amazonses.com", input.Region),
			}),
		})
		if err != nil {
			return nil, err
		}

		_, err = route53.NewRecord(ctx, "email-spf", &route53.RecordArgs{
			ZoneId: pulumi.String(envelopeZone.Id),
			Name:   pulumi.String(envelopeDomain),
			Type:   pulumi.String("TXT"),
			Records: pulumi.ToStringArray([]string{
				"v=spf1 include:amazonses.com -all",
			}),
		})
		if err != nil {
			return nil, err
		}

		zone, err := util.LookupZone(ctx, input.Domains[0])
		if err != nil {
			return nil, err
		}
		emailIdentity.DkimSigningAttributes.Tokens().ApplyT(func(selectors []string) (string, error) {
			for i, selector := range selectors {
				_, err = route53.NewRecord(ctx, fmt.Sprintf("email-dkim-%d", i), &route53.RecordArgs{
					ZoneId: pulumi.String(zone.Id),
					Name:   pulumi.String(fmt.Sprintf("%s._domainkey.%s", selector, input.Domains[0])),
					Type:   pulumi.String("CNAME"),
					Records: pulumi.ToStringArray([]string{
						fmt.Sprintf("%s.dkim.amazonses.com", selector),
					}),
				})
				if err != nil {
					return "", err
				}
			}
			return "", nil
		})
		_, err = route53.NewRecord(ctx, "email-dmarc", &route53.RecordArgs{
			ZoneId: pulumi.String(zone.Id),
			Name:   pulumi.String(fmt.Sprintf("_dmarc.%s", input.Domains[0])),
			Type:   pulumi.String("TXT"),
			Records: pulumi.ToStringArray([]string{
				"v=DMARC1; p=reject; adkim=s; aspf=r;", // spf is relaxed because SES envelope must be a subdomain of the identity.
			}),
		})
		if err != nil {
			return nil, err
		}
	} else {
		ctx.Export(fmt.Sprintf("REQUIRED_MX_BOUNCE"), pulumi.Sprintf("MX %s \"%s\"",
			pulumi.String(envelopeDomain),
			pulumi.String(fmt.Sprintf("10 feedback-smtp.%s.amazonses.com", input.Region)),
		))
		ctx.Export(fmt.Sprintf("REQUIRED_TXT_SPF"), pulumi.Sprintf("TXT %s \"%s\"",
			pulumi.String(envelopeDomain),
			pulumi.String("v=spf1 include:amazonses.com -all"),
		))
		emailIdentity.DkimSigningAttributes.Tokens().ApplyT(func(selectors []string) (string, error) {
			for i, selector := range selectors {
				ctx.Export(fmt.Sprintf("REQUIRED_CNAME_DKIM_%d", i), pulumi.Sprintf("CNAME %s %s",
					pulumi.String(fmt.Sprintf("%s._domainkey.%s", selector, input.Domains[0])),
					pulumi.String(fmt.Sprintf("%s.dkim.amazonses.com", selector)),
				))
			}
			return "", nil
		})
		ctx.Export(fmt.Sprintf("REQUIRED_TXT_DMARC"), pulumi.Sprintf("TXT %s \"%s\"",
			pulumi.String(fmt.Sprintf("_dmarc.%s", input.Domains[0])),
			pulumi.String("v=DMARC1; p=reject; adkim=s; aspf=r;"), // spf is relaxed because the SES envelope must be a subdomain of the identity.
		))
	}

	emailPolicy, err := iam.NewPolicy(ctx, "email", &iam.PolicyArgs{
		Name: pulumi.String("zen-email-send"),
		Policy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"ses:SendEmail"
				],
				"Resource": "%s"
			}]
		}`, emailIdentity.Arn),
	})
	if err != nil {
		return nil, err
	}

	return &DeployOutput{
		EmailName:      pulumi.Sprintf("noreply@%s", emailIdentity.EmailIdentity),
		EmailPolicyArn: emailPolicy.Arn,
	}, nil
}
