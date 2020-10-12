/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"devt.de/krotik/ecal/cli/tool"
	"devt.de/krotik/ecal/config"
)

/*
TODO:
- create executable binary (pack into single binary)
- debug server support (vscode)
*/

func main() {

	// Initialize the default command line parser

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)

	// Define default usage message

	flag.Usage = func() {

		// Print usage for tool selection

		fmt.Println(fmt.Sprintf("Usage of %s <tool>", os.Args[0]))
		fmt.Println()
		fmt.Println(fmt.Sprintf("ECAL %v - Event Condition Action Language", config.ProductVersion))
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println()
		fmt.Println("    console   Interactive console (default)")
		fmt.Println("    run       Execute ECAL code")
		fmt.Println("    debug     Run a debug server")
		fmt.Println("    pack      Create a single executable from ECAL code")
		fmt.Println()
		fmt.Println(fmt.Sprintf("Use %s <command> -help for more information about a given command.", os.Args[0]))
		fmt.Println()
	}

	// Parse the command bit

	err := flag.CommandLine.Parse(os.Args[1:])

	if len(flag.Args()) > 0 {

		arg := flag.Args()[0]

		if arg == "console" {
			err = tool.Interpret(true)
		} else if arg == "run" {
			err = tool.Interpret(false)
		} else {
			flag.Usage()
		}

	} else if err == nil {

		err = tool.Interpret(true)
	}

	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
	}
}
