package deploy

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type proxyInput struct {
	SchedulerDomain pulumi.StringOutput
	ManagerDomain   pulumi.StringOutput
	BucketDomain    pulumi.StringOutput
}

type proxyOutput struct {
	ProxyDomain pulumi.StringOutput
}

func (o *Operator) deployProxy(ctx *pulumi.Context, input *proxyInput) (*proxyOutput, error) {
	viewerCertificate := cloudfront.DistributionViewerCertificateArgs{
		CloudfrontDefaultCertificate: pulumi.BoolPtr(true),
	}
	if o.certificateArn != "" {
		viewerCertificate = cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:      pulumi.String(o.certificateArn),
			MinimumProtocolVersion: pulumi.String("TLSv1.2"),
			SslSupportMethod:       pulumi.String("sni-only"),
		}
	} else if len(o.domains) > 0 {
		sans := []string{}
		if len(o.domains) > 1 {
			sans = o.domains[1:]
		}
		validations := acm.CertificateValidationOptionArray{}
		for _, domain := range o.domains {
			validations = append(validations, acm.CertificateValidationOptionArgs{
				DomainName:       pulumi.String(domain),
				ValidationDomain: pulumi.String(domain),
			})
		}
		cert, err := acm.NewCertificate(ctx, "proxy", &acm.CertificateArgs{
			Region:                  aws.RegionUSEast1,
			KeyAlgorithm:            pulumi.String("RSA_2048"),
			DomainName:              pulumi.String(o.domains[0]),
			SubjectAlternativeNames: pulumi.ToStringArray(sans),
			ValidationMethod:        pulumi.String("DNS"),
			ValidationOptions:       validations,
		})
		if err != nil {
			return nil, err
		}
		validationFqdns := []pulumi.StringOutput{}
		cert.DomainValidationOptions.ApplyT(func(input acm.CertificateDomainValidationOption)) // TODO
		for i, domain := range o.domains {
			zone, err := lookupZone(ctx, domain)
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

	proxy, err := cloudfront.NewDistribution(ctx, "proxy", &cloudfront.DistributionArgs{
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
				CustomOriginConfig: cloudfront.DistributionOriginCustomOriginConfigPtr(&cloudfront.DistributionOriginCustomOriginConfigArgs{
					HttpsPort:            pulumi.Int(443),
					OriginProtocolPolicy: pulumi.String("https-only"),
					OriginSslProtocols:   pulumi.ToStringArray([]string{"TLSv1.2"}),
				}),
				DomainName: input.SchedulerDomain,
			},
			cloudfront.DistributionOriginArgs{
				OriginId: pulumi.String("manager-api"),
				CustomOriginConfig: cloudfront.DistributionOriginCustomOriginConfigPtr(&cloudfront.DistributionOriginCustomOriginConfigArgs{
					HttpsPort:            pulumi.Int(443),
					OriginProtocolPolicy: pulumi.String("https-only"),
					OriginSslProtocols:   pulumi.ToStringArray([]string{"TLSv1.2"}),
				}),
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
				PathPattern:          pulumi.String("/api/scheduler"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD"}),
				Compress:             pulumi.BoolPtr(false),
				TargetOriginId:       pulumi.String("scheduler-api"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        apiCachePolicy.ID(),
			},
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:          pulumi.String("/api/manager"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD"}),
				Compress:             pulumi.BoolPtr(false),
				TargetOriginId:       pulumi.String("manager-api"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        apiCachePolicy.ID(),
			},
		},
		DefaultRootObject: pulumi.String("index.html"),
		IsIpv6Enabled:     pulumi.BoolPtr(true),
		Aliases:           pulumi.ToStringArray(o.domains),
		ViewerCertificate: viewerCertificate,
	})
	if err != nil {
		return nil, err
	}

	if o.certificateArn == "" {
		return &proxyOutput{ProxyDomain: proxy.DomainName}, nil
	}
	// if hosted zone is in route53, records are automatically created.
	for i, domain := range o.domains {
		zone, err := lookupZone(ctx, domain)
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
	return &proxyOutput{ProxyDomain: proxy.DomainName}, nil
}

// lookupZone checks if there is a route53 zone for the provided domain.
// It traverses each domain segment to check for a zone.
func lookupZone(ctx *pulumi.Context, domain string) (*route53.LookupZoneResult, error) {
	var (
		err      error
		zoneName string = domain
		zone     *route53.LookupZoneResult
	)
	for {
		if lZone, lErr := route53.LookupZone(ctx, &route53.LookupZoneArgs{
			Name:        pulumi.StringRef(zoneName),
			PrivateZone: pulumi.BoolRef(false),
		}); lErr != nil {
			err = errors.Join(err, lErr)
		} else {
			zone = lZone
			break
		}
		segments := strings.Split(zoneName, ".")
		if len(segments) < 3 { // 3 segments minimum, the tld is never a hosted zone
			return nil, fmt.Errorf("no route53 hosted zone found for domain '%s': %v", domain, err)
		}
		zoneName = segments[1]
	}
	return zone, nil
}
