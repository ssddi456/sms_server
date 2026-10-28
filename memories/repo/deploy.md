---
brief: 部署方式、远端目标与验证手段
---

# 部署

- 推荐脚本：`scripts/deploy_remote_podman.ps1`（SSH 直连远端 Podman 构建，无需本地 Docker）
- 远端目标：`root@120.48.167.92:/opt/sms-server`，端口 22，密钥 `$HOME/.ssh/id_rsa_new`
- 完整命令：`powershell -ExecutionPolicy Bypass -File .\scripts\deploy_remote_podman.ps1 -RemoteHost 120.48.167.92 -RemoteUser root -RemotePath /opt/sms-server -IdentityFile "$HOME\.ssh\id_rsa_new" -HealthTimeout 180`
- 脚本打包时排除 `frontend/static/app`，由远端 Dockerfile 的 Node 阶段重新构建前端，因此只改 `src/` 即可
- 健康检查：`http://127.0.0.1:8322/sms/healthz`，成功返回 `{"code":"SUCCESS","data":{"status":"healthy"}}`
- 远端验证：`ssh ... 'podman exec sms-server grep -c <关键字> /app/frontend/static/app/static/js/*.js'`（注意 PowerShell 用单引号包裹整条远程命令，避免引号被本地解析）