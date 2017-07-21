package cpio

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"./readme.txt", "This archive contains some text files."},
	{"./gophers.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
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
	//f, err := ioutil.TempFile("", "cpio-test-")
	f, err := os.OpenFile("./testdata/out_svr4.cpio", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}

	defer func() {
		f.Close()
		//os.Remove(f.Name())
	}()

	w := NewWriter(f)
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("error closing write: %v", err)
		}
	}()

	for _, file := range files {
		path := filepath.Join("./testdata", file.Name)
		infile, err := os.Open(path)
		if err != nil {
			t.Fatalf("cannot open test file: %v", err)
		}

		fi, err := infile.Stat()
		if err != nil {
			t.Fatalf("cannot stat test file: %v", err)
		}

		hdr := &Header{
			Name: file.Name,
			Mode: fi.Mode(),
			Size: fi.Size(),
		}

		if err := w.WriteHeader(hdr); err != nil {
			t.Fatalf("error writing header: %v", err)
		}

		if n, err := io.Copy(w, infile); err != nil {
			t.Fatalf("error writing file body: %v", err)
		} else if n != fi.Size() {
			t.Fatalf("bad write length: expect %v, got %v", fi.Size(), n)
		}
	}
}
