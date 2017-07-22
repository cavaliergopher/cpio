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
	ErrHeader = errors.New("cpio: invalid cpio header")
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
	Inode    int64       // inode number
	Mode     os.FileMode // permission and mode bits
	UID      int         // user id of the owner
	GID      int         // group id of the owner
	Links    int         // number of inbound links
	ModTime  time.Time   // modified time
	Size     int64       // size in bytes
	Name     string      // filename
	Checksum Checksum    // computed checksum
}

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// FileInfoHeader creates a partially-populated Header from fi.
// Because os.FileInfo's Name method returns only the base name of
// the file it describes, it may be necessary to modify the Name field
// of the returned header to provide the full path name of the file.
func FileInfoHeader(fi os.FileInfo) (*Header, error) {
	if fi == nil {
		return nil, errors.New("cpio: FileInfo is nil")
	}
	if sys, ok := fi.Sys().(*Header); ok {
		// This FileInfo came from a Header (not the OS). Return a copy of the
		// original Header.
		h := &Header{}
		*h = *sys
		return h, nil
	}
	h := &Header{
		Name:    fi.Name(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		Size:    fi.Size(),
	}
	return h, nil
}

// ReadHeader creates a new Header, reading from r.
func readHeader(r io.Reader) (*Header, error) {
	// currently only SVR4 format is supported
	return readSVR4Header(r)
}
