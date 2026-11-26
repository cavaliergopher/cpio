package cpio_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/cavaliergopher/cpio"
)

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestReaderFileEOF(t *testing.T) {
	readErr := errors.New("read failed")
	tests := []struct {
		name string
		size int64
		body io.Reader
		want string
		err  error
	}{
		{"missing", 6, strings.NewReader(""), "", io.ErrUnexpectedEOF},
		{"partial", 6, strings.NewReader("ab"), "ab", io.ErrUnexpectedEOF},
		{"partial with EOF", 6, iotest.DataErrReader(strings.NewReader("ab")), "ab", io.ErrUnexpectedEOF},
		{"complete", 6, strings.NewReader("abcdef"), "abcdef", nil},
		{"complete with EOF", 6, iotest.DataErrReader(strings.NewReader("abcdef")), "abcdef", nil},
		{"short reads", 6, iotest.OneByteReader(strings.NewReader("abcdef")), "abcdef", nil},
		{"empty", 0, strings.NewReader(""), "", nil},
		{"other error", 6, io.MultiReader(strings.NewReader("ab"), errorReader{readErr}), "ab", readErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header bytes.Buffer
			w := cpio.NewWriter(&header)
			if err := w.WriteHeader(&cpio.Header{Name: "file.txt", Mode: 0600, Size: tt.size}); err != nil {
				t.Fatal(err)
			}
			r := cpio.NewReader(io.MultiReader(&header, tt.body))
			if _, err := r.Next(); err != nil {
				t.Fatal(err)
			}
			body, err := ioutil.ReadAll(r)
			if err != tt.err {
				t.Errorf("ReadAll error = %v, want %v", err, tt.err)
			}
			if string(body) != tt.want {
				t.Errorf("ReadAll body = %q, want %q", body, tt.want)
			}
			if tt.err == nil {
				if n, err := r.Read(make([]byte, 1)); n != 0 || err != io.EOF {
					t.Errorf("Read after end of file = (%d, %v), want (0, EOF)", n, err)
				}
			}
		})
	}
}
