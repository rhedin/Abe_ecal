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
	"strconv"
	"strings"

	"devt.de/krotik/ecal/engine"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/stdlib"
	"devt.de/krotik/ecal/util"
)

/*
InbuildFuncMap contains the mapping of inbuild functions.
*/
var InbuildFuncMap = map[string]util.ECALFunction{
	"range":           &rangeFunc{&inbuildBaseFunc{}},
	"new":             &newFunc{&inbuildBaseFunc{}},
	"len":             &lenFunc{&inbuildBaseFunc{}},
	"del":             &delFunc{&inbuildBaseFunc{}},
	"add":             &addFunc{&inbuildBaseFunc{}},
	"concat":          &concatFunc{&inbuildBaseFunc{}},
	"dumpenv":         &dumpenvFunc{&inbuildBaseFunc{}},
	"doc":             &docFunc{&inbuildBaseFunc{}},
	"raise":           &raise{&inbuildBaseFunc{}},
	"addEvent":        &addevent{&inbuildBaseFunc{}},
	"addEventAndWait": &addeventandwait{&addevent{&inbuildBaseFunc{}}},
}

/*
inbuildBaseFunc is the base structure for inbuild functions providing some
utility functions.
*/
type inbuildBaseFunc struct {
}

/*
AssertNumParam converts a general interface{} parameter into a number.
*/
func (ibf *inbuildBaseFunc) AssertNumParam(index int, val interface{}) (float64, error) {
	var err error

	resNum, ok := val.(float64)

	if !ok {

		resNum, err = strconv.ParseFloat(fmt.Sprint(val), 64)
		if err != nil {
			err = fmt.Errorf("Parameter %v should be a number", index)
		}
	}

	return resNum, err
}

/*
AssertMapParam converts a general interface{} parameter into a map.
*/
func (ibf *inbuildBaseFunc) AssertMapParam(index int, val interface{}) (map[interface{}]interface{}, error) {

	valMap, ok := val.(map[interface{}]interface{})

	if ok {
		return valMap, nil
	}

	return nil, fmt.Errorf("Parameter %v should be a map", index)
}

/*
AssertListParam converts a general interface{} parameter into a list.
*/
func (ibf *inbuildBaseFunc) AssertListParam(index int, val interface{}) ([]interface{}, error) {

	valList, ok := val.([]interface{})

	if ok {
		return valList, nil
	}

	return nil, fmt.Errorf("Parameter %v should be a list", index)
}

// Range
// =====

/*
rangeFunc is an interator function which returns a range of numbers.
*/
type rangeFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *rangeFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var currVal, to float64
	var err error

	lenargs := len(args)
	from := 0.
	step := 1.

	if lenargs == 0 {
		err = fmt.Errorf("Need at least an end range as first parameter")
	}

	if err == nil {

		if stepVal, ok := is[instanceID+"step"]; ok {

			step = stepVal.(float64)
			from = is[instanceID+"from"].(float64)
			to = is[instanceID+"to"].(float64)
			currVal = is[instanceID+"currVal"].(float64)

			is[instanceID+"currVal"] = currVal + step

			// Check for end of iteration

			if (from < to && currVal > to) || (from > to && currVal < to) || from == to {
				err = util.ErrEndOfIteration
			}

		} else {

			if lenargs == 1 {
				to, err = rf.AssertNumParam(1, args[0])
			} else {
				from, err = rf.AssertNumParam(1, args[0])

				if err == nil {
					to, err = rf.AssertNumParam(2, args[1])
				}

				if err == nil && lenargs > 2 {
					step, err = rf.AssertNumParam(3, args[2])
				}
			}

			if err == nil {
				is[instanceID+"from"] = from
				is[instanceID+"to"] = to
				is[instanceID+"step"] = step
				is[instanceID+"currVal"] = from

				currVal = from
			}
		}
	}

	if err == nil {
		err = util.ErrIsIterator // Identify as iterator
	}

	return currVal, err
}

