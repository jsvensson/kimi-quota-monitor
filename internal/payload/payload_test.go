package payload

import (
	"encoding/json"
	"testing"

	"github.com/jsvensson/kimi-quota-monitor/internal/kimi"
)

// n is a helper for building *kimi.Number values in tests.
func n(v int64) *kimi.Number {
	num := kimi.Number(v)
	return &num
}

// TestFromUsages verifies that a full Usages response maps each window
// label to the current remaining value, e.g. {"5h":155,"7d":3770}.
// The API reports the 5-hour window as 300 minutes.
func TestFromUsages(t *testing.T) {
	t.Parallel()

	u := kimi.Usages{
		Usage: kimi.Quota{Limit: 5000, Used: n(1230), Remaining: n(3770)},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
			Detail: kimi.Quota{Limit: 200, Used: n(45), Remaining: n(155)},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]int64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if want := int64(3770); got[LabelWeekly] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelWeekly, got[LabelWeekly], want)
	}
	if want := int64(155); got[LabelFiveHour] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelFiveHour, got[LabelFiveHour], want)
	}
}

// TestFromUsagesFiveHourInHours verifies that a 5-hour window reported
// in TIME_UNIT_HOUR also matches.
func TestFromUsagesFiveHourInHours(t *testing.T) {
	t.Parallel()

	u := kimi.Usages{
		Usage: kimi.Quota{Limit: 5000, Used: n(1230), Remaining: n(3770)},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 5, TimeUnit: "TIME_UNIT_HOUR"},
			Detail: kimi.Quota{Limit: 200, Used: n(45), Remaining: n(155)},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]int64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if want := int64(155); got[LabelFiveHour] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelFiveHour, got[LabelFiveHour], want)
	}
}

// TestFromUsagesNoFiveHourWindow verifies that the 5h entry is omitted
// when the response has no 5-hour rate-limit window. The weekly value
// is still emitted as a single number.
func TestFromUsagesNoFiveHourWindow(t *testing.T) {
	t.Parallel()

	u := kimi.Usages{
		Usage:  kimi.Quota{Limit: 5000, Used: n(1230), Remaining: n(3770)},
		Limits: []kimi.Rate{{Window: kimi.Window{Duration: 1, TimeUnit: "TIME_UNIT_DAY"}}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]int64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if _, ok := got[LabelFiveHour]; ok {
		t.Errorf("payload contains %q, want it omitted", LabelFiveHour)
	}
	if want := int64(3770); got[LabelWeekly] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelWeekly, got[LabelWeekly], want)
	}
}

// TestFromUsagesRemainingFallback verifies that the remaining value is
// derived from limit - used when the API omits the remaining field.
func TestFromUsagesRemainingFallback(t *testing.T) {
	t.Parallel()

	u := kimi.Usages{
		Usage: kimi.Quota{Limit: 100, Used: n(93), Remaining: n(7)},
		Limits: []kimi.Rate{{
			Window: kimi.Window{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"},
			Detail: kimi.Quota{Limit: 100, Used: n(100)},
		}},
	}

	data, err := FromUsages(u)
	if err != nil {
		t.Fatalf("FromUsages() error = %v", err)
	}

	var got map[string]int64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if want := int64(7); got[LabelWeekly] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelWeekly, got[LabelWeekly], want)
	}
	if want := int64(0); got[LabelFiveHour] != want {
		t.Errorf("payload[%q] = %v, want %v", LabelFiveHour, got[LabelFiveHour], want)
	}
}
