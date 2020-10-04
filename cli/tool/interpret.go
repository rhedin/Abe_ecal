/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package tool

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"devt.de/krotik/common/fileutil"
	"devt.de/krotik/common/termutil"
	"devt.de/krotik/ecal/config"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/util"
)

/*
Interpret starts the ECAL code interpreter from a CLI application which
calls the interpret function as a sub executable. Starts an interactive console
if the interactive flag is set.
*/
func Interpret(interactive bool) error {
	var err error

	wd, _ := os.Getwd()

	idir := flag.String("dir", wd, "Root directory for ECAL interpreter")
	ilogFile := flag.String("logfile", "", "Log to a file")
	ilogLevel := flag.String("loglevel", "Info", "Logging level (Debug, Info, Error)")
	showHelp := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Println()
		if !interactive {
			fmt.Println(fmt.Sprintf("Usage of %s run [options] <file>", os.Args[0]))
		} else {
			fmt.Println(fmt.Sprintf("Usage of %s [options]", os.Args[0]))
		}
		fmt.Println()
		flag.PrintDefaults()
		fmt.Println()
	}

	if len(os.Args) > 2 {
		flag.CommandLine.Parse(os.Args[2:])

		if *showHelp {
			flag.Usage()
			return nil
		}
	}

	var clt termutil.ConsoleLineTerminal
	var logger util.Logger

	clt, err = termutil.NewConsoleLineTerminal(os.Stdout)

	// Create the logger

	if err == nil {

		// Check if we should log to a file

		if ilogFile != nil && *ilogFile != "" {
			var logWriter io.Writer
			logFileRollover := fileutil.SizeBasedRolloverCondition(1000000) // Each file can be up to a megabyte
			logWriter, err = fileutil.NewMultiFileBuffer(*ilogFile, fileutil.ConsecutiveNumberIterator(10), logFileRollover)
			logger = util.NewBufferLogger(logWriter)

		} else {

			// Log to the console by default

			logger = util.NewStdOutLogger()
		}

		// Set the log level

		if err == nil {
			if ilogLevel != nil && *ilogLevel != "" {
				logger, err = util.NewLogLevelLogger(logger, *ilogLevel)
			}

		}
	}

	// Get the import locator

	importLocator := &util.FileImportLocator{Root: *idir}

	if err == nil {

		name := "ECAL console"

		// Create interpreter

		erp := interpreter.NewECALRuntimeProvider(name, importLocator, logger)

		// Create global variable scope

		vs := scope.NewScope(scope.GlobalScope)

		// TODO Execute file

		if interactive {

			// Preload stdlib packages and functions

			// TODO stdlibPackages, stdlibConst, stdlibFuncs := stdlib.GetStdlibSymbols()

			// Drop into interactive shell

			if err == nil {
				isExitLine := func(s string) bool {
					return s == "exit" || s == "q" || s == "quit" || s == "bye" || s == "\x04"
				}

				// Add history functionality without file persistence

				clt, err = termutil.AddHistoryMixin(clt, "",
					func(s string) bool {
						return isExitLine(s)
					})

				if err == nil {

					if err = clt.StartTerm(); err == nil {
						var line string

						defer clt.StopTerm()

						fmt.Println(fmt.Sprintf("ECAL %v", config.ProductVersion))
						fmt.Println("Type 'q' or 'quit' to exit the shell and '?' to get help")

						line, err = clt.NextLine()
						for err == nil && !isExitLine(line) {
							trimmedLine := strings.TrimSpace(line)

							// Process the entered line

							if line == "?" {

								// Show help

								clt.WriteString(fmt.Sprintf("ECAL %v\n", config.ProductVersion))
								clt.WriteString(fmt.Sprintf("\n"))
								clt.WriteString(fmt.Sprintf("Console supports all normal ECAL statements and the following special commands:\n"))
								clt.WriteString(fmt.Sprintf("\n"))
								clt.WriteString(fmt.Sprintf("    @syms - List all available inbuild functions and available stdlib packages of ECAL.\n"))
								clt.WriteString(fmt.Sprintf("    @stdl - List all available constants and functions of a stdlib package.\n"))
								clt.WriteString(fmt.Sprintf("    @lk   - Do a full text search through all docstrings.\n"))
								clt.WriteString(fmt.Sprintf("\n"))

							} else if strings.HasPrefix(trimmedLine, "@syms") {
								args := strings.Split(trimmedLine, " ")[1:]

								// TODO Implement

								clt.WriteString(fmt.Sprint("syms:", args))

							} else if line == "!reset" {

							} else {
								var ierr error
								var ast *parser.ASTNode
								var res interface{}

								if ast, ierr = parser.ParseWithRuntime("console input", line, erp); ierr == nil {

									if ierr = ast.Runtime.Validate(); ierr == nil {

										if res, ierr = ast.Runtime.Eval(vs, make(map[string]interface{})); ierr == nil && res != nil {
											clt.WriteString(fmt.Sprintln(res))
										}
									}
								}

								if ierr != nil {
									clt.WriteString(fmt.Sprintln(ierr.Error()))
								}
							}

							line, err = clt.NextLine()
						}
					}
				}
			}
		}
	}

	return err
}
