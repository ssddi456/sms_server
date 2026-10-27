import type { HistoryPayload, SmsRecord } from "../types";

export async function fetchHistory(page: number, limit: number, reportKey: string): Promise<{
  items: SmsRecord[];
  hasMore: boolean;
}> {
  const response = await fetch(`/sms/history?limit=${limit}&page=${page}`, {
    method: "GET",
    headers: {
      "X-Report-Key": reportKey
    }
  });

  if (response.status === 401) {
    throw new Error("鉴权失败，请检查 key 是否正确。");
  }
  if (!response.ok) {
    throw new Error(`请求失败：HTTP ${response.status}`);
  }

  const payload = (await response.json()) as HistoryPayload;
  const items = payload.data?.items ?? [];
  return {
    items,
    hasMore: items.length >= limit
  };
}
