package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// DeterministicReader is a dummy reader for deterministic key generation (FOR TESTING ONLY)
// In production, use crypto/rand.Reader
type DeterministicReader struct {
	seed int64
}

func (r *DeterministicReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		// Simple LCG
		r.seed = (r.seed*1103515245 + 12345) & 0x7fffffff
		p[i] = byte(r.seed)
	}
	return len(p), nil
}

func generateKey(id int) (*rsa.PrivateKey, error) {
	// Use ID as seed to generate same key for same ID every time
	reader := &DeterministicReader{seed: int64(id + 1000)} // offset to avoid trivial seeds
	// Small key size for performance in this prototype (usually 2048+)
	return rsa.GenerateKey(reader, 1024)
}

func sign(privKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

func verify(pubKey *rsa.PublicKey, data []byte, signature []byte) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
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
