// Copyright (C) 2015  The GoHBase Authors.  All rights reserved.
// This file is part of GoHBase.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package gohbase

import (
	"reflect"
	"testing"

	"github.com/tsuna/gohbase/region"
	"github.com/tsuna/gohbase/regioninfo"
)

func TestMetaCache(t *testing.T) {
	client := NewClient("~invalid.quorum~") // We shouldn't connect to ZK.
	reg := client.getRegion([]byte("test"), []byte("theKey"))
	if reg != nil {
		t.Errorf("Found region %#v even though the cache was empty?!", reg)
	}

	// Inject an entry in the cache.  This entry covers the entire key range.
	wholeTable := &regioninfo.Info{
		Table:      []byte("test"),
		RegionName: []byte("test,,1234567890042.56f833d5569a27c7a43fbf547b4924a4."),
		StopKey:    []byte(""),
	}
	regClient := &region.Client{}
	client.addRegionToCache(wholeTable, regClient)

	reg = client.getRegion([]byte("test"), []byte("theKey"))
	if !reflect.DeepEqual(reg, wholeTable) {
		t.Errorf("Found region %#v but expected %#v", reg, wholeTable)
	}
	reg = client.getRegion([]byte("test"), []byte("")) // edge case.
	if !reflect.DeepEqual(reg, wholeTable) {
		t.Errorf("Found region %#v but expected %#v", reg, wholeTable)
	}

	// Clear our client.
	client = NewClient("~invalid.quorum~")

	// Inject 3 entries in the cache.
	region1 := &regioninfo.Info{
		Table:      []byte("test"),
		RegionName: []byte("test,,1234567890042.56f833d5569a27c7a43fbf547b4924a4."),
		StopKey:    []byte("foo"),
	}
	client.addRegionToCache(region1, regClient)

	region2 := &regioninfo.Info{
		Table:      []byte("test"),
		RegionName: []byte("test,foo,1234567890042.56f833d5569a27c7a43fbf547b4924a4."),
		StopKey:    []byte("gohbase"),
	}
	client.addRegionToCache(region2, regClient)

	region3 := &regioninfo.Info{
		Table:      []byte("test"),
		RegionName: []byte("test,gohbase,1234567890042.56f833d5569a27c7a43fbf547b4924a4."),
		StopKey:    []byte(""),
	}
	client.addRegionToCache(region3, regClient)

	testcases := []struct {
		key string
		reg *regioninfo.Info
	}{
		{key: "theKey", reg: region3},
		{key: "", reg: region1},
		{key: "bar", reg: region1},
		{key: "fon\xFF", reg: region1},
		{key: "foo", reg: region2},
		{key: "foo\x00", reg: region2},
		{key: "gohbase", reg: region3},
	}
	for i, testcase := range testcases {
		reg = client.getRegion([]byte("test"), []byte(testcase.key))
		if !reflect.DeepEqual(reg, testcase.reg) {
			t.Errorf("[#%d] Found region %#v but expected %#v", i, reg, testcase.reg)
		}
	}

	// Change the last region (maybe it got split).
	region3 = &regioninfo.Info{
		Table:      []byte("test"),
		RegionName: []byte("test,gohbase,1234567890042.56f833d5569a27c7a43fbf547b4924a4."),
		StopKey:    []byte("zab"),
	}
	client.addRegionToCache(region3, regClient)

	reg = client.getRegion([]byte("test"), []byte("theKey"))
	if !reflect.DeepEqual(reg, region3) {
		t.Errorf("Found region %#v but expected %#v", reg, region3)
	}
	reg = client.getRegion([]byte("test"), []byte("zoo"))
	if reg != nil {
		t.Errorf("Shouldn't have found any region yet found %#v", reg)
	}
}
