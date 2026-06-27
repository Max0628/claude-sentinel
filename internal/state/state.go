package state

import (
	"encoding/json"
	"os"
)

type State struct {
	LastSessionResetAt   string `json:"last_session_reset_at"`
	SessionAlertSentFor  string `json:"session_alert_sent_for"`
	WeeklyResetAt        string `json:"weekly_reset_at"`
	WeeklyThresholdsSent []int  `json:"weekly_thresholds_sent"`
	LastPaceAlertSentAt  string `json:"last_pace_alert_sent_at"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var s State
	if len(data) == 0 || string(data) == "{}" {
		return &State{}, nil
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}, nil
	}

	return &s, nil
}

func Save(path string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *State) HasSentWeeklyThreshold(threshold int) bool {
	for _, t := range s.WeeklyThresholdsSent {
		if t == threshold {
			return true
		}
	}
	return false
}

func (s *State) AddWeeklyThreshold(threshold int) {
	s.WeeklyThresholdsSent = append(s.WeeklyThresholdsSent, threshold)
}

func (s *State) ResetWeekly(newResetAt string) {
	s.WeeklyResetAt = newResetAt
	s.WeeklyThresholdsSent = nil
}
