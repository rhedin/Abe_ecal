/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package util

import (
	"fmt"
	"testing"

	"devt.de/krotik/ecal/parser"
)

func TestRuntimeError(t *testing.T) {

	ast, _ := parser.Parse("foo", "a")

	err1 := NewRuntimeError("foo", fmt.Errorf("foo"), "bar", ast)

	if err1.Error() != "ECAL error in foo: foo (bar) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err1)
		return
	}

	ast.Token = nil

	err2 := NewRuntimeError("foo", fmt.Errorf("foo"), "bar", ast)

	if err2.Error() != "ECAL error in foo: foo (bar)" {
		t.Error("Unexpected result:", err2)
		return
	}
}
