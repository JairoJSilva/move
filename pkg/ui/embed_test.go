package ui

import (
	"io"
	"testing"
)

func TestEmbeddedUI(t *testing.T) {
	if !HasEmbeddedUI() {
		t.Fatal("HasEmbeddedUI returned false, expected true")
	}

	sub, err := GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}

	f, err := sub.Open("index.html")
	if err != nil {
		t.Fatalf("Failed to open index.html: %v", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("index.html is empty")
	}

	t.Logf("Successfully read index.html (%d bytes)", len(content))
}
