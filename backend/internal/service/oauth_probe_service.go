package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
)

// OAuthProbeService periodically probes OAuth/Setup-Token accounts to:
// - keep OAuth account availability observable even when not scheduled,
// - refresh upstream quota snapshots (e.g. OpenAI Codex 5h/7d windows),
// - reduce stale/expired quota display issues on the account page.
//
// NOTE: This service is disabled by default because each probe may consume real quota.
type OAuthProbeService struct {
	accountRepo         AccountRepository
	httpUpstream        HTTPUpstream
	openAITokenProvider *OpenAITokenProvider
	geminiTokenProvider *GeminiTokenProvider
	claudeTokenProvider *ClaudeTokenProvider
	accountUsageService *AccountUsageService
	cfg                 *config.OAuthProbeConfig
	securityCfg         *config.SecurityConfig

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewOAuthProbeService(
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	openAITokenProvider *OpenAITokenProvider,
	geminiTokenProvider *GeminiTokenProvider,
	claudeTokenProvider *ClaudeTokenProvider,
	accountUsageService *AccountUsageService,
	cfg *config.Config,
) *OAuthProbeService {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &OAuthProbeService{
		accountRepo:         accountRepo,
		httpUpstream:        httpUpstream,
		openAITokenProvider: openAITokenProvider,
		geminiTokenProvider: geminiTokenProvider,
		claudeTokenProvider: claudeTokenProvider,
		accountUsageService: accountUsageService,
		cfg:                 &cfg.OAuthProbe,
		securityCfg:         &cfg.Security,
		stopCh:              make(chan struct{}),
	}
}

func (s *OAuthProbeService) Start() {
	if s.cfg == nil || !s.cfg.Enabled {
		log.Println("[OAuthProbe] Service disabled by configuration")
		return
	}

	s.wg.Add(1)
	go s.probeLoop()

	log.Printf("[OAuthProbe] Service started (check every %d minutes, idle_threshold=%d minutes, timeout=%ds, concurrency=%d, max_accounts_per_cycle=%d)",
		s.cfg.CheckIntervalMinutes,
		s.cfg.IdleThresholdMinutes,
		s.cfg.RequestTimeoutSeconds,
		s.cfg.MaxConcurrency,
		s.cfg.MaxAccountsPerCycle,
	)
}

func (s *OAuthProbeService) Stop() {
	if s == nil {
		return
	}
	close(s.stopCh)
	s.wg.Wait()
	log.Println("[OAuthProbe] Service stopped")
}

func (s *OAuthProbeService) probeLoop() {
	defer s.wg.Done()

	interval := time.Duration(s.cfg.CheckIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run once on startup to populate usage windows quickly.
	s.processProbeCycle()

	for {
		select {
		case <-ticker.C:
			s.processProbeCycle()
		case <-s.stopCh:
			return
		}
	}
}

func (s *OAuthProbeService) processProbeCycle() {
	if s == nil || s.accountRepo == nil {
		return
	}
	ctx := context.Background()

	accounts, err := s.accountRepo.ListActive(ctx)
	if err != nil {
		log.Printf("[OAuthProbe] Failed to list accounts: %v", err)
		return
	}

	// Filter to OAuth/Setup-Token accounts on OpenAI/Gemini/Claude only.
	targets := make([]Account, 0, len(accounts))
	idleThreshold := time.Duration(s.cfg.IdleThresholdMinutes) * time.Minute
	if idleThreshold < 0 {
		idleThreshold = 0
	}
	for _, a := range accounts {
		if !a.IsOAuth() {
			continue
		}
		switch a.Platform {
		case PlatformOpenAI, PlatformGemini, PlatformAnthropic:
			// Only probe idle accounts as a "fallback" mechanism.
			// If the account was used recently, skip to avoid consuming extra quota.
			if idleThreshold > 0 && a.LastUsedAt != nil && time.Since(*a.LastUsedAt) < idleThreshold {
				continue
			}
			targets = append(targets, a)
		}
	}

	if len(targets) == 0 {
		log.Printf("[OAuthProbe] Cycle complete: total=%d, targets=0", len(accounts))
		return
	}

	maxPerCycle := s.cfg.MaxAccountsPerCycle
	if maxPerCycle > 0 && len(targets) > maxPerCycle {
		targets = targets[:maxPerCycle]
	}

	concurrency := s.cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	var (
		mu          sync.Mutex
		okCount     int
		failCount   int
		openaiCount int
		geminiCount int
		claudeCount int
		openaiOK    int
		geminiOK    int
		claudeOK    int
	)

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := range targets {
		account := targets[i]
		wg.Add(1)
		sem <- struct{}{}

		go func(a Account) {
			defer wg.Done()
			defer func() { <-sem }()

			probeOK := false
			switch a.Platform {
			case PlatformOpenAI:
				mu.Lock()
				openaiCount++
				mu.Unlock()
				if err := s.probeOpenAIOAuth(ctx, &a); err == nil {
					probeOK = true
					mu.Lock()
					openaiOK++
					mu.Unlock()
				}
			case PlatformGemini:
				mu.Lock()
				geminiCount++
				mu.Unlock()
				if err := s.probeGeminiOAuth(ctx, &a); err == nil {
					probeOK = true
					mu.Lock()
					geminiOK++
					mu.Unlock()
				}
			case PlatformAnthropic:
				mu.Lock()
				claudeCount++
				mu.Unlock()
				if err := s.probeClaudeOAuth(ctx, &a); err == nil {
					probeOK = true
					mu.Lock()
					claudeOK++
					mu.Unlock()
				}
			}

			mu.Lock()
			if probeOK {
				okCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}(account)
	}

	wg.Wait()
	log.Printf("[OAuthProbe] Cycle complete: total=%d, targets=%d, ok=%d, failed=%d (openai=%d ok=%d, gemini=%d ok=%d, claude=%d ok=%d)",
		len(accounts), len(targets), okCount, failCount,
		openaiCount, openaiOK, geminiCount, geminiOK, claudeCount, claudeOK,
	)
}

func (s *OAuthProbeService) probeOpenAIOAuth(parent context.Context, account *Account) error {
	if account == nil || account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return nil
	}
	if s.httpUpstream == nil {
		return fmt.Errorf("http upstream not configured")
	}

	timeout := time.Duration(s.cfg.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	token := strings.TrimSpace(account.GetOpenAIAccessToken())
	if s.openAITokenProvider != nil {
		if t, err := s.openAITokenProvider.GetAccessToken(ctx, account); err == nil && strings.TrimSpace(t) != "" {
			token = t
		}
	}
	if token == "" {
		return fmt.Errorf("missing access token")
	}

	nonce, _ := randomHexString(8)
	payload := createOpenAITestPayload(openai.DefaultTestModel, true, nonce)
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatgptCodexAPIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Host = "chatgpt.com"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accept", "text/event-stream")
	req.Header.Set("OpenAI-Beta", "responses=experimental")
	if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
		req.Header.Set("chatgpt-account-id", chatgptAccountID)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		s.saveProbeResult(context.Background(), account.ID, false, 0, err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Persist Codex quota snapshot (headers only).
	s.persistOpenAICodexSnapshot(ctx, account.ID, resp.Header)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		err := fmt.Errorf("openai oauth probe status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
		s.saveProbeResult(context.Background(), account.ID, false, resp.StatusCode, err)
		return err
	}

	s.saveProbeResult(context.Background(), account.ID, true, resp.StatusCode, nil)
	return nil
}

func (s *OAuthProbeService) persistOpenAICodexSnapshot(ctx context.Context, accountID int64, headers http.Header) {
	if s == nil || s.accountRepo == nil {
		return
	}
	if snapshot := extractCodexUsageHeaders(headers); snapshot != nil {
		derived := deriveCodexUsageSnapshot(snapshot)
		if derived == nil || len(derived.updates) == 0 {
			return
		}
		_ = s.accountRepo.UpdateExtra(ctx, accountID, derived.updates)
		if resetAt := codexRateLimitResetAt(derived); resetAt != nil && resetAt.After(time.Now()) {
			_ = s.accountRepo.SetRateLimited(ctx, accountID, *resetAt)
		}
	}
}

func (s *OAuthProbeService) probeGeminiOAuth(parent context.Context, account *Account) error {
	if account == nil || account.Platform != PlatformGemini || account.Type != AccountTypeOAuth {
		return nil
	}
	if s.httpUpstream == nil {
		return fmt.Errorf("http upstream not configured")
	}
	if s.geminiTokenProvider == nil {
		return fmt.Errorf("gemini token provider not configured")
	}

	timeout := time.Duration(s.cfg.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	accessToken, err := s.geminiTokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		s.saveProbeResult(context.Background(), account.ID, false, 0, err)
		return err
	}

	nonce, _ := randomHexString(8)
	payload := createGeminiTestPayload(nonce)

	projectID := strings.TrimSpace(account.GetCredential("project_id"))
	modelID := geminicli.DefaultTestModel

	var req *http.Request
	if projectID == "" {
		baseURL := account.GetCredential("base_url")
		if strings.TrimSpace(baseURL) == "" {
			baseURL = geminicli.AIStudioBaseURL
		}
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return err
		}
		fullURL := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", strings.TrimRight(normalizedBaseURL, "/"), modelID)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
	} else {
		var inner map[string]any
		if err := json.Unmarshal(payload, &inner); err != nil {
			return err
		}
		wrapped := map[string]any{
			"model":   modelID,
			"project": projectID,
			"request": inner,
		}
		wrappedBytes, _ := json.Marshal(wrapped)

		normalizedBaseURL, err := s.validateUpstreamBaseURL(geminicli.GeminiCliBaseURL)
		if err != nil {
			return err
		}
		fullURL := fmt.Sprintf("%s/v1internal:streamGenerateContent?alt=sse", strings.TrimRight(normalizedBaseURL, "/"))
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(wrappedBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("User-Agent", geminicli.GeminiCLIUserAgent)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		s.saveProbeResult(context.Background(), account.ID, false, 0, err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Persist Gemini (simulated) quota snapshot to DB extra.
	if s.accountUsageService != nil {
		updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = s.accountUsageService.SyncUsageSnapshotToExtra(updateCtx, account, usageSnapshotSourceProbe)
		cancel()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		err := fmt.Errorf("gemini oauth probe status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
		s.saveProbeResult(context.Background(), account.ID, false, resp.StatusCode, err)
		return err
	}

	s.saveProbeResult(context.Background(), account.ID, true, resp.StatusCode, nil)
	return nil
}

func (s *OAuthProbeService) probeClaudeOAuth(parent context.Context, account *Account) error {
	if account == nil || account.Platform != PlatformAnthropic || !account.IsOAuth() {
		return nil
	}
	if s.httpUpstream == nil {
		return fmt.Errorf("http upstream not configured")
	}

	timeout := time.Duration(s.cfg.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	authToken := strings.TrimSpace(account.GetCredential("access_token"))
	if account.Type == AccountTypeOAuth && s.claudeTokenProvider != nil {
		if t, err := s.claudeTokenProvider.GetAccessToken(ctx, account); err == nil && strings.TrimSpace(t) != "" {
			authToken = t
		}
	}
	if authToken == "" {
		return fmt.Errorf("missing access token")
	}

	nonce, _ := randomHexString(8)
	payload, err := createTestPayload(claude.DefaultTestModel, nonce)
	if err != nil {
		return err
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testClaudeAPIURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", claude.DefaultBetaHeader)
	for key, value := range claude.DefaultHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, account.IsTLSFingerprintEnabled())
	if err != nil {
		s.saveProbeResult(context.Background(), account.ID, false, 0, err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Persist Claude quota snapshot (5h/7d windows) to DB extra.
	// This calls Anthropic's OAuth usage API, which should not consume message quota.
	if s.accountUsageService != nil {
		updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = s.accountUsageService.SyncUsageSnapshotToExtra(updateCtx, account, usageSnapshotSourceProbe)
		cancel()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		err := fmt.Errorf("claude oauth probe status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
		s.saveProbeResult(context.Background(), account.ID, false, resp.StatusCode, err)
		return err
	}

	s.saveProbeResult(context.Background(), account.ID, true, resp.StatusCode, nil)
	return nil
}

func (s *OAuthProbeService) saveProbeResult(ctx context.Context, accountID int64, ok bool, statusCode int, err error) {
	if s == nil || s.accountRepo == nil || accountID <= 0 {
		return
	}

	msg := ""
	if err != nil {
		msg = strings.TrimSpace(err.Error())
		if len(msg) > 256 {
			msg = msg[:256]
		}
	}

	updates := map[string]any{
		"oauth_probe": map[string]any{
			"updated_at":  time.Now().UTC().Format(time.RFC3339),
			"ok":          ok,
			"status_code": statusCode,
			"error":       msg,
		},
	}

	updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = s.accountRepo.UpdateExtra(updateCtx, accountID, updates)
}

func (s *OAuthProbeService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.securityCfg == nil {
		return urlvalidator.ValidateURLFormat(raw, true)
	}
	if !s.securityCfg.URLAllowlist.Enabled {
		return urlvalidator.ValidateURLFormat(raw, s.securityCfg.URLAllowlist.AllowInsecureHTTP)
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.securityCfg.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.securityCfg.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", err
	}
	return normalized, nil
}
