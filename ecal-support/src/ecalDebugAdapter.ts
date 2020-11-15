/**
 * Debug Adapter for VS Code to support the ECAL debugger.
 *
 * See the debugger extension guide:
 * https://code.visualstudio.com/api/extension-guides/debugger-extension
 */

import {
  logger,
  Logger,
  LoggingDebugSession,
  Thread,
  Source,
  Breakpoint,
  InitializedEvent,
  BreakpointEvent,
  StoppedEvent,
} from "vscode-debugadapter";
import { DebugProtocol } from "vscode-debugprotocol";
import { WaitGroup } from "@jpwilliams/waitgroup";
import { ECALDebugClient } from "./ecalDebugClient";
import * as vscode from "vscode";
import { ClientBreakEvent, DebugStatus } from "./types";

/**
 * ECALDebugArguments are the arguments which VSCode can pass to the debug adapter.
 * This defines the parameter which a VSCode instance using the ECAL extention can pass to the
 * debug adapter from a lauch configuration ('.vscode/launch.json') in a project folder.
 */
interface ECALDebugArguments extends DebugProtocol.LaunchRequestArguments {
  host: string; // Host of the ECAL debug server
  port: number; // Port of the ECAL debug server
  dir: string; // Root directory for ECAL interpreter
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
   * Client to the ECAL debug server
   */
  private client: ECALDebugClient;

  /**
   * Output channel for log messages
   */
  private extout: vscode.OutputChannel = vscode.window.createOutputChannel(
    "ECAL Debug Session"
  );

  /**
   * WaitGroup to wait the finish of the configuration sequence
   */
  private wgConfig = new WaitGroup();

  private config: ECALDebugArguments = {} as ECALDebugArguments;

  private unconfirmedBreakpoints: DebugProtocol.Breakpoint[] = [];

  private bpCount: number = 1;
  private bpIds: Record<string, number> = {};

  public sendEvent(event: DebugProtocol.Event): void {
    super.sendEvent(event);
    console.error("#### Sending event:", event);
  }

  /**
   * Create a new debug adapter which is used for one debug session.
   */
  public constructor() {
    super("mock-debug.txt");

    this.extout.appendLine("Creating Debug Session");
    this.client = new ECALDebugClient(new LogChannelAdapter(this.extout));

    // Add event handlers

    this.client.on("pauseOnBreakpoint", (e: ClientBreakEvent) => {
      console.log("#### send StoppedEvent event:", e.tid, typeof e.tid);
      this.sendEvent(new StoppedEvent("breakpoint", e.tid));
    });

    this.client.on("status", (e: DebugStatus) => {
      try {
        if (this.unconfirmedBreakpoints.length > 0) {
          for (const toConfirm of this.unconfirmedBreakpoints) {
            for (const [breakpointString, ok] of Object.entries(
              e.breakpoints
            )) {
              const line = parseInt(breakpointString.split(":")[1]);
              if (ok) {
                if (
                  toConfirm.line === line &&
                  toConfirm.source?.name === breakpointString
                ) {
                  console.log("Confirmed breakpoint:", breakpointString);
                  toConfirm.verified = true;
                  this.sendEvent(new BreakpointEvent("changed", toConfirm));
                }
              }
            }
          }
          this.unconfirmedBreakpoints = [];
        }
      } catch (e) {
        console.error(e);
      }
    });

    // Lines and columns start at 1
    this.setDebuggerLinesStartAt1(true);
    this.setDebuggerColumnsStartAt1(true);

    // Increment the config WaitGroup counter for configurationDoneRequest()
    this.wgConfig.add(1);
  }

  /**
   * Called as the first step in the DAP. The client (e.g. VSCode)
   * interrogates the debug adapter on the features which it provides.
   */
  protected initializeRequest(
    response: DebugProtocol.InitializeResponse,
    args: DebugProtocol.InitializeRequestArguments
  ): void {
    console.log("##### initializeRequest:", args);

    response.body = response.body || {};

    // The adapter implements the configurationDoneRequest.
    response.body.supportsConfigurationDoneRequest = true;

    // make VS Code to send cancelRequests
    response.body.supportsCancelRequest = true;

    // make VS Code send the breakpointLocations request
    response.body.supportsBreakpointLocationsRequest = true;

    // make VS Code provide "Step in Target" functionality
    response.body.supportsStepInTargetsRequest = true;

    this.sendResponse(response);

    this.sendEvent(new InitializedEvent());
  }

  /**
   * Called as part of the "configuration Done" step in the DAP. The client (e.g. VSCode) has
   * finished the initialization of the debug adapter.
   */
  protected configurationDoneRequest(
    response: DebugProtocol.ConfigurationDoneResponse,
    args: DebugProtocol.ConfigurationDoneArguments
  ): void {
    console.log("##### configurationDoneRequest");

    super.configurationDoneRequest(response, args);
    this.wgConfig.done();
  }

