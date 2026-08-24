// Package payload maps Kimi quota data to the MQTT payload format
// that the ESP32 consumer expects.
package payload

import (
	"encoding/json"

	"github.com/jsvensson/kimi-quota-monitor/internal/kimi"
)

// LabelFiveHour is the payload label for the 5-hour rolling window.
const LabelFiveHour = "5h"

// LabelWeekly is the payload label for the weekly quota window.
const LabelWeekly = "7d"

// windowValue is the per-window payload: the remaining percentage and
// the Unix epoch timestamp when the window resets.
type windowValue struct {
	Pct      int64 `json:"pct"`
	ResetsAt int64 `json:"resets_at"`
}

// FromUsages builds the MQTT payload from a Usages response.
// The payload maps window labels to {"pct": <remaining>, "resets_at": <unix epoch>}.
// The 5h entry is omitted if the response has no 5-hour rate-limit window.
func FromUsages(u kimi.Usages) ([]byte, error) {
	p := map[string]windowValue{
		LabelWeekly: {
			Pct:      u.Usage.RemainingAmount(),
			ResetsAt: u.Usage.ResetTime.Unix(),
		},
	}
	if q, ok := fiveHourQuota(u); ok {
		p[LabelFiveHour] = windowValue{
			Pct:      q.RemainingAmount(),
			ResetsAt: q.ResetTime.Unix(),
		}
	}
	return json.Marshal(p)
}

// fiveHourMinutes is the length of the 5-hour rolling window in minutes.
const fiveHourMinutes = 5 * 60

// fiveHourQuota returns the quota of the 5-hour rolling rate-limit window.
// The API may report the window in any time unit (e.g. 300 minutes),
// so windows are matched by their normalized length in minutes.
func fiveHourQuota(u kimi.Usages) (kimi.Quota, bool) {
	for _, l := range u.Limits {
		if m, ok := l.Window.Minutes(); ok && m == fiveHourMinutes {
			return l.Detail, true
		}
	}
	return kimi.Quota{}, false
}
