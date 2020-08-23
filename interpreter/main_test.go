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
	"flag"
	"fmt"
	"os"
	"testing"

	"devt.de/krotik/common/datautil"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
)

// Main function for all tests in this package

func TestMain(m *testing.M) {
	flag.Parse()

	// Run the tests

	res := m.Run()

	// Check if all nodes have been tested

	for n := range providerMap {
		if _, ok := usedNodes[n]; !ok {
			fmt.Println("Not tested node: ", n)
		}
	}

	os.Exit(res)
}

// Used nodes map which is filled during unit testing. Prefilled only with nodes
// which should not be encountered in ASTs.
//
var usedNodes = map[string]bool{
	parser.NodeEOF: true,
}

func UnitTestEval(input string, vs parser.Scope) (interface{}, error) {
	return UnitTestEvalAndAST(input, vs, "")
}

func UnitTestEvalAndAST(input string, vs parser.Scope, expectedAST string) (interface{}, error) {
	var traverseAST func(n *parser.ASTNode)

	traverseAST = func(n *parser.ASTNode) {
		if n.Name == "" {
			panic(fmt.Sprintf("Node found with empty string name: %s", n))
		}

		usedNodes[n.Name] = true
		for _, cn := range n.Children {
			traverseAST(cn)
		}
	}

	// Parse the input

	ast, err := parser.ParseWithRuntime("ECALEvalTest", input, NewECALRuntimeProvider("ECALTestRuntime"))
	if err != nil {
		return nil, err
	}

	traverseAST(ast)

	if expectedAST != "" && ast.String() != expectedAST {
		return nil, fmt.Errorf("Unexpected AST result:\n%v", ast.String())
	}

	// Validate input

	if err := ast.Runtime.Validate(); err != nil {
		return nil, err
	}

	if vs == nil {
		vs = scope.NewScope(scope.GlobalScope)
	}

	return ast.Runtime.Eval(vs, make(map[string]interface{}))
}

/*
addLogFunction adds a simple log function to a given Scope.
*/
func addLogFunction(vs parser.Scope) *datautil.RingBuffer {
	buf := datautil.NewRingBuffer(20)
	vs.SetValue("log", &TestLogger{buf})
	return buf
}

/*
TestLogger is a simple logger function which can be added to tess.
*/
type TestLogger struct {
	buf *datautil.RingBuffer
}

func (tl *TestLogger) Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error) {
	tl.buf.Add(fmt.Sprint(args...))
	return nil, nil
}

func (tl *TestLogger) DocString() (string, error) {
	return "testlogger docstring", nil
}

func (tl *TestLogger) String() string {
	return "TestLogger"
}

func (tl *TestLogger) MarshalJSON() ([]byte, error) {
	return []byte(tl.String()), nil
}
