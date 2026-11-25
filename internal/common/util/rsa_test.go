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

package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestKeyFile(t *testing.T, path string) []byte {
	t.Helper()
	key, err := os.ReadFile(path)
	require.NoError(t, err)
	return key
}

func TestLoadRSAPublicKey(t *testing.T) {
	publicKeyPEM := loadTestKeyFile(t, "./../auth/credentials/keytest.pub")

	t.Run("valid public key", func(t *testing.T) {
		key, err := LoadRSAPublicKey(publicKeyPEM)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, 2048, key.Size()*8)
	})

	t.Run("invalid public key", func(t *testing.T) {
		_, err := LoadRSAPublicKey([]byte("invalid PEM data"))
		assert.ErrorContains(t, err, "failed to parse PEM block containing the public key")
	})

	t.Run("incorrect PEM type for public key", func(t *testing.T) {
		privateKeyPEM := loadTestKeyFile(t, "./../auth/credentials/keytest")
		_, err := LoadRSAPublicKey(privateKeyPEM)
		assert.ErrorContains(t, err, "failed to parse PEM block containing the public key")
	})
}

func TestLoadRSAPrivateKey(t *testing.T) {
	privateKeyPEM := loadTestKeyFile(t, "./../auth/credentials/keytest")

	t.Run("valid private key", func(t *testing.T) {
		key, err := LoadRSAPrivateKey(privateKeyPEM)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, 2048, key.Size()*8)
	})

	t.Run("invalid private key", func(t *testing.T) {
		_, err := LoadRSAPrivateKey([]byte("invalid PEM data"))
		assert.ErrorContains(t, err, "failed to parse PEM block containing the private key")
	})

	t.Run("incorrect PEM type for private key", func(t *testing.T) {
		publicKeyPEM := loadTestKeyFile(t, "./../auth/credentials/keytest.pub")
		_, err := LoadRSAPrivateKey(publicKeyPEM)
		assert.ErrorContains(t, err, "failed to parse PEM block containing the private key")
	})
}
