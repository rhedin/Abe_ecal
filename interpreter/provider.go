/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

// TODO:
// Import resolve
// Event function: event
// Context supporting final
// Event handling

package interpreter

import (
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/util"
)

/*
ecalRuntimeNew is used to instantiate ECAL runtime components.
*/
type ecalRuntimeNew func(*ECALRuntimeProvider, *parser.ASTNode) parser.Runtime

/*
providerMap contains the mapping of AST nodes to runtime components for ECAL ASTs.
*/
var providerMap = map[string]ecalRuntimeNew{

	parser.NodeEOF: invalidRuntimeInst,

	parser.NodeSTRING:     stringValueRuntimeInst, // String constant
	parser.NodeNUMBER:     numberValueRuntimeInst, // Number constant
	parser.NodeIDENTIFIER: identifierRuntimeInst,  // Idendifier

	// Constructed tokens

	parser.NodeSTATEMENTS: statementsRuntimeInst, // List of statements
	parser.NodeFUNCCALL:   voidRuntimeInst,       // Function call
	parser.NodeCOMPACCESS: voidRuntimeInst,       // Composition structure access
	parser.NodeLIST:       listValueRuntimeInst,  // List value
	parser.NodeMAP:        mapValueRuntimeInst,   // Map value
	parser.NodePARAMS:     voidRuntimeInst,       // Function parameters
	parser.NodeGUARD:      guardRuntimeInst,      // Guard expressions for conditional statements

	// Condition operators

	parser.NodeGEQ: greaterequalOpRuntimeInst,
	parser.NodeLEQ: lessequalOpRuntimeInst,
	parser.NodeNEQ: notequalOpRuntimeInst,
	parser.NodeEQ:  equalOpRuntimeInst,
	parser.NodeGT:  greaterOpRuntimeInst,
	parser.NodeLT:  lessOpRuntimeInst,

	// Separators

	parser.NodeKVP:    voidRuntimeInst, // Key-value pair
	parser.NodePRESET: voidRuntimeInst, // Preset value

	// Arithmetic operators

	parser.NodePLUS: plusOpRuntimeInst,

	parser.NodeMINUS:  minusOpRuntimeInst,
	parser.NodeTIMES:  timesOpRuntimeInst,
	parser.NodeDIV:    divOpRuntimeInst,
	parser.NodeMODINT: modintOpRuntimeInst,
	parser.NodeDIVINT: divintOpRuntimeInst,

	// Assignment statement

	parser.NodeASSIGN: assignmentRuntimeInst,
	/*

		// Import statement

		parser.NodeIMPORT

		// Sink definition

		parser.NodeSINK
		parser.NodeKINDMATCH
		parser.NodeSCOPEMATCH
		parser.NodeSTATEMATCH
		parser.NodePRIORITY
		parser.NodeSUPPRESSES
	*/
	// Function definition

	parser.NodeFUNC:   funcRuntimeInst,
	parser.NodeRETURN: returnRuntimeInst,

	// Boolean operators

	parser.NodeOR:  orOpRuntimeInst,
	parser.NodeAND: andOpRuntimeInst,
	parser.NodeNOT: notOpRuntimeInst,

	// Condition operators

	parser.NodeLIKE:      likeOpRuntimeInst,
	parser.NodeIN:        inOpRuntimeInst,
	parser.NodeHASPREFIX: beginswithOpRuntimeInst,
	parser.NodeHASSUFFIX: endswithOpRuntimeInst,
	parser.NodeNOTIN:     notinOpRuntimeInst,

	// Constant terminals

	parser.NodeFALSE: falseRuntimeInst,
	parser.NodeTRUE:  trueRuntimeInst,
	parser.NodeNULL:  nullRuntimeInst,

	// Conditional statements

	parser.NodeIF: ifRuntimeInst,

	// Loop statements

	parser.NodeLOOP:     loopRuntimeInst,
	parser.NodeBREAK:    breakRuntimeInst,
	parser.NodeCONTINUE: continueRuntimeInst,
}

/*
ECALRuntimeProvider is the factory object producing runtime objects for ECAL ASTs.
*/
type ECALRuntimeProvider struct {
	Name string // Name to identify the input
}

/*
NewECALRuntimeProvider returns a new instance of a ECAL runtime provider.
*/
func NewECALRuntimeProvider(name string) *ECALRuntimeProvider {
	return &ECALRuntimeProvider{name}
}

/*
Runtime returns a runtime component for a given ASTNode.
*/
func (erp *ECALRuntimeProvider) Runtime(node *parser.ASTNode) parser.Runtime {

	if instFunc, ok := providerMap[node.Name]; ok {
		return instFunc(erp, node)
	}

	return invalidRuntimeInst(erp, node)
}

/*
NewRuntimeError creates a new RuntimeError object.
*/
func (erp *ECALRuntimeProvider) NewRuntimeError(t error, d string, node *parser.ASTNode) error {
	return util.NewRuntimeError(erp.Name, t, d, node)
}
