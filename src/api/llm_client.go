package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// Configurable LLM endpoint & model (map decision, see .scratch/llm-refinements/issues/
// 01-configurable-llm-endpoint-and-model.md). Shared by all three LLM features (chatbot,
// who-to-sign, scout-report) so the SSRF-hardening logic below lives in exactly one
// place rather than being triplicated across handlers.

const llmBaseURLHeader = "X-OpenAI-Base-URL"
const llmModelHeader = "X-OpenAI-Model"

// llmRequestConfig is read once per HTTP request from the optional override headers.
// Empty fields mean "use the caller's own default" -- each header is independently
// optional, per the map decision.
type llmRequestConfig struct {
	baseURL string
	model   string
}

func readLLMRequestConfig(r *http.Request) llmRequestConfig {
	return llmRequestConfig{
		baseURL: r.Header.Get(llmBaseURLHeader),
		model:   r.Header.Get(llmModelHeader),
	}
}

// resolveModel returns the request's override model if set. Otherwise, when
// LOCAL_LLM_BASE_URL is active (see localLLMBaseURL), it returns the local model name --
// the caller's hardcoded default (e.g. "gpt-5.6-luna") would be meaningless to a local
// Ollama server, which only knows the model name it was created with. With neither set,
// it falls through to the caller's hardcoded default, i.e. real OpenAI.
func (c llmRequestConfig) resolveModel(defaultModel string) string {
	if c.model != "" {
		return c.model
	}
	if localLLMBaseURL() != "" {
		return localLLMModel()
	}
	return defaultModel
}

// newLLMClient builds the openai.Client used to talk to OpenAI or, when the caller has
// configured one, a custom OpenAI-compatible endpoint (Azure OpenAI, a local proxy,
// OpenRouter, etc.). cfg.baseURL comes from the per-request X-OpenAI-Base-URL header --
// i.e. user-suppliable input -- and is routed through the SSRF-hardened client below.
// When that header is absent, and only outside production, we fall back to
// LOCAL_LLM_BASE_URL (see localLLMBaseURL) for pointing dev builds at a local model
// server; that value is operator-set server config, not request input, so it uses a
// plain client instead of the hardened one. With neither set, this is identical to the
// SDK's own default client construction (talks to real OpenAI).
func newLLMClient(apiKey string, cfg llmRequestConfig) (*openai.Client, error) {
	if cfg.baseURL != "" {
		return newHardenedLLMClient(apiKey, cfg.baseURL)
	}
	if local := localLLMBaseURL(); local != "" {
		httpClient := &http.Client{Timeout: localLLMTimeout()}
		client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(local), option.WithHTTPClient(httpClient))
		return &client, nil
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &client, nil
}

// localLLMBaseURL returns the operator-configured local LLM endpoint for development,
// e.g. http://localhost:11434/v1 for a local Ollama server. Only honored outside
// production (ENVIRONMENT defaults to "development", see otel_config.go) so this can
// never activate on the deployed multi-tenant app even if the env var were somehow set
// there. Unlike the header-based baseURL, this is not attacker-reachable input -- it's
// a value the person running the server put in their own environment -- so it
// deliberately skips the https-only/private-IP hardening below, which exists to stop a
// malicious *request* from redirecting the backend at an internal service.
func localLLMBaseURL() string {
	if os.Getenv("ENVIRONMENT") == "production" {
		return ""
	}
	return os.Getenv("LOCAL_LLM_BASE_URL")
}

// localLLMModel returns the model name to request from the local dev LLM endpoint,
// read from LOCAL_LLM_MODEL (default "ornith" -- the name launch_dev.ps1/.sh create the
// local model under). Only consulted by resolveModel when localLLMBaseURL is active.
func localLLMModel() string {
	if model := os.Getenv("LOCAL_LLM_MODEL"); model != "" {
		return model
	}
	return "ornith"
}

// localLLMTimeout returns the HTTP client timeout for the local dev LLM path. Without
// this, a local Ollama model that hangs mid-generation (observed with ling-tiny -- see
// local_llm_failures.md #2) leaves the request -- and the UI's spinner -- stuck forever,
// since (unlike newHardenedLLMClient) nothing here previously bounded it.
//
// This must stay well under 120s: main.go's RequestTimeoutMiddlewareFunc already puts a
// 120s context deadline on /api/chatbot/, /api/who-to-sign/, and /api/scout-report/, and
// that middleware has its own bug -- when its deadline fires it abandons the still-running
// handler goroutine instead of waiting for it, and (for these SSE routes, which already
// sent a 200 and started streaming before the LLM call even begins) its http.Error() call
// lands as unparseable plain text mid-stream rather than a proper "event: error" frame, so
// the frontend can't render it and the chat UI looks hung forever even though the
// connection is actually closed. Confirmed live: a chatbot call against ling-tiny cut off
// at exactly 120.001s with the literal body "Request timeout after 2m0s" and HTTP 200.
// Racing that broken path is worse than avoiding it, so this timeout -- and
// callResponsesWithRetry skipping its retry on a timeout, see below -- must let our own
// call fail cleanly (a real "event: error" SSE frame) comfortably before 120s, every time,
// including the tool-calling loops' several rounds. 45s balances that against the local
// model actually getting to finish a normal call; override via LOCAL_LLM_TIMEOUT_SECONDS
// for a deliberately slower setup, but keep it under 120s or you're back to the same bug.
func localLLMTimeout() time.Duration {
	if raw := os.Getenv("LOCAL_LLM_TIMEOUT_SECONDS"); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 45 * time.Second
}

