/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

/*
Package interpreter contains the ECAL interpreter.
*/
package interpreter

import (
	"fmt"
	"strings"
	"sync"

	"devt.de/krotik/common/errorutil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
ecalDebugger is the inbuild default debugger.
*/
type ecalDebugger struct {
	breakPoints         map[string]bool                // Break points (active or not)
	interrogationStates map[uint64]*interrogationState // Collection of threads which are interrogated
	callStacks          map[uint64][]*parser.ASTNode   // Call stacks of threads
	sources             map[string]bool                // All known sources
	breakOnStart        bool                           // Flag to stop at the start of the next execution
	globalScope         parser.Scope                   // Global variable scope which can be used to transfer data
	lock                *sync.RWMutex                  // Lock for this debugger

}

/*
interrogationState contains state information of a thread interrogation.
*/
type interrogationState struct {
	cond         *sync.Cond        // Condition on which the thread is waiting when suspended
	running      bool              // Flag if the thread is running or waiting
	cmd          interrogationCmd  // Next interrogation command for the thread
	stepOutStack []*parser.ASTNode // Target stack when doing a step out
	node         *parser.ASTNode   // Node on which the thread was last stopped
	vs           parser.Scope      // Variable scope of the thread when it was last stopped
}

/*
interrogationCmd represents a command for a thread interrogation.
*/
type interrogationCmd int

/*
Interrogation commands
*/
const (
	Stop     interrogationCmd = iota // Stop the execution (default)
	StepIn                           // Step into the next function
	StepOut                          // Step out of the current function
	StepOver                         // Step over the next function
	Resume                           // Resume execution - do not break again on the same line
)

/*
newInterrogationState creates a new interrogation state.
*/
func newInterrogationState(node *parser.ASTNode, vs parser.Scope) *interrogationState {
	return &interrogationState{
		sync.NewCond(&sync.Mutex{}),
		false,
		Stop,
		nil,
		node,
		vs,
	}
}

/*
NewDebugger returns a new debugger object.
*/
func NewECALDebugger(globalVS parser.Scope) util.ECALDebugger {
	return &ecalDebugger{
		breakPoints:         make(map[string]bool),
		interrogationStates: make(map[uint64]*interrogationState),
		callStacks:          make(map[uint64][]*parser.ASTNode),
		sources:             make(map[string]bool),
		breakOnStart:        false,
		globalScope:         globalVS,
		lock:                &sync.RWMutex{},
	}
}

/*
HandleInput handles a given debug instruction from a console.
*/
func (ed *ecalDebugger) HandleInput(input string) (interface{}, error) {
	var res interface{}
	var err error

	args := strings.Fields(input)

	if cmd, ok := DebugCommandsMap[args[0]]; ok {
		if len(args) > 1 {
			res, err = cmd.Run(ed, args[1:])
		} else {
			res, err = cmd.Run(ed, nil)
		}
	} else {
		err = fmt.Errorf("Unknown command: %v", args[0])
	}

	return res, err
}

/*
Break on the start of the next execution.
*/
func (ed *ecalDebugger) BreakOnStart(flag bool) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakOnStart = flag
}

/*
VisitState is called for every state during the execution of a program.
*/
func (ed *ecalDebugger) VisitState(node *parser.ASTNode, vs parser.Scope, tid uint64) util.TraceableRuntimeError {

	ed.lock.RLock()
	_, ok := ed.callStacks[tid]
	ed.lock.RUnlock()

	if !ok {

		// Make the debugger aware of running threads

		ed.lock.Lock()
		ed.callStacks[tid] = make([]*parser.ASTNode, 0, 10)
		ed.lock.Unlock()
	}

	if node.Token != nil { // Statements are excluded here
		targetIdentifier := fmt.Sprintf("%v:%v", node.Token.Lsource, node.Token.Lline)

		ed.lock.RLock()
		is, ok := ed.interrogationStates[tid]
		_, sourceKnown := ed.sources[node.Token.Lsource]
		ed.lock.RUnlock()

		if !sourceKnown {
			ed.RecordSource(node.Token.Lsource)
		}

		if ok {

			// The thread is being interrogated

			switch is.cmd {
			case Resume:
				if is.node.Token.Lline != node.Token.Lline {

					// Remove the resume command once we are on a different line

					ed.lock.Lock()
					delete(ed.interrogationStates, tid)
					ed.lock.Unlock()

					return ed.VisitState(node, vs, tid)
				}
			case Stop, StepIn, StepOver:

				if is.node.Token.Lline != node.Token.Lline || is.cmd == Stop {
					is.node = node
					is.vs = vs
					is.running = false

					is.cond.L.Lock()
					is.cond.Wait()
					is.cond.L.Unlock()
				}
			}

		} else if active, ok := ed.breakPoints[targetIdentifier]; (ok && active) || ed.breakOnStart {

			// A globally defined breakpoint has been hit - note the position
			// in the thread specific map and wait

			is := newInterrogationState(node, vs)

			ed.lock.Lock()
			ed.breakOnStart = false
			ed.interrogationStates[tid] = is
			ed.lock.Unlock()

			is.cond.L.Lock()
			is.cond.Wait()
			is.cond.L.Unlock()
		}
	}

	return nil
}

