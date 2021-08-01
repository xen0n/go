// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong

import (
	"testing"
)

func TestChopSImm(t *testing.T) {
	var tests = []struct {
		imm        int64
		wantTop    int32
		wantHigher int32
		wantUpper  int32
		wantLow    uint32
	}{
		{0x0, 0, 0, 0, 0},
		{0x1, 0, 0, 0, 1},
		{0x7ff, 0, 0, 0, 0x7ff},
		{0x800, 0, 0, 0, 0x800},
		{0xfff, 0, 0, 0, 0xfff},
		{0x0000_1000, 0, 0, 1, 0},
		{0x7fff_ffff, 0, 0, 0x7ffff, 0xfff},
		{0x8000_0000, 0, 0, -0x80000, 0},
		{0xffff_ffff, 0, 0, -1, 0xfff},
		{0x0000_0001_0000_0000, 0, 1, 0, 0},
		{0x000f_ffff_ffff_ffff, 0, -1, -1, 0xfff},
		{0x0010_0000_0000_0000, 1, 0, 0, 0},
		{0x0010_0001_0000_1001, 1, 1, 1, 1},
		{0x7fff_ffff_ffff_ffff, 0x7ff, -1, -1, 0xfff},
		{-0x8000_0000_0000_0000, -0x800, 0, 0, 0},
		{-0x7ff8_0000_0000_0000, -0x800, -0x80000, 0, 0},
		{-0x0123_4567_89ab_cdef, -0x13, -0x34568, 0x76543, 0x211},
		{-1, -1, -1, -1, 0xfff},
	}

	for _, test := range tests {
		got := chopSImm(test.imm)
		want := choppedSImm{
			low:    test.wantLow,
			upper:  test.wantUpper,
			higher: test.wantHigher,
			top:    test.wantTop,
		}
		if got != want {
			t.Errorf("chopSImm(%d) = %v, want %v", test.imm, got, want)
		}
	}
}
