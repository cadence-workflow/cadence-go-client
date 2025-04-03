// Copyright (c) 2017-2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"fmt"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"go.uber.org/cadence/internal/common/auth"
)

// OAuthAuthorizerConfig allows to configure external OAuth provider
// This is machine to machine / service to service 2-legged OAuth2 flow
type OAuthAuthorizerConfig struct {
	// ClientID to be used for acquiring token
	ClientID string `yaml:"clientID"`

	// ClientSecret to be used for acquiring token
	ClientSecret string `yaml:"clientSecret"`

	// TokenURL is the endpoint used to get token from provider
	TokenURL string `yaml:"tokenURL"`

	// Scope specifies optional requested permissions
	Scopes []string `yaml:"scopes"`

	// EndpointParams specifies additional parameters for requests to the token endpoint.
	// This needs to be provided for some OAuth providers
	EndpointParams map[string]string `yaml:"endpointParams"`
}

var _ auth.AuthorizationProvider = (*OAuthProvider)(nil)

type OAuthProvider struct {
	tokenSource oauth2.TokenSource
	config      clientcredentials.Config
}

func NewOAuthAuthorizationProvider(config OAuthAuthorizerConfig) *OAuthProvider {
	oauthConfig := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.TokenURL,
		Scopes:       config.Scopes,
	}

	if config.EndpointParams != nil {
		v := url.Values{}
		for name, value := range config.EndpointParams {
			v.Set(name, value)
		}
		oauthConfig.EndpointParams = v
	}

	return &OAuthProvider{
		tokenSource: oauthConfig.TokenSource(context.Background()),
		config:      oauthConfig,
	}
}

func (o *OAuthProvider) GetAuthToken() ([]byte, error) {
	token, err := o.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token source: %w", err)
	}

	return []byte(token.AccessToken), nil
}
