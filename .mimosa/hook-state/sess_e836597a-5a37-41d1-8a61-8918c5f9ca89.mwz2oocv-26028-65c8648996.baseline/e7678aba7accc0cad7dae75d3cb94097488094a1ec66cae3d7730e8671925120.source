package llmpool

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Autumn-27/norma/llm"
)

// Member is one LLM profile in the chain, already built into a provider (wrapped
// with the recorder by the caller, so a failed attempt is still recorded under
// its own profile name).
type Member struct {
	ID       int64  // llm_profiles.id
	Name     string // profile name, for logs / UI
	Model    string
	Format   string // "anthropic" | "openai"
	Priority int    // the profile's configured priority (display only)
	Active   bool   // is_default (display only)
	// Rank is the ordering key the caller assigned: higher goes first, and members
	// sharing a Rank take turns leading (load-spreading across duplicate keys).
	// The caller encodes "active profile heads the chain" as a Rank above every
	// user-settable priority, so this type needs no policy of its own.
	Rank int
	// WindowTokens is the profile's context window in tokens. A member whose
	// window can't hold the request is skipped rather than made to fail on it.
	WindowTokens int

	// Prov is the construction-time provider. Reads go through provider() so a
	// mid-flight re-resolution (cur) is honored safely.
	Prov llm.Provider

	// provMu guards cur: the 404 re-resolve swaps the provider mid-flight, and
	// pools are shared across concurrent tasks, so reads and the swap must be
	// synchronized. Prov stays the construction-time value; cur starts nil and
	// falls back to it until a re-resolution installs something better.
	provMu sync.RWMutex
	cur    llm.Provider
}

// provider returns the member's current provider (the swapped one when a
// re-resolution has replaced it).
func (m *Member) provider() llm.Provider {
	m.provMu.RLock()
	defer m.provMu.RUnlock()
	if m.cur != nil {
		return m.cur
	}
	return m.Prov
}

// swapProvider installs a re-resolved provider for all future calls.
func (m *Member) swapProvider(p llm.Provider) {
	m.provMu.Lock()
	defer m.provMu.Unlock()
	m.cur = p
}

// RankActive is the Rank the caller gives the chain head (the active profile, or
// an explicitly bound one) so it always outranks any configured priority.
const RankActive = int(^uint(0)>>1) - 1

// ErrExhausted is returned when every member of the chain failed.
var ErrExhausted = errors.New("LLM failover: all profiles unavailable")

// Pool is an llm.Provider that fails over across an ordered chain of members.
// It is safe for concurrent use: member order is immutable after construction
// and all mutable state lives either in the shared Registry or behind per-member
// locks.
type Pool struct {
	members []*Member // in chain order (active first, then priority DESC)
	health  *Registry
	rr      atomic.Uint64 // rotates the starting point within an equal-priority group

	// Resolve, when set, is invoked after a member 404s: it maps the failure to
	// a rebuilt provider whose model id was re-resolved against the provider's
	// live catalog. Returning ok=true swaps the member's provider and the
	// request retries that member once with the corrected model. Nil disables
	// re-resolution (the member just fails over as before).
	Resolve func(m *Member, reqErr error) (llm.Provider, bool)
}

// New builds a Pool over members (already in chain order). Returns nil when the
// chain is empty. A single-member chain is still a valid Pool — it just behaves
// exactly like the bare provider.
func New(members []*Member, health *Registry) *Pool {
	if len(members) == 0 {
		return nil
	}
	if health == nil {
		health = NewRegistry(nil, nil)
	}
	return &Pool{members: members, health: health}
}

// Members returns the chain in order (read-only).
func (p *Pool) Members() []*Member { return p.members }

// Head returns the first member of the chain.
func (p *Pool) Head() *Member { return p.members[0] }

