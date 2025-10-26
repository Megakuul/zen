// package util contains some basic helpers for specific tasks in deployment.
package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// LookupZone checks if there is a route53 zone for the provided domain.
// It traverses each domain segment to check for a zone.
func LookupZone(ctx *pulumi.Context, domain string) (*route53.LookupZoneResult, error) {
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
		zoneName = strings.Join(segments[1:], ".")
	}
	return zone, nil
}
