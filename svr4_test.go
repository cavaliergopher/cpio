package cpio

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"./readme.txt", "This archive contains some text files."},
	{"./gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"./todo.txt", "Get animal handling license."},
}

func TestRead(t *testing.T) {
	f, err := os.Open("testdata/test_svr4.cpio")
	if err != nil {
		t.Fatalf("error opening test file: %v", err)
	}
	defer f.Close()

	r := NewReader(f)
	for i := 1; i > 0; i++ {
		_, err := r.Next()
		if err != nil {
			if err != io.EOF {
				t.Errorf("error moving to next header: %v", err)
			}
			i = -1
			return
		}

		// TODO: validate header fields
	}
}

func TestWrite(t *testing.T) {
	f, err := ioutil.TempFile("", "cpio-test-")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	w := NewWriter(f)
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("error closing write: %v", err)
		}
	}()

	for _, file := range files {
		hdr := &Header{
			name: file.Name,
			mode: 0644,
			size: int64(len(file.Body)),
		}

		if err := w.WriteHeader(hdr); err != nil {
			t.Fatalf("error writing header: %v", err)
		}

		if n, err := w.Write([]byte(file.Body)); err != nil {
			t.Fatalf("error writing file body: %v", err)
		} else if n != len(file.Body) {
			t.Fatalf("bad write length: expect %v, got %v", len(file.Body), n)
		}
	}
}
