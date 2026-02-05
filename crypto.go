package main

import (
	"crypto/ed25519"
	"fmt"
	mrand "math/rand"
)

// DeterministicReader is a dummy reader for deterministic key generation (FOR TESTING ONLY)
// In production, use crypto/rand.Reader
type DeterministicReader struct {
	src mrand.Source
}

func (r *DeterministicReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = byte(r.src.Int63())
	}
	return len(p), nil
}

func generateKey(id int) (ed25519.PrivateKey, error) {
	// Use ID as seed to generate same key for same ID every time
	src := mrand.NewSource(int64(id + 1000))
	reader := &DeterministicReader{src: src}
	_, priv, err := ed25519.GenerateKey(reader)
	return priv, err
}

func sign(privKey ed25519.PrivateKey, data []byte) ([]byte, error) {
	// Ed25519 signs the message itself, usually.
	// But to match previous logic (hashing first), we can hash first.
	// However, Ed25519 is fast enough to sign full message, or we can sign hash.
	// Standard Ed25519 signs the message.
	// Let's stick to previous behavior: sign the hash?
	// RSA PKCS1v15 usually signs the hash.
	// Ed25519 signs the message.
	// But if 'data' is already large? 'digestPrePrepare' returns a small string.
	// So we can sign 'data' directly.
	return ed25519.Sign(privKey, data), nil
}

func verify(pubKey ed25519.PublicKey, data []byte, signature []byte) error {
	if ed25519.Verify(pubKey, data, signature) {
		return nil
	}
	return fmt.Errorf("invalid signature")
}

// Helper to construct data for signing
func digestPrePrepare(view int, seq int, digest string) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s", view, seq, digest))
}

func digestPrepare(view int, seq int, digest string, nodeID int) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s:%d", view, seq, digest, nodeID))
}

func digestCommit(view int, seq int, digest string, nodeID int) []byte {
	return []byte(fmt.Sprintf("%d:%d:%s:%d", view, seq, digest, nodeID))
}
