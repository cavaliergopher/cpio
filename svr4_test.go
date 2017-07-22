package cpio

import (
	"io"
	"os"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"./gophers.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"./readme.txt", "This archive contains some text files."},
	{"./todo.txt", "Get animal handling license."},
}

func TestRead(t *testing.T) {
	f, err := os.Open("testdata/test_svr4_crc.cpio")
	if err != nil {
		t.Fatalf("error opening test file: %v", err)
	}
	defer f.Close()

	r := NewReader(f)
	for {
		hdr, err := r.Next()
		if err != nil {
			if err != io.EOF {
				t.Errorf("error moving to next header: %v", err)
			}
			return
		}

		// TODO: validate header fields
		t.Logf("%v", hdr)
	}
}

func TestSVR4CRC(t *testing.T) {
	f, err := os.Open("testdata/test_svr4_crc.cpio")
	if err != nil {
		t.Fatalf("error opening test file: %v", err)
	}
	defer f.Close()

	w := NewHash()
	r := NewReader(f)
	for {
		hdr, err := r.Next()
		if err != nil {
			if err != io.EOF {
				t.Errorf("error moving to next header: %v", err)
			}
			return
		}

		w.Reset()
		_, err = io.CopyN(w, r, hdr.Size)
		if err != nil {
			t.Fatalf("error writing to checksum hash: %v", err)
		}

		sum := Checksum(w.Sum32())
		if sum != hdr.Checksum {
			t.Errorf("expected checksum %v, got %v for %v", hdr.Checksum, sum, hdr.Name)
		}
	}
}
