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
	"math"
	"testing"
)

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

	if f, _ := GetStdlibFunc("fmt.Println"); f != fmtFuncMap["Println"] {
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
