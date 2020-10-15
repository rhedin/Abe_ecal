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

/*
OutputTerminal is a generic output terminal which can write strings.
*/
type OutputTerminal interface {

	/*
	   WriteString write a string on this terminal.
	*/
	WriteString(s string)
}

/*
Debugger is a debugging object which can be used to inspect and modify a running
ECAL environment.
*/
type Debugger interface {
}
