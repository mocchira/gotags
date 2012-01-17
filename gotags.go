//  gotags generates a tags file for the Go Programming Language in the format used by exuberant-ctags.
//  Copyright (C) 2009  Michael R. Elkins <me@sigpipe.org>
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Date: 2009-11-12
// Updated: 2010-08-21
// Updated: 2010-11-21 Tomas Heran <tomas.heran@gmail.com>
// Updated: 2012-01-14 Yoshiyuki Kanno <nekotaroh@gmail.com> - implemented for weekly and recognized "pkg.Symbol" pattern as a tagname

//
// usage: gotags filename [ filename... ] > tags
//    ex: find ${GOROOT}/src/pkg -type f -name "*.go"|xargs gotags > tags

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"unicode"
	"unicode/utf8"
)

var (
	tags = make(sort.StringSlice, 0)
)

func isPrivate(n string) bool {
	dst := utf8.NewString(n)
	return unicode.IsLower(dst.At(0))
}

func output_tag(pkgName *ast.Ident, fset *token.FileSet, name *ast.Ident, kind byte, hasRecv bool) {
	position := fset.Position(name.NamePos)
	switch kind {
	case PKG:
		tags = append(tags, (fmt.Sprintf("%s\t%s\t%d;\"\t%c",
			name.Name, position.Filename, position.Line, kind)))
	case FUNC:
		var fname string
		if hasRecv || isPrivate(name.Name) {
			fname = name.Name
		} else {
			fname = pkgName.Name + "." + name.Name
		}
		tags = append(tags, (fmt.Sprintf("%s\t%s\t%d;\"\t%c",
			fname, position.Filename, position.Line, kind)))
	default:
		var vname string
		if isPrivate(name.Name) {
			vname = name.Name
		} else {
			vname = pkgName.Name + "." + name.Name
		}
		tags = append(tags, (fmt.Sprintf("%s\t%s\t%d;\"\t%c",
			vname, position.Filename, position.Line, kind)))
	}
}

func main() {
	parse_files()

	println("!_TAG_FILE_SORTED\t1\t")
	tags.Sort()
	for _, s := range tags {
		println(s)
	}
}

const FUNC, TYPE, VAR, PKG = 'f', 't', 'v', 'p'

func parse_files() {
	for _, f := range os.Args[1:] {
		fileset := token.NewFileSet()
		fi, _ := os.Lstat(f)
		fileset.AddFile(f, fileset.Base(), int(fi.Size()))
		tree, err := parser.ParseFile(fileset, f, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing file %s - %s\n", f, err.Error())
			os.Exit(1)
		}
		output_tag(tree.Name, fileset, tree.Name, PKG, false)
		for _, node := range tree.Decls {
			switch n := node.(type) {
			case *ast.FuncDecl:
				output_tag(tree.Name, fileset, n.Name, FUNC, n.Recv != nil)
			case *ast.GenDecl:
				do_gen_decl(tree.Name, fileset, n)
			}
		}
	}

}

func do_gen_decl(pkgName *ast.Ident, fset *token.FileSet, node *ast.GenDecl) {
	for _, v := range node.Specs {
		switch n := v.(type) {
		case *ast.TypeSpec:
			output_tag(pkgName, fset, n.Name, TYPE, false)

		case *ast.ValueSpec:
			for _, vv := range n.Names {
				output_tag(pkgName, fset, vv, VAR, false)
			}
		default:
		}
	}
}
