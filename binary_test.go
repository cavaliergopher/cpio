package cpio

import (
	"io"
	"os"
	"testing"
)

func TestReadBinary(t *testing.T) {
	f, err := os.Open("testdata/test_binary.cpio")
	if err != nil {
		t.Fatalf("error opening test file: %v", err)
	}
	defer f.Close()

	r := NewReader(f)
	for {
		_, err := r.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Errorf("error moving to the next header: %v", err)
			return
		}
		// TODO: validate header fields
	}
}
