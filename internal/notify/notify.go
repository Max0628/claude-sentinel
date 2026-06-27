package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var taipei = time.FixedZone("CST", 8*60*60)

var weekdays = []string{"日", "一", "二", "三", "四", "五", "六"}

func formatTime(iso string) string {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		return iso
	}
	t = t.In(taipei)
	return fmt.Sprintf("%02d/%02d(%s) %02d:%02d",
		t.Month(), t.Day(), weekdays[t.Weekday()], t.Hour(), t.Minute())
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%d天%d時", days, hours)
	}
	return fmt.Sprintf("%d時", hours)
}

func remaining(utilization float64) int {
	r := 100 - int(utilization)
	if r < 0 {
		return 0
	}
	return r
}

func send(webhookURL, message string) error {
	payload, _ := json.Marshal(map[string]string{"content": message})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("send discord: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}
	return nil
}

func SessionLow(webhookURL string, sessionUtil, weeklyUtil float64, sessionResetsAt string) error {
	msg := fmt.Sprintf(
		"Claude Pro\n\nSession 剩餘    %d%%\n本週剩餘        %d%%\n下次 Reset      %s",
		remaining(sessionUtil),
		remaining(weeklyUtil),
		formatTime(sessionResetsAt),
	)
	return send(webhookURL, msg)
}

func SessionReset(webhookURL string, weeklyUtil float64) error {
	msg := fmt.Sprintf(
		"Claude Pro Session 已重置\n\nSession 剩餘    100%%\n本週剩餘        %d%%",
		remaining(weeklyUtil),
	)
	return send(webhookURL, msg)
}

func WeeklyThreshold(webhookURL string, threshold int, weeklyUtil float64, weeklyResetsAt string) error {
	prefix := ""
	suffix := ""
	if threshold >= 80 {
		prefix = "❗ "
		suffix = "\n\n本週請保守使用 Claude。"
	}

	msg := fmt.Sprintf(
		"%sClaude Pro 本週用量%s\n\n本週已使用      %d%%\n本週 Reset      %s%s",
		prefix,
		thresholdLabel(threshold),
		int(weeklyUtil),
		formatTime(weeklyResetsAt),
		suffix,
	)
	return send(webhookURL, msg)
}

func thresholdLabel(threshold int) string {
	if threshold >= 80 {
		return "警告"
	}
	return "提醒"
}

// PaceAlert sends the pace notification with emotional copy based on projected remaining %.
func PaceAlert(webhookURL string, sessionUtil float64, sessionResetsAt string, weeklyUtil float64, weeklyResetsAt string, projectedRemaining float64, dailySuggestion float64) error {
	tier := paceTier(projectedRemaining)
	copy := paceMessages[tier][rand.Intn(10)]

	weeklyResetTime, _ := time.Parse(time.RFC3339Nano, weeklyResetsAt)
	timeUntilWeeklyReset := time.Until(weeklyResetTime)

	var weeklyProjectionLine string
	if projectedRemaining <= 0 {
		weeklyProjectionLine = "週末預期會用完"
	} else {
		weeklyProjectionLine = fmt.Sprintf("預測週末剩餘    %d%%", int(projectedRemaining))
	}

	now := time.Now().In(taipei)
	nowStr := fmt.Sprintf("%d/%d %02d:%02d", now.Month(), now.Day(), now.Hour(), now.Minute())

	msg := fmt.Sprintf(
		"▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬\n%s\n\n5小時 Session 剩餘  %d%%\n下次 Reset  %s\n\n本週已用    %d%%\n%s\n週 Reset  %s（+%s）\n\n今天建議再用  %d%%\n\n%s",
		nowStr,
		remaining(sessionUtil),
		formatTime(sessionResetsAt),
		int(weeklyUtil),
		weeklyProjectionLine,
		formatTime(weeklyResetsAt),
		formatDuration(timeUntilWeeklyReset),
		int(dailySuggestion),
		copy,
	)
	return send(webhookURL, msg)
}

// paceTier returns 0-9 index into paceMessages based on projected remaining %.
func paceTier(projectedRemaining float64) int {
	switch {
	case projectedRemaining < 10:
		return 0
	case projectedRemaining < 20:
		return 1
	case projectedRemaining < 30:
		return 2
	case projectedRemaining < 40:
		return 3
	case projectedRemaining < 50:
		return 4
	case projectedRemaining < 60:
		return 5
	case projectedRemaining < 70:
		return 6
	case projectedRemaining < 80:
		return 7
	case projectedRemaining < 90:
		return 8
	default:
		return 9
	}
}
