package sonzai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// testSign is a helper that produces a Sonzai-Signature header value.
// It mirrors the server-side Sign function from the platform API.
func testSign(secret string, ts int64, body []byte) string {
	signedPayload := fmt.Sprintf("%d.%s", ts, body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

func TestVerifyWebhookSignature_Roundtrip(t *testing.T) {
	secret := "whsec_dGVzdHNlY3JldGtleTEyMzQ1Njc4OTAxMjM0NTY="
	payload := []byte(`{"event_type":"on_diary_generated","agent_id":"abc"}`)
	ts := time.Now().Unix()

	// Sign with the raw key (strip prefix, matching server behavior).
	rawKey := secret[len("whsec_"):]
	header := testSign(rawKey, ts, payload)

	if err := VerifyWebhookSignature(payload, header, secret); err != nil {
		t.Fatalf("expected valid signature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_DeterministicOutput(t *testing.T) {
	secret := "whsec_fixedkey123"
	rawKey := "fixedkey123"
	payload := []byte(`{"test":true}`)
	var ts int64 = 1700000000

	header := testSign(rawKey, ts, payload)

	// Verify the header is deterministic.
	header2 := testSign(rawKey, ts, payload)
	if header != header2 {
		t.Fatalf("signatures not deterministic: %q vs %q", header, header2)
	}

	if err := VerifyWebhookSignatureWithTolerance(payload, header, secret, 0); err != nil {
		t.Fatalf("expected valid signature with tolerance=0, got: %v", err)
	}
}

func TestVerifyWebhookSignature_TamperedPayload(t *testing.T) {
	secret := "whsec_testsecret"
	rawKey := "testsecret"
	ts := time.Now().Unix()

	original := []byte(`{"event_type":"on_wakeup_ready"}`)
	header := testSign(rawKey, ts, original)

	tampered := []byte(`{"event_type":"on_wakeup_ready","injected":true}`)
	err := VerifyWebhookSignature(tampered, header, secret)
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_StaleTimestamp(t *testing.T) {
	secret := "whsec_testsecret"
	rawKey := "testsecret"
	payload := []byte(`{"test":true}`)

	// Timestamp from 10 minutes ago.
	staleTS := time.Now().Add(-10 * time.Minute).Unix()
	header := testSign(rawKey, staleTS, payload)

	err := VerifyWebhookSignature(payload, header, secret)
	if err != ErrTimestampExpired {
		t.Fatalf("expected ErrTimestampExpired, got: %v", err)
	}
}

func TestVerifyWebhookSignature_FutureTimestamp(t *testing.T) {
	secret := "whsec_testsecret"
	rawKey := "testsecret"
	payload := []byte(`{"test":true}`)

	// Timestamp 10 minutes in the future.
	futureTS := time.Now().Add(10 * time.Minute).Unix()
	header := testSign(rawKey, futureTS, payload)

	err := VerifyWebhookSignature(payload, header, secret)
	if err != ErrTimestampExpired {
		t.Fatalf("expected ErrTimestampExpired for future timestamp, got: %v", err)
	}
}

func TestVerifyWebhookSignature_SkipTimestampCheck(t *testing.T) {
	secret := "whsec_testsecret"
	rawKey := "testsecret"
	payload := []byte(`{"test":true}`)

	// Very old timestamp - should pass with tolerance=0.
	var oldTS int64 = 1000000000
	header := testSign(rawKey, oldTS, payload)

	if err := VerifyWebhookSignatureWithTolerance(payload, header, secret, 0); err != nil {
		t.Fatalf("expected valid with tolerance=0, got: %v", err)
	}
}

func TestVerifyWebhookSignature_MultipleSignatures(t *testing.T) {
	secret := "whsec_newsecret"
	rawKey := "newsecret"
	payload := []byte(`{"test":true}`)
	ts := time.Now().Unix()

	// Simulate a header with an old signature (wrong key) and the correct one.
	oldMac := hmac.New(sha256.New, []byte("oldsecret"))
	oldMac.Write([]byte(fmt.Sprintf("%d.%s", ts, payload)))
	oldSig := hex.EncodeToString(oldMac.Sum(nil))

	newMac := hmac.New(sha256.New, []byte(rawKey))
	newMac.Write([]byte(fmt.Sprintf("%d.%s", ts, payload)))
	newSig := hex.EncodeToString(newMac.Sum(nil))

	header := fmt.Sprintf("t=%d,v1=%s,v1=%s", ts, oldSig, newSig)

	if err := VerifyWebhookSignature(payload, header, secret); err != nil {
		t.Fatalf("expected valid with multiple signatures, got: %v", err)
	}
}

func TestVerifyWebhookSignature_EmptySecret(t *testing.T) {
	payload := []byte(`{"test":true}`)
	header := "t=1700000000,v1=abc123"

	err := VerifyWebhookSignature(payload, header, "")
	if err != ErrInvalidSecret {
		t.Fatalf("expected ErrInvalidSecret, got: %v", err)
	}
}

func TestVerifyWebhookSignature_PrefixOnlySecret(t *testing.T) {
	payload := []byte(`{"test":true}`)
	header := "t=1700000000,v1=abc123"

	err := VerifyWebhookSignature(payload, header, "whsec_")
	if err != ErrInvalidSecret {
		t.Fatalf("expected ErrInvalidSecret for prefix-only secret, got: %v", err)
	}
}

func TestVerifyWebhookSignature_EmptyHeader(t *testing.T) {
	payload := []byte(`{"test":true}`)

	err := VerifyWebhookSignature(payload, "", "whsec_test")
	if err != ErrMissingSignature {
		t.Fatalf("expected ErrMissingSignature, got: %v", err)
	}
}

func TestVerifyWebhookSignature_MalformedHeader(t *testing.T) {
	payload := []byte(`{"test":true}`)

	// No timestamp.
	err := VerifyWebhookSignature(payload, "v1=abc123", "whsec_test")
	if err == nil {
		t.Fatal("expected error for header without timestamp")
	}
}

func TestVerifyWebhookSignature_SecretWithoutPrefix(t *testing.T) {
	// Secrets without the whsec_ prefix should also work.
	rawKey := "rawsecretwithoutprefix"
	payload := []byte(`{"test":true}`)
	ts := time.Now().Unix()

	header := testSign(rawKey, ts, payload)

	if err := VerifyWebhookSignature(payload, header, rawKey); err != nil {
		t.Fatalf("expected valid with raw secret, got: %v", err)
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"test":true}`)
	ts := time.Now().Unix()

	header := testSign("correctsecret", ts, payload)

	err := VerifyWebhookSignature(payload, header, "whsec_wrongsecret")
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}
