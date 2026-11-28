package llmpool

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRetryDelayBounds(t *testing.T) {
	// Full jitter: floor = ceil/2, ceiling doubles per attempt, capped.
	for attempt := 0; attempt < 8; attempt++ {
		for i := 0; i < 200; i++ {
			d := RetryDelay(attempt)
			ceil := RetryBase << attempt
			if ceil > RetryCap {
				ceil = RetryCap
			}
			if d < ceil/2 || d > ceil {
				t.Fatalf("attempt %d: delay %v outside [%v, %v]", attempt, d, ceil/2, ceil)
			}
		}
	}
	if RetryDelay(100) > RetryCap {
		t.Fatal("uncapped delay")
	}
}

func TestExplainErrorBYOK(t *testing.T) {
	raw := fmt.Errorf("openai: status 404: %s",
		`{"error":{"message":"No endpoints found that can be served by your own provider keys. Your workspace is restricted to BYOK capacity, but you haven't added any provider keys. Add provider keys at https://openrouter.ai/workspaces/default/byok so your requests can be served."}}`)
	err := ExplainError(raw)
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected ActionableError, got %T: %v", err, err)
	}
	if !strings.Contains(ae.Fix, "openrouter.ai/workspaces/default/byok") {
		t.Errorf("fix must name the BYOK key page: %q", ae.Fix)
	}
	if !errors.Is(err, raw) {
		t.Error("must unwrap to the raw error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "BYOK") {
		t.Errorf("message must name the restriction: %q", msg)
	}
}

func TestExplainError404(t *testing.T) {
	raw := errors.New(`openai: status 404: {"error":{"message":"The model \"umbasa-old\" does not exist"}}`)
	err := ExplainError(raw)
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if !strings.Contains(ae.What, "umbasa-old") {
		t.Errorf("must name the failing model: %q", ae.What)
	}
	if !strings.Contains(ae.Fix, "catalog") {
		t.Errorf("must point at the catalog check: %q", ae.Fix)
	}
}

func TestExplainErrorPassthrough(t *testing.T) {
	// Network errors and unknown statuses are not decorated.
	raw := errors.New("dial tcp 10.0.0.1:443: connect: connection refused")
	if got := ExplainError(raw); got != raw {
		t.Errorf("passthrough changed the error: %v", got)
	}
	// Already actionable: idempotent.
	ae := &ActionableError{What: "x", Fix: "y", Cause: raw}
	if got := ExplainError(ae); got != ae {
		t.Error("double wrap")
	}
}

func TestEmptyPoolNamedReason(t *testing.T) {
	if New(nil, nil) != nil {
		t.Fatal("empty pool must be nil so callers keep the bare provider path")
	}
	// The named reason exists and reads as operator guidance, not a stack trace.
	if !strings.Contains(ErrEmptyPool.Error(), "no usable LLM profile") {
		t.Errorf("ErrEmptyPool wording: %q", ErrEmptyPool)
	}
}

func TestErrExhaustedWrapsExplained(t *testing.T) {
	byok := fmt.Errorf("openai: status 404: workspace is restricted to byok capacity, add keys at https://openrouter.ai/workspaces/default/byok")
	wrapped := fmt.Errorf("%w: %w", ErrExhausted, ExplainError(byok))
	if !errors.Is(wrapped, ErrExhausted) {
		t.Error("must unwrap to ErrExhausted")
	}
	var ae *ActionableError
	if !errors.As(wrapped, &ae) {
		t.Error("must carry the actionable explanation")
	}
	if !strings.Contains(wrapped.Error(), "LLM failover: all profiles unavailable") {
		t.Errorf("must name the exhausted chain: %q", wrapped)
	}
}