  /**
   * The client (e.g. VSCode) asks the debug adapter to start the debuggee communication.
   */
  protected async launchRequest(
    response: DebugProtocol.LaunchResponse,
    args: ECALDebugArguments
  ) {
    console.log("##### launchRequest:", args);

    this.config = args; // Store the configuration

    // Setup logging either verbose or just on errors

    logger.setup(
      args.trace ? Logger.LogLevel.Verbose : Logger.LogLevel.Error,
      false
    );

    await this.wgConfig.wait(); // Wait for configuration sequence to finish

    this.extout.appendLine(`Configuration loaded: ${JSON.stringify(args)}`);

    await this.client.conect(args.host, args.port);

    console.log("##### launchRequest result:", response.body);

    this.sendResponse(response);
  }

  protected async setBreakPointsRequest(
    response: DebugProtocol.SetBreakpointsResponse,
    args: DebugProtocol.SetBreakpointsArguments
  ): Promise<void> {
    console.log("##### setBreakPointsRequest:", args);

    let breakpoints: DebugProtocol.Breakpoint[] = [];

    if (args.source.path?.indexOf(this.config.dir) === 0) {
      const sourcePath = args.source.path.slice(this.config.dir.length + 1);

      // Clear all breakpoints of the file

      await this.client.clearBreakpoints(sourcePath);

      // Send all breakpoint requests to the debug server

      for (const sbp of args.breakpoints || []) {
        await this.client.setBreakpoint(`${sourcePath}:${sbp.line}`);
      }

      // Confirm that the breakpoints have been set

      const status = await this.client.status();

      if (status) {
        breakpoints = (args.lines || []).map((line) => {
          const breakpointString = `${sourcePath}:${line}`;
          const bp: DebugProtocol.Breakpoint = new Breakpoint(
            status.breakpoints[breakpointString],
            line,
            undefined,
            new Source(breakpointString, args.source.path)
          );
          bp.id = this.getBreakPointId(breakpointString);
          return bp;
        });
      } else {
        for (const sbp of args.breakpoints || []) {
          const breakpointString = `${sourcePath}:${sbp.line}`;
          const bp: DebugProtocol.Breakpoint = new Breakpoint(
            false,
            sbp.line,
            undefined,
            new Source(breakpointString, args.source.path)
          );
          bp.id = this.getBreakPointId(breakpointString);
          breakpoints.push(bp);
        }
        this.unconfirmedBreakpoints = breakpoints;
        console.log(
          "Breakpoints to be confirmed:",
          this.unconfirmedBreakpoints
        );
      }
    }

    response.body = {
      breakpoints,
    };

    console.error("##### setBreakPointsRequest result:", response.body);

    this.sendResponse(response);
  }

  protected async breakpointLocationsRequest(
    response: DebugProtocol.BreakpointLocationsResponse,
    args: DebugProtocol.BreakpointLocationsArguments
  ) {
    let breakpoints: DebugProtocol.BreakpointLocation[] = [];

    if (args.source.path?.indexOf(this.config.dir) === 0) {
      const sourcePath = args.source.path.slice(this.config.dir.length + 1);
      const status = await this.client.status();

      if (status) {
        for (const [breakpointString, v] of Object.entries(
          status.breakpoints
        )) {
          if (v) {
            const line = parseInt(breakpointString.split(":")[1]);
            if (`${sourcePath}:${line}` === breakpointString) {
              breakpoints.push({
                line,
              });
            }
          }
        }
      }
    }
    response.body = {
      breakpoints,
    };

    this.sendResponse(response);
  }

  protected async threadsRequest(
    response: DebugProtocol.ThreadsResponse
  ): Promise<void> {
    console.log("##### threadsRequest");

    const status = await this.client.status();
    const threads = [];

    if (status) {
      for (const tid of Object.keys(status.threads)) {
        threads.push(new Thread(parseInt(tid), `Thread ${tid}`));
      }
    } else {
      threads.push(new Thread(1, "Thread 1"));
    }

    response.body = {
      threads,
    };

    console.log("##### threadsRequest result:", response.body);

    this.sendResponse(response);
  }

  protected stackTraceRequest(
    response: DebugProtocol.StackTraceResponse,
    args: DebugProtocol.StackTraceArguments
  ): void {
    console.error("##### stackTraceRequest:", args);

    response.body = {
      stackFrames: [],
    };
    this.sendResponse(response);
  }

  protected scopesRequest(
    response: DebugProtocol.ScopesResponse,
    args: DebugProtocol.ScopesArguments
  ): void {
    console.error("##### scopesRequest:", args);

    response.body = {
      scopes: [],
    };
    this.sendResponse(response);
  }

  protected async variablesRequest(
    response: DebugProtocol.VariablesResponse,
    args: DebugProtocol.VariablesArguments,
    request?: DebugProtocol.Request
  ) {
    console.error("##### variablesRequest", args, request);

    response.body = {
      variables: [],
    };
    this.sendResponse(response);
  }

