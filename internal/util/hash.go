package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashString berechnet den SHA256-Hash eines Strings und gibt ihn hex-kodiert zurück.
func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HashBytes berechnet den SHA256-Hash von Byte-Daten und gibt ihn hex-kodiert zurück.
func HashBytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
