/**
 * Debug Adapter for VS Code to support the ECAL debugger.
 *
 * See the debugger extension guide:
 * https://code.visualstudio.com/api/extension-guides/debugger-extension
 */

import { logger, Logger, LoggingDebugSession, InitializedEvent, Thread } from 'vscode-debugadapter'
import { DebugProtocol } from 'vscode-debugprotocol'
import { WaitGroup } from '@jpwilliams/waitgroup'

/**
 * ECALDebugArguments are the arguments which VSCode can pass to the debug adapter.
 * This defines the parameter which a VSCode instance using the ECAL extention can pass to the
 * debug adapter from a lauch configuration ('.vscode/launch.json') in a project folder.
 */
interface ECALDebugArguments extends DebugProtocol.LaunchRequestArguments {

    dir: string; // Root directory for ECAL interpreter
    serverURL: string // URL of the ECAL debug server
    executeOnEntry?: boolean; // Flag if the debugged script should be executed when the debug session is started
    trace?: boolean; // Flag to enable verbose logging of the adapter protocol
}

/**
 * Debug adapter implementation.
 *
 * Uses: https://github.com/microsoft/vscode-debugadapter-node
 *
 * See the Debug Adapter Protocol (DAP) documentation:
 * https://microsoft.github.io/debug-adapter-protocol/overview#How_it_works
 */
export class ECALDebugSession extends LoggingDebugSession {
    /**
     * WaitGroup to wait the finish of the configuration sequence
     */
    private wgConfig = new WaitGroup();

    /**
     * Create a new debug adapter which is used for one debug session.
     */
    public constructor () {
      super('mock-debug.txt')
      console.error('##### constructor')
      // Lines and columns start at 1
      this.setDebuggerLinesStartAt1(true)
      this.setDebuggerColumnsStartAt1(true)

      // Increment the config WaitGroup counter for configurationDoneRequest()
      this.wgConfig.add(1)
    }

    /**
     * Called as the first step in the DAP. The client (e.g. VSCode)
     * interrogates the debug adapter on the features which it provides.
     */
    protected initializeRequest (response: DebugProtocol.InitializeResponse, args: DebugProtocol.InitializeRequestArguments): void {
      console.log('##### initializeRequest:', args)

      response.body = response.body || {}

      // The adapter implements the configurationDoneRequest.
      response.body.supportsConfigurationDoneRequest = true

      this.sendResponse(response)

      this.sendEvent(new InitializedEvent())
    }

    /**
     * Called as part of the "configuration Done" step in the DAP. The client (e.g. VSCode) has
     * finished the initialization of the debug adapter.
     */
    protected configurationDoneRequest (response: DebugProtocol.ConfigurationDoneResponse, args: DebugProtocol.ConfigurationDoneArguments): void {
      console.error('##### configurationDoneRequest')

      super.configurationDoneRequest(response, args)
      this.wgConfig.done()
    }

    /**
     * The client (e.g. VSCode) asks the debug adapter to start the debuggee communication.
     */
    protected async launchRequest (response: DebugProtocol.LaunchResponse, args: ECALDebugArguments) {
      console.error('##### launchRequest:', args)

      // Setup logging either verbose or just on errors

      logger.setup(args.trace ? Logger.LogLevel.Verbose : Logger.LogLevel.Error, false)

      await this.wgConfig.wait() // Wait for configuration sequence to finish

      this.sendResponse(response)
    }

    protected async setBreakPointsRequest (response: DebugProtocol.SetBreakpointsResponse, args: DebugProtocol.SetBreakpointsArguments): Promise<void> {
      console.error('##### setBreakPointsRequest:', args)

      response.body = {
        breakpoints: []
      }
      this.sendResponse(response)
    }

    protected threadsRequest (response: DebugProtocol.ThreadsResponse): void {
      console.error('##### threadsRequest')

      // runtime supports no threads so just return a default thread.
      response.body = {
        threads: [
          new Thread(1, 'thread 1')
        ]
      }
      this.sendResponse(response)
    }

    protected stackTraceRequest (response: DebugProtocol.StackTraceResponse, args: DebugProtocol.StackTraceArguments): void {
      console.error('##### stackTraceRequest:', args)

      response.body = {
        stackFrames: []
      }
      this.sendResponse(response)
    }

