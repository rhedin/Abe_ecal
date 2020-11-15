export interface EcalAstNode {
  allowescapes: boolean;
  children: EcalAstNode[];
  id: number;
  identifier: boolean;
  line: number;
  linepos: number;
  name: string;
  pos: number;
  source: string;
  value: any;
}

export interface ThreadInspection {
  callstack: string[];
  threadRunning: boolean;
  code?: string;
  node?: EcalAstNode;
  vs?: any;
}

export interface ClientBreakEvent {
  tid: number;
  inspection: ThreadInspection;
}

export interface ThreadStatus {
  callstack: string[];
  threadRunning?: boolean;
}

export interface DebugStatus {
  breakonstart: boolean;
  breakpoints: Record<string, boolean>;
  sources: string[];
  threads: Record<number, ThreadStatus>;
}

/**
 * Log output stream for this client.
 */
export interface LogOutputStream {
  log(value: string): void;
  error(value: string): void;
}
