/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package stdlib

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestGetPkgDocString(t *testing.T) {

	mathFuncMap["Println"] = &ECALFunctionAdapter{reflect.ValueOf(fmt.Println), "foo"}

	f, _ := GetStdlibFunc("math.Println")

	if s, _ := f.DocString(); s != "foo" {
		t.Error("Unexpected result:", s)
		return
	}

	doc, _ := GetPkgDocString("math")

	if doc == "" {
		t.Error("Unexpected result:", doc)
		return
	}
}

func TestSymbols(t *testing.T) {
	p, c, f := GetStdlibSymbols()
	if len(p) == 0 || len(c) == 0 || len(f) == 0 {
		t.Error("Should have some entries in symbol lists:", p, c, f)
		return
	}
}

func TestSplitModuleAndName(t *testing.T) {

	if m, n := splitModuleAndName("fmt.Println"); m != "fmt" || n != "Println" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName(""); m != "" || n != "" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName("a"); m != "a" || n != "" {
		t.Error("Unexpected result:", m, n)
		return
	}

	if m, n := splitModuleAndName("my.FuncCall"); m != "my" || n != "FuncCall" {
		t.Error("Unexpected result:", m, n)
		return
	}
}

func TestGetStdLibItems(t *testing.T) {

	mathFuncMap["Println"] = &ECALFunctionAdapter{reflect.ValueOf(fmt.Println), "foo"}

	if f, _ := GetStdlibFunc("math.Println"); f != mathFuncMap["Println"] {
		t.Error("Unexpected resutl: functions should lookup correctly")
		return
	}

	if c, ok := GetStdlibFunc("foo"); c != nil || ok {
		t.Error("Unexpected resutl: constants should lookup correctly")
		return
	}

	if c, _ := GetStdlibConst("math.Pi"); c != math.Pi {
		t.Error("Unexpected resutl: constants should lookup correctly")
		return
	}

	if c, ok := GetStdlibConst("foo"); c != nil || ok {
		t.Error("Unexpected resutl: constants should lookup correctly")
		return
	}
}
