// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build loong64 && linux

package cpu

// This is initialized by archauxv and should not be changed after it is
// initialized.
var HWCap uint

// HWCAP bits. These are exposed by the Linux kernel.
const (
	hwcap_LOONGARCH_CPUCFG   = 1 << 0
	hwcap_LOONGARCH_LAM      = 1 << 1
	hwcap_LOONGARCH_UAL      = 1 << 2
	hwcap_LOONGARCH_FPU      = 1 << 3
	hwcap_LOONGARCH_LSX      = 1 << 4
	hwcap_LOONGARCH_LASX     = 1 << 5
	hwcap_LOONGARCH_CRC32    = 1 << 6
	hwcap_LOONGARCH_COMPLEX  = 1 << 7
	hwcap_LOONGARCH_CRYPTO   = 1 << 8
	hwcap_LOONGARCH_LVZ      = 1 << 9
	hwcap_LOONGARCH_LBT_X86  = 1 << 10
	hwcap_LOONGARCH_LBT_ARM  = 1 << 11
	hwcap_LOONGARCH_LBT_MIPS = 1 << 12
)

func hwcapInit() {
	Loong64.HasCPUCFG = isSet(HWCap, hwcap_LOONGARCH_CPUCFG)

	// These are not taken from CPUCFG data regardless of availability of
	// CPUCFG, because the CPUCFG data only reflects capabilities of the
	// hardware, but not kernel support.
	//
	// As of 2023, we do not know for sure if the CPUCFG data can be
	// patched in software, nor does any known LoongArch kernel do that.
	Loong64.HasLSX = isSet(HWCap, hwcap_LOONGARCH_LSX)
	Loong64.HasLASX = isSet(HWCap, hwcap_LOONGARCH_LASX)
	Loong64.HasCRC32 = isSet(HWCap, hwcap_LOONGARCH_CRC32)
	Loong64.HasLBTX86 = isSet(HWCap, hwcap_LOONGARCH_LBT_X86)
	Loong64.HasLBTARM = isSet(HWCap, hwcap_LOONGARCH_LBT_ARM)
}

func isSet(hwc uint, value uint) bool {
	return hwc&value != 0
}
