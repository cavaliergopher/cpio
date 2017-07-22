package cpio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	// headerEOF is the value of the filename of the last header
	// ("TRAILER!!!\x00") in a CPIO archive.
	headerEOF = []byte{0x54, 0x52, 0x41, 0x49, 0x4C, 0x45, 0x52, 0x21, 0x21, 0x21, 0x00}
)

var (
	ErrHeader = errors.New("archive/cpio: invalid cpio header")
)

// Checksum is the sum of all bytes	in the file data. This sum is computed
// treating all bytes as unsigned values and using unsigned arithmetic. Only
// the least-significant 32 bits of the sum are stored. Use NewHash to compute
// the actual checksum of an archived file.
type Checksum uint32

func (c Checksum) String() string {
	return fmt.Sprintf("%04X", uint32(c))
}

// A Header represents a single header in a CPIO archive.
type Header struct {
	DeviceID int
	Inode    int64
	Mode     os.FileMode
	UID      int
	GID      int
	Links    int
	ModTime  time.Time
	Size     int64
	Name     string
	Checksum Checksum
}

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// ReadHeader creates a new Header, reading from r.
func readHeader(r io.Reader) (*Header, error) {
	// currently only SVR4 format supported
	return readSVR4Header(r)
}
