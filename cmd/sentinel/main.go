package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/tashuchiu/claude-sentinel/internal/credentials"
	"github.com/tashuchiu/claude-sentinel/internal/notify"
	"github.com/tashuchiu/claude-sentinel/internal/quota"
	"github.com/tashuchiu/claude-sentinel/internal/state"
)

const (
	defaultCredentialsPath  = "/app/credentials.json"
	defaultStatePath        = "/app/state.json"
	weeklyMinutes           = float64(7 * 24 * 60)
	paceAlertInterval = 20 * time.Minute
)

func main() {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("DISCORD_WEBHOOK_URL is not set")
	}

	credPath := getenv("CREDENTIALS_PATH", defaultCredentialsPath)
	statePath := getenv("STATE_PATH", defaultStatePath)

	creds, err := credentials.Load(credPath)
	if err != nil {
		log.Fatalf("load credentials: %v", err)
	}

	usage, err := quota.Fetch(creds.AccessToken)
	if err != nil {
		if errors.Is(err, quota.ErrUnauthorized) {
			log.Println("access token expired, refreshing...")
			creds, err = credentials.Refresh(credPath)
			if err != nil {
				notify.AuthFailed(webhookURL)
				log.Fatalf("refresh token: %v", err)
			}
			usage, err = quota.Fetch(creds.AccessToken)
			if err != nil {
				log.Fatalf("fetch quota after refresh: %v", err)
			}
			fmt.Println("token refreshed successfully")
		} else {
			log.Fatalf("fetch quota: %v", err)
		}
	}

	s, err := state.Load(statePath)
	if err != nil {
		log.Fatalf("load state: %v", err)
	}

	isFirstRun := s.LastSessionResetAt == ""

	// Rule 2: session reset
	if !isFirstRun && !sameTimestamp(usage.FiveHour.ResetsAt, s.LastSessionResetAt) {
		if err := notify.SessionReset(webhookURL, usage.SevenDay.Utilization); err != nil {
			log.Printf("notify session reset: %v", err)
		} else {
			fmt.Println("notified: session reset")
		}
		s.SessionAlertSentFor = ""
	}
	s.LastSessionResetAt = usage.FiveHour.ResetsAt

	// Rule 1: session low (>= 80% used)
	if usage.FiveHour.Utilization >= 80 && s.SessionAlertSentFor != usage.FiveHour.ResetsAt {
		if err := notify.SessionLow(webhookURL, usage.FiveHour.Utilization, usage.SevenDay.Utilization, usage.FiveHour.ResetsAt); err != nil {
			log.Printf("notify session low: %v", err)
		} else {
			fmt.Println("notified: session low")
			s.SessionAlertSentFor = usage.FiveHour.ResetsAt
		}
	}

	// Rule 3: weekly thresholds every 10%
	if !sameTimestamp(usage.SevenDay.ResetsAt, s.WeeklyResetAt) {
		s.ResetWeekly(usage.SevenDay.ResetsAt)
	}

	if !isFirstRun {
		weeklyThresholds := []int{10, 20, 30, 40, 50, 60, 70, 80, 90}
		for _, threshold := range weeklyThresholds {
			if int(usage.SevenDay.Utilization) >= threshold && !s.HasSentWeeklyThreshold(threshold) {
				if err := notify.WeeklyThreshold(webhookURL, threshold, usage.SevenDay.Utilization, usage.SevenDay.ResetsAt); err != nil {
					log.Printf("notify weekly threshold %d%%: %v", threshold, err)
				} else {
					fmt.Printf("notified: weekly %d%%\n", threshold)
					s.AddWeeklyThreshold(threshold)
				}
			}
		}
	}

	// Rule 4: pace alert every 30 minutes (while weekly < 90%)
	if !isFirstRun && usage.SevenDay.Utilization < 90 {
		if shouldSendPaceAlert(s.LastPaceAlertSentAt, paceAlertInterval) {
			elapsed := time.Since(weeklyStartTime(usage.SevenDay.ResetsAt))
			projected, daily := calcPace(usage.SevenDay.Utilization, elapsed, usage.SevenDay.ResetsAt)
			if err := notify.PaceAlert(webhookURL, usage.FiveHour.Utilization, usage.FiveHour.ResetsAt, usage.SevenDay.Utilization, usage.SevenDay.ResetsAt, projected, daily); err != nil {
				log.Printf("notify pace alert: %v", err)
			} else {
				fmt.Printf("notified: pace alert (projected remaining %.0f%%)\n", projected)
				s.LastPaceAlertSentAt = time.Now().UTC().Format(time.RFC3339)
			}
		}
	}

	if err := state.Save(statePath, s); err != nil {
		log.Fatalf("save state: %v", err)
	}
}

func weeklyStartTime(resetsAt string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, resetsAt)
	if err != nil {
		return time.Now()
	}
	return t.Add(-7 * 24 * time.Hour)
}

func calcPace(currentUtil float64, elapsed time.Duration, resetsAt string) (projectedRemaining, dailySuggestion float64) {
	elapsedMinutes := elapsed.Minutes()
	if elapsedMinutes <= 0 {
		return 100, 100
	}

	resetsAtTime, _ := time.Parse(time.RFC3339Nano, resetsAt)
	remainingMinutes := time.Until(resetsAtTime).Minutes()
	if remainingMinutes < 0 {
		remainingMinutes = 0
	}

	ratePerMinute := currentUtil / elapsedMinutes
	projectedAdditional := ratePerMinute * remainingMinutes
	projectedTotalUsed := math.Min(currentUtil+projectedAdditional, 100)
	projectedRemaining = 100 - projectedTotalUsed

	remainingDays := remainingMinutes / (24 * 60)
	if remainingDays > 0 {
		dailySuggestion = (100 - currentUtil) / remainingDays
	} else {
		dailySuggestion = 100 - currentUtil
	}

	return projectedRemaining, dailySuggestion
}

func shouldSendPaceAlert(lastSentAt string, interval time.Duration) bool {
	if lastSentAt == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, lastSentAt)
	if err != nil {
		return true
	}
	return time.Since(t) >= interval
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// sameTimestamp compares two RFC3339 timestamps rounded to the nearest minute.
// The API returns slightly different sub-second values for the same logical reset time,
// sometimes landing on opposite sides of a second boundary (e.g. 11:19:59.949 vs 11:20:00.003).
func sameTimestamp(a, b string) bool {
	if a == b {
		return true
	}
	ta, err1 := time.Parse(time.RFC3339Nano, a)
	tb, err2 := time.Parse(time.RFC3339Nano, b)
	if err1 != nil || err2 != nil {
		return false
	}
	return ta.Round(time.Minute).Equal(tb.Round(time.Minute))
}
