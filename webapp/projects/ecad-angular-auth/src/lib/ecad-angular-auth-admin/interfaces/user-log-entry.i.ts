export interface UserLogEntry {
    id: string;
    ts: string;
    event: string;
    user_id: string;
    target_id: string;
    addr: string;
    msg: string;
    data: LogData;
}

export interface LogData {
    addr: string;
    email: string;
    event: string;
    id: string;
    user_id: string;
}