/*
VisitStepInState is called before entering a function call.
*/
func (ed *ecalDebugger) VisitStepInState(node *parser.ASTNode, vs parser.Scope, tid uint64) util.TraceableRuntimeError {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	var err util.TraceableRuntimeError

	threadCallStack := ed.callStacks[tid]

	is, ok := ed.interrogationStates[tid]

	if ok {

		if is.cmd == Stop {

			// Special case a parameter of a function was resolved by another
			// function call - the debugger should stop before entering

			ed.lock.Unlock()
			err = ed.VisitState(node, vs, tid)
			ed.lock.Lock()
		}

		if err == nil {
			// The thread is being interrogated

			switch is.cmd {
			case StepIn:
				is.cmd = Stop
			case StepOver:
				is.cmd = StepOut
				is.stepOutStack = ed.callStacks[tid]
			}
		}
	}

	ed.callStacks[tid] = append(threadCallStack, node)

	return err
}

/*
VisitStepOutState is called after returning from a function call.
*/
func (ed *ecalDebugger) VisitStepOutState(node *parser.ASTNode, vs parser.Scope, tid uint64) util.TraceableRuntimeError {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	threadCallStack := ed.callStacks[tid]
	lastIndex := len(threadCallStack) - 1

	ok, cerr := threadCallStack[lastIndex].Equals(node, false) // Sanity check step in node must be the same as step out node
	errorutil.AssertTrue(ok,
		fmt.Sprintf("Unexpected callstack when stepping out - callstack: %v - funccall: %v - comparison error: %v",
			threadCallStack, node, cerr))

	ed.callStacks[tid] = threadCallStack[:lastIndex] // Remove the last item

	is, ok := ed.interrogationStates[tid]

	if ok {

		// The thread is being interrogated

		switch is.cmd {
		case StepOver, StepOut:

			if len(ed.callStacks[tid]) == len(is.stepOutStack) {
				is.cmd = Stop
			}
		}
	}

	return nil
}

/*
RecordSource records a code source.
*/
func (ed *ecalDebugger) RecordSource(source string) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.sources[source] = true
}

/*
SetBreakPoint sets a break point.
*/
func (ed *ecalDebugger) SetBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakPoints[fmt.Sprintf("%v:%v", source, line)] = true
}

/*
DisableBreakPoint disables a break point but keeps the code reference.
*/
func (ed *ecalDebugger) DisableBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	ed.breakPoints[fmt.Sprintf("%v:%v", source, line)] = false
}

/*
RemoveBreakPoint removes a break point.
*/
func (ed *ecalDebugger) RemoveBreakPoint(source string, line int) {
	ed.lock.Lock()
	defer ed.lock.Unlock()
	delete(ed.breakPoints, fmt.Sprintf("%v:%v", source, line))
}

