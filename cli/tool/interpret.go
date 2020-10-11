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
	"io/ioutil"
	"os"
	"strings"

	"devt.de/krotik/common/fileutil"
	"devt.de/krotik/common/stringutil"
	"devt.de/krotik/common/termutil"
	"devt.de/krotik/ecal/config"
	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/parser"
	"devt.de/krotik/ecal/scope"
	"devt.de/krotik/ecal/stdlib"
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

	if interactive {
		fmt.Println(fmt.Sprintf("ECAL %v", config.ProductVersion))
	}

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
				if logger, err = util.NewLogLevelLogger(logger, *ilogLevel); err == nil && interactive {
					fmt.Print(fmt.Sprintf("Log level: %v - ", logger.(*util.LogLevelLogger).Level()))
				}
			}
		}
	}

	if err == nil {

		// Get the import locator

		if interactive {
			fmt.Println(fmt.Sprintf("Root directory: %v", *idir))
		}

		importLocator := &util.FileImportLocator{Root: *idir}

		name := "console"

		// Create interpreter

		erp := interpreter.NewECALRuntimeProvider(name, importLocator, logger)

		// Create global variable scope

		vs := scope.NewScope(scope.GlobalScope)

		// Execute file if given

		if cargs := flag.Args(); len(cargs) > 0 {
			var ast *parser.ASTNode
			var initFile []byte

			initFileName := flag.Arg(0)
			initFile, err = ioutil.ReadFile(initFileName)

			if ast, err = parser.ParseWithRuntime(initFileName, string(initFile), erp); err == nil {
				if err = ast.Runtime.Validate(); err == nil {
					_, err = ast.Runtime.Eval(vs, make(map[string]interface{}))
				}
			}
		}

		if err == nil {

			if interactive {

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
									clt.WriteString(fmt.Sprintf("    @sym [glob] - List all available inbuild functions and available stdlib packages of ECAL.\n"))
									clt.WriteString(fmt.Sprintf("    @std <package> [glob] - List all available constants and functions of a stdlib package.\n"))
									clt.WriteString(fmt.Sprintf("\n"))
									clt.WriteString(fmt.Sprintf("Add an argument after a list command to do a full text search. The search string should be in glob format.\n"))

								} else if strings.HasPrefix(trimmedLine, "@sym") {
									displaySymbols(clt, strings.Split(trimmedLine, " ")[1:])

								} else if strings.HasPrefix(trimmedLine, "@std") {
									displayPackage(clt, strings.Split(trimmedLine, " ")[1:])

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
	}

	return err
}

/*
displaySymbols lists all available inbuild functions and available stdlib packages of ECAL.
*/
func displaySymbols(clt termutil.ConsoleLineTerminal, args []string) {

	tabData := []string{"Inbuild function", "Description"}

	for name, f := range interpreter.InbuildFuncMap {
		ds, _ := f.DocString()

		if len(args) > 0 && !matchesFulltextSearch(clt, fmt.Sprintf("%v %v", name, ds), args[0]) {
			continue
		}

		tabData = fillTableRow(tabData, name, ds)
	}

	if len(tabData) > 2 {
		clt.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}

	packageNames, _, _ := stdlib.GetStdlibSymbols()

	tabData = []string{"Package name", "Description"}

	for _, p := range packageNames {
		ps, _ := stdlib.GetPkgDocString(p)

		if len(args) > 0 && !matchesFulltextSearch(clt, fmt.Sprintf("%v %v", p, ps), args[0]) {
			continue
		}

		tabData = fillTableRow(tabData, p, ps)
	}

	if len(tabData) > 2 {
		clt.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}
}

/*
displayPackage list all available constants and functions of a stdlib package.
*/
func displayPackage(clt termutil.ConsoleLineTerminal, args []string) {

	_, constSymbols, funcSymbols := stdlib.GetStdlibSymbols()

	tabData := []string{"Constant", "Value"}

	for _, s := range constSymbols {

		if len(args) > 0 && !strings.HasPrefix(s, args[0]) {
			continue
		}

		val, _ := stdlib.GetStdlibConst(s)

		tabData = fillTableRow(tabData, s, fmt.Sprint(val))
	}

	if len(tabData) > 2 {
		clt.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}

	tabData = []string{"Function", "Description"}

	for _, f := range funcSymbols {
		if len(args) > 0 && !strings.HasPrefix(f, args[0]) {
			continue
		}

		fObj, _ := stdlib.GetStdlibFunc(f)
		fDoc, _ := fObj.DocString()

		fDoc = strings.Replace(fDoc, "\n", " ", -1)
		fDoc = strings.Replace(fDoc, "\t", " ", -1)

		if len(args) > 1 && !matchesFulltextSearch(clt, fmt.Sprintf("%v %v", f, fDoc), args[1]) {
			continue
		}

		tabData = fillTableRow(tabData, f, fDoc)
	}

	if len(tabData) > 2 {
		clt.WriteString(stringutil.PrintGraphicStringTable(tabData, 2, 1,
			stringutil.SingleDoubleLineTable))
	}
}
