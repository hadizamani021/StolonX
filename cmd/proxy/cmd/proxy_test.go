// Copyright 2015 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package cmd

import (
	"testing"

	"github.com/sorintlab/stolon/internal/cluster"
)

func TestProxyMasterEndpoint(t *testing.T) {
	db := &cluster.DB{
		UID: "db1",
		Spec: &cluster.DBSpec{
			KeeperUID: "k1",
		},
		Status: cluster.DBStatus{
			ListenAddress:         "10.0.0.1",
			Port:                  "5432",
			InternalListenAddress: "172.16.0.1",
			InternalPort:          "6432",
		},
	}
	keeperEU := &cluster.Keeper{
		UID: "k1",
		Status: cluster.KeeperStatus{
			Region: "eu-west",
		},
	}

	tests := []struct {
		name          string
		proxyRegion   string
		db            *cluster.DB
		masterKeeper  *cluster.Keeper
		wantHost      string
		wantPort      string
		wantInternal  bool
	}{
		{
			name:         "nil db",
			db:           nil,
			wantHost:     "",
			wantPort:     "",
			wantInternal: false,
		},
		{
			name:         "external when proxy region unset",
			proxyRegion:  "",
			db:           db,
			masterKeeper: keeperEU,
			wantHost:     "10.0.0.1",
			wantPort:     "5432",
			wantInternal: false,
		},
		{
			name:         "external when keeper region empty",
			proxyRegion:  "eu-west",
			db:           db,
			masterKeeper: &cluster.Keeper{UID: "k1", Status: cluster.KeeperStatus{}},
			wantHost:     "10.0.0.1",
			wantPort:     "5432",
			wantInternal: false,
		},
		{
			name:         "external when regions mismatch",
			proxyRegion:  "us-east",
			db:           db,
			masterKeeper: keeperEU,
			wantHost:     "10.0.0.1",
			wantPort:     "5432",
			wantInternal: false,
		},
		{
			name:         "external when internal address unset",
			proxyRegion:  "eu-west",
			db: &cluster.DB{
				Spec:   db.Spec,
				Status: cluster.DBStatus{ListenAddress: "10.0.0.1", Port: "5432"},
			},
			masterKeeper: keeperEU,
			wantHost:     "10.0.0.1",
			wantPort:     "5432",
			wantInternal: false,
		},
		{
			name:        "internal when regions match and internal set",
			proxyRegion: "eu-west",
			db:          db,
			masterKeeper: &cluster.Keeper{
				UID:    "k1",
				Status: cluster.KeeperStatus{Region: "eu-west"},
			},
			wantHost:     "172.16.0.1",
			wantPort:     "6432",
			wantInternal: true,
		},
		{
			name:        "internal port falls back to external port",
			proxyRegion: "eu-west",
			db: &cluster.DB{
				Spec: db.Spec,
				Status: cluster.DBStatus{
					ListenAddress:         "10.0.0.1",
					Port:                  "5432",
					InternalListenAddress: "172.16.0.1",
				},
			},
			masterKeeper: keeperEU,
			wantHost:     "172.16.0.1",
			wantPort:     "5432",
			wantInternal: true,
		},
		{
			name:         "nil master keeper uses external",
			proxyRegion:  "eu-west",
			db:           db,
			masterKeeper: nil,
			wantHost:     "10.0.0.1",
			wantPort:     "5432",
			wantInternal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, internal := proxyMasterEndpoint(tt.proxyRegion, tt.db, tt.masterKeeper)
			if host != tt.wantHost || port != tt.wantPort || internal != tt.wantInternal {
				t.Fatalf("proxyMasterEndpoint() = (%q, %q, %v), want (%q, %q, %v)",
					host, port, internal, tt.wantHost, tt.wantPort, tt.wantInternal)
			}
		})
	}
}
