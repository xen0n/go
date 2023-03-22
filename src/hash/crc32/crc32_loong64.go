// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// LoongArch64-specific hardware-assisted CRC32 algorithms. See crc32.go for a
// description of the interface that each architecture-specific file
// implements.

package crc32

func castagnoliUpdate(crc uint32, p []byte) uint32
func ieeeUpdate(crc uint32, p []byte) uint32

func archAvailableCastagnoli() bool {
	return true
}

func archInitCastagnoli() {
}

func archUpdateCastagnoli(crc uint32, p []byte) uint32 {
	return ^castagnoliUpdate(^crc, p)
}

func archAvailableIEEE() bool {
	return true
}

func archInitIEEE() {
}

func archUpdateIEEE(crc uint32, p []byte) uint32 {
	return ^ieeeUpdate(^crc, p)
}
