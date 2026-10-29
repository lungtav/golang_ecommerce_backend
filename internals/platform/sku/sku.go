package sku

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func Generate() (string, error) {
	b := make([]byte, 6)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate SKU: %w", err)
	}

	return "SKU-" + strings.ToUpper(hex.EncodeToString(b)), nil
}