  protected continueRequest(
    response: DebugProtocol.ContinueResponse,
    args: DebugProtocol.ContinueArguments
  ): void {
    console.error("##### continueRequest", args);
    this.sendResponse(response);
  }

  protected reverseContinueRequest(
    response: DebugProtocol.ReverseContinueResponse,
    args: DebugProtocol.ReverseContinueArguments
  ): void {
    console.error("##### reverseContinueRequest", args);
    this.sendResponse(response);
  }

  protected nextRequest(
    response: DebugProtocol.NextResponse,
    args: DebugProtocol.NextArguments
  ): void {
    console.error("##### nextRequest", args);
    this.sendResponse(response);
  }

  protected stepBackRequest(
    response: DebugProtocol.StepBackResponse,
    args: DebugProtocol.StepBackArguments
  ): void {
    console.error("##### stepBackRequest", args);
    this.sendResponse(response);
  }

  protected stepInTargetsRequest(
    response: DebugProtocol.StepInTargetsResponse,
    args: DebugProtocol.StepInTargetsArguments
  ) {
    console.error("##### stepInTargetsRequest", args);
    response.body = {
      targets: [],
    };
    this.sendResponse(response);
  }

  protected stepInRequest(
    response: DebugProtocol.StepInResponse,
    args: DebugProtocol.StepInArguments
  ): void {
    console.error("##### stepInRequest", args);
    this.sendResponse(response);
  }

  protected stepOutRequest(
    response: DebugProtocol.StepOutResponse,
    args: DebugProtocol.StepOutArguments
  ): void {
    console.error("##### stepOutRequest", args);
    this.sendResponse(response);
  }

  protected async evaluateRequest(
    response: DebugProtocol.EvaluateResponse,
    args: DebugProtocol.EvaluateArguments
  ): Promise<void> {
    let result: any;

    try {
      result = await this.client.sendCommandString(`${args.expression}\r\n`);

      if (typeof result !== "string") {
        result = JSON.stringify(result, null, "  ");
      }
    } catch (e) {
      result = String(e);
    }

    response.body = {
      result,
      variablesReference: 0,
    };

    this.sendResponse(response);
  }

  protected dataBreakpointInfoRequest(
    response: DebugProtocol.DataBreakpointInfoResponse,
    args: DebugProtocol.DataBreakpointInfoArguments
  ): void {
    console.error("##### dataBreakpointInfoRequest", args);

    response.body = {
      dataId: null,
      description: "cannot break on data access",
      accessTypes: undefined,
      canPersist: false,
    };

    this.sendResponse(response);
  }

  protected setDataBreakpointsRequest(
    response: DebugProtocol.SetDataBreakpointsResponse,
    args: DebugProtocol.SetDataBreakpointsArguments
  ): void {
    console.error("##### setDataBreakpointsRequest", args);

    response.body = {
      breakpoints: [],
    };

    this.sendResponse(response);
  }

  protected completionsRequest(
    response: DebugProtocol.CompletionsResponse,
    args: DebugProtocol.CompletionsArguments
  ): void {
    console.error("##### completionsRequest", args);

    response.body = {
      targets: [
        {
          label: "item 10",
          sortText: "10",
        },
        {
          label: "item 1",
          sortText: "01",
        },
        {
          label: "item 2",
          sortText: "02",
        },
        {
          label: "array[]",
          selectionStart: 6,
          sortText: "03",
        },
        {
          label: "func(arg)",
          selectionStart: 5,
          selectionLength: 3,
          sortText: "04",
        },
      ],
    };
    this.sendResponse(response);
  }

  protected cancelRequest(
    response: DebugProtocol.CancelResponse,
    args: DebugProtocol.CancelArguments
  ) {
    console.error("##### cancelRequest", args);
    this.sendResponse(response);
  }

  protected customRequest(
    command: string,
    response: DebugProtocol.Response,
    args: any
  ) {
    console.error("##### customRequest", args);

    if (command === "toggleFormatting") {
      this.sendResponse(response);
    } else {
      super.customRequest(command, response, args);
    }
  }

  public shutdown() {
    console.log("#### Shutdown");
    this.client
      ?.shutdown()
      .then(() => {
        this.extout.appendLine("Debug Session has finished");
      })
      .catch((e) => {
        this.extout.appendLine(
          `Debug Session has finished with an error: ${e}`
        );
      });
  }

  /**
   * Map a given breakpoint string to a breakpoint ID.
   */
  private getBreakPointId(breakpointString: string): number {
    let id = this.bpIds[breakpointString];
    if (!id) {
      id = this.bpCount++;
      this.bpIds[breakpointString] = id;
    }
    return id;
  }
}

class LogChannelAdapter {
  private out: vscode.OutputChannel;

  constructor(out: vscode.OutputChannel) {
    this.out = out;
  }

  log(value: string): void {
    this.out.appendLine(value);
  }

  error(value: string): void {
    this.out.appendLine(`Error: ${value}`);
    setTimeout(() => {
      this.out.show(true);
    }, 500);
  }
}
