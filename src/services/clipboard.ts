/**
 * 剪贴板工具。
 *
 * 优先使用异步 Clipboard API；在非安全上下文（例如通过 http://IP:port 访问）
 * 下 navigator.clipboard 不可用，此时退回 textarea + execCommand 的兜底方案，
 * 保证局域网 / 内网 IP 访问时点击复制同样生效。
 */

function legacyCopy(text: string): boolean {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.top = "-1000px";
  textarea.style.left = "-1000px";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);

  let ok = false;
  try {
    textarea.select();
    textarea.setSelectionRange(0, textarea.value.length);
    ok = document.execCommand("copy");
  } catch {
    ok = false;
  } finally {
    document.body.removeChild(textarea);
  }

  return ok;
}

/** 将文本写入剪贴板，返回是否成功。空文本视为失败。 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (!text) {
    return false;
  }

  if (window.isSecureContext && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // 权限被拒绝等情况，继续走兜底方案
    }
  }

  return legacyCopy(text);
}