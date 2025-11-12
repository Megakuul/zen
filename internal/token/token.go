// package token provides a tokenCtrl used to issue and verify access tokens.
package token

import (
	"context"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
)

type Controller struct {
	Issuer    string
	KmsConfig *jwtkms.Config
}

func New(issuer string, kms *jwtkms.Config) *Controller {
	return &Controller{
		Issuer:    issuer,
		KmsConfig: kms,
	}
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Email   string `json:"email,omitempty"`
	Refresh bool   `json:"refresh,omitempty"`
}

func (c *Controller) Issue(ctx context.Context, subject, email string, refresh bool, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwtkms.SigningMethodECDSA256, &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Audience:  jwt.ClaimStrings{c.Issuer}, // rp and resource server are the same entity, so aud == iss
			Issuer:    c.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   subject,
		},
		Email:   email,
		Refresh: refresh,
	})
	signedToken, err := token.SignedString(c.KmsConfig.WithContext(ctx))
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	return signedToken, nil
}

func (c *Controller) Verify(ctx context.Context, token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return c.KmsConfig, nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	if !slices.Contains(claims.Audience, c.Issuer) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("token was not issued for this audience"))
	}
	return claims, nil
}
