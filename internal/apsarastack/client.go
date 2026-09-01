package apsarastack

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var readonlyPrefixes = []string{"Describe", "Get", "List", "Search", "Query"}

// Client is a minimal POP RPC client for Apsara Stack read-only APIs.
type Client struct {
	AccessKeyID  string
	AccessKeySecret string
	Region       string
	RateLimiter  *RateLimiter
	httpClient   *http.Client
}

func NewClient(accessKeyID, accessKeySecret, region string) *Client {
	if region == "" {
		region = DefaultRegion
	}
	return &Client{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Region:          region,
		RateLimiter:     NewRateLimiter(20, 1.0),
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

func LoadCredentialsFromEnv() (accessKeyID, accessKeySecret, region string, ok bool) {
	accessKeyID = firstNonEmpty(
		os.Getenv("APSARASTACK_ACCESS_KEY_ID"),
		os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
	)
	accessKeySecret = firstNonEmpty(
		os.Getenv("APSARASTACK_ACCESS_KEY_SECRET"),
		os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
	)
	region = firstNonEmpty(
		os.Getenv("APSARASTACK_REGION"),
		os.Getenv("ALIBABA_CLOUD_REGION"),
		DefaultRegion,
	)
	return accessKeyID, accessKeySecret, region, accessKeyID != "" && accessKeySecret != ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ServiceEndpoint returns the POP endpoint for a cloud product.
// Apsara Stack endpoints are deployment-specific; obtain from Uni-manager Service Registration Dashboard.
func ServiceEndpoint(service string) string {
	key := "APSARASTACK_" + strings.ToUpper(service) + "_ENDPOINT"
	if v := os.Getenv(key); v != "" {
		return strings.TrimPrefix(v, "https://")
	}
	if v := os.Getenv("APSARASTACK_POP_ENDPOINT"); v != "" {
		return strings.TrimPrefix(v, "https://")
	}
	// Placeholder defaults — override via env in Apsara Stack environments.
	switch service {
	case "ecs":
		return "ecs." + envOrDefault("APSARASTACK_REGION", DefaultRegion) + ".example.stack.local"
	case "slb":
		return "slb." + envOrDefault("APSARASTACK_REGION", DefaultRegion) + ".example.stack.local"
	case "rds":
		return "rds." + envOrDefault("APSARASTACK_REGION", DefaultRegion) + ".example.stack.local"
	case "cms":
		return "metrics." + envOrDefault("APSARASTACK_REGION", DefaultRegion) + ".example.stack.local"
	default:
		return service + ".example.stack.local"
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (c *Client) isReadonlyAction(action string) bool {
	for _, p := range readonlyPrefixes {
		if strings.HasPrefix(action, p) {
			return true
		}
	}
	return false
}

func percentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

func randomNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func sign(secret, method, canonicalizedQuery string) string {
	stringToSign := method + "&" + percentEncode("/") + "&" + percentEncode(canonicalizedQuery)
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// CallRPC invokes a read-only POP RPC API (GET + HMAC-SHA1 signature).
func (c *Client) CallRPC(endpoint, action, version string, params map[string]string) (map[string]interface{}, error) {
	if !c.isReadonlyAction(action) {
		return nil, fmt.Errorf("action '%s' is not in the read-only allowlist", action)
	}

	c.RateLimiter.WaitIfNeeded()

	all := map[string]string{
		"Format":           "JSON",
		"Version":          version,
		"AccessKeyId":      c.AccessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   randomNonce(),
		"Action":           action,
		"RegionId":         c.Region,
	}
	for k, v := range params {
		all[k] = v
	}

	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"="+percentEncode(all[k]))
	}
	canonicalized := strings.Join(pairs, "&")
	all["Signature"] = sign(c.AccessKeySecret, "GET", canonicalized)

	query := url.Values{}
	for k, v := range all {
		query.Set(k, v)
	}

	reqURL := "https://" + endpoint + "/?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	if code, ok := envelope["Code"].(string); ok && code != "" {
		msg, _ := envelope["Message"].(string)
		return nil, fmt.Errorf("Apsara Stack API error: %s - %s", code, msg)
	}
	return envelope, nil
}
