package tests

import (
	"godocgen/internal/util"
	"testing"
)

func TestHashConsistency(t *testing.T) {
	content1 := "graph TD; A-->B;"
	content2 := "graph TD; A-->B;"
	content3 := "graph TD; A-->C;"

	h1 := util.HashString(content1)
	h2 := util.HashString(content2)
	h3 := util.HashString(content3)

	if h1 != h2 {
		t.Error("Hashes for identical content should be the same")
	}
	if h1 == h3 {
		t.Error("Hashes for different content should be different")
	}
}