// Stream implements llm.Provider with failover.
//
// The one hard rule: a member may only be abandoned BEFORE it has yielded any
// event. Once text or a tool_use has reached the caller, re-sending the same
// request to another model would duplicate output and corrupt the conversation
// history — so a mid-stream failure is surfaced as-is and left to the agent
// harness's resume logic. Fortunately the failures this exists for (402 no
// credit, 401 bad key, 429, 5xx) all surface during request establishment,
// before the body is read, so they always land in the safe window.
func (p *Pool) Stream(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	order := p.order(req)
	return func(yield func(llm.StreamEvent, error) bool) {
		var lastErr error
		for i, m := range order {
			failed := p.callWithRetry(ctx, m, req, func(ev llm.StreamEvent, err error) bool {
				return yield(ev, err)
			})
			if failed == nil {
				p.health.Pass(m.ID) // completed (or failed in a non-failover way)
				return
			}
			lastErr = failed
			p.logFailure(m, failed)
			hard := isHardFailure(failed)
			if p.health.Trip(m.ID, trimErr(failed), hard) {
				log.Printf("[llmpool] profile %q(%s) id=%d breaker tripped: status=%d: %s", m.Name, m.Model, m.ID, statusOf(failed), trimErr(failed))
			}
			if i+1 < len(order) {
				n := order[i+1]
				log.Printf("[llmpool] LLM failover: %q(%s) → %q(%s), reason: %s",
					m.Name, m.Model, n.Name, n.Model, trimErr(failed))
			}
		}
		if lastErr == nil {
			lastErr = ErrExhausted
		}
		explained := ExplainError(lastErr)
		log.Printf("[llmpool] failover chain exhausted (all %d profiles failed), last error: %s", len(order), trimErr(explained))
		yield(llm.StreamEvent{}, fmt.Errorf("%w: %w", ErrExhausted, explained))
	}
}

// callWithRetry drives one member's Stream, transparently handling the two
// recoverable failure classes: transient 429/5xx (retried in place with
// jittered backoff) and a 404 whose model id can be re-resolved (provider
// swapped, retried once). Returns nil on success — including a mid-stream
// failure AFTER output started, which is yielded to the caller as-is (the
// caller's yield returning false aborts). A non-nil return is a pre-output
// failure: failover-eligible, the member is abandoned.
func (p *Pool) callWithRetry(ctx context.Context, m *Member, req llm.CompletionRequest, yield func(llm.StreamEvent, error) bool) error {
	resolved := false
	for attempt := 0; ; attempt++ {
		var failed error
		emitted := false
		for ev, err := range m.provider().Stream(ctx, req) {
			if err != nil && !emitted && shouldFailover(ctx, err) {
				failed = err
				break // safe window: nothing reached the caller yet
			}
			emitted = true
			if !yield(ev, err) {
				return nil // caller stopped consuming (cancel / early exit)
			}
			if err != nil {
				return nil // terminal error already handed to the caller
			}
		}
		if failed == nil {
			return nil
		}
		// 404: one re-resolution attempt against the live catalog.
		if statusOf(failed) == 404 && !resolved && p.Resolve != nil {
			resolved = true
			if prov, ok := p.Resolve(m, failed); ok {
				m.swapProvider(prov)
				log.Printf("[llmpool] re-resolved model for %q(%s), retrying with corrected id", m.Name, m.Model)
				continue
			}
		}
		// Transient: jittered backoff, bounded by RetryAttempts.
		if !isTransient(failed) || attempt >= RetryAttempts-1 {
			return failed
		}
		d := RetryDelay(attempt)
		log.Printf("[llmpool] profile %q(%s) status=%d transient, retry in %s (%d/%d)",
			m.Name, m.Model, statusOf(failed), d, attempt+1, RetryAttempts-1)
		if !backoffWait(ctx, d) {
			return nil // cancelled while waiting: surface nothing further
		}
	}
}

