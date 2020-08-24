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

import "devt.de/krotik/ecal/parser"

/*
Processor models a top level execution instance for ECAL.
*/
type Processor interface {
}

/*
ECALImportLocator is used to resolve imports.
*/
type ECALImportLocator interface {

	/*
		Resolve a given import path and parse the imported file into an AST.
	*/
	Resolve(path string) (string, error)
}

/*
ECALFunction models a callable function in ECAL.
*/
type ECALFunction interface {

	/*
		Run executes this function. The envirnment provides a unique instanceID for
		every code location in the running code, the variable scope of the function,
		an instance state which can be used in combinartion with the instanceID
		to store instance specific state (e.g. for iterator functions) and a list
		of argument values which were passed to the function by the calling code.
	*/
	Run(instanceID string, vs parser.Scope, is map[string]interface{}, args []interface{}) (interface{}, error)

	/*
	   DocString returns a descriptive text about this function.
	*/
	DocString() (string, error)
}

/*
Logger is required external object to which the interpreter releases its log messages.
*/
type Logger interface {

	/*
	   LogError adds a new error log message.
	*/
	LogError(v ...interface{})

	/*
	   LogInfo adds a new info log message.
	*/
	LogInfo(v ...interface{})

	/*
	   LogDebug adds a new debug log message.
	*/
	LogDebug(v ...interface{})
}
