export interface UserLogEntry {
  id: string;
  ts: string;
  event: string;
  source_id: string;
  target_id: string;
  source_type: 'membership' | 'user' | 'tenant';
  target_type: 'membership' | 'user' | 'tenant';
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