// Complete implements llm.Provider with failover for non-streaming calls. A
// non-streaming request is atomic — it never delivers partial output — so every
// failover-eligible failure lands in the safe window and the next member can be
// tried without risk of duplicated output. Mirrors Stream's health-tripping and
// chain-exhaustion behavior.
func (p *Pool) Complete(ctx context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	order := p.order(req)
	var lastErr error
	for i, m := range order {
		msg, sr, usage, err := p.completeWithRetry(ctx, m, req)
		if err == nil {
			p.health.Pass(m.ID)
			return msg, sr, usage, nil
		}
		if !shouldFailover(ctx, err) {
			// Non-failover error (e.g. ctx cancel, deterministic 4xx): surface as-is
			// without tripping health, matching Stream's non-failover path.
			p.health.Pass(m.ID)
			return llm.Message{}, "", llm.Usage{}, err
		}
		lastErr = err
		p.logFailure(m, err)
		hard := isHardFailure(err)
		if p.health.Trip(m.ID, trimErr(err), hard) {
			log.Printf("[llmpool] profile %q(%s) id=%d breaker tripped: status=%d: %s", m.Name, m.Model, m.ID, statusOf(err), trimErr(err))
		}
		if i+1 < len(order) {
			n := order[i+1]
			log.Printf("[llmpool] LLM failover: %q(%s) → %q(%s), reason: %s",
				m.Name, m.Model, n.Name, n.Model, trimErr(err))
		}
	}
	if lastErr == nil {
		lastErr = ErrExhausted
	}
	explained := ExplainError(lastErr)
	log.Printf("[llmpool] failover chain exhausted (all %d profiles failed), last error: %s", len(order), trimErr(explained))
	return llm.Message{}, "", llm.Usage{}, fmt.Errorf("%w: %w", ErrExhausted, explained)
}

// completeWithRetry mirrors callWithRetry for the non-streaming surface: a 404
// gets one re-resolution, transient statuses get jittered backoff retries.
// Non-streaming calls are atomic — a retry can never duplicate output.
func (p *Pool) completeWithRetry(ctx context.Context, m *Member, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	resolved := false
	for attempt := 0; ; attempt++ {
		msg, sr, usage, err := m.provider().Complete(ctx, req)
		if err == nil {
			return msg, sr, usage, nil
		}
		if statusOf(err) == 404 && !resolved && p.Resolve != nil {
			resolved = true
			if prov, ok := p.Resolve(m, err); ok {
				m.swapProvider(prov)
				log.Printf("[llmpool] re-resolved model for %q(%s), retrying with corrected id", m.Name, m.Model)
				continue
			}
		}
		if !isTransient(err) || attempt >= RetryAttempts-1 {
			return llm.Message{}, "", llm.Usage{}, err
		}
		d := RetryDelay(attempt)
		log.Printf("[llmpool] profile %q(%s) status=%d transient, retry in %s (%d/%d)",
			m.Name, m.Model, statusOf(err), d, attempt+1, RetryAttempts-1)
		if !backoffWait(ctx, d) {
			return llm.Message{}, "", llm.Usage{}, ctx.Err()
		}
	}
}

// logFailure records the three things the operator needs on every failed
// member attempt: which profile, which model id, and what status the provider
// returned. Deduplicated against the breaker-trip line by being the single
// "attempt failed" record; the trip line marks state changes only.
func (p *Pool) logFailure(m *Member, err error) {
	log.Printf("[llmpool] member failed: profile=%q id=%d model=%q status=%d: %s",
		m.Name, m.ID, m.Model, statusOf(err), trimErr(err))
}

// isTransient reports whether err is worth an in-place retry: rate limits and
// server errors, nothing else. 4xx (other than 429/408) fails fast.
func isTransient(err error) bool {
	switch code := statusOf(err); {
	case code == 429, code == 408:
		return true
	case code >= 500:
		return true
	}
	return false
}