/*
ExtractValue copies a value from a suspended thread into the
global variable scope.
*/
func (ed *ecalDebugger) ExtractValue(threadId uint64, varName string, destVarName string) error {
	if ed.globalScope == nil {
		return fmt.Errorf("Cannot access global scope")
	}

	err := fmt.Errorf("Cannot find suspended thread %v", threadId)

	ed.lock.Lock()
	defer ed.lock.Unlock()

	is, ok := ed.interrogationStates[threadId]

	if ok && !is.running {
		var val interface{}
		var ok bool

		if val, ok, err = is.vs.GetValue(varName); ok {
			err = ed.globalScope.SetValue(destVarName, val)
		} else if err == nil {
			err = fmt.Errorf("No such value %v", varName)
		}
	}

	return err
}

/*
InjectValue copies a value from an expression (using the global variable scope) into
a suspended thread.
*/
func (ed *ecalDebugger) InjectValue(threadId uint64, varName string, expression string) error {
	if ed.globalScope == nil {
		return fmt.Errorf("Cannot access global scope")
	}

	err := fmt.Errorf("Cannot find suspended thread %v", threadId)

	ed.lock.Lock()
	defer ed.lock.Unlock()

	is, ok := ed.interrogationStates[threadId]

	if ok && !is.running {
		var ast *parser.ASTNode
		var val interface{}

		// Eval expression

		ast, err = parser.ParseWithRuntime("InjectValueExpression", expression,
			NewECALRuntimeProvider("InjectValueExpression2", nil, nil))

		if err == nil {
			if err = ast.Runtime.Validate(); err == nil {

				ivs := scope.NewScopeWithParent("InjectValueExpressionScope", ed.globalScope)
				val, err = ast.Runtime.Eval(ivs, make(map[string]interface{}), 999)

				if err == nil {
					err = is.vs.SetValue(varName, val)
				}
			}
		}
	}

	return err
}

/*
Continue will continue a suspended thread.
*/
func (ed *ecalDebugger) Continue(threadId uint64, contType util.ContType) {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	if is, ok := ed.interrogationStates[threadId]; ok && !is.running {

		switch contType {
		case util.Resume:
			is.cmd = Resume
		case util.StepIn:
			is.cmd = StepIn
		case util.StepOver:
			is.cmd = StepOver
		case util.StepOut:
			is.cmd = StepOut
			stack := ed.callStacks[threadId]
			is.stepOutStack = stack[:len(stack)-1]
		}

		is.running = true

		is.cond.L.Lock()
		is.cond.Broadcast()
		is.cond.L.Unlock()
	}
}

/*
Status returns the current status of the debugger.
*/
func (ed *ecalDebugger) Status() interface{} {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	var sources []string

	threadStates := make(map[string]map[string]interface{})

	res := map[string]interface{}{
		"breakpoints":  ed.breakPoints,
		"breakonstart": ed.breakOnStart,
		"threads":      threadStates,
	}

	for k := range ed.sources {
		sources = append(sources, k)
	}
	res["sources"] = sources

	for k, v := range ed.callStacks {
		s := map[string]interface{}{
			"callStack": ed.prettyPrintCallStack(v),
		}

		if is, ok := ed.interrogationStates[k]; ok {
			s["threadRunning"] = is.running
		}

		threadStates[fmt.Sprint(k)] = s
	}

	return res
}

/*
Describe decribes a thread currently observed by the debugger.
*/
func (ed *ecalDebugger) Describe(threadId uint64) interface{} {
	ed.lock.RLock()
	defer ed.lock.RUnlock()

	var res map[string]interface{}

	threadCallStack, ok1 := ed.callStacks[threadId]

	if is, ok2 := ed.interrogationStates[threadId]; ok1 && ok2 {

		res = map[string]interface{}{
			"threadRunning": is.running,
			"callStack":     ed.prettyPrintCallStack(threadCallStack),
		}

		if !is.running {

			codeString, _ := parser.PrettyPrint(is.node)
			res["code"] = codeString
			res["node"] = is.node.ToJSONObject()
			res["vs"] = is.vs.ToJSONObject()
		}
	}

	return res
}

/*
Describe decribes a thread currently observed by the debugger.
*/
func (ed *ecalDebugger) prettyPrintCallStack(threadCallStack []*parser.ASTNode) []string {
	cs := []string{}
	for _, s := range threadCallStack {
		pp, _ := parser.PrettyPrint(s)
		cs = append(cs, fmt.Sprintf("%v (%v:%v)",
			pp, s.Token.Lsource, s.Token.Lline))
	}
	return cs
}
