package tencentcloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var readonlyPrefixes = []string{"Describe", "Get", "List", "Search", "Query"}

// Client is a minimal TC3 client for read-only APIs.
type Client struct {
	SecretID   string
	SecretKey  string
	Region     string
	Service    string
	Host       string
	RateLimiter *RateLimiter
	httpClient *http.Client
}

func NewClient(secretID, secretKey, region string) *Client {
	return &Client{
		SecretID:    secretID,
		SecretKey:   secretKey,
		Region:      region,
		Service:     "cvm",
		Host:        "cvm.tencentcloudapi.com",
		RateLimiter: NewRateLimiter(20, 1.0),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) isReadonlyAction(action string) bool {
	for _, p := range readonlyPrefixes {
		if strings.HasPrefix(action, p) {
			return true
		}
	}
	return false
}

func (c *Client) sign(payload string, timestamp int64, date string) string {
	algorithm := "TC3-HMAC-SHA256"
	canonicalHeaders := "content-type:application/json; charset=utf-8\nhost:" + c.Host + "\n"
	signedHeaders := "content-type;host"
	hashedPayload := sha256Hex(payload)
	canonicalRequest := fmt.Sprintf(
		"POST\n/\n\n%s\n%s\n%s",
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	)
	credentialScope := date + "/" + c.Service + "/tc3_request"
	hashedCanonical := sha256Hex(canonicalRequest)
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s", algorithm, timestamp, credentialScope, hashedCanonical)

	secretDate := hmacSHA256([]byte("TC3"+c.SecretKey), date)
	secretService := hmacSHA256(secretDate, c.Service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	return fmt.Sprintf(
		"%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, c.SecretID, credentialScope, signedHeaders, signature,
	)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key []byte, msg string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))
	return mac.Sum(nil)
}

// Call invokes a read-only Tencent Cloud API action.
func (c *Client) Call(action string, params map[string]interface{}, version string) (map[string]interface{}, error) {
	if !c.isReadonlyAction(action) {
		return nil, fmt.Errorf("action '%s' is not in the read-only allowlist", action)
	}

	c.RateLimiter.WaitIfNeeded()

	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	payload := string(payloadBytes)

	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	authorization := c.sign(payload, timestamp, date)

	req, err := http.NewRequest(http.MethodPost, "https://"+c.Host, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", c.Host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Region", c.Region)

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

	response, ok := envelope["Response"].(map[string]interface{})
	if !ok {
		return envelope, nil
	}
	if errObj, ok := response["Error"].(map[string]interface{}); ok {
		code, _ := errObj["Code"].(string)
		msg, _ := errObj["Message"].(string)
		return nil, fmt.Errorf("Tencent Cloud API error: %s - %s", code, msg)
	}
	return response, nil
}
