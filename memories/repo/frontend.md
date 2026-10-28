---
brief: 前端结构、构建、产物不入库与本地环境限制
---

# 前端

- 新版 React UI：`src/`，入口 `src/main.tsx`，由 rsbuild 构建到 `frontend/static/app/`，访问 `/sms/assets/app/index.html`
- 构建：`npm run build:ui`
- **构建产物不入库**：`.gitignore` 含 `frontend/static/app/`（rsbuild `cleanDistPath` 每次换 hash 文件名，入库会产生删除+新增噪音）
- 产物由 Dockerfile 的 Node 阶段（`npm ci` + `npm run build:ui`）自动生成；`.dockerignore` 也排除了它
- 因此只改 `src/` 即可，部署时远端自动重新构建；本地运行服务前必须先 `npm run build:ui`
- 旧版遗留模板：`frontend/templates/index.html` + `frontend/static/js/app.js`（部分功能已有 `copy-code` 按钮），新功能优先改 React
- 剪贴板工具：`src/services/clipboard.ts` 的 `copyToClipboard()`，HTTP 非安全上下文自动回退 `execCommand`
- 历史记录表格：`src/modules/features/HistoryPanel.tsx`，验证码 tag 为 `button.code-badge.clickable`，点击复制并显示"已复制 ✓"1.6 秒
- 本地限制：`npx tsc --noEmit` 因缺 `@types/node` 报 TS2688（既有问题）；本机 `go run` 因 CGO 缺失无法启动（go-sqlite3 需 cgo），本地冒烟需靠远端构建