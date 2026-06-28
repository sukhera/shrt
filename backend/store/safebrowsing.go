package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// safeBrowsingEndpoint is the Google Safe Browsing v4 threatMatches:find URL.
// The API key is appended as a query parameter.
const safeBrowsingEndpoint = "https://safebrowsing.googleapis.com/v4/threatMatches:find"

// sbRequest is the threatMatches:find request body.
type sbRequest struct {
	Client     sbClient     `json:"client"`
	ThreatInfo sbThreatInfo `json:"threatInfo"`
}

type sbClient struct {
	ClientID      string `json:"clientId"`
	ClientVersion string `json:"clientVersion"`
}

type sbThreatInfo struct {
	ThreatTypes      []string  `json:"threatTypes"`
	PlatformTypes    []string  `json:"platformTypes"`
	ThreatEntryTypes []string  `json:"threatEntryTypes"`
	ThreatEntries    []sbEntry `json:"threatEntries"`
}

type sbEntry struct {
	URL string `json:"url"`
}

// sbResponse holds the matches returned by the API. An empty matches list means
// the URL is clean.
type sbResponse struct {
	Matches []json.RawMessage `json:"matches"`
}

// checkSafeBrowsing returns ErrUnsafeURL if rawURL is flagged by Google Safe
// Browsing. When no API key is configured the check is skipped (with a warning)
// so local development works without a key. Transport or API errors are logged
// and treated as "not flagged" — Safe Browsing must never block link creation on
// its own outage.
func (s *Store) checkSafeBrowsing(ctx context.Context, rawURL string) error {
	if s.cfg.SafeBrowsingAPIKey == "" {
		slog.Warn("safe browsing skipped: no API key configured")
		return nil
	}

	body := sbRequest{
		Client: sbClient{ClientID: "shrt", ClientVersion: "1.0"},
		ThreatInfo: sbThreatInfo{
			ThreatTypes:      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "POTENTIALLY_HARMFUL_APPLICATION"},
			PlatformTypes:    []string{"ANY_PLATFORM"},
			ThreatEntryTypes: []string{"URL"},
			ThreatEntries:    []sbEntry{{URL: rawURL}},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		slog.Warn("safe browsing: marshal request failed", "err", err)
		return nil
	}

	url := safeBrowsingEndpoint + "?key=" + s.cfg.SafeBrowsingAPIKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		slog.Warn("safe browsing: build request failed", "err", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		slog.Warn("safe browsing: request failed, allowing url", "err", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("safe browsing: non-200 response, allowing url", "status", resp.StatusCode)
		return nil
	}

	var result sbResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Warn("safe browsing: decode response failed, allowing url", "err", err)
		return nil
	}

	if len(result.Matches) > 0 {
		return fmt.Errorf("%w: %s", ErrUnsafeURL, rawURL)
	}
	return nil
}
