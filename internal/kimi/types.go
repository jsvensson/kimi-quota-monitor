// Package kimi implements a client for the Kimi Code quota API.
package kimi

import (
	"bytes"
	"strconv"
	"time"
)

// Usages is the response of GET /usages.
type Usages struct {
	SubType string `json:"subType"`
	Usage   Quota  `json:"usage"`
	Limits  []Rate `json:"limits"`
}

// Quota describes a quota window with a limit and the remaining amount.
type Quota struct {
	Limit     Number    `json:"limit"`
	Used      Number    `json:"used"`
	Remaining Number    `json:"remaining"`
	ResetTime time.Time `json:"resetTime"`
}

// UsedAmount returns the used amount of the quota.
// It prefers the explicit "used" field and falls back to limit - remaining.
func (q Quota) UsedAmount() int64 {
	if q.Used > 0 {
		return int64(q.Used)
	}
	return int64(q.Limit - q.Remaining)
}

// Rate describes a rolling rate-limit window.
type Rate struct {
	Window Window `json:"window"`
	Detail Quota  `json:"detail"`
}

// Window describes the duration of a rolling rate-limit window.
type Window struct {
	Duration Number `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

// Minutes returns the window duration in minutes. ok is false for
// unknown time units or non-positive durations.
func (w Window) Minutes() (minutes int64, ok bool) {
	if w.Duration <= 0 {
		return 0, false
	}
	var multiplier int64
	switch w.TimeUnit {
	case "TIME_UNIT_MINUTE":
		multiplier = 1
	case "TIME_UNIT_HOUR":
		multiplier = 60
	case "TIME_UNIT_DAY":
		multiplier = 24 * 60
	default:
		return 0, false
	}
	return int64(w.Duration) * multiplier, true
}

// Number is an int64 that unmarshals from a JSON number or a JSON string.
// The API is protobuf-based and serializes int64 fields as strings.
type Number int64

// UnmarshalJSON decodes a JSON number or string into n.
func (n *Number) UnmarshalJSON(data []byte) error {
	s := string(bytes.Trim(data, `"`))
	if len(s) == 0 || s == "null" {
		*n = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*n = Number(v)
	return nil
}
