package authentication

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/megakuul/zen/internal/auth"
	"github.com/megakuul/zen/internal/model/email"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/manager/authentication"
)

const (
	refreshTokenName = "refresh_token"
	accessTokenTTL   = 15 * time.Minute // 15 minutes
	refreshTokenTTL  = 720 * time.Hour  // 30 days
)

type Service struct {
	logger     *slog.Logger
	tokenCtrl  *token.Controller
	authCtrl   *auth.Controller
	emailModel *email.Model
}

func New(logger *slog.Logger, token *token.Controller, auth *auth.Controller, email *email.Model) *Service {
	return &Service{
		logger:     logger,
		tokenCtrl:  token,
		authCtrl:   auth,
		emailModel: email,
	}
}

func (s *Service) Login(ctx context.Context, r *connect.Request[authentication.LoginRequest]) (*connect.Response[authentication.LoginResponse], error) {
	refreshCookie := findRefreshCookie(r.Header())
	if refreshCookie != nil {
		claims, err := s.tokenCtrl.Verify(ctx, refreshCookie.Value)
		if err != nil {
			return nil, err
		} else if !claims.Refresh {
			return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("invalid token; expected refresh_token"))
		}
		token, err := s.tokenCtrl.Issue(ctx, claims.Subject, r.Msg.Verifier.Email, false, time.Now().Add(accessTokenTTL))
		if err != nil {
			return nil, err
		}
		return &connect.Response[authentication.LoginResponse]{
			Msg: &authentication.LoginResponse{
				Token: token,
			},
		}, nil
	}
	if r.Msg.Verifier.Email == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("you are not logged in"))
	}

	registration, found, err := s.emailModel.GetRegistration(ctx, r.Msg.Verifier.Email)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("email is not registered"))
	}

	verified, err := s.authCtrl.Authenticate(ctx, r.Msg.Verifier)
	if err != nil {
		return nil, err
	} else if !verified {
		return connect.NewResponse(&authentication.LoginResponse{}), nil
	}

	accessToken, err := s.tokenCtrl.Issue(ctx, registration.User, r.Msg.Verifier.Email, false, time.Now().Add(accessTokenTTL))
	if err != nil {
		return nil, err
	}
	resp := connect.NewResponse(&authentication.LoginResponse{
		Token: accessToken,
	})
	if r.Msg.AutoRefresh {
		refreshToken, err := s.tokenCtrl.Issue(ctx, registration.User, r.Msg.Verifier.Email, true, time.Now().Add(refreshTokenTTL))
		if err != nil {
			return nil, err
		}
		cookie := http.Cookie{
			Name:     refreshTokenName,
			Expires:  time.Now().Add(refreshTokenTTL),
			Path:     "/api/", // TODO make this more specific
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Value:    refreshToken,
		}
		resp.Header().Add("Set-Cookie", cookie.String())
	}
	return resp, nil
}

func (s *Service) Logout(ctx context.Context, r *connect.Request[authentication.LogoutRequest]) (*connect.Response[authentication.LogoutResponse], error) {
	resp := connect.NewResponse(&authentication.LogoutResponse{})
	refreshCookie := findRefreshCookie(r.Header())
	if refreshCookie == nil {
		return resp, nil
	}
	refreshCookie.Expires = time.Now().Add(-8760 * time.Hour) // expire cookie
	refreshCookie.MaxAge = -1
	resp.Header().Add("Set-Cookie", refreshCookie.String())
	return resp, nil
}

func findRefreshCookie(headers http.Header) *http.Cookie {
	cookieHeader := headers.Get("Cookie")
	if cookieHeader != "" {
		cookies, err := http.ParseCookie(cookieHeader)
		if err != nil {
			return nil
		}
		for _, cookie := range cookies {
			if cookie.Name == refreshTokenName {
				return cookie
			}
		}
	}
	return nil
}
