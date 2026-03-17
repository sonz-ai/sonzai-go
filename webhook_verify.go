package sonzai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors for webhook signature verification.
var (
	ErrInvalidSignature = errors.New("sonzai: webhook signature is invalid")
	ErrTimestampExpired = errors.New("sonzai: webhook timestamp is outside tolerance")
	ErrMissingSignature = errors.New("sonzai: missing or empty signature header")
	ErrInvalidSecret    = errors.New("sonzai: invalid or empty webhook secret")
)

const (
	// defaultTimestampTolerance is the maximum allowed age of a webhook signature.
	defaultTimestampTolerance = 5 * time.Minute

	// webhookSecretPrefix is the expected prefix for signing secrets.
	webhookSecretPrefix = "whsec_"
)

// VerifyWebhookSignature verifies a Sonzai webhook payload signature using the
// default timestamp tolerance of 5 minutes.
//
// Usage in a webhook handler:
//
//	body, _ := io.ReadAll(r.Body)
//	sig := r.Header.Get("Sonzai-Signature")
//	if err := sonzai.VerifyWebhookSignature(body, sig, webhookSecret); err != nil {
//	    http.Error(w, "Invalid signature", 401)
//	    return
//	}
func VerifyWebhookSignature(payload []byte, sigHeader, secret string) error {
	return VerifyWebhookSignatureWithTolerance(payload, sigHeader, secret, defaultTimestampTolerance)
}

// VerifyWebhookSignatureWithTolerance verifies a Sonzai webhook payload signature
// with a custom timestamp tolerance. Set tolerance to 0 to skip timestamp checking.
//
// The signature header format is: t={unix_timestamp},v1={hex_hmac_sha256}
// Multiple signatures may be present (comma-separated v1= values) to support
// secret rotation.
func VerifyWebhookSignatureWithTolerance(payload []byte, sigHeader, secret string, tolerance time.Duration) error {
	if secret == "" {
		return ErrInvalidSecret
	}
	if sigHeader == "" {
		return ErrMissingSignature
	}

	// Extract the raw secret key (strip whsec_ prefix if present).
	signingKey := secret
	if strings.HasPrefix(secret, webhookSecretPrefix) {
		signingKey = secret[len(webhookSecretPrefix):]
	}
	if signingKey == "" {
		return ErrInvalidSecret
	}

	// Parse the signature header: t={ts},v1={sig1},v1={sig2},...
	ts, sigs, err := parseSignatureHeader(sigHeader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMissingSignature, err)
	}

	if len(sigs) == 0 {
		return ErrMissingSignature
	}

	// Check timestamp tolerance.
	if tolerance > 0 {
		signedAt := time.Unix(ts, 0)
		now := time.Now()
		if now.Sub(signedAt) > tolerance || signedAt.Sub(now) > tolerance {
			return ErrTimestampExpired
		}
	}

	// Compute expected signature: HMAC-SHA256("{timestamp}.{payload}")
	signedPayload := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Check if any of the provided signatures match (supports rotation).
	for _, sig := range sigs {
		if hmac.Equal([]byte(sig), []byte(expectedSig)) {
			return nil
		}
	}

	return ErrInvalidSignature
}

// parseSignatureHeader parses a header of the form "t={ts},v1={sig1},v1={sig2}".
func parseSignatureHeader(header string) (timestamp int64, signatures []string, err error) {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "t=") {
			tsStr := part[2:]
			timestamp, err = strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("invalid timestamp %q: %w", tsStr, err)
			}
		} else if strings.HasPrefix(part, "v1=") {
			sig := part[3:]
			if sig != "" {
				signatures = append(signatures, sig)
			}
		}
	}
	if timestamp == 0 {
		return 0, nil, fmt.Errorf("missing timestamp in header")
	}
	return timestamp, signatures, nil
}
