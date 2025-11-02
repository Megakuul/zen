package proxy

import (
	"fmt"
	"strings"

	"github.com/megakuul/zen/internal/deploy/util"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployInput struct {
	Domains         []string
	AutoDns         bool
	CertificateArn  string
	WebRouter       pulumi.String
	SchedulerDomain pulumi.StringOutput
	ManagerDomain   pulumi.StringOutput
	BucketDomain    pulumi.StringOutput
}

type DeployOutput struct {
	ProxyDomain pulumi.StringOutput
}

func Deploy(ctx *pulumi.Context, input *DeployInput) (*DeployOutput, error) {
	viewerCertificate := cloudfront.DistributionViewerCertificateArgs{
		CloudfrontDefaultCertificate: pulumi.BoolPtr(true),
	}
	if !input.AutoDns {
		viewerCertificate = cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:      pulumi.String(input.CertificateArn),
			MinimumProtocolVersion: pulumi.String("TLSv1.2"),
			SslSupportMethod:       pulumi.String("sni-only"),
		}
	} else {
		sans := []string{}
		if len(input.Domains) > 1 {
			sans = input.Domains[1:]
		}
		validations := acm.CertificateValidationOptionArray{}
		for _, domain := range input.Domains {
			validations = append(validations, acm.CertificateValidationOptionArgs{
				DomainName:       pulumi.String(domain),
				ValidationDomain: pulumi.String(domain),
			})
		}
		cert, err := acm.NewCertificate(ctx, "proxy", &acm.CertificateArgs{
			Region:                  aws.RegionUSEast1,
			KeyAlgorithm:            pulumi.String("RSA_2048"),
			DomainName:              pulumi.String(input.Domains[0]),
			SubjectAlternativeNames: pulumi.ToStringArray(sans),
			ValidationMethod:        pulumi.String("DNS"),
			ValidationOptions:       validations,
		})
		if err != nil {
			return nil, err
		}
		validationFqdns := []pulumi.StringOutput{}
		for i, domain := range input.Domains {
			zone, err := util.LookupZone(ctx, domain)
			if err != nil {
				return nil, err
			}
			validationRecord, err := route53.NewRecord(ctx, fmt.Sprintf("proxy-validation-%d", i), &route53.RecordArgs{
				ZoneId: pulumi.String(zone.Id),
				Name:   cert.DomainValidationOptions.Index(pulumi.Int(i)).ResourceRecordName().Elem(),
				Type:   cert.DomainValidationOptions.Index(pulumi.Int(i)).ResourceRecordType().Elem(),
				Ttl:    pulumi.Int(60),
				Records: pulumi.StringArray{
					cert.DomainValidationOptions.Index(pulumi.Int(i)).ResourceRecordValue().Elem(),
				},
			})
			if err != nil {
				return nil, err
			}
			validationFqdns = append(validationFqdns, validationRecord.Fqdn)
		}
		certValidation, err := acm.NewCertificateValidation(ctx, "proxy", &acm.CertificateValidationArgs{
			CertificateArn:        cert.Arn,
			Region:                aws.RegionUSEast1,
			ValidationRecordFqdns: pulumi.ToStringArrayOutput(validationFqdns),
		})
		if err != nil {
			return nil, err
		}
		viewerCertificate = cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:      certValidation.CertificateArn,
			MinimumProtocolVersion: pulumi.String("TLSv1.2"),
			SslSupportMethod:       pulumi.String("sni-only"),
		}
	}

	webCachePolicy, err := cloudfront.NewCachePolicy(ctx, "proxy-web", &cloudfront.CachePolicyArgs{
		Name:       pulumi.String("zen-proxy-web"),
		Comment:    pulumi.String("full cache policy to serve the static website assets"),
		MinTtl:     pulumi.IntPtr(1),
		MaxTtl:     pulumi.IntPtr(31536000), // 1 year
		DefaultTtl: pulumi.IntPtr(86400),    // 1 day
		ParametersInCacheKeyAndForwardedToOrigin: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginArgs{
			EnableAcceptEncodingBrotli: pulumi.BoolPtr(true),
			EnableAcceptEncodingGzip:   pulumi.BoolPtr(true),
			CookiesConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginCookiesConfigArgs{
				CookieBehavior: pulumi.String("none"),
			},
			HeadersConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginHeadersConfigArgs{
				HeaderBehavior: pulumi.String("none"),
			},
			QueryStringsConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginQueryStringsConfigArgs{
				QueryStringBehavior: pulumi.String("none"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	leaderboardCachePolicy, err := cloudfront.NewCachePolicy(ctx, "proxy-leaderboard", &cloudfront.CachePolicyArgs{
		Name:       pulumi.String("zen-proxy-leaderboard"),
		Comment:    pulumi.String("short-lived cache policy to serve the rolling leaderboard assets"),
		MinTtl:     pulumi.IntPtr(1),
		MaxTtl:     pulumi.IntPtr(300), // 5 minutes
		DefaultTtl: pulumi.IntPtr(180), // 3 minutes
		ParametersInCacheKeyAndForwardedToOrigin: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginArgs{
			EnableAcceptEncodingBrotli: pulumi.BoolPtr(true),
			EnableAcceptEncodingGzip:   pulumi.BoolPtr(true),
			CookiesConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginCookiesConfigArgs{
				CookieBehavior: pulumi.String("none"),
			},
			HeadersConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginHeadersConfigArgs{
				HeaderBehavior: pulumi.String("none"),
			},
			QueryStringsConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginQueryStringsConfigArgs{
				QueryStringBehavior: pulumi.String("none"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	apiCachePolicy, err := cloudfront.NewCachePolicy(ctx, "proxy-api", &cloudfront.CachePolicyArgs{
		Name:       pulumi.String("zen-proxy-api"),
		Comment:    pulumi.String("disabled cache policy to serve the dynamic api"),
		MinTtl:     pulumi.IntPtr(0),
		MaxTtl:     pulumi.IntPtr(0),
		DefaultTtl: pulumi.IntPtr(0),
		ParametersInCacheKeyAndForwardedToOrigin: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginArgs{
			EnableAcceptEncodingBrotli: pulumi.BoolPtr(false),
			EnableAcceptEncodingGzip:   pulumi.BoolPtr(false),
			CookiesConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginCookiesConfigArgs{
				CookieBehavior: pulumi.String("none"),
			},
			HeadersConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginHeadersConfigArgs{
				HeaderBehavior: pulumi.String("none"),
			},
			QueryStringsConfig: cloudfront.CachePolicyParametersInCacheKeyAndForwardedToOriginQueryStringsConfigArgs{
				QueryStringBehavior: pulumi.String("none"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	apiOriginPolicy, err := cloudfront.NewOriginRequestPolicy(ctx, "proxy-api", &cloudfront.OriginRequestPolicyArgs{
		Name:    pulumi.String("zen-proxy-api"),
		Comment: pulumi.String("allows all data (headers, queries, cookies) to be forwarded to the api"),
		CookiesConfig: cloudfront.OriginRequestPolicyCookiesConfigArgs{
			CookieBehavior: pulumi.String("all"),
		},
		HeadersConfig: cloudfront.OriginRequestPolicyHeadersConfigArgs{
			HeaderBehavior: pulumi.String("allViewer"),
		},
		QueryStringsConfig: cloudfront.OriginRequestPolicyQueryStringsConfigArgs{
			QueryStringBehavior: pulumi.String("all"),
		},
	})
	if err != nil {
		return nil, err
	}

	oac, err := cloudfront.NewOriginAccessControl(ctx, "proxy", &cloudfront.OriginAccessControlArgs{
		Name:                          pulumi.String("zen-proxy-oac"),
		OriginAccessControlOriginType: pulumi.String("s3"),
		SigningBehavior:               pulumi.String("always"),
		SigningProtocol:               pulumi.String("sigv4"),
	})
	if err != nil {
		return nil, err
	}

	router, err := cloudfront.NewFunction(ctx, "proxy", &cloudfront.FunctionArgs{
		Name:    pulumi.String("zen-proxy-router"),
		Comment: pulumi.String("Sveltekit router used to correctly route pages"),
		Runtime: pulumi.String("cloudfront-js-2.0"),
		Code:    input.WebRouter,
	})
	if err != nil {
		return nil, err
	}

	proxy, err := cloudfront.NewDistribution(ctx, "proxy", &cloudfront.DistributionArgs{
		Enabled:    pulumi.Bool(true),
		PriceClass: pulumi.String("PriceClass_All"),
		Origins: cloudfront.DistributionOriginArray{
			cloudfront.DistributionOriginArgs{
				OriginId:              pulumi.String("web"),
				OriginPath:            pulumi.String("/web"),
				OriginAccessControlId: oac.ID(),
				DomainName:            input.BucketDomain,
			},
			cloudfront.DistributionOriginArgs{
				OriginId:              pulumi.String("leaderboard"),
				OriginAccessControlId: oac.ID(),
				DomainName:            input.BucketDomain,
			},
			cloudfront.DistributionOriginArgs{
				OriginId: pulumi.String("scheduler-api"),
				CustomOriginConfig: &cloudfront.DistributionOriginCustomOriginConfigArgs{
					HttpsPort:            pulumi.Int(443),
					OriginProtocolPolicy: pulumi.String("https-only"),
					OriginSslProtocols:   pulumi.ToStringArray([]string{"TLSv1.2"}),
				},
				DomainName: input.SchedulerDomain,
			},
			cloudfront.DistributionOriginArgs{
				OriginId: pulumi.String("manager-api"),
				CustomOriginConfig: &cloudfront.DistributionOriginCustomOriginConfigArgs{
					HttpsPort:            pulumi.Int(443),
					OriginProtocolPolicy: pulumi.String("https-only"),
					OriginSslProtocols:   pulumi.ToStringArray([]string{"TLSv1.2"}),
				},
				DomainName: input.ManagerDomain,
			},
		},
		DefaultCacheBehavior: cloudfront.DistributionDefaultCacheBehaviorArgs{
			AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
			CachedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
			Compress:             pulumi.BoolPtr(true),
			TargetOriginId:       pulumi.String("web"),
			ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
			CachePolicyId:        webCachePolicy.ID(),
			FunctionAssociations: cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArray{
				cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArgs{
					EventType:   pulumi.String("viewer-request"),
					FunctionArn: router.Arn,
				},
			},
		},
		OrderedCacheBehaviors: cloudfront.DistributionOrderedCacheBehaviorArray{
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:          pulumi.String("/leaderboard"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
				CachedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
				Compress:             pulumi.BoolPtr(true),
				TargetOriginId:       pulumi.String("leaderboard"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        leaderboardCachePolicy.ID(),
			},
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:           pulumi.String("/api/scheduler"),
				AllowedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:         pulumi.ToStringArray([]string{"GET", "HEAD"}),
				Compress:              pulumi.BoolPtr(false),
				TargetOriginId:        pulumi.String("scheduler-api"),
				ViewerProtocolPolicy:  pulumi.String("redirect-to-https"),
				CachePolicyId:         apiCachePolicy.ID(),
				OriginRequestPolicyId: apiOriginPolicy.ID(),
			},
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:           pulumi.String("/api/manager"),
				AllowedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:         pulumi.ToStringArray([]string{"GET", "HEAD"}),
				Compress:              pulumi.BoolPtr(false),
				TargetOriginId:        pulumi.String("manager-api"),
				ViewerProtocolPolicy:  pulumi.String("redirect-to-https"),
				CachePolicyId:         apiCachePolicy.ID(),
				OriginRequestPolicyId: apiOriginPolicy.ID(),
			},
		},
		DefaultRootObject: pulumi.String("index.html"),
		IsIpv6Enabled:     pulumi.BoolPtr(true),
		Aliases:           pulumi.ToStringArray(input.Domains),
		ViewerCertificate: viewerCertificate,
		Restrictions: cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
				Locations: pulumi.ToStringArray([]string{}),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if !input.AutoDns {
		for _, domain := range input.Domains {
			ctx.Export(
				fmt.Sprintf("REQUIRED_CNAME_%s", strings.ToUpper(domain)),
				pulumi.Sprintf("CNAME %s %s", pulumi.String(domain), proxy.DomainName),
			)
		}
		return &DeployOutput{ProxyDomain: proxy.DomainName}, nil
	}
	// if hosted zone is in route53, records are automatically created.
	for i, domain := range input.Domains {
		zone, err := util.LookupZone(ctx, domain)
		if err != nil {
			return nil, err
		}
		_, err = route53.NewRecord(ctx, fmt.Sprintf("proxy-traffic-a-%d", i), &route53.RecordArgs{
			ZoneId: pulumi.String(zone.Id),
			Name:   pulumi.String(domain),
			Type:   pulumi.String("A"),
			Aliases: route53.RecordAliasArray{
				route53.RecordAliasArgs{
					ZoneId: pulumi.String("Z2FDTNDATAQYW2"), // cloudfront hosted zone id
					Name:   proxy.DomainName,
				},
			},
		})
		if err != nil {
			return nil, err
		}

		_, err = route53.NewRecord(ctx, fmt.Sprintf("proxy-traffic-aaaa-%d", i), &route53.RecordArgs{
			ZoneId: pulumi.String(zone.Id),
			Name:   pulumi.String(domain),
			Type:   pulumi.String("AAAA"),
			Aliases: route53.RecordAliasArray{
				route53.RecordAliasArgs{
					ZoneId: pulumi.String("Z2FDTNDATAQYW2"), // cloudfront hosted zone id
					Name:   proxy.DomainName,
				},
			},
		})
		if err != nil {
			return nil, err
		}
	}
	return &DeployOutput{ProxyDomain: proxy.DomainName}, nil
}