/*
DocString returns a descriptive string.
*/
func (rf *rangeFunc) DocString() (string, error) {
	return "Range function which can be used to iterate over number ranges. Parameters are start, end and step.", nil
}

// New
// ===

/*
newFunc instantiates a new object.
*/
type newFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *newFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res interface{}

	err := fmt.Errorf("Need a map as first parameter")

	if len(args) > 0 {
		var argMap map[interface{}]interface{}
		if argMap, err = rf.AssertMapParam(1, args[0]); err == nil {
			obj := make(map[interface{}]interface{})
			res = obj

			_, err = rf.addSuperClasses(vs, is, obj, argMap)

			if initObj, ok := obj["init"]; ok {
				if initFunc, ok := initObj.(*function); ok {

					initvs := scope.NewScope(fmt.Sprintf("newfunc: %v", instanceID))
					initis := make(map[string]interface{})

					_, err = initFunc.Run(instanceID, initvs, initis, args[1:])
				}
			}
		}
	}

	return res, err
}

/*
addSuperClasses adds super class functions to a given object.
*/
func (rf *newFunc) addSuperClasses(vs parser.Scope, is map[string]interface{},
	obj map[interface{}]interface{}, template map[interface{}]interface{}) (interface{}, error) {

	var err error

	var initFunc interface{}
	var initSuperList []interface{}

	// First loop into the base classes (i.e. top-most classes)

	if super, ok := template["super"]; ok {
		if superList, ok := super.([]interface{}); ok {
			for _, superObj := range superList {
				var superInit interface{}

				if superTemplate, ok := superObj.(map[interface{}]interface{}); ok {
					superInit, err = rf.addSuperClasses(vs, is, obj, superTemplate)
					initSuperList = append(initSuperList, superInit) // Build up the list of super functions
				}
			}
		} else {
			err = fmt.Errorf("Property _super must be a list of super classes")
		}
	}

	// Copy all properties from template to obj

	for k, v := range template {

		// Save previous init function

		if funcVal, ok := v.(*function); ok {
			newFunction := &function{funcVal.name, nil, obj, funcVal.declaration}
			if k == "init" {
				newFunction.super = initSuperList
				initFunc = newFunction
			}
			obj[k] = newFunction
		} else {
			obj[k] = v
		}
	}

	return initFunc, err
}

/*
DocString returns a descriptive string.
*/
func (rf *newFunc) DocString() (string, error) {
	return "New creates a new object instance.", nil
}

// Len
// ===

