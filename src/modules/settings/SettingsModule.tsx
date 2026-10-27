interface SettingsModuleProps {
  draftKey: string;
  savedKey: string;
  webhookUrl: string;
  legacyWebhookUrl: string;
  onDraftKeyChange: (value: string) => void;
  onSave: () => void;
  onClear: () => void;
  onCopy: (text: string, successMessage: string) => Promise<void>;
}

export function SettingsModule({
  draftKey,
  savedKey,
  webhookUrl,
  legacyWebhookUrl,
  onDraftKeyChange,
  onSave,
  onClear,
  onCopy
}: SettingsModuleProps) {
  const hasSavedKey = savedKey.trim().length > 0;

  return (
    <section className="module-grid">
      <article className="card">
        <h2>查询 Key 设置</h2>
        <p className="muted">未设置 key 时，功能模块不可用。请先保存 key。</p>
        <div className="field-group">
          <label htmlFor="report-key-input">Report Key</label>
          <input
            id="report-key-input"
            type="password"
            value={draftKey}
            onChange={(event) => onDraftKeyChange(event.target.value)}
            placeholder="请输入 REPORT_AUTH_KEY 或轮换 key"
          />
        </div>
        <div className="row-actions">
          <button type="button" className="button primary" onClick={onSave}>
            保存
          </button>
          <button type="button" className="button" onClick={onClear}>
            清空
          </button>
        </div>
        <div className={`status-tip ${hasSavedKey ? "ok" : "warn"}`}>
          {hasSavedKey ? "key 已保存，可使用功能模块。" : "尚未保存 key。"}
        </div>
      </article>

      <article className="card">
        <h2>Webhook 地址（已拼接 key）</h2>
        <p className="muted">适合仅支持 URL 参数的第三方上报。</p>

        <div className="field-group">
          <label htmlFor="webhook-url-input">主地址</label>
          <div className="inline-input-action">
            <input id="webhook-url-input" value={webhookUrl} readOnly />
            <button
              type="button"
              className="button"
              onClick={() => onCopy(webhookUrl, "已复制 webhook 地址。")}
            >
              复制
            </button>
          </div>
        </div>

        <div className="field-group">
          <label htmlFor="legacy-webhook-url-input">兼容地址</label>
          <div className="inline-input-action">
            <input id="legacy-webhook-url-input" value={legacyWebhookUrl} readOnly />
            <button
              type="button"
              className="button"
              onClick={() => onCopy(legacyWebhookUrl, "已复制兼容 webhook 地址。")}
            >
              复制
            </button>
          </div>
        </div>
      </article>
    </section>
  );
}
