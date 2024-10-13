package message

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"
)

func TestCalDigest(t *testing.T) {
	req := &Request{
		ReqType: ReqVerifyBlock,
		Content: []byte("Test block content"),
		ReqTime: time.Now(),
	}

	expectedDigest := func() [32]byte {
		b, _ := json.Marshal(req)
		return sha256.Sum256(b)
	}()

	req.CalDigest()

	if req.Digest != expectedDigest {
		t.Errorf("Expected Digest %x, got %x", expectedDigest, req.Digest)
	}
}

func TestDoubleCallCalDigest(t *testing.T) {
	req := &Request{
		ReqType: ReqVerifyBlock,
		Content: []byte("Test block content"),
		ReqTime: time.Now(),
	}

	req.CalDigest()
	firstDigest := req.Digest

	req.CalDigest()

	if firstDigest != req.Digest {
		t.Errorf("Expected Digest %x, got %x", firstDigest, req.Digest)
	}
}
