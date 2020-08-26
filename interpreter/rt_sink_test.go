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

	"devt.de/krotik/ecal/scope"
)

func TestEventProcessing(t *testing.T) {

	vs := scope.NewScope(scope.GlobalScope)

	_, err := UnitTestEvalAndAST(
		`
/*
My cool rule
*/
sink rule1
    kindmatch [ "core.*" ],
	scopematch [ "data.write" ],
	statematch { "val" : NULL },
	priority 10,
	suppresses [ "rule2" ]
	{
        log("rule1 < ", event)
	}
`, vs,
		`
sink # 
My cool rule

  identifier: rule1
  kindmatch
    list
      string: 'core.*'
  scopematch
    list
      string: 'data.write'
  statematch
    map
      kvp
        string: 'val'
        null
  priority
    number: 10
  suppresses
    list
      string: 'rule2'
  statements
    identifier: log
      funccall
        string: 'rule1 < '
        identifier: event
`[1:])

	if err != nil {
		t.Error(err)
		return
	}

	// Nothing defined in the global scope

	if vs.String() != `
GlobalScope {
}`[1:] {
		t.Error("Unexpected result: ", vs)
		return
	}

	if res := fmt.Sprint(testprocessor.Rules()["rule1"]); res !=
		`Rule:rule1 [My cool rule] (Priority:10 Kind:[core.*] Scope:[data.write] StateMatch:{"val":null} Suppress:[rule2])` {
		t.Error("Unexpected result:", res)
		return
	}

	_, err = UnitTestEval(
		`
sink rule1
    kindmatch [ "web.page.index" ],
	scopematch [ "request.read" ],
	{
        log("rule1 > Handling request: ", event.kind)
        addEvent("Rule1Event1", "not_existing", event.state)
        addEvent("Rule1Event2", "web.log", event.state)
	}

sink rule2
    kindmatch [ "web.page.*" ],
    priority 1,  # Ensure this rule is always executed after rule1
	{
        log("rule2 > Tracking user:", event.state.user)
	}

sink rule3
    kindmatch [ "web.log" ],
	{
        log("rule3 > Logging user:", event.state.user)
	}

res := addEventAndWait("request", "web.page.index", {
	"user" : "foo"
}, {
	"request.read" : true
})
log("ErrorResult:", res, len(res) == 0)
`, vs)

	if err != nil {
		t.Error(err)
		return
	}

	if testlogger.String() != `
rule1 > Handling request: web.page.index
rule2 > Tracking user:foo
rule3 > Logging user:foo
ErrorResult:[] true`[1:] {
		t.Error("Unexpected result:", testlogger.String())
		return
	}
}