// backoffWait sleeps d, returning false when ctx ends first.
func backoffWait(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// order picks the members to try, in order: skip those in a cooling-off window
// and those whose context window can't hold this request, then rotate within each
// equal-priority group so same-priority profiles share the load. Never returns an
// empty slice — if everything is filtered out, the head of the chain is tried
// anyway, since stalling the engine is worse than one more failed request.
func (p *Pool) order(req llm.CompletionRequest) []*Member {
	est := estimateTokens(req)
	var open []*Member
	for _, m := range p.members {
		if p.health.IsOpen(m.ID) {
			continue
		}
		if m.WindowTokens > 0 && est > m.WindowTokens {
			continue // would 400 on length — not a useful failover target
		}
		open = append(open, m)
	}
	if len(open) == 0 {
		return p.members[:1] // last resort: probe the head rather than stall
	}
	return rotateGroups(open, p.rr.Add(1)-1)
}

// rotateGroups rotates each run of equal-ranked members by n, so profiles sharing
// a priority take turns going first (free load-spreading across duplicate keys).
// The chain head has its own Rank and always stays at the front.
func rotateGroups(in []*Member, n uint64) []*Member {
	out := make([]*Member, 0, len(in))
	for i := 0; i < len(in); {
		j := i + 1
		for j < len(in) && in[j].Rank == in[i].Rank {
			j++
		}
		g := in[i:j]
		if len(g) > 1 {
			off := int(n % uint64(len(g)))
			for k := range g {
				out = append(out, g[(k+off)%len(g)])
			}
		} else {
			out = append(out, g...)
		}
		i = j
	}
	return out
}

// statusRe pulls the HTTP status out of the SDK's error text, which is formatted
// as "<prefix>: status <code>: <body>" (norma/llm/retry.go). The SDK exposes no
// typed error, so the string is what we have.
var statusRe = regexp.MustCompile(`status (\d{3})`)

// statusOf returns the HTTP status carried by err, or 0 if it isn't one.
func statusOf(err error) int {
	m := statusRe.FindStringSubmatch(err.Error())
	if m == nil {
		return 0
	}
	code, _ := strconv.Atoi(m[1])
	return code
}

// shouldFailover reports whether err justifies trying the next profile.
//
// Never fails over on context cancellation — that's the user stopping a task or a
// task-level timeout, and burning a backup key on it would both waste credit and
// pollute the run's termination diagnosis. Never on 400 either: a malformed or
// over-long request fails identically everywhere.
func shouldFailover(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	switch code := statusOf(err); {
	case code == 0:
		return true // no status → transport-level failure (reset / DNS / timeout)
	case code == 400:
		return false // bad or over-long request: identical everywhere
	case code == 401, code == 402, code == 403, code == 404, code == 408, code == 429:
		return true
	case code >= 500:
		return true
	}
	return false
}

// isHardFailure reports whether the failure is deterministic (the profile will
// keep failing until a human fixes it) rather than transient. Hard failures open
// the breaker on the first occurrence.
func isHardFailure(err error) bool {
	switch statusOf(err) {
	case 401, 402, 403, 404:
		return true
	}
	return false
}

// trimErr shortens an error for logs/UI — provider bodies can be long.
func trimErr(err error) string {
	s := strings.TrimSpace(strings.ReplaceAll(err.Error(), "\n", " "))
	if len(s) > 300 {
		s = s[:300] + "…"
	}
	return s
}

// estimateTokens roughly sizes a request so members whose context window clearly
// can't hold it are skipped. Deliberately crude (~3.5 chars/token) and biased to
// over-estimate slightly; it only needs to tell "fits" from "nowhere near".
func estimateTokens(req llm.CompletionRequest) int {
	n := 0
	for _, s := range req.System {
		n += len(s)
	}
	for _, m := range req.Messages {
		n += blocksLen(m.Content)
	}
	for _, t := range req.Tools {
		// InputSchema is a decoded map, so its serialized size isn't available
		// cheaply — charge a flat ~120 chars per top-level property instead.
		n += len(t.Name) + len(t.Description) + len(t.InputSchema)*120
	}
	return n * 2 / 7 // ≈ len/3.5
}

func blocksLen(bs []llm.ContentBlock) int {
	n := 0
	for _, b := range bs {
		n += len(b.Text) + len(b.Thinking) + len(b.Input) + len(b.Name)
		if len(b.Content) > 0 {
			n += blocksLen(b.Content)
		}
	}
	return n
}
