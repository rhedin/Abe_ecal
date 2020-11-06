export interface ThreadInspection {
    callstack: string[]
    threadRunning: boolean
}

export interface ThreadStatus {
    callstack: string[]
    threadRunning?: boolean
}

export interface DebugStatus {
    breakonstart: boolean,
    breakpoints: any,
    sources: string[],
    threads: Record<number, ThreadStatus>
}

/**
 * Log output stream for this client.
 */
export interface LogOutputStream {
    log(value: string): void;
    error(value: string): void;
}
