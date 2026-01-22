package service

import (
	"time"
)

type codexDerivedUsageSnapshot struct {
	updates map[string]any

	updatedAt time.Time

	fiveHourUsedPercent       *float64
	fiveHourResetAfterSeconds *int
	fiveHourWindowMinutes     *int
	fiveHourResetAt           *time.Time

	sevenDayUsedPercent       *float64
	sevenDayResetAfterSeconds *int
	sevenDayWindowMinutes     *int
	sevenDayResetAt           *time.Time
}

func deriveCodexUsageSnapshot(snapshot *OpenAICodexUsageSnapshot) *codexDerivedUsageSnapshot {
	if snapshot == nil {
		return nil
	}

	updates := make(map[string]any)

	if snapshot.PrimaryUsedPercent != nil {
		updates["codex_primary_used_percent"] = *snapshot.PrimaryUsedPercent
	}
	if snapshot.PrimaryResetAfterSeconds != nil {
		updates["codex_primary_reset_after_seconds"] = *snapshot.PrimaryResetAfterSeconds
	}
	if snapshot.PrimaryWindowMinutes != nil {
		updates["codex_primary_window_minutes"] = *snapshot.PrimaryWindowMinutes
	}
	if snapshot.SecondaryUsedPercent != nil {
		updates["codex_secondary_used_percent"] = *snapshot.SecondaryUsedPercent
	}
	if snapshot.SecondaryResetAfterSeconds != nil {
		updates["codex_secondary_reset_after_seconds"] = *snapshot.SecondaryResetAfterSeconds
	}
	if snapshot.SecondaryWindowMinutes != nil {
		updates["codex_secondary_window_minutes"] = *snapshot.SecondaryWindowMinutes
	}
	if snapshot.PrimaryOverSecondaryPercent != nil {
		updates["codex_primary_over_secondary_percent"] = *snapshot.PrimaryOverSecondaryPercent
	}
	if snapshot.UpdatedAt != "" {
		updates["codex_usage_updated_at"] = snapshot.UpdatedAt
	}

	updatedAt := time.Now()
	if snapshot.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, snapshot.UpdatedAt); err == nil {
			updatedAt = t
		}
	}

	// Normalize to canonical 5h/7d fields based on window_minutes
	// This fixes the issue where OpenAI's primary/secondary naming is ambiguous across accounts.
	//
	// IMPORTANT: We can only reliably determine window type from window_minutes field.
	// reset_after_seconds is remaining time, not window size, so it cannot be used for comparison.

	var primaryWindowMins, secondaryWindowMins int
	var hasPrimaryWindow, hasSecondaryWindow bool

	if snapshot.PrimaryWindowMinutes != nil {
		primaryWindowMins = *snapshot.PrimaryWindowMinutes
		hasPrimaryWindow = true
	}
	if snapshot.SecondaryWindowMinutes != nil {
		secondaryWindowMins = *snapshot.SecondaryWindowMinutes
		hasSecondaryWindow = true
	}

	var use5hFromPrimary, use7dFromPrimary bool
	var use5hFromSecondary, use7dFromSecondary bool

	if hasPrimaryWindow && hasSecondaryWindow {
		// Both window sizes known: compare and assign smaller to 5h, larger to 7d.
		if primaryWindowMins < secondaryWindowMins {
			use5hFromPrimary = true
			use7dFromSecondary = true
		} else {
			use5hFromSecondary = true
			use7dFromPrimary = true
		}
	} else if hasPrimaryWindow {
		// Only primary window size known: classify by absolute threshold.
		if primaryWindowMins <= 360 {
			use5hFromPrimary = true
		} else {
			use7dFromPrimary = true
		}
	} else if hasSecondaryWindow {
		// Only secondary window size known: classify by absolute threshold.
		if secondaryWindowMins <= 360 {
			use5hFromSecondary = true
		} else {
			use7dFromSecondary = true
		}
	} else {
		// No window_minutes available: cannot reliably determine window types.
		// Fall back to legacy assumption (may be incorrect):
		// assume primary=7d, secondary=5h based on historical observation.
		if snapshot.SecondaryUsedPercent != nil || snapshot.SecondaryResetAfterSeconds != nil || snapshot.SecondaryWindowMinutes != nil {
			use5hFromSecondary = true
		}
		if snapshot.PrimaryUsedPercent != nil || snapshot.PrimaryResetAfterSeconds != nil || snapshot.PrimaryWindowMinutes != nil {
			use7dFromPrimary = true
		}
	}

	d := &codexDerivedUsageSnapshot{
		updates:   updates,
		updatedAt: updatedAt,
	}

	if use5hFromPrimary {
		d.fiveHourUsedPercent = snapshot.PrimaryUsedPercent
		d.fiveHourResetAfterSeconds = snapshot.PrimaryResetAfterSeconds
		d.fiveHourWindowMinutes = snapshot.PrimaryWindowMinutes
		if snapshot.PrimaryUsedPercent != nil {
			updates["codex_5h_used_percent"] = *snapshot.PrimaryUsedPercent
		}
		if snapshot.PrimaryResetAfterSeconds != nil {
			updates["codex_5h_reset_after_seconds"] = *snapshot.PrimaryResetAfterSeconds
		}
		if snapshot.PrimaryWindowMinutes != nil {
			updates["codex_5h_window_minutes"] = *snapshot.PrimaryWindowMinutes
		}
	} else if use5hFromSecondary {
		d.fiveHourUsedPercent = snapshot.SecondaryUsedPercent
		d.fiveHourResetAfterSeconds = snapshot.SecondaryResetAfterSeconds
		d.fiveHourWindowMinutes = snapshot.SecondaryWindowMinutes
		if snapshot.SecondaryUsedPercent != nil {
			updates["codex_5h_used_percent"] = *snapshot.SecondaryUsedPercent
		}
		if snapshot.SecondaryResetAfterSeconds != nil {
			updates["codex_5h_reset_after_seconds"] = *snapshot.SecondaryResetAfterSeconds
		}
		if snapshot.SecondaryWindowMinutes != nil {
			updates["codex_5h_window_minutes"] = *snapshot.SecondaryWindowMinutes
		}
	}

	if use7dFromPrimary {
		d.sevenDayUsedPercent = snapshot.PrimaryUsedPercent
		d.sevenDayResetAfterSeconds = snapshot.PrimaryResetAfterSeconds
		d.sevenDayWindowMinutes = snapshot.PrimaryWindowMinutes
		if snapshot.PrimaryUsedPercent != nil {
			updates["codex_7d_used_percent"] = *snapshot.PrimaryUsedPercent
		}
		if snapshot.PrimaryResetAfterSeconds != nil {
			updates["codex_7d_reset_after_seconds"] = *snapshot.PrimaryResetAfterSeconds
		}
		if snapshot.PrimaryWindowMinutes != nil {
			updates["codex_7d_window_minutes"] = *snapshot.PrimaryWindowMinutes
		}
	} else if use7dFromSecondary {
		d.sevenDayUsedPercent = snapshot.SecondaryUsedPercent
		d.sevenDayResetAfterSeconds = snapshot.SecondaryResetAfterSeconds
		d.sevenDayWindowMinutes = snapshot.SecondaryWindowMinutes
		if snapshot.SecondaryUsedPercent != nil {
			updates["codex_7d_used_percent"] = *snapshot.SecondaryUsedPercent
		}
		if snapshot.SecondaryResetAfterSeconds != nil {
			updates["codex_7d_reset_after_seconds"] = *snapshot.SecondaryResetAfterSeconds
		}
		if snapshot.SecondaryWindowMinutes != nil {
			updates["codex_7d_window_minutes"] = *snapshot.SecondaryWindowMinutes
		}
	}

	// Compute absolute reset timestamps using snapshot updated_at + reset_after_seconds.
	if d.fiveHourResetAfterSeconds != nil {
		resetAt := updatedAt.Add(time.Duration(*d.fiveHourResetAfterSeconds) * time.Second)
		d.fiveHourResetAt = &resetAt
		updates["codex_5h_reset_at"] = resetAt.UTC().Format(time.RFC3339)
	}
	if d.sevenDayResetAfterSeconds != nil {
		resetAt := updatedAt.Add(time.Duration(*d.sevenDayResetAfterSeconds) * time.Second)
		d.sevenDayResetAt = &resetAt
		updates["codex_7d_reset_at"] = resetAt.UTC().Format(time.RFC3339)
	}

	return d
}

func codexRateLimitResetAt(derived *codexDerivedUsageSnapshot) *time.Time {
	if derived == nil {
		return nil
	}
	if derived.sevenDayUsedPercent != nil && *derived.sevenDayUsedPercent >= 100 && derived.sevenDayResetAt != nil {
		return derived.sevenDayResetAt
	}
	if derived.fiveHourUsedPercent != nil && *derived.fiveHourUsedPercent >= 100 && derived.fiveHourResetAt != nil {
		return derived.fiveHourResetAt
	}
	return nil
}
