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
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"devt.de/krotik/ecal/interpreter"
	"devt.de/krotik/ecal/util"
)

/*
CLIDebugInterpreter is a commandline interpreter with debug capabilities for ECAL.
*/
type CLIDebugInterpreter struct {
	*CLIInterpreter

	// Parameter these can either be set programmatically or via CLI args

	DebugServerAddr *string // Debug server address
	RunDebugServer  *bool   // Run a debug server
	Interactive     *bool   // Flag if the interpreter should open a console in the current tty.
}

/*
NewCLIDebugInterpreter wraps an existing CLIInterpreter object and adds capabilities.
*/
func NewCLIDebugInterpreter(i *CLIInterpreter) *CLIDebugInterpreter {
	return &CLIDebugInterpreter{i, nil, nil, nil}
}

/*
ParseArgs parses the command line arguments.
*/
func (i *CLIDebugInterpreter) ParseArgs() bool {

	if i.Interactive != nil {
		return false
	}

	i.DebugServerAddr = flag.String("serveraddr", "localhost:33274", "Debug server address") // Think BERTA
	i.RunDebugServer = flag.Bool("server", false, "Run a debug server")
	i.Interactive = flag.Bool("interactive", true, "Run interactive console")

	return i.CLIInterpreter.ParseArgs()
}

/*
Interpret starts the ECAL code interpreter with debug capabilities.
*/
func (i *CLIDebugInterpreter) Interpret() error {

	if i.ParseArgs() {
		return nil
	}

	i.CLIInterpreter.CustomWelcomeMessage = "Running in debug mode - "
	if *i.RunDebugServer {
		i.CLIInterpreter.CustomWelcomeMessage += fmt.Sprintf("with debug server on %v - ", *i.DebugServerAddr)
	}
	i.CLIInterpreter.CustomWelcomeMessage += "prefix debug commands with ##"
	i.CustomHelpString = "    @dbg [glob] - List all available debug commands.\n"

	err := i.CreateRuntimeProvider("debug console")

	if err == nil {

		if *i.RunDebugServer {
			debugServer := &debugTelnetServer{*i.DebugServerAddr, "ECALDebugServer: ",
				nil, true, i, i.RuntimeProvider.Logger}
			go debugServer.Run()
			time.Sleep(500 * time.Millisecond) // Too lazy to do proper signalling
			defer func() {
				if debugServer.listener != nil {
					debugServer.listen = false
					debugServer.listener.Close() // Attempt to cleanup
				}
			}()
		}

		err = i.CLIInterpreter.Interpret(*i.Interactive)
	}

	return err
}

/*
debugTelnetServer is a simple telnet server to send and receive debug data.
*/
type debugTelnetServer struct {
	address     string
	logPrefix   string
	listener    *net.TCPListener
	listen      bool
	interpreter *CLIDebugInterpreter
	logger      util.Logger
}

/*
Run runs the debug server.
*/
func (s *debugTelnetServer) Run() {
	tcpaddr, err := net.ResolveTCPAddr("tcp", s.address)

	if err == nil {

		if s.listener, err = net.ListenTCP("tcp", tcpaddr); err == nil {

			s.logger.LogInfo(s.logPrefix,
				"Running Debug Server on ", tcpaddr.String())

			for s.listen {
				var conn net.Conn

				if conn, err = s.listener.Accept(); err == nil {

					go s.HandleConnection(conn)

				} else if s.listen {
					s.logger.LogError(s.logPrefix, err)
					err = nil
				}
			}
		}
	}

	if s.listen && err != nil {
		s.logger.LogError(s.logPrefix, "Could not start debug server - ", err)
	}
}

/*
HandleConnection handles an incoming connection.
*/
func (s *debugTelnetServer) HandleConnection(conn net.Conn) {
	tid := s.interpreter.RuntimeProvider.NewThreadID()
	inputReader := bufio.NewReader(conn)
	outputTerminal := interpreter.OutputTerminal(&bufioWriterShim{bufio.NewWriter(conn)})

	line := ""

	s.logger.LogDebug(s.logPrefix, "Connect ", conn.RemoteAddr())

	for {
		var err error

		if line, err = inputReader.ReadString('\n'); err == nil {
			line = strings.TrimSpace(line)

			if line == "exit" || line == "q" || line == "quit" || line == "bye" || line == "\x04" {
				break
			}

			s.interpreter.HandleInput(outputTerminal, line, tid)
		}

		if err != nil {
			s.logger.LogDebug(s.logPrefix, "Disconnect ", conn.RemoteAddr(), " - ", err)
			break
		}
	}

	conn.Close()
}

/*
bufioWriterShim is a shim to allow a bufio.Writer to be used as an OutputTerminal.
*/
type bufioWriterShim struct {
	writer *bufio.Writer
}

/*
WriteString write a string to the writer.
*/
func (shim *bufioWriterShim) WriteString(s string) {
	shim.writer.WriteString(s)
	shim.writer.Flush()
}
