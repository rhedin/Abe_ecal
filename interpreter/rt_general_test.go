/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package interpreter

import (
	"testing"

	"devt.de/krotik/ecal/parser"
)

func TestGeneralErrorCases(t *testing.T) {

	n, _ := parser.Parse("a", "a")
	inv := &invalidRuntime{newBaseRuntime(NewECALRuntimeProvider("a"), n)}

	if err := inv.Validate().Error(); err != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := inv.Eval(nil, nil); err.Error() != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}
}
