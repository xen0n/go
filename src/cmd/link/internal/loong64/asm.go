// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loong64

import (
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"log"
)

func gentext(ctxt *ld.Link, ldr *loader.Loader) {
}

func genSymsLate(ctxt *ld.Link, ldr *loader.Loader) {
	if ctxt.LinkMode != ld.LinkExternal {
		return
	}

	log.Fatalf("genSymsLate not implemented")
}

func elfreloc1(ctxt *ld.Link, out *ld.OutBuf, ldr *loader.Loader, s loader.Sym, r loader.ExtReloc, ri int, sectoff int64) bool {
	log.Fatalf("elfreloc1 not implemented")
	return false
}

func elfsetupplt(ctxt *ld.Link, plt, gotplt *loader.SymbolBuilder, dynamic loader.Sym) {
	log.Fatalf("elfsetupplt")
}

func machoreloc1(*sys.Arch, *ld.OutBuf, *loader.Loader, loader.Sym, loader.ExtReloc, int64) bool {
	log.Fatalf("machoreloc1 not implemented")
	return false
}

func archreloc(target *ld.Target, ldr *loader.Loader, syms *ld.ArchSyms, r loader.Reloc, s loader.Sym, val int64) (o int64, nExtReloc int, ok bool) {
	log.Fatalf("archreloc not implemented")
	return val, 0, false
}

func archrelocvariant(*ld.Target, *loader.Loader, loader.Reloc, sym.RelocVariant, loader.Sym, int64, []byte) int64 {
	log.Fatalf("archrelocvariant")
	return -1
}
