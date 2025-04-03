// Copyright (c) 2021 Uber Technologies Inc.
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type fakeOKTokenSource struct{}
type fakeFailingTokenSource struct{}

func (f *fakeOKTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "token"}, nil
}

func (f *fakeFailingTokenSource) Token() (*oauth2.Token, error) { return nil, errors.New("error") }

func TestNewOAuthAuthorizationProviderEmptyConfigFails(t *testing.T) {
	p := NewOAuthAuthorizationProvider(OAuthAuthorizerConfig{})
	_, err := p.GetAuthToken()
	assert.Error(t, err, "empty config will always fail")
}

func TestNewOAuthAuthorizationProviderExtraParams(t *testing.T) {
	p := NewOAuthAuthorizationProvider(OAuthAuthorizerConfig{
		EndpointParams: map[string]string{"test1": "test2"},
	})
	assert.NotEmpty(t, p.config.EndpointParams)
}

func TestNewOAuthAuthorizationProviderFailingTokenSourceResultsInError(t *testing.T) {
	p := &OAuthProvider{tokenSource: &fakeFailingTokenSource{}}
	token, err := p.GetAuthToken()
	assert.Nil(t, token)
	assert.Error(t, err, "error from tokensource will result in error")
}

func TestNewOAuthAuthorizationProviderTokenSourceReturnsToken(t *testing.T) {
	p := &OAuthProvider{tokenSource: &fakeOKTokenSource{}}
	token, err := p.GetAuthToken()
	assert.Equal(t, []byte("token"), token)
	assert.NoError(t, err)
}
