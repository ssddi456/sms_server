import { useEffect, useMemo, useState } from "react";
import QRCode from "qrcode";

interface DownloadApkButtonProps {
  apkFileName: string;
  onCopy: (text: string, successMessage: string) => Promise<void>;
}

export function DownloadApkButton({ apkFileName, onCopy }: DownloadApkButtonProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [qrCodeDataUrl, setQrCodeDataUrl] = useState("");
  const [qrError, setQrError] = useState("");

  const apkUrl = `${window.location.origin}/downloads/${encodeURIComponent(apkFileName)}`;

  const isMobile = useMemo(() => {
    const ua = navigator.userAgent.toLowerCase();
    const mobileUA = /android|iphone|ipad|ipod|mobile|harmonyos/.test(ua);
    return mobileUA || window.matchMedia("(max-width: 768px)").matches;
  }, []);

  useEffect(() => {
    if (!isDialogOpen || isMobile) {
      return;
    }
    let cancelled = false;
    setQrCodeDataUrl("");
    setQrError("");
    QRCode.toDataURL(apkUrl, { width: 220, margin: 1 })
      .then((dataUrl) => {
        if (!cancelled) {
          setQrCodeDataUrl(dataUrl);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setQrError("二维码生成失败，请复制下载地址。");
        }
      });
    return () => {
      cancelled = true;
    };
  }, [apkUrl, isDialogOpen, isMobile]);

  const handleDownloadClick = () => {
    if (isMobile) {
      window.location.href = apkUrl;
      return;
    }
    setIsDialogOpen(true);
  };

  return (
    <>
      <button type="button" className="button primary" onClick={handleDownloadClick}>
        下载 APK
      </button>

      {isDialogOpen ? (
        <div className="dialog-mask" role="presentation" onClick={() => setIsDialogOpen(false)}>
          <div className="dialog" role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
            <h3>下载 SmsForwarder</h3>
            <div className="inline-input-action">
              <input value={apkUrl} readOnly />
              <button
                type="button"
                className="button"
                onClick={() => onCopy(apkUrl, "已复制 APK 下载地址。")}
              >
                复制
              </button>
            </div>
            {qrCodeDataUrl ? (
              <div className="qrcode-box">
                <img src={qrCodeDataUrl} alt="APK 下载二维码" />
              </div>
            ) : (
              <p className="muted">{qrError || "二维码生成中..."}</p>
            )}
            <div className="row-actions">
              <a className="button primary" href={apkUrl} target="_blank" rel="noopener noreferrer">
                打开下载链接
              </a>
              <button type="button" className="button" onClick={() => setIsDialogOpen(false)}>
                关闭
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
