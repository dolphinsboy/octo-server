# Changelog

## [v1.1] - 2026-03-04
### Security
- Bot token 吊销时正确撤销 IM token（cleanupBotConnection）
- 删除机器人时清理 IM 连接、Redis 心跳和事件队列
- WuKongIM 管理 API (5300) 限制为 127.0.0.1 访问
- robotList API 权限升级为 SuperAdmin (#36 → #37)
- Android 文件下载路径遍历防护（sanitizeFileName）(#21)
- Android 文件选择 100MB 大小限制 + 危险扩展名黑名单 (#22)

### Fixed
- Web Unicode emoji 显示为方块 — 添加 Segoe UI Emoji / Noto Color Emoji 字体回退 (#14)

### Infrastructure
- 仓库迁移到 dmwork-org GitHub 组织
- 添加 Feature Request / PR 模板
- 建立 Milestone v1.1 + Labels 体系
- OpenClaw adapter 升级到 0.2.17（BodyForAgent + 滑动窗口历史）

### Previous (v1.0 → v1.1)
- 邮箱验证码注册登录 (#35)
- Bot register 支持 force_refresh (#34)
- 全局搜索优雅降级 (#33)
- 本地默认头像生成 (#31)
- Bot HTTP API 压测脚本 (#30)
- API 测试 28/28 通过 (#29)
- 支持同时 @多个 Bot (#23)
- 安全加固: bcrypt 密码 + Webhook HMAC (#17)
- Bot 增强: @群聊路由 + 入群回调 + 自动已读 (#16)
- 文件模块安全增强 (#11)

## [v1.0] - 2026-02-28
- 初始版本
- 基于悟空IM (WuKongIM) 的即时通讯平台
- Web / Android 客户端
- Bot 系统 (BotFather 模式)