    protected scopesRequest (response: DebugProtocol.ScopesResponse, args: DebugProtocol.ScopesArguments): void {
      console.error('##### scopesRequest:', args)

      response.body = {
        scopes: []
      }
      this.sendResponse(response)
    }

    protected async variablesRequest (response: DebugProtocol.VariablesResponse, args: DebugProtocol.VariablesArguments, request?: DebugProtocol.Request) {
      console.error('##### variablesRequest', args, request)

      response.body = {
        variables: []
      }
      this.sendResponse(response)
    }

    protected continueRequest (response: DebugProtocol.ContinueResponse, args: DebugProtocol.ContinueArguments): void {
      console.error('##### continueRequest', args)
      this.sendResponse(response)
    }

    protected reverseContinueRequest (response: DebugProtocol.ReverseContinueResponse, args: DebugProtocol.ReverseContinueArguments): void {
      console.error('##### reverseContinueRequest', args)
      this.sendResponse(response)
    }

    protected nextRequest (response: DebugProtocol.NextResponse, args: DebugProtocol.NextArguments): void {
      console.error('##### nextRequest', args)
      this.sendResponse(response)
    }

    protected stepBackRequest (response: DebugProtocol.StepBackResponse, args: DebugProtocol.StepBackArguments): void {
      console.error('##### stepBackRequest', args)
      this.sendResponse(response)
    }

    protected stepInTargetsRequest (response: DebugProtocol.StepInTargetsResponse, args: DebugProtocol.StepInTargetsArguments) {
      console.error('##### stepInTargetsRequest', args)
      response.body = {
        targets: []
      }
      this.sendResponse(response)
    }

    protected stepInRequest (response: DebugProtocol.StepInResponse, args: DebugProtocol.StepInArguments): void {
      console.error('##### stepInRequest', args)
      this.sendResponse(response)
    }

    protected stepOutRequest (response: DebugProtocol.StepOutResponse, args: DebugProtocol.StepOutArguments): void {
      console.error('##### stepOutRequest', args)
      this.sendResponse(response)
    }

    protected async evaluateRequest (response: DebugProtocol.EvaluateResponse, args: DebugProtocol.EvaluateArguments): Promise<void> {
      console.error('##### evaluateRequest', args)

      response.body = {
        result: 'evaluate',
        variablesReference: 0
      }
      this.sendResponse(response)
    }

    protected dataBreakpointInfoRequest (response: DebugProtocol.DataBreakpointInfoResponse, args: DebugProtocol.DataBreakpointInfoArguments): void {
      console.error('##### dataBreakpointInfoRequest', args)

      response.body = {
        dataId: null,
        description: 'cannot break on data access',
        accessTypes: undefined,
        canPersist: false
      }

      this.sendResponse(response)
    }

    protected setDataBreakpointsRequest (response: DebugProtocol.SetDataBreakpointsResponse, args: DebugProtocol.SetDataBreakpointsArguments): void {
      console.error('##### setDataBreakpointsRequest', args)

      response.body = {
        breakpoints: []
      }

      this.sendResponse(response)
    }

    protected completionsRequest (response: DebugProtocol.CompletionsResponse, args: DebugProtocol.CompletionsArguments): void {
      console.error('##### completionsRequest', args)

      response.body = {
        targets: [
          {
            label: 'item 10',
            sortText: '10'
          },
          {
            label: 'item 1',
            sortText: '01'
          },
          {
            label: 'item 2',
            sortText: '02'
          },
          {
            label: 'array[]',
            selectionStart: 6,
            sortText: '03'
          },
          {
            label: 'func(arg)',
            selectionStart: 5,
            selectionLength: 3,
            sortText: '04'
          }
        ]
      }
      this.sendResponse(response)
    }

    protected cancelRequest (response: DebugProtocol.CancelResponse, args: DebugProtocol.CancelArguments) {
      console.error('##### cancelRequest', args)
      this.sendResponse(response)
    }

    protected customRequest (command: string, response: DebugProtocol.Response, args: any) {
      console.error('##### customRequest', args)

      if (command === 'toggleFormatting') {
        this.sendResponse(response)
      } else {
        super.customRequest(command, response, args)
      }
    }
}
