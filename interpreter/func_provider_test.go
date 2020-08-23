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
	"fmt"
	"testing"
)

func TestStdlib(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`fmt.Sprint([1,2,3])`, nil,
		`
identifier: fmt
  identifier: Sprint
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err != nil || res != "[1 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`fmt.Sprint(math.Pi)`, nil,
		`
identifier: fmt
  identifier: Sprint
    funccall
      identifier: math
        identifier: Pi
`[1:])

	if err != nil || res != "3.141592653589793" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Negative case

	res, err = UnitTestEvalAndAST(
		`a.fmtSprint([1,2,3])`, nil,
		`
identifier: a
  identifier: fmtSprint
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Unknown construct (Unknown function: fmtSprint) (Line:1 Pos:3)" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}

func TestSimpleFunctions(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`len([1,2,3])`, nil,
		`
identifier: len
  funccall
    list
      number: 1
      number: 2
      number: 3
`[1:])

	if err != nil || res != 3. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`len({"a":1, 2:"b"})`, nil,
		`
identifier: len
  funccall
    map
      kvp
        string: 'a'
        number: 1
      kvp
        number: 2
        string: 'b'
`[1:])

	if err != nil || res != 2. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`del([1,2,3], 1)`, nil,
		`
identifier: del
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 1
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`del({
  "a" : 1
  "b" : 2
  "c" : 3
}, "b")`, nil,
		`
identifier: del
  funccall
    map
      kvp
        string: 'a'
        number: 1
      kvp
        string: 'b'
        number: 2
      kvp
        string: 'c'
        number: 3
    string: 'b'
`[1:])

	if err != nil || fmt.Sprint(res) != "map[a:1 c:3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 2 3 4]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4, 0)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
    number: 0
`[1:])

	if err != nil || fmt.Sprint(res) != "[4 1 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`add([1,2,3], 4, 1)`, nil,
		`
identifier: add
  funccall
    list
      number: 1
      number: 2
      number: 3
    number: 4
    number: 1
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 4 2 3]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`concat([1,2,3], [4,5,6], [7,8,9])`, nil,
		`
identifier: concat
  funccall
    list
      number: 1
      number: 2
      number: 3
    list
      number: 4
      number: 5
      number: 6
    list
      number: 7
      number: 8
      number: 9
`[1:])

	if err != nil || fmt.Sprint(res) != "[1 2 3 4 5 6 7 8 9]" {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`dumpenv()`, nil,
		`
identifier: dumpenv
  funccall
`[1:])

	if err != nil || fmt.Sprint(res) != `GlobalScope {
}` {
		t.Error("Unexpected result: ", res, err)
		return
	}

	// Negative case

	res, err = UnitTestEvalAndAST(
		`a.len([1,2,3])`, nil,
		`
identifier: a
  identifier: len
    funccall
      list
        number: 1
        number: 2
        number: 3
`[1:])

	if err == nil ||
		err.Error() != "ECAL error in ECALTestRuntime: Unknown construct (Unknown function: len) (Line:1 Pos:3)" {
		t.Error("Unexpected result: ", res, err)
		return
	}

}
