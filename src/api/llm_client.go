package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
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

// resolveModel returns the request's override model if set, otherwise the caller's
// hardcoded default (e.g. chatbotModel, whoToSignModel, scoutReportModel).
func (c llmRequestConfig) resolveModel(defaultModel string) string {
	if c.model != "" {
		return c.model
	}
	return defaultModel
}

// newLLMClient builds the openai.Client used to talk to OpenAI or, when the caller has
// configured one, a custom OpenAI-compatible endpoint (Azure OpenAI, a local proxy,
// OpenRouter, etc.). When baseURL is empty this is identical to the SDK's own default
// client construction. When baseURL is set, the underlying http.Client uses a hardened
// Transport -- https-only, and every dial (including on redirect) is validated against
// a private/loopback/link-local/reserved IP blocklist at actual connection time, not
// just as a pre-request check, which defeats DNS rebinding between check and connect.
// This app is deployed as public/multi-tenant, so a user-suppliable server-side base
// URL is a real SSRF vector: without this, a malicious user could point the backend at
// an internal service or a cloud metadata endpoint (e.g. 169.254.169.254) and read back
// the response through the OpenAI-shaped API surface. There is deliberately no
// hostname allowlist -- any public host clears validation, so legitimate custom
// endpoints keep working without this app maintaining a curated list.
func newLLMClient(apiKey, baseURL string) (*openai.Client, error) {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL == "" {
		client := openai.NewClient(opts...)
		return &client, nil
	}

	httpClient, err := hardenedHTTPClientForBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	opts = append(opts, option.WithBaseURL(baseURL), option.WithHTTPClient(httpClient))
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
