package service

import (
	"testing"
	"time"
)

func TestDeriveCodexUsageSnapshot_Assigns5hAnd7dByWindowMinutes(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := now.Format(time.RFC3339)

	primaryUsed := 80.0
	primaryReset := 100
	primaryWindow := 10080

	secondaryUsed := 50.0
	secondaryReset := 200
	secondaryWindow := 300

	snapshot := &OpenAICodexUsageSnapshot{
		PrimaryUsedPercent:         &primaryUsed,
		PrimaryResetAfterSeconds:   &primaryReset,
		PrimaryWindowMinutes:       &primaryWindow,
		SecondaryUsedPercent:       &secondaryUsed,
		SecondaryResetAfterSeconds: &secondaryReset,
		SecondaryWindowMinutes:     &secondaryWindow,
		UpdatedAt:                  updatedAt,
	}

	derived := deriveCodexUsageSnapshot(snapshot)
	if derived == nil {
		t.Fatalf("expected derived snapshot, got nil")
	}

	if derived.fiveHourUsedPercent == nil || *derived.fiveHourUsedPercent != secondaryUsed {
		t.Fatalf("expected 5h used=%v, got=%v", secondaryUsed, derefFloat(derived.fiveHourUsedPercent))
	}
	if derived.sevenDayUsedPercent == nil || *derived.sevenDayUsedPercent != primaryUsed {
		t.Fatalf("expected 7d used=%v, got=%v", primaryUsed, derefFloat(derived.sevenDayUsedPercent))
	}

	if v, ok := derived.updates["codex_5h_used_percent"].(float64); !ok || v != secondaryUsed {
		t.Fatalf("expected updates.codex_5h_used_percent=%v, got=%v", secondaryUsed, derived.updates["codex_5h_used_percent"])
	}
	if v, ok := derived.updates["codex_7d_used_percent"].(float64); !ok || v != primaryUsed {
		t.Fatalf("expected updates.codex_7d_used_percent=%v, got=%v", primaryUsed, derived.updates["codex_7d_used_percent"])
	}

	fiveHourResetAt := now.Add(time.Duration(secondaryReset) * time.Second).UTC().Format(time.RFC3339)
	if v, ok := derived.updates["codex_5h_reset_at"].(string); !ok || v != fiveHourResetAt {
		t.Fatalf("expected updates.codex_5h_reset_at=%q, got=%v", fiveHourResetAt, derived.updates["codex_5h_reset_at"])
	}

	sevenDayResetAt := now.Add(time.Duration(primaryReset) * time.Second).UTC().Format(time.RFC3339)
	if v, ok := derived.updates["codex_7d_reset_at"].(string); !ok || v != sevenDayResetAt {
		t.Fatalf("expected updates.codex_7d_reset_at=%q, got=%v", sevenDayResetAt, derived.updates["codex_7d_reset_at"])
	}
}

func TestCodexRateLimitResetAt_Prefers7dOver5h(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := now.Format(time.RFC3339)

	fiveHourUsed := 100.0
	fiveHourReset := 60
	fiveHourWindow := 300

	sevenDayUsed := 100.0
	sevenDayReset := 3600
	sevenDayWindow := 10080

	// Put 7d in primary, 5h in secondary.
	snapshot := &OpenAICodexUsageSnapshot{
		PrimaryUsedPercent:         &sevenDayUsed,
		PrimaryResetAfterSeconds:   &sevenDayReset,
		PrimaryWindowMinutes:       &sevenDayWindow,
		SecondaryUsedPercent:       &fiveHourUsed,
		SecondaryResetAfterSeconds: &fiveHourReset,
		SecondaryWindowMinutes:     &fiveHourWindow,
		UpdatedAt:                  updatedAt,
	}

	derived := deriveCodexUsageSnapshot(snapshot)
	if derived == nil {
		t.Fatalf("expected derived snapshot, got nil")
	}

	resetAt := codexRateLimitResetAt(derived)
	if resetAt == nil {
		t.Fatalf("expected resetAt, got nil")
	}

	want := now.Add(time.Duration(sevenDayReset) * time.Second)
	if !resetAt.Equal(want) {
		t.Fatalf("expected resetAt=%v, got=%v", want, *resetAt)
	}
}

func TestCodexRateLimitResetAt_Uses5hWhen7dNotFull(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := now.Format(time.RFC3339)

	fiveHourUsed := 100.0
	fiveHourReset := 120
	fiveHourWindow := 300

	sevenDayUsed := 80.0
	sevenDayReset := 3600
	sevenDayWindow := 10080

	snapshot := &OpenAICodexUsageSnapshot{
		PrimaryUsedPercent:         &sevenDayUsed,
		PrimaryResetAfterSeconds:   &sevenDayReset,
		PrimaryWindowMinutes:       &sevenDayWindow,
		SecondaryUsedPercent:       &fiveHourUsed,
		SecondaryResetAfterSeconds: &fiveHourReset,
		SecondaryWindowMinutes:     &fiveHourWindow,
		UpdatedAt:                  updatedAt,
	}

	derived := deriveCodexUsageSnapshot(snapshot)
	if derived == nil {
		t.Fatalf("expected derived snapshot, got nil")
	}

	resetAt := codexRateLimitResetAt(derived)
	if resetAt == nil {
		t.Fatalf("expected resetAt, got nil")
	}

	want := now.Add(time.Duration(fiveHourReset) * time.Second)
	if !resetAt.Equal(want) {
		t.Fatalf("expected resetAt=%v, got=%v", want, *resetAt)
	}
}

func derefFloat(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}