/*
lenFunc returns the size of a list or map.
*/
type lenFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *lenFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res float64

	err := fmt.Errorf("Need a list or a map as first parameter")

	if len(args) > 0 {
		argList, ok1 := args[0].([]interface{})
		argMap, ok2 := args[0].(map[interface{}]interface{})

		if ok1 {
			res = float64(len(argList))
			err = nil
		} else if ok2 {
			res = float64(len(argMap))
			err = nil
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *lenFunc) DocString() (string, error) {
	return "Len returns the size of a list or map.", nil
}

// Del
// ===

/*
delFunc removes an element from a list or map.
*/
type delFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *delFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res interface{}

	err := fmt.Errorf("Need a list or a map as first parameter and an index or key as second parameter")

	if len(args) == 2 {

		if argList, ok1 := args[0].([]interface{}); ok1 {
			var index float64

			index, err = rf.AssertNumParam(2, args[1])
			if err == nil {
				res = append(argList[:int(index)], argList[int(index+1):]...)
			}
		}

		if argMap, ok2 := args[0].(map[interface{}]interface{}); ok2 {
			key := fmt.Sprint(args[1])
			delete(argMap, key)
			res = argMap
			err = nil
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *delFunc) DocString() (string, error) {
	return "Del removes an item from a list or map.", nil
}

// Add
// ===

/*
addFunc adds an element to a list.
*/
type addFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *addFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res interface{}

	err := fmt.Errorf("Need a list as first parameter and a value as second parameter")

	if len(args) > 1 {
		var argList []interface{}

		if argList, err = rf.AssertListParam(1, args[0]); err == nil {
			if len(args) == 3 {
				var index float64

				if index, err = rf.AssertNumParam(3, args[2]); err == nil {
					argList = append(argList, 0)
					copy(argList[int(index+1):], argList[int(index):])
					argList[int(index)] = args[1]
					res = argList
				}
			} else {
				res = append(argList, args[1])
			}
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *addFunc) DocString() (string, error) {
	return "Add adds an item to a list. The item is added at the optionally given index or at the end if no index is specified.", nil
}

// Concat
// ======

/*
concatFunc joins one or more lists together.
*/
type concatFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *concatFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res interface{}

	err := fmt.Errorf("Need at least two lists as parameters")

	if len(args) > 1 {
		var argList []interface{}

		resList := make([]interface{}, 0)
		err = nil

		for _, a := range args {
			if err == nil {
				if argList, err = rf.AssertListParam(1, a); err == nil {
					resList = append(resList, argList...)
				}
			}
		}

		if err == nil {
			res = resList
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *concatFunc) DocString() (string, error) {
	return "Concat joins one or more lists together. The result is a new list.", nil
}

// dumpenv
// =======

/*
dumpenvFunc returns the current variable environment as a string.
*/
type dumpenvFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *dumpenvFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	return vs.String(), nil
}

/*
DocString returns a descriptive string.
*/
func (rf *dumpenvFunc) DocString() (string, error) {
	return "Dumpenv returns the current variable environment as a string.", nil
}

// doc
// ===

/*
docFunc returns the docstring of a function.
*/
type docFunc struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *docFunc) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var res interface{}
	err := fmt.Errorf("Need a function as parameter")

	if len(args) > 0 {

		funcObj, ok := args[0].(util.ECALFunction)

		if args[0] == nil {

			// Try to lookup by the given identifier

			c := is["astnode"].(*parser.ASTNode).Children[0].Children[0]
			astring := c.Token.Val

			if len(c.Children) > 0 {
				astring = fmt.Sprintf("%v.%v", astring, c.Children[0].Token.Val)
			}

			// Check for stdlib function

			if funcObj, ok = stdlib.GetStdlibFunc(astring); !ok {

				// Check for inbuild function

				funcObj, ok = InbuildFuncMap[astring]
			}
		}

		if ok {
			res, err = funcObj.DocString()
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *docFunc) DocString() (string, error) {
	return "Doc returns the docstring of a function.", nil
}

// raise
// =====

/*
raise returns an error. Outside of sinks this will stop the code execution
if the error is not handled by try / except. Inside a sink only the specific sink
will fail. This error can be used to break trigger sequences of sinks if
FailOnFirstErrorInTriggerSequence is set.
*/
type raise struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *raise) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	var err error
	var detailMsg string
	var detail interface{}

	if len(args) > 0 {
		err = fmt.Errorf("%v", args[0])
		if len(args) > 1 {
			if args[1] != nil {
				detailMsg = fmt.Sprint(args[1])
			}
			if len(args) > 2 {
				detail = args[2]
			}
		}
	}

	erp := is["erp"].(*ECALRuntimeProvider)
	node := is["astnode"].(*parser.ASTNode)

	return nil, &util.RuntimeErrorWithDetail{
		RuntimeError: erp.NewRuntimeError(err, detailMsg, node).(*util.RuntimeError),
		Environment:  vs,
		Data:         detail,
	}

}

/*
DocString returns a descriptive string.
*/
func (rf *raise) DocString() (string, error) {
	return "Raise returns an error object.", nil
}

// addEvent
// ========

/*
addevent adds an event to trigger sinks. This function will return immediately
and not wait for the event cascade to finish. Use this function for event cascades.
*/
type addevent struct {
	*inbuildBaseFunc
}

/*
Run executes this function.
*/
func (rf *addevent) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	return rf.addEvent(func(proc engine.Processor, event *engine.Event, scope *engine.RuleScope) (interface{}, error) {
		var monitor engine.Monitor

		parentMonitor, ok := is["monitor"]

		if scope != nil || !ok {
			monitor = proc.NewRootMonitor(nil, scope)
		} else {
			monitor = parentMonitor.(engine.Monitor).NewChildMonitor(0)
		}

		_, err := proc.AddEvent(event, monitor)
		return nil, err
	}, is, args)
}

