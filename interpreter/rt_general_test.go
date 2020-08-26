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
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

func TestGeneralErrorCases(t *testing.T) {

	n, _ := parser.Parse("a", "a")
	inv := &invalidRuntime{newBaseRuntime(NewECALRuntimeProvider("a", nil, nil), n)}

	if err := inv.Validate().Error(); err != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}

	if _, err := inv.Eval(nil, nil); err.Error() != "ECAL error in a: Invalid construct (Unknown node: identifier) (Line:1 Pos:1)" {
		t.Error("Unexpected result:", err)
		return
	}
}

func TestImporting(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)
	il := &util.MemoryImportLocator{make(map[string]string)}

	il.Files["foo/bar"] = `
b := 123
`

	res, err := UnitTestEvalAndASTAndImport(
		`
	   import "foo/bar" as foobar
	   a := foobar.b`, vs,
		`
statements
  import
    string: 'foo/bar'
    identifier: foobar
  :=
    identifier: a
    identifier: foobar
      identifier: b
`[1:], il)

	if vsRes := vs.String(); err != nil || res != nil || vsRes != `GlobalScope {
    a (float64) : 123
    foobar (map[interface {}]interface {}) : {"b":123}
}` {
		t.Error("Unexpected result: ", vsRes, res, err)
		return
	}
}

func TestLogging(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
log("Hello")
debug("foo")
error("bar")
`, vs,
		`
statements
  identifier: log
    funccall
      string: 'Hello'
  identifier: debug
    funccall
      string: 'foo'
  identifier: error
    funccall
      string: 'bar'
`[1:])

	if err != nil {
		t.Error("Unexpected result: ", err)
		return
	}

	if testlogger.String() != `Hello
debug: foo
error: bar` {
		t.Error("Unexpected result: ", testlogger.String())
		return
	}
}
