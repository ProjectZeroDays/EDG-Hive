package llmpool

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
	"time"
)

// repair.go turns provider failures into either (a) a retry with backoff and
// jitter, or (b) an actionable error that names the fix — instead of a raw
// status code bubbling up to the planner.

// RetryBase / RetryAttempts bound the pool-level retry of one member on
// transient statuses (429 / 5xx). The SDK already retries request
// establishment; this layer catches what slips through (mid-window resets after
// the SDK's ladder is exhausted) and, more importantly, adds the jitter the
// fixed SDK ladder lacks, so concurrent tasks don't re-hit a recovering
// provider in lockstep.
const (
	RetryAttempts = 3
	RetryBase     = 500 * time.Millisecond
	RetryCap      = 8 * time.Second
)

// RetryDelay is the full-jittered backoff for attempt n (0-based):
// uniform random in [base/2, min(base·2ⁿ, cap)]. The floor keeps tests and
// recovery fast; the ceiling keeps a wedged provider from stalling a request
// for minutes.
func RetryDelay(attempt int) time.Duration {
	ceil := RetryBase << attempt
	if ceil > RetryCap {
		ceil = RetryCap
	}
	floor := ceil / 2
	return floor + time.Duration(rand.Int64N(int64(ceil-floor+1)))
}

// ErrEmptyPool is returned (wrapped) when a role has zero usable profiles —
// nothing configured, or everything is tripped, excluded or invalid. It exists
// so callers can tell "no model" apart from "model broken" and say so.
var ErrEmptyPool = errors.New(
	"no usable LLM profile: none configured, or every configured profile is disabled, tripped, or invalid — check System → LLM Profiles")

// ActionableError wraps a provider failure with the concrete fix, so the
// planner and the UI can show "do X" instead of a status code.
type ActionableError struct {
	What  string // what happened, one line
	Fix   string // the action that resolves it
	Cause error  // the original provider error
}

func (e *ActionableError) Error() string {
	if e.Fix == "" {
		return e.What
	}
	return e.What + " — " + e.Fix
}
func (e *ActionableError) Unwrap() error { return e.Cause }

// byokNeedle matches the OpenRouter workspace restriction: the account is on
// BYOK-only capacity with no provider keys attached, so NO model on the
// free/default routes can be served.
const byokNeedle = "restricted to byok capacity"

// openRouterBYOKFix names the key and the limit, per the part spec.
const openRouterBYOKFix = "OpenRouter workspace is BYOK-restricted: attach provider keys at https://openrouter.ai/workspaces/default/byok (or switch this profile to a provider that serves its own models)"

// ExplainError maps a raw provider error to an actionable one where the fix is
// known. Unknown errors pass through unchanged.
func ExplainError(err error) error {
	if err == nil {
		return nil
	}
	var ae *ActionableError
	if errors.As(err, &ae) {
		return err // already actionable
	}
	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), byokNeedle) {
		return &ActionableError{
			What:  "OpenRouter refused the request: workspace restricted to BYOK capacity",
			Fix:   openRouterBYOKFix,
			Cause: err,
		}
	}
	if code := statusOf(err); code == 404 {
		// Name the model when the provider body carries it: OpenAI-style
		// `The model "x" does not exist`, backtick variant, Ollama's
		// `model 'x' not found`.
		what := "model id not found on provider (404)"
		if model := modelNamedIn(msg); model != "" {
			what = fmt.Sprintf("model %q not found on provider (404)", model)
		}
		return &ActionableError{
			What:  what,
			Fix:   "check the model id against the provider's catalog (System → LLM → Test), or let the boot repair pass resolve it automatically",
			Cause: err,
		}
	}
	return err
}

var modelNamedRe = regexp.MustCompile(`(?:model ["'\x60]([^"'\x60]+)["'\x60]|["'\x60]([^"'\x60]+)["'\x60] does not exist)`)

// modelNamedIn extracts a model id from common provider 404 bodies.
func modelNamedIn(s string) string {
	m := modelNamedRe.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}
