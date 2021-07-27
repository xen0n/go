// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
)

func Init() (*sys.Arch, ld.Arch) {
	arch := sys.ArchLoong64

	theArch := ld.Arch{
		Funcalign:  funcAlign,
		Maxalign:   maxAlign,
		Minalign:   minAlign,
		Dwarfregsp: dwarfRegSP,
		Dwarfreglr: dwarfRegLR,

		Archinit:         archinit,
		Archreloc:        archreloc,
		Archrelocvariant: archrelocvariant,
		Elfreloc1:        elfreloc1,
		ElfrelocSize:     24,
		Elfsetupplt:      elfsetupplt,
		Gentext:          gentext,
		GenSymsLate:      genSymsLate,
		Machoreloc1:      machoreloc1,

		Linuxdynld:     "XXX",
		Freebsddynld:   "XXX",
		Netbsddynld:    "XXX",
		Openbsddynld:   "XXX",
		Dragonflydynld: "XXX",
		Solarisdynld:   "XXX",
	}

	return arch, theArch
}

func archinit(ctxt *ld.Link) {
	switch ctxt.HeadType {
	case objabi.Hlinux:
		ld.Elfinit(ctxt)
		ld.HEADR = ld.ELFRESERVE
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 0x10000 + int64(ld.HEADR)
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 0x10000
		}
	default:
		ld.Exitf("unknown -H option: %v", ctxt.HeadType)
	}
}