// isLocalLLMMode reports whether requests are being routed to a local dev LLM server
// (see localLLMBaseURL). The three tool-calling loops (chatbot.go, who_to_sign.go,
// scout_report.go) use it to decide how to carry conversation state across rounds —
// see appendResponseOutputToHistory below for why.
func isLocalLLMMode() bool {
	return localLLMBaseURL() != ""
}

// appendResponseOutputToHistory folds one Responses API turn's output items into a
// conversation history built as explicit input items, for local-LLM tool-calling loops
// that resend the whole conversation each round instead of relying on
// PreviousResponseID (see local_llm_failures.md #3: a live probe against Ollama's
// /v1/responses showed it never actually persists a response server-side — even with
// store:true explicitly set on the request, a follow-up call with previous_response_id
// comes back having lost every prior turn, including the original system prompt, and
// the model then either returns nothing or hallucinates generic content ungrounded in
// the real conversation. Real OpenAI doesn't have this problem, so this path is only
// exercised in local-dev mode; see isLocalLLMMode).
//
// Only assistant messages and function calls are carried forward — reasoning items are
// dropped. The same live probe confirmed the model stays grounded without them, and
// carrying every round's full chain-of-thought forward would balloon prompt size on
// every subsequent round for no verified benefit.
func appendResponseOutputToHistory(history []responses.ResponseInputItemUnionParam, output []responses.ResponseOutputItemUnion) []responses.ResponseInputItemUnionParam {
	for _, item := range output {
		switch item.Type {
		case "function_call":
			param := item.AsFunctionCall().ToParam()
			history = append(history, responses.ResponseInputItemUnionParam{OfFunctionCall: &param})
		case "message":
			param := item.AsMessage().ToParam()
			history = append(history, responses.ResponseInputItemUnionParam{OfOutputMessage: &param})
		}
	}
	return history
}

// newHardenedLLMClient builds a client for a user-suppliable custom endpoint. The
// underlying http.Client uses a hardened Transport -- https-only, and every dial
// (including on redirect) is validated against a private/loopback/link-local/reserved
// IP blocklist at actual connection time, not just as a pre-request check, which
// defeats DNS rebinding between check and connect. This app is deployed as
// public/multi-tenant, so a user-suppliable server-side base URL is a real SSRF vector:
// without this, a malicious user could point the backend at an internal service or a
// cloud metadata endpoint (e.g. 169.254.169.254) and read back the response through the
// OpenAI-shaped API surface. There is deliberately no hostname allowlist -- any public
// host clears validation, so legitimate custom endpoints keep working without this app
// maintaining a curated list.
func newHardenedLLMClient(apiKey, baseURL string) (*openai.Client, error) {
	httpClient, err := hardenedHTTPClientForBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	opts := []option.RequestOption{option.WithAPIKey(apiKey), option.WithBaseURL(baseURL), option.WithHTTPClient(httpClient)}
	client := openai.NewClient(opts...)
	return &client, nil
}

func hardenedHTTPClientForBaseURL(rawURL string) (*http.Client, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, fmt.Errorf("custom LLM endpoint must be a valid https:// URL")
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return safeDialContext(ctx, dialer, network, addr)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to a non-https URL was blocked")
			}
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			// No IP check needed here -- DialContext above re-validates the redirect
			// target's resolved IP on the connection it makes for this next request.
			return nil
		},
	}, nil
}

// safeDialContext resolves addr's host, rejects it if any resolved IP is
// private/loopback/link-local/reserved, then dials the resolved IP directly (not the
// hostname) so the connection target is exactly what was just validated -- this closes
// the gap a second, independent DNS resolution inside a plain dial would leave open to
// a DNS-rebinding attacker who returns a safe IP for the check and an unsafe one for
// the real connection.
func safeDialContext(ctx context.Context, dialer *net.Dialer, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}
	for _, ip := range ips {
		if isBlockedLLMEndpointIP(ip.IP) {
			return nil, fmt.Errorf("connection to %s is blocked (private/reserved address)", ip.IP)
		}
	}

	var lastErr error
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, lastErr
}

// isBlockedLLMEndpointIP rejects the private/loopback/link-local/reserved ranges an
// SSRF payload would target -- including the AWS/GCP/Azure metadata address
// 169.254.169.254, already covered by IsLinkLocalUnicast (169.254.0.0/16) but checked
// explicitly for clarity and defense-in-depth.
func isBlockedLLMEndpointIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 169 && ip4[1] == 254 {
		return true
	}
	return false
}
