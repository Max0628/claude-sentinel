# claude-sentinel

監控 Claude Pro 訂閱額度，在重要事件發生時主動發送 Discord 通知——包括 session 即將用盡、session 重置、週用量門檻、以及每 20 分鐘一次的用量進度督促。

服務以 systemd 排程的 Podman container 形式運行在 home server 上，每 3 分鐘檢查一次。

---

## 前置條件

- 安裝了 Podman 的 Ubuntu server
- 已安裝並登入 Claude Code（`~/.claude/.credentials.json` 必須存在）
- Discord Webhook URL

---

## 設定

建立設定目錄與 env 檔案：

```bash
mkdir -p ~/.config/claude-sentinel
```

建立 `~/.config/claude-sentinel/env`：

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/你的_WEBHOOK_URL
```

建立初始 state 檔案：

```bash
echo '{}' > ~/.config/claude-sentinel/state.json
```

---

## 建置

在 repo 根目錄執行：

```bash
podman build -t claude-sentinel:latest .
```

---

## 部署（systemd）

複製 systemd unit 檔案：

```bash
cp deploy/systemd/claude-sentinel.service ~/.config/systemd/user/
cp deploy/systemd/claude-sentinel.timer ~/.config/systemd/user/
```

啟用並啟動：

```bash
systemctl --user daemon-reload
systemctl --user enable --now claude-sentinel.timer
```

---

## 管理服務

```bash
# 查看 timer 狀態
systemctl --user status claude-sentinel.timer

# 查看最近執行的 log
journalctl --user -u claude-sentinel.service -n 50

# 手動執行一次（立即檢查）
systemctl --user start claude-sentinel.service

# 停止 timer
systemctl --user disable --now claude-sentinel.timer
```

---

## 通知規則

| # | 事件 | 條件 | 發送頻率 |
|---|------|------|---------|
| 1 | Session 用量偏高 | 5 小時使用率 ≥ 80% | 每個 session 一次 |
| 2 | Session 重置 | `resets_at` 時間戳改變 | 每次重置 |
| 3 | 週用量門檻通知 | 7 天使用率每達到 10%、20%、30%、40%、50%、60%、70%、80%、90% | 每個門檻各一次 |
| 4 | 週用量進度督促 | 週用量 < 90%，且已進入週期超過 24 小時 | 每 20 分鐘一次 |

80% 以上的週用量門檻通知加上紅色驚嘆號（❗）。

Rule 4 從週期開始滿 24 小時後才啟動，避免週初數據不穩定造成誤報。

### Rule 4 督促強度

根據「預測週末剩餘 %」分為 10 個情緒等級。預測公式：

```
速率         = 目前已用% ÷ 已過分鐘數
剩餘預期用量  = 速率 × 剩餘分鐘數
預測週末剩餘  = 100% - MIN(目前已用% + 剩餘預期用量, 100%)
```

| 等級 | 預測週末剩餘 | 語氣 |
|------|------------|------|
| 1 | 0–10% | 平靜肯定，無 emoji |
| 2 | 10–20% | 輕微鼓勵，無 emoji |
| 3 | 20–30% | 溫和推進，無 emoji |
| 4 | 30–40% | 明確要求，無 emoji |
| 5 | 40–50% | 情緒化警告，無 emoji |
| 6 | 50–60% | 憤怒 + emoji |
| 7 | 60–70% | 強烈憤怒 + emoji + 輕微髒話 |
| 8 | 70–80% | 爆炸 + emoji + 髒話 |
| 9 | 80–90% | 最強情緒 + emoji + 重度髒話 |
| 10 | 90–100% | 冷靜諷刺與最大力道混用 |

每個等級有 10 種文案，每次隨機選取一種。

---

## 通知內容範例

### Rule 1 — Session 用量偏高

```
Claude Pro

Session 剩餘    18%
本週剩餘        64%
下次 Reset      06/28 (六) 18:40
```

### Rule 2 — Session 重置

```
Claude Pro Session 已重置

Session 剩餘    100%
本週剩餘        45%
```

本週剩餘隨每次重置一起顯示，讓你知道在密集使用時週額度正在快速消耗。

### Rule 3 — 週用量通知（10%–70%）

```
Claude Pro 本週用量提醒

本週已使用      40%
本週 Reset      06/30 (二) 12:00
```

### Rule 3 — 週用量通知（80%–90%）

```
❗ Claude Pro 本週用量警告

本週已使用      82%
本週 Reset      06/30 (二) 12:00

本週請保守使用 Claude。
```

### Rule 4 — 週用量進度督促（每 20 分鐘）

```
Claude Pro 週用量進度

本週已用        24%
預測週末剩餘    0%
距 Reset        6 天 20 小時

今天建議再用    11%

完美節奏，繼續。
```

```
Claude Pro 週用量進度

本週已用        8%
預測週末剩餘    72%
距 Reset        4 天 12 小時

今天建議再用    23%

🤬 幹！你到底在幹嘛？70% 要沒了！
```

---

## 疑難排解

**沒有收到通知**
- 確認 `~/.config/claude-sentinel/env` 裡的 Discord Webhook URL 正確
- 手動執行看看：`systemctl --user start claude-sentinel.service`，再用 `journalctl` 查看 log

**API 認證失敗**
- `~/.claude/.credentials.json` 裡的 access token 可能過期
- 在這台機器上開啟 Claude Code，讓它自動重新整理 token

**重複收到相同通知**
- 重置 state 檔案：`echo '{}' > ~/.config/claude-sentinel/state.json`
