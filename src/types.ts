export interface SmsRecord {
  id: number;
  from: string;
  sms: string;
  extracted_code?: string;
  receive_time?: string;
}

export interface HistoryPayload {
  result: string;
  code: string;
  message: string;
  data?: {
    page: number;
    limit: number;
    items: SmsRecord[];
  };
}
