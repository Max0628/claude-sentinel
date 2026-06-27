# claude-sentinel — 專案說明

## 這是什麼

一個 Go 服務，每 10 分鐘輪詢 Claude Pro OAuth 使用量 API，並在三種事件發生時發送 Discord 通知：session 用量偏高（≥80%）、session 重置、週用量偏高（≥80%）。

## 架構

**部署模式：** systemd timer → `podman run --rm`（一次性執行，與 daily_log 相同模式）

**執行時掛載的檔案：**
- `~/.claude/.credentials.json` — 唯讀掛載；包含 OAuth API 所需的 `accessToken`
- `~/.config/claude-sentinel/env` — Discord Webhook URL 等設定（不進 repo）
- `~/.config/claude-sentinel/state.json` — 在每次執行之間持久化通知狀態

## OAuth API

```
GET https://api.anthropic.com/api/oauth/usage
Authorization: Bearer <accessToken>
anthropic-beta: oauth-2025-04-20
User-Agent: claude-code/2.1.191
```

使用的 response 欄位：
- `five_hour.utilization` — session 使用率（0–100）
- `five_hour.resets_at` — ISO 8601 時間戳；改變即代表新 session 開始
- `seven_day.utilization` — 週使用率（0–100）
- `seven_day.resets_at` — 週期的 ISO 8601 時間戳

## State 檔案結構

```json
{
  "last_session_reset_at": "2026-06-27T07:00:00Z",
  "session_alert_sent_for": "2026-06-27T07:00:00Z",
  "weekly_reset_at": "2026-07-04T04:00:00Z",
  "weekly_thresholds_sent": [10, 20],
  "last_pace_alert_sent_at": "2026-06-27T08:00:00Z"
}
```

- `last_session_reset_at` — 上次記錄的 `five_hour.resets_at`；改變時觸發重置通知
- `session_alert_sent_for` — 已發送低用量警告的 session（以 `resets_at` 識別）
- `weekly_reset_at` — 記錄週期識別用，改變時清空 thresholds
- `weekly_thresholds_sent` — 本週期已發送的門檻清單（10, 20, ... 90）
- `last_pace_alert_sent_at` — 上次發送督促通知的時間；間隔 ≥ 1 小時才再發

## 通知規則

| 規則 | 觸發條件 | 去重 key |
|------|---------|---------|
| Session 用量偏高 | `five_hour.utilization >= 80` | `five_hour.resets_at` |
| Session 重置 | `five_hour.resets_at` 改變 | 儲存值 vs 當前值比對 |
| 週用量偏高 | `seven_day.utilization >= 80` | `seven_day.resets_at` |

## Go Package 結構

```
cmd/sentinel/main.go          — 程式進入點
internal/credentials/         — 讀取 ~/.claude/.credentials.json
internal/quota/               — 呼叫 OAuth API、解析 response
internal/state/               — 讀寫 state.json
internal/notify/notify.go     — 發送 Discord webhook、格式化訊息
internal/notify/messages.go   — 10 × 10 督促文案（不 hardcode 在邏輯裡）
```

## 建置與執行

```bash
# 建置 image
podman build -t claude-sentinel:latest .

# 手動執行一次
podman run --rm \
  --env-file ~/.config/claude-sentinel/env \
  -v ~/.claude/.credentials.json:/app/credentials.json:ro \
  -v ~/.config/claude-sentinel/state.json:/app/state.json:rw \
  claude-sentinel:latest
```

## 程式風格

- 優先使用 standard library，除非外部套件明顯值得引入
- 不使用全域狀態；將 config 和 state 明確傳遞
- package 保持小而單一職責
- `internal/` 下的邏輯使用標準 Go table-driven tests
