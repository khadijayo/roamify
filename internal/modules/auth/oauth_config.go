package auth

import (
	"github.com/khadijayo/roamify/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewGoogleOAuthConfig builds the oauth2.Config from the app config.
// Called once in NewHandler so the config is ready when the route fires.
func NewGoogleOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		Endpoint: google.Endpoint,
	}
}
