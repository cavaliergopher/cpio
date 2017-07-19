package cpio

import (
	"io"
	"os"
	"time"
)

// A Header represents a single header in a CPIO archive.
type Header struct {
	DeviceID int
	Inode    int64
	Mode     os.FileMode
	Uid      int
	Gid      int
	Links    int
	ModTime  time.Time
	Size     int64
	Name     string
}

// IsDir reports whether the header describes a directory. That is, it tests for
// the ModeDir bit being set in Mode().
//func (h *Header) IsDir() bool {
//	return h.mode&040000 != 0
//}

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
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
