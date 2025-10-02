package zen

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type proxyInput struct {
	SchedulerDomain pulumi.StringOutput
	ManagerDomain   pulumi.StringOutput
	CertificateArn  pulumi.StringOutput
	BucketArn       pulumi.StringOutput
}

type proxyOutput struct {
	ProxyDomain pulumi.StringOutput
}

func (o *Operator) deployProxy(ctx *pulumi.Context, input *proxyInput) (*proxyOutput, error) {
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
		Name: pulumi.String("zen-proxy-oac"),
		OriginAccessControlOriginType: pulumi.String("s3"),
		SigningBehavior: pulumi.String("always"),
		SigningProtocol: pulumi.String("sigv4"),
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
				OriginAccessControlId: oac.Name,
			},
			cloudfront.DistributionOriginArgs{
				OriginId:              pulumi.String("leaderboard"),
				OriginAccessControlId: oac.Name,
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
			CachePolicyId:        webCachePolicy.Name,
		},
		OrderedCacheBehaviors: cloudfront.DistributionOrderedCacheBehaviorArray{
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:          pulumi.String("/leaderboard"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
				CachedMethods:        pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
				Compress:             pulumi.BoolPtr(true),
				TargetOriginId:       pulumi.String("leaderboard"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        leaderboardCachePolicy.Name,
			},
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:          pulumi.String("/api/scheduler"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:        pulumi.ToStringArray([]string{}),
				Compress:             pulumi.BoolPtr(false),
				TargetOriginId:       pulumi.String("scheduler-api"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        apiCachePolicy.Name,
			},
			cloudfront.DistributionOrderedCacheBehaviorArgs{
				PathPattern:          pulumi.String("/api/manager"),
				AllowedMethods:       pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS", "POST"}),
				CachedMethods:        pulumi.ToStringArray([]string{}),
				Compress:             pulumi.BoolPtr(false),
				TargetOriginId:       pulumi.String("manager-api"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				CachePolicyId:        apiCachePolicy.Name,
			},
		},
		DefaultRootObject: pulumi.String("index.html"),
		IsIpv6Enabled:     pulumi.BoolPtr(true),
		Aliases:           pulumi.ToStringArray(o.domains),
		ViewerCertificate: cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn:            input.CertificateArn,
			CloudfrontDefaultCertificate: pulumi.BoolPtr(true),
		},
	})
	if err != nil {
		return nil, err
	}
	return &proxyOutput{
		ProxyDomain: proxy.DomainName,
	}, nil
}
