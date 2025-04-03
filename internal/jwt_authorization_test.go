// Copyright (c) 2021 Uber Technologies, Inc.
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
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/internal/common/auth"
	"go.uber.org/cadence/internal/common/util"
)

func TestCorrectTokenCreation(t *testing.T) {
	key, err := os.ReadFile("./common/auth/credentials/keytest")
	assert.NoError(t, err)
	authorizer := NewAdminJwtAuthorizationProvider(key)
	authToken, err := authorizer.GetAuthToken()
	assert.NoError(t, err)

	// Decrypting token, it should be enough to make sure authtoken is not empty, this is one step ahead of that
	publicKeyStr, err := os.ReadFile("./common/auth/credentials/keytest.pub")
	assert.NoError(t, err)
	publicKey, err := util.LoadRSAPublicKey(publicKeyStr)
	assert.NoError(t, err)
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))
	var claims auth.JWTClaims

	_, err = parser.ParseWithClaims(string(authToken), &claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, true, claims.Admin)
	assert.Equal(t, "", claims.Groups)
	assert.Equal(t, int64(60*10), claims.TTL)
}
func TestIncorrectPrivateKeyForTokenCreation(t *testing.T) {
	authorizer := NewAdminJwtAuthorizationProvider([]byte{})
	_, err := authorizer.GetAuthToken()
	assert.EqualError(t, err, "failed to parse PEM block containing the private key")
}
