package payload

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jsvensson/kimi-quota-monitor/internal/kimi"
)

// n is a helper for building *kimi.Number values in tests.
func n(v int64) *kimi.Number {
	num := kimi.Number(v)
	return &num
}

// TestFromUsages verifies that a full Usages response maps each window
// label to {"pct": <remaining>, "resets_at": <unix epoch>}.
// The API reports the 5-hour window as 300 minutes.
func TestFromUsages(t *testing.T) {
	t.Parallel()

	weeklyReset := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	fiveHourReset := time.Date(2026, 8, 24, 20, 0, 0, 0, time.UTC)

	u := kimi.Usages{
		Usage: kimi.Quota{
			Limit:     5000,
			Used:      n(1230),
			Remaining: n(3770),
			ResetTime: weeklyReset,
		},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
			Detail: kimi.Quota{
				Limit:     200,
				Used:      n(45),
				Remaining: n(155),
				ResetTime: fiveHourReset,
			},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]struct {
		Pct      int64 `json:"pct"`
		ResetsAt int64 `json:"resets_at"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got[LabelWeekly].Pct != 3770 || got[LabelWeekly].ResetsAt != weeklyReset.Unix() {
		t.Errorf("payload[%q] = %+v, want pct 3770 and resets_at %d", LabelWeekly, got[LabelWeekly], weeklyReset.Unix())
	}
	if got[LabelFiveHour].Pct != 155 || got[LabelFiveHour].ResetsAt != fiveHourReset.Unix() {
		t.Errorf("payload[%q] = %+v, want pct 155 and resets_at %d", LabelFiveHour, got[LabelFiveHour], fiveHourReset.Unix())
	}
}

// TestFromUsagesFiveHourInHours verifies that a 5-hour window reported
// in TIME_UNIT_HOUR also matches.
func TestFromUsagesFiveHourInHours(t *testing.T) {
	t.Parallel()

	reset := time.Date(2026, 8, 24, 20, 0, 0, 0, time.UTC)
	u := kimi.Usages{
		Usage: kimi.Quota{
			Limit:     5000,
			Used:      n(1230),
			Remaining: n(3770),
			ResetTime: reset,
		},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 5, TimeUnit: "TIME_UNIT_HOUR"},
			Detail: kimi.Quota{
				Limit:     200,
				Used:      n(45),
				Remaining: n(155),
				ResetTime: reset,
			},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]struct {
		Pct      int64 `json:"pct"`
		ResetsAt int64 `json:"resets_at"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got[LabelFiveHour].Pct != 155 {
		t.Errorf("payload[%q].pct = %d, want 155", LabelFiveHour, got[LabelFiveHour].Pct)
	}
}

// TestFromUsagesNoFiveHourWindow verifies that the 5h entry is omitted
// when the response has no 5-hour rate-limit window.
func TestFromUsagesNoFiveHourWindow(t *testing.T) {
	t.Parallel()

	reset := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	u := kimi.Usages{
		Usage: kimi.Quota{
			Limit:     5000,
			Used:      n(1230),
			Remaining: n(3770),
			ResetTime: reset,
		},
		Limits: []kimi.Rate{{Window: kimi.Window{Duration: 1, TimeUnit: "TIME_UNIT_DAY"}}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]struct {
		Pct      int64 `json:"pct"`
		ResetsAt int64 `json:"resets_at"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if _, ok := got[LabelFiveHour]; ok {
		t.Errorf("payload contains %q, want it omitted", LabelFiveHour)
	}
	if got[LabelWeekly].Pct != 3770 {
		t.Errorf("payload[%q].pct = %d, want 3770", LabelWeekly, got[LabelWeekly].Pct)
	}
}

// TestFromUsagesRemainingFallback verifies that the remaining value is
// derived from limit - used when the API omits the remaining field.
func TestFromUsagesRemainingFallback(t *testing.T) {
	t.Parallel()

	weeklyReset := time.Date(2026, 8, 27, 23, 16, 16, 0, time.UTC)
	fiveHourReset := time.Date(2026, 8, 24, 20, 16, 16, 0, time.UTC)

	u := kimi.Usages{
		Usage: kimi.Quota{
			Limit:     100,
			Used:      n(93),
			Remaining: n(7),
			ResetTime: weeklyReset,
		},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
			Detail: kimi.Quota{
				Limit:     100,
				Used:      n(100),
				ResetTime: fiveHourReset,
			},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]struct {
		Pct      int64 `json:"pct"`
		ResetsAt int64 `json:"resets_at"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got[LabelWeekly].Pct != 7 || got[LabelWeekly].ResetsAt != weeklyReset.Unix() {
		t.Errorf("payload[%q] = %+v, want pct 7 and resets_at %d", LabelWeekly, got[LabelWeekly], weeklyReset.Unix())
	}
	if got[LabelFiveHour].Pct != 0 || got[LabelFiveHour].ResetsAt != fiveHourReset.Unix() {
		t.Errorf("payload[%q] = %+v, want pct 0 and resets_at %d", LabelFiveHour, got[LabelFiveHour], fiveHourReset.Unix())
	}
}
