// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"fmt"
)

// Notes:
//  - Boolean types occupy the entire register. 0=false, 1=true.

// Suffixes encode the bit width of various instructions:
//
// D (double word) = 64 bit int
// W (word)        = 32 bit int
// H (half word)   = 16 bit int
// B (byte)        = 8 bit int
// S (single)      = 32 bit float
// D (double)      = 64 bit float
// L               = 64 bit int, used when the opcode starts with F

const (
	loong64REG_ZERO     = 0
	loong64REG_LR       = 1
	loong64REG_TP       = 2
	loong64REG_SP       = 3
	loong64REG_TMP      = 20
	loong64REG_RESERVED = 21
	loong64REG_CTXT     = 30
	loong64REG_G        = 31
)

func loong64RegName(r int) string {
	switch {
	case r == loong64REG_G:
		return "g"
	case r == loong64REG_SP:
		return "SP"
	case 0 <= r && r <= 31:
		return fmt.Sprintf("R%d", r)
	case 32 <= r && r <= 63:
		return fmt.Sprintf("F%d", r-32)
	default:
		panic(fmt.Sprintf("unknown register %d", r))
	}
}

func init() {
	var regNamesLoong64 []string
	var gpMask, fpMask, gpgMask, gpspMask, gpspsbMask, gpspsbgMask regMask
	regNamed := make(map[string]regMask)

	// Build the list of register names, creating an appropriately indexed
	// regMask for the gp and fp registers as we go.
	//
	// If name is specified, use it rather than the loong reg number.
	addreg := func(r int, name string) regMask {
		mask := regMask(1) << uint(len(regNamesLoong64))
		if name == "" {
			name = loong64RegName(r)
		}
		regNamesLoong64 = append(regNamesLoong64, name)
		regNamed[name] = mask
		return mask
	}

	// General purpose registers.
	for r := 0; r <= 31; r++ {
		switch r {
		case loong64REG_LR, loong64REG_RESERVED:
			// LR is not used by regalloc and R21 is reserved by ABI,
			// so we skip it to leave room for pseudo-register SB.
			continue
		}

		mask := addreg(r, "")

		// Add general purpose registers to gpMask.
		switch r {
		// ZERO, TP and TMP are not in any gp mask.
		case loong64REG_ZERO, loong64REG_TP, loong64REG_TMP:
		case loong64REG_G:
			gpgMask |= mask
			gpspsbgMask |= mask
		case loong64REG_SP:
			gpspMask |= mask
			gpspsbMask |= mask
			gpspsbgMask |= mask
		default:
			gpMask |= mask
			gpgMask |= mask
			gpspMask |= mask
			gpspsbMask |= mask
			gpspsbgMask |= mask
		}
	}

	// Floating pointer registers.
	for r := 32; r <= 63; r++ {
		mask := addreg(r, "")
		fpMask |= mask
	}

	// Pseudo-register: SB
	mask := addreg(-1, "SB")
	gpspsbMask |= mask
	gpspsbgMask |= mask

	if len(regNamesLoong64) > 64 {
		// regMask is only 64 bits.
		panic("Too many Loong64 registers")
	}

	regCtxt := regNamed["R30"]
	callerSave := gpMask | fpMask | regNamed["g"]
	_ = regCtxt
	_ = callerSave

	var (
		gp21 = regInfo{inputs: []regMask{gpMask, gpMask}, outputs: []regMask{gpMask}}
	)

	Loong64ops := []opData{
		{name: "ADDD", argLength: 2, reg: gp21, asm: "ADDD", commutative: true}, // arg0 + arg1
	}

	Loong64blocks := []blockData{
		// Natively supported branches.
		{name: "BEQZ", controls: 1},
		{name: "BNEZ", controls: 1},
		{name: "BEQ", controls: 2},
		{name: "BNE", controls: 2},
		{name: "BLE", controls: 2},
		{name: "BGT", controls: 2},
		{name: "BLEU", controls: 2},
		{name: "BGTU", controls: 2},

		// FP condition code branches.
		{name: "BCEQZ", controls: 1},
		{name: "BCNEZ", controls: 1},

		// Sugar.
		{name: "BLEZ", controls: 1},
		{name: "BGEZ", controls: 1},
		{name: "BLTZ", controls: 1},
		{name: "BGTZ", controls: 1},
	}

	archs = append(archs, arch{
		name:            "Loong64",
		pkg:             "cmd/internal/obj/loong",
		genfile:         "../../loong64/ssa.go",
		ops:             Loong64ops,
		blocks:          Loong64blocks,
		regnames:        regNamesLoong64,
		gpregmask:       gpMask,
		fpregmask:       fpMask,
		framepointerreg: -1, // not used
	})
}
