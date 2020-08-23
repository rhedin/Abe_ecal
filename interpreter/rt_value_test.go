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

	"devt.de/krotik/ecal/scope"
)

func TestSimpleValues(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`4`, nil,
		`
number: 4
`[1:])

	if err != nil || res != 4. {
		t.Error("Unexpected result: ", res, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`"test"`, nil,
		`
string: 'test'
`[1:])

	if err != nil || res != "test" {
		t.Error("Unexpected result: ", res, err)
		return
	}
}

func TestCompositionValues(t *testing.T) {

	res, err := UnitTestEvalAndAST(
		`{"a":1, "b": 2, "c" : 3}`, nil,
		`
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
`[1:])

	if resStr := scope.EvalToString(res); err != nil || resStr != `{"a":1,"b":2,"c":3}` {
		t.Error("Unexpected result: ", resStr, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`{"a":1, "b": {"a":1, "b": 2, "c" : 3}, "c" : 3}`, nil,
		`
map
  kvp
    string: 'a'
    number: 1
  kvp
    string: 'b'
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
  kvp
    string: 'c'
    number: 3
`[1:])

	if resStr := scope.EvalToString(res); err != nil || resStr != `{"a":1,"b":{"a":1,"b":2,"c":3},"c":3}` {
		t.Error("Unexpected result: ", resStr, err)
		return
	}

	res, err = UnitTestEvalAndAST(
		`{"a":1, "b": [1, [2, 3], 4], "c" : 3}`, nil,
		`
map
  kvp
    string: 'a'
    number: 1
  kvp
    string: 'b'
    list
      number: 1
      list
        number: 2
        number: 3
      number: 4
  kvp
    string: 'c'
    number: 3
`[1:])

	if resStr := scope.EvalToString(res); err != nil || resStr != `{"a":1,"b":[1,[2,3],4],"c":3}` {
		t.Error("Unexpected result: ", resStr, err)
		return
	}

}
