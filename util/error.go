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
Package util contains utility definitions and functions for the event condition language ECAL.
*/
package util

import (
	"errors"
	"fmt"

	"devt.de/krotik/ecal/parser"
)

/*
RuntimeError is a runtime related error.
*/
type RuntimeError struct {
	Source string          // Name of the source which was given to the parser
	Type   error           // Error type (to be used for equal checks)
	Detail string          // Details of this error
	Node   *parser.ASTNode // AST Node where the error occurred
	Line   int             // Line of the error
	Pos    int             // Position of the error
}

/*
Runtime related error types.
*/
var (
	ErrRuntimeError     = errors.New("Runtime error")
	ErrUnknownConstruct = errors.New("Unknown construct")
	ErrInvalidConstruct = errors.New("Invalid construct")
	ErrInvalidState     = errors.New("Invalid state")
	ErrVarAccess        = errors.New("Cannot access variable")
	ErrNotANumber       = errors.New("Operand is not a number")
	ErrNotABoolean      = errors.New("Operand is not a boolean")
	ErrNotAList         = errors.New("Operand is not a list")
	ErrNotAMap          = errors.New("Operand is not a map")
	ErrNotAListOrMap    = errors.New("Operand is not a list nor a map")

	// ErrReturn is not an error. It is used to return when executing a function
	ErrReturn = errors.New("*** return ***")

	// Error codes for loop operations
	ErrIsIterator        = errors.New("Function is an iterator")
	ErrEndOfIteration    = errors.New("End of iteration was reached")
	ErrContinueIteration = errors.New("End of iteration step - Continue iteration")
)

/*
NewRuntimeError creates a new RuntimeError object.
*/
func NewRuntimeError(source string, t error, d string, node *parser.ASTNode) error {
	if node.Token != nil {
		return &RuntimeError{source, t, d, node, node.Token.Lline, node.Token.Lpos}
	}
	return &RuntimeError{source, t, d, node, 0, 0}
}

/*
Error returns a human-readable string representation of this error.
*/
func (re *RuntimeError) Error() string {
	ret := fmt.Sprintf("ECAL error in %s: %v (%v)", re.Source, re.Type, re.Detail)

	if re.Line != 0 {

		// Add line if available

		ret = fmt.Sprintf("%s (Line:%d Pos:%d)", ret, re.Line, re.Pos)
	}

	return ret
}
