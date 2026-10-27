import { useEffect, useState } from "react";
import { fetchHistory } from "../../services/api";
import type { SmsRecord } from "../../types";

interface HistoryPanelProps {
  reportKey: string;
  onRequireSettings: () => void;
}

const LIMIT = 20;

export function HistoryPanel({ reportKey, onRequireSettings }: HistoryPanelProps) {
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<SmsRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [hasMore, setHasMore] = useState(false);
  const [reloadTick, setReloadTick] = useState(0);
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());

  const toggleExpanded = (id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  useEffect(() => {
    if (!reportKey) {
      setItems([]);
      setError("");
      setHasMore(false);
      setPage(1);
      return;
    }
    let cancelled = false;
    async function loadHistory() {
      setLoading(true);
      setError("");
      try {
        const result = await fetchHistory(page, LIMIT, reportKey);
        if (cancelled) return;
        setItems(result.items);
        setHasMore(result.hasMore);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "加载失败");
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void loadHistory();
    return () => {
      cancelled = true;
    };
  }, [page, reportKey, reloadTick]);

  const formatTime = (value?: string) => {
    if (!value) return "-";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString("zh-CN", { hour12: false });
  };

  if (!reportKey) {
    return (
      <div className="empty-tip">
        <p>请先到设置模块保存 key，再查询历史记录。</p>
        <button type="button" className="button primary" onClick={onRequireSettings}>
          前往设置
        </button>
      </div>
    );
  }

  return (
    <article className="card">
      <div className="card-head">
        <h2>短信历史记录</h2>
        <button type="button" className="button" onClick={() => setReloadTick((value) => value + 1)}>
          刷新
        </button>
      </div>

      <div className="history-list">
        <div className="history-head">
          <span>ID</span>
          <span>发送方</span>
          <span>短信内容</span>
          <span>验证码</span>
          <span>接收时间</span>
        </div>

        {loading ? (
          <div className="table-message">加载中...</div>
        ) : error ? (
          <div className="table-message error">{error}</div>
        ) : items.length === 0 ? (
          <div className="table-message">暂无数据</div>
        ) : (
          items.map((item) => (
            <div className="history-row" key={item.id}>
              <span className="h-id" data-label="ID">
                {item.id}
              </span>
              <span className="h-from" data-label="发送方">
                {item.from || "-"}
              </span>
              <span
                className={`h-sms${expandedIds.has(item.id) ? " expanded" : ""}`}
                data-label="短信内容"
                onClick={() => toggleExpanded(item.id)}
              >
                <span className="sms-text" title={item.sms || ""}>
                  {item.sms || "-"}
                </span>
              </span>
              <span className="h-code" data-label="验证码">
                {item.extracted_code ? (
                  <span className="code-badge">{item.extracted_code}</span>
                ) : (
                  <span className="code-badge empty">-</span>
                )}
              </span>
              <span className="h-time" data-label="接收时间">
                {formatTime(item.receive_time)}
              </span>
            </div>
          ))
        )}
      </div>

      <div className="pager">
        <span>第 {page} 页</span>
        <div className="row-actions">
          <button type="button" className="button" disabled={page <= 1} onClick={() => setPage((current) => current - 1)}>
            上一页
          </button>
          <button type="button" className="button" disabled={!hasMore} onClick={() => setPage((current) => current + 1)}>
            下一页
          </button>
        </div>
      </div>
    </article>
  );
}
