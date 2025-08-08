// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testdata

import (
	"time"

	gogo "github.com/gogo/protobuf/types"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

const (
	DomainID          = "DomainID"
	DomainName        = "DomainName"
	DomainDescription = "DomainDescription"
	DomainOwnerEmail  = "DomainOwnerEmail"
	DomainDataKey     = "DomainDataKey"
	DomainDataValue   = "DomainDataValue"

	ClusterName1 = "ClusterName1"
	ClusterName2 = "ClusterName2"

	BadBinaryReason   = "BadBinaryReason"
	BadBinaryOperator = "BadBinaryOperator"

	HistoryArchivalURI    = "HistoryArchivalURI"
	VisibilityArchivalURI = "VisibilityArchivalURI"

	DeleteBadBinary = "DeleteBadBinary"

	FailoverVersion1 = 301
	FailoverVersion2 = 302
)

var (
	DomainRetention = gogo.DurationProto(3 * time.Hour * 24)

	BadBinaryInfo = apiv1.BadBinaryInfo{
		Reason:      BadBinaryReason,
		Operator:    BadBinaryOperator,
		CreatedTime: Timestamp1,
	}
	BadBinaryInfoMap = map[string]*apiv1.BadBinaryInfo{
		"BadBinary1": &BadBinaryInfo,
	}
	BadBinaries = apiv1.BadBinaries{
		Binaries: BadBinaryInfoMap,
	}
	DomainData = map[string]string{DomainDataKey: DomainDataValue}
	Domain     = apiv1.Domain{
		Id:                               DomainID,
		Name:                             DomainName,
		Status:                           DomainStatus,
		Description:                      DomainDescription,
		OwnerEmail:                       DomainOwnerEmail,
		Data:                             DomainData,
		WorkflowExecutionRetentionPeriod: DomainRetention,
		BadBinaries:                      &BadBinaries,
		HistoryArchivalStatus:            ArchivalStatus,
		HistoryArchivalUri:               HistoryArchivalURI,
		VisibilityArchivalStatus:         ArchivalStatus,
		VisibilityArchivalUri:            VisibilityArchivalURI,
		ActiveClusterName:                ClusterName1,
		Clusters:                         ClusterReplicationConfigurationArray,
		ActiveClusters:                   ActiveClusters,
	}
	ClusterReplicationConfiguration = apiv1.ClusterReplicationConfiguration{
		ClusterName: ClusterName1,
	}
	ClusterReplicationConfigurationArray = []*apiv1.ClusterReplicationConfiguration{
		&ClusterReplicationConfiguration,
	}
	ActiveClustersByRegion = map[string]string{
		"Region1": ClusterName1,
		"Region2": ClusterName2,
	}
	ActiveClusters = &apiv1.ActiveClusters{
		RegionToCluster: map[string]*apiv1.ActiveClusterInfo{
			"Region1": &apiv1.ActiveClusterInfo{
				ActiveClusterName: ClusterName1,
				FailoverVersion:   0,
			},
			"Region2": &apiv1.ActiveClusterInfo{
				ActiveClusterName: ClusterName2,
				FailoverVersion:   0,
			},
		},
	}
)
