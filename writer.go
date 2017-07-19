package cpio

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrWriteTooLong    = errors.New("archive/cpio: write too long")
	ErrWriteAfterClose = errors.New("archive/cpio: write after close")
)

// zeroed buffer to copy for padding
var zeroBlock = make([]byte, 64)

// A Writer provides sequential writing of a CPIO archive. A CPIO archive
// consists of a sequence of files. Call WriteHeader to begin a new file, and
// then call Write to supply that file's data, writing at most hdr.Size bytes
// in total.
type Writer struct {
	w      io.Writer
	nb     int64 // number of unwritten bytes for current file entry
	pad    int64 // amount of padding to write after current file entry
	closed bool
	inode  int64
}

// NewWriter creates a new Writer writing to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// WriteHeader writes hdr and prepares to accept the file's contents.
// WriteHeader calls Flush if it is not the first header. Calling after a Close
// will return ErrWriteAfterClose.
func (w *Writer) WriteHeader(hdr *Header) (err error) {
	if w.closed {
		err = ErrWriteAfterClose
		return
	}

	if err = w.Flush(); err != nil {
		return
	}

	w.nb = hdr.size
	w.pad = (4 - (hdr.size % 4)) % 4

	// TODO: padding should be decided by header version

	w.inode++
	hdr.inode = w.inode
	_, err = writeHeader(w.w, hdr)
	return
}

// Write writes to the current entry in the CPIO archive. Write returns the
// error ErrWriteTooLong if more than hdr.Size bytes are written after
// WriteHeader.
func (w *Writer) Write(b []byte) (n int, err error) {
	if w.closed {
		err = ErrWriteAfterClose
		return
	}
	overwrite := false
	if int64(len(b)) > w.nb {
		b = b[0:w.nb]
		overwrite = true
	}
	n, err = w.w.Write(b)
	w.nb -= int64(n)
	if err == nil && overwrite {
		err = ErrWriteTooLong
		return
	}
	return
}

// Flush finishes writing the current file (optional).
func (w *Writer) Flush() error {
	if w.nb > 0 {
		return fmt.Errorf("archive/cpio: missed writing %d bytes", w.nb)
	}

	if w.pad == 0 {
		return nil
	}

	if _, err := w.w.Write(zeroBlock[:w.pad]); err != nil {
		return err
	}

	w.nb = 0
	w.pad = 0
	return nil
}

// Close closes the CPIO archive, flushing any unwritten data to the underlying
// writer.
func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	hdr := &Header{
		name:  string(svr4EOFHeader),
		links: 1,
	}
	if err := w.WriteHeader(hdr); err != nil {
		return fmt.Errorf("error writing final header: %v", err)
	}
	w.closed = true
	return nil
}
