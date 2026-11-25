// Copyright (c) 2017 Uber Technologies, Inc.
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

// below are the metadata which will be embedded as part
// of headers in every rpc call made by this client to
// cadence server.

// Update to the metadata below is typically done
// by the cadence team as part of a major feature or
// behavior change

// LibraryVersion is a historical way to report the "library release" version,
// prior to go modules providing a far more consistent way to do so.
// It is sent in a header on every request.
//
// deprecated: This cannot accurately report pre-release version information,
// and it is easy for it to drift from the release version (especially if an old
// commit is tagged, to avoid branching, as this behaves poorly with go modules).
//
// Ideally it would be replaced by runtime/debug.ReadBuildInfo()... but that is
// not guaranteed to exist, and even if this is a fallback it still needs to be
// maintained and may be inherently out of date at any time.
//
// Due to all of this unreliability, this should be used as strictly informational
// metadata, e.g. for caller version monitoring, never behavioral (use
// FeatureVersion or feature flags instead).
const LibraryVersion = "1.3.0"

// FeatureVersion is a semver that informs the server of what high-level behaviors
// this client supports.
// This is sent in a header on every request.
//
// If you wish to tie new behavior to a client release, rather than a feature
// flag, increment the major/minor/patch as seems appropriate here.
//
// It can in principle be inferred from the release version in nearly all
// "normal" scenarios, but release versions are not always available
// (debug.BuildInfo is not guaranteed) and non-released versions do not have any
// way to safely infer behavior.  So it is a hard-coded string instead.
const FeatureVersion = "1.7.0"
