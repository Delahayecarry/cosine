package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"cosine/config"
)

// CodeState represents OAuth callback parameters
type CodeState struct {
	Code  string `json:"code" form:"code"`
	State string `json:"state" form:"state"`
}

func (c *CodeState) Check() error {
	if c.Code == "" {
		return errors.New("code is empty")
	}
	return nil
}

// LinuxDoUserInfo represents user information from linux.do
type LinuxDoUserInfo struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	TrustLevel int    `json:"trust_level"`
	Silenced   bool   `json:"silenced"`
}

// linux.do OAuth2 Endpoint & user info URL
var (
	linuxDoEndpoint = oauth2.Endpoint{
		AuthURL:   "https://connect.linux.do/oauth2/authorize",
		TokenURL:  "https://connect.linux.do/oauth2/token",
		AuthStyle: oauth2.AuthStyleInHeader,
	}

	linuxDoUserURL = "https://connect.linux.do/api/user"
)

func linuxDoAuthConf(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.LinuxDo.ClientID,
		ClientSecret: cfg.LinuxDo.ClientSecret,
		RedirectURL:  cfg.LinuxDo.BackendBaseURL + "/api/auth/linuxdo/callback",
		Endpoint:     linuxDoEndpoint,
	}
}

// LinuxDoAuthCodeURL generates authorization URL for frontend
func LinuxDoAuthCodeURL(cfg *config.Config, state string) string {
	authConf := linuxDoAuthConf(cfg)
	return authConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange code for token
func linuxDoToken(ctx context.Context, cfg *config.Config, code string) (*oauth2.Token, error) {
	authConf := linuxDoAuthConf(cfg)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	return authConf.Exchange(ctx, code)
}

// Get user info with access_token
func getLinuxDoUserInfo(ctx context.Context, accessToken string) (*LinuxDoUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", linuxDoUserURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	userInfo := new(LinuxDoUserInfo)
	err = json.NewDecoder(resp.Body).Decode(userInfo)
	return userInfo, err
}

// GetLinuxDoUser is the main function to get user info from code
func GetLinuxDoUser(ctx context.Context, cfg *config.Config, cs *CodeState) (*LinuxDoUserInfo, error) {
	if err := cs.Check(); err != nil {
		return nil, err
	}

	token, err := linuxDoToken(ctx, cfg, cs.Code)
	if err != nil {
		return nil, err
	}

	return getLinuxDoUserInfo(ctx, token.AccessToken)
}
