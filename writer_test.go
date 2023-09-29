package cpio_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/cavaliergopher/cpio"
)

func store(w *cpio.Writer, fn string) error {
	fi, err := os.Lstat(fn)
	if err != nil {
		return err
	}
	var link string
	if fi.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(fn); err != nil {
			return err
		}
	}
	hdr, err := cpio.FileInfoHeader(fi, link)
	if err != nil {
		return err
	}
	if hdr.Mode&^cpio.ModePerm == cpio.ModeSymlink {
		hdr.Size = 0 // FIXME: should be done in FileInfoHeader
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	if fi.Mode().IsRegular() {
		f, err := os.Open(fn)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
	}
	return err
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := cpio.NewWriter(&buf)
	if err := store(w, "testdata/etc"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := store(w, "testdata/etc/hosts"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWriter_Symlink(t *testing.T) {
	var buf bytes.Buffer
	w := cpio.NewWriter(&buf)
	func() {
		defer func() {
			if err := w.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
		}()
		if err := store(w, "testdata/checklist.txt"); err != nil {
			t.Fatalf("store: %v", err)
		}
	}()
	r := cpio.NewReader(&buf)
	hdr, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if hdr.Mode&^cpio.ModePerm != cpio.ModeSymlink {
		t.Fatalf("file has mode %s, expected %s (ModeSymlink)", hdr.Mode, cpio.FileMode(cpio.ModeSymlink))
	}
	if hdr.Linkname == "" {
		t.Fatal("Empty Linkname on ModeSymlink file")
	}
}
