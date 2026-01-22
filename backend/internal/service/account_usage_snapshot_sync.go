package service

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	claudeUsageSnapshotKey     = "claude_usage_snapshot"
	claudeUsageUpdatedAtKey    = "claude_usage_updated_at"
	claudeUsageSnapshotSource  = "claude_usage_source"
	geminiUsageSnapshotKey     = "gemini_usage_snapshot"
	geminiUsageUpdatedAtKey    = "gemini_usage_updated_at"
	geminiUsageSnapshotSource  = "gemini_usage_source"
	usageSnapshotSourceGateway = "gateway"
	usageSnapshotSourceTest    = "test"
	usageSnapshotSourceProbe   = "probe"
)

func (s *AccountUsageService) SyncUsageSnapshotToExtra(ctx context.Context, account *Account, source string) error {
	if s == nil || s.accountRepo == nil || account == nil || account.ID <= 0 {
		return nil
	}

	now := time.Now().UTC()
	if strings.TrimSpace(source) == "" {
		source = usageSnapshotSourceGateway
	}

	switch account.Platform {
	case PlatformAnthropic:
		usage, err := s.computeClaudeUsageSnapshot(ctx, account, now)
		if err != nil {
			return err
		}
		if usage == nil {
			return nil
		}
		if usage.UpdatedAt == nil {
			usage.UpdatedAt = &now
		}

		updates := map[string]any{
			claudeUsageSnapshotKey:    usage,
			claudeUsageUpdatedAtKey:   now.Format(time.RFC3339),
			claudeUsageSnapshotSource: source,
		}
		if err := s.accountRepo.UpdateExtra(ctx, account.ID, updates); err != nil {
			return err
		}

		// If the quota is already full, persist a real reset window to account.rate_limit_reset_at
		// so the scheduler can skip the account until the window ends.
		if resetAt := claudeRateLimitResetAt(usage); resetAt != nil && resetAt.After(time.Now()) {
			_ = s.accountRepo.SetRateLimited(ctx, account.ID, *resetAt)
		}

		return nil

	case PlatformGemini:
		usage, err := s.computeGeminiUsageSnapshot(ctx, account, now)
		if err != nil {
			return err
		}
		if usage == nil {
			return nil
		}
		if usage.UpdatedAt == nil {
			usage.UpdatedAt = &now
		}

		updates := map[string]any{
			geminiUsageSnapshotKey:    usage,
			geminiUsageUpdatedAtKey:   now.Format(time.RFC3339),
			geminiUsageSnapshotSource: source,
		}
		return s.accountRepo.UpdateExtra(ctx, account.ID, updates)
	default:
		return nil
	}
}

// MaybeSyncUsageSnapshotToExtra performs a best-effort sync with an in-memory throttle.
// NOTE: This is per-process only; multi-instance deployments still sync independently.
func (s *AccountUsageService) MaybeSyncUsageSnapshotToExtra(ctx context.Context, account *Account, source string, minInterval time.Duration) error {
	if s == nil || account == nil || account.ID <= 0 {
		return nil
	}
	if minInterval <= 0 {
		return s.SyncUsageSnapshotToExtra(ctx, account, source)
	}

	now := time.Now()
	if v, ok := s.usageSnapshotSyncCache.Load(account.ID); ok {
		if last, ok := v.(time.Time); ok && now.Sub(last) < minInterval {
			return nil
		}
	}
	s.usageSnapshotSyncCache.Store(account.ID, now)
	return s.SyncUsageSnapshotToExtra(ctx, account, source)
}

func (s *AccountUsageService) computeClaudeUsageSnapshot(ctx context.Context, account *Account, now time.Time) (*UsageInfo, error) {
	if account == nil || account.Platform != PlatformAnthropic {
		return nil, nil
	}
	// API Key accounts do not support Claude OAuth usage API.
	if account.Type == AccountTypeAPIKey {
		return nil, nil
	}

	// OAuth accounts: call Anthropic OAuth usage API.
	if account.CanGetUsage() {
		apiResp, err := s.fetchOAuthUsageRaw(ctx, account)
		if err != nil {
			return nil, err
		}
		usage := s.buildUsageInfo(apiResp, &now)
		s.addWindowStats(ctx, account, usage)
		return usage, nil
	}

	// Setup Token accounts: estimate 5h window from session_window fields.
	if account.Type == AccountTypeSetupToken {
		usage := s.estimateSetupTokenUsage(account)
		usage.UpdatedAt = &now
		s.addWindowStats(ctx, account, usage)
		return usage, nil
	}

	return nil, fmt.Errorf("unsupported anthropic account type: %s", account.Type)
}

func (s *AccountUsageService) computeGeminiUsageSnapshot(ctx context.Context, account *Account, now time.Time) (*UsageInfo, error) {
	if account == nil || account.Platform != PlatformGemini {
		return nil, nil
	}

	usage, err := s.getGeminiUsage(ctx, account)
	if err != nil {
		return nil, err
	}
	if usage != nil && usage.UpdatedAt == nil {
		usage.UpdatedAt = &now
	}
	return usage, nil
}

func claudeRateLimitResetAt(usage *UsageInfo) *time.Time {
	if usage == nil {
		return nil
	}
	if usage.SevenDay != nil && usage.SevenDay.ResetsAt != nil && usage.SevenDay.Utilization >= 100 {
		return usage.SevenDay.ResetsAt
	}
	if usage.FiveHour != nil && usage.FiveHour.ResetsAt != nil && usage.FiveHour.Utilization >= 100 {
		return usage.FiveHour.ResetsAt
	}
	return nil
}
