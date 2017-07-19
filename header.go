package cpio

import (
	"io"
	"os"
	"time"
)

// compile time check
var _ os.FileInfo = &Header{}

// A Header represents a single header in a CPIO archive.
type Header struct {
	deviceID int
	inode    int64
	mode     os.FileMode
	uid      int
	gid      int
	links    int
	modTime  time.Time
	size     int64
	name     string
}

func (h *Header) Name() string {
	return h.name
}

func (h *Header) Size() int64 {
	return h.size
}

// ModTime reports the last modified timestamp of the file described in the
// header.
func (h *Header) ModTime() time.Time {
	return h.modTime
}

func (h *Header) Mode() os.FileMode {
	return h.mode
}

// IsDir reports whether the header describes a directory. That is, it tests for
// the ModeDir bit being set in Mode().
func (h *Header) IsDir() bool {
	return h.mode&040000 != 0
}

func (h *Header) Uid() int {
	return h.uid
}

func (h *Header) Gid() int {
	return h.gid
}

func (h *Header) Inode() int64 {
	return h.inode
}

func (h *Header) DeviceID() int {
	return h.deviceID
}

func (h *Header) Sys() interface{} {
	return nil
}

// ReadHeader creates a new Header, reading from r.
func readHeader(r io.Reader) (*Header, error) {
	// currently only SVR4 format supported
	return readSVR4Header(r)
}

func writeHeader(w io.Writer, hdr *Header) (n int, err error) {
	// currently only SVR4 format supported
	return writeSVR4Header(w, hdr)
}
