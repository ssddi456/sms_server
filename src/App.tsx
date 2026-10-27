import { useMemo, useState } from "react";
import { DownloadApkButton } from "./components/DownloadApkButton";
import { FeaturesModule } from "./modules/features/FeaturesModule";
import { SettingsModule } from "./modules/settings/SettingsModule";

const REPORT_KEY_STORAGE = "sms_report_auth_key";
const DEFAULT_APK_FILE_NAME = "SmsF_3.5.0.260224_100055_universal_release.apk";

type ModuleTab = "features" | "settings";

function buildWebhookUrl(path: string, key: string): string {
  const base = `${window.location.origin}${path}`;
  if (!key) {
    return base;
  }
  return `${base}?report_key=${encodeURIComponent(key)}`;
}

export function App() {
  const storedKey = localStorage.getItem(REPORT_KEY_STORAGE)?.trim() ?? "";
  const [draftKey, setDraftKey] = useState(storedKey);
  const [savedKey, setSavedKey] = useState(storedKey);
  const [statusMessage, setStatusMessage] = useState("");
  const [activeModule, setActiveModule] = useState<ModuleTab>(savedKey ? "features" : "settings");

  const apkFileName = useMemo(
    () => window.__SMS_UI_CONFIG__?.apkFileName?.trim() || DEFAULT_APK_FILE_NAME,
    []
  );

  const webhookUrl = useMemo(() => buildWebhookUrl("/sms/webhook", draftKey.trim()), [draftKey]);
  const legacyWebhookUrl = useMemo(
    () => buildWebhookUrl("/v1/sms/receive", draftKey.trim()),
    [draftKey]
  );

  async function copyText(text: string, successMessage: string) {
    if (!text) {
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      setStatusMessage(successMessage);
    } catch {
      setStatusMessage("复制失败，请手动复制。");
    }
  }

  function handleSave() {
    const normalized = draftKey.trim();
    if (!normalized) {
      setStatusMessage("key 不能为空。");
      return;
    }
    localStorage.setItem(REPORT_KEY_STORAGE, normalized);
    setSavedKey(normalized);
    setStatusMessage("key 已保存。");
    setActiveModule("features");
  }

  function handleClear() {
    setDraftKey("");
    setSavedKey("");
    localStorage.removeItem(REPORT_KEY_STORAGE);
    setStatusMessage("已清空 key。");
    setActiveModule("settings");
  }

  return (
    <div className="page">
      <header className="header card">
        <div className="brand">
          <h1>SMS 历史记录</h1>
          <p className="subtitle">短信验证码接收与历史查询</p>
        </div>
        <div className="header-actions">
          <DownloadApkButton apkFileName={apkFileName} onCopy={copyText} />
          <a className="button" href="/sms/healthz" target="_blank" rel="noreferrer">
            健康检查
          </a>
        </div>
      </header>

      <nav className="module-switch">
        <button
          type="button"
          className={`tab-button ${activeModule === "features" ? "active" : ""}`}
          onClick={() => setActiveModule("features")}
        >
          功能模块
        </button>
        <button
          type="button"
          className={`tab-button ${activeModule === "settings" ? "active" : ""}`}
          onClick={() => setActiveModule("settings")}
        >
          设置模块
        </button>
      </nav>

      {!savedKey ? (
        <div className="status-tip warn">未设置 key，请先在设置模块保存 key 后再使用功能模块。</div>
      ) : null}
      {statusMessage ? <div className="status-tip ok">{statusMessage}</div> : null}

      {activeModule === "features" ? (
        <FeaturesModule
          reportKey={savedKey}
          onRequireSettings={() => setActiveModule("settings")}
        />
      ) : (
        <SettingsModule
          draftKey={draftKey}
          savedKey={savedKey}
          webhookUrl={webhookUrl}
          legacyWebhookUrl={legacyWebhookUrl}
          onDraftKeyChange={setDraftKey}
          onSave={handleSave}
          onClear={handleClear}
          onCopy={copyText}
        />
      )}
    </div>
  );
}
