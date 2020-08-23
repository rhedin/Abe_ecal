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
)

func TestLogging(t *testing.T) {

	ml := NewMemoryLogger(5)

	ml.LogDebug(nil, "test")
	ml.LogInfo(nil, "test")

	if ml.String() != `debug: test
info: test` {
		t.Error("Unexpected result:", ml.String())
		return
	}

	if res := fmt.Sprint(ml.Slice()); res != "[debug: test info: test]" {
		t.Error("Unexpected result:", res)
		return
	}

	ml.Reset()

	ml.LogError(nil, "test1")

	if res := fmt.Sprint(ml.Slice()); res != "[error: test1]" {
		t.Error("Unexpected result:", res)
		return
	}

	if res := ml.Size(); res != 1 {
		t.Error("Unexpected result:", res)
		return
	}

	// Test that the functions can be called

	nl := NewNullLogger()
	nl.LogDebug(nil, "test")
	nl.LogInfo(nil, "test")
	nl.LogError(nil, "test")

	sol := NewStdOutLogger()
	sol.stdlog = func(v ...interface{}) {}
	sol.LogDebug(nil, "test")
	sol.LogInfo(nil, "test")
	sol.LogError(nil, "test")
}