func (rf *addevent) addEvent(addFunc func(engine.Processor, *engine.Event, *engine.RuleScope) (interface{}, error),
	is map[string]interface{}, args []interface{}) (interface{}, error) {

	var res interface{}
	var stateMap map[interface{}]interface{}

	erp := is["erp"].(*ECALRuntimeProvider)
	proc := erp.Processor

	if proc.Stopped() {
		proc.Start()
	}

	err := fmt.Errorf("Need at least three parameters: name, kind and state")

	if len(args) > 2 {

		if stateMap, err = rf.AssertMapParam(3, args[2]); err == nil {
			var scope *engine.RuleScope

			event := engine.NewEvent(
				fmt.Sprint(args[0]),
				strings.Split(fmt.Sprint(args[1]), "."),
				stateMap,
			)

			if len(args) > 3 {
				var scopeMap map[interface{}]interface{}

				// Add optional scope - if not specified it is { "": true }

				if scopeMap, err = rf.AssertMapParam(4, args[3]); err == nil {
					var scopeData = map[string]bool{}

					for k, v := range scopeMap {
						b, _ := strconv.ParseBool(fmt.Sprint(v))
						scopeData[fmt.Sprint(k)] = b
					}

					scope = engine.NewRuleScope(scopeData)
				}
			}

			if err == nil {
				res, err = addFunc(proc, event, scope)
			}
		}
	}

	return res, err
}

/*
DocString returns a descriptive string.
*/
func (rf *addevent) DocString() (string, error) {
	return "AddEvent adds an event to trigger sinks. This function will return " +
		"immediately and not wait for the event cascade to finish.", nil
}

// addEventAndWait
// ===============

/*
addeventandwait adds an event to trigger sinks. This function will return once
the event cascade has finished and return all errors.
*/
type addeventandwait struct {
	*addevent
}

/*
Run executes this function.
*/
func (rf *addeventandwait) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	return rf.addEvent(func(proc engine.Processor, event *engine.Event, scope *engine.RuleScope) (interface{}, error) {
		var res []interface{}
		rm := proc.NewRootMonitor(nil, scope)
		m, err := proc.AddEventAndWait(event, rm)

		if m != nil {
			allErrors := m.(*engine.RootMonitor).AllErrors()

			for _, e := range allErrors {

				errors := map[interface{}]interface{}{}
				for k, v := range e.ErrorMap {
					se := v.(*util.RuntimeErrorWithDetail)

					// Note: The variable scope of the sink (se.environment)
					// was also captured - for now it is not exposed to the
					// language environment

					errors[k] = map[interface{}]interface{}{
						"error":  se.Error(),
						"type":   se.Type.Error(),
						"detail": se.Detail,
						"data":   se.Data,
					}
				}

				item := map[interface{}]interface{}{
					"event": map[interface{}]interface{}{
						"name":  e.Event.Name(),
						"kind":  strings.Join(e.Event.Kind(), "."),
						"state": e.Event.State(),
					},
					"errors": errors,
				}

				res = append(res, item)
			}
		}

		return res, err
	}, is, args)
}

/*
DocString returns a descriptive string.
*/
func (rf *addeventandwait) DocString() (string, error) {
	return "AddEventAndWait adds an event to trigger sinks. This function will " +
		"return once the event cascade has finished.", nil
}
