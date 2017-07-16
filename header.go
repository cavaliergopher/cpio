package cpio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	asciiV2Magic = []byte{0x30, 0x37, 0x30, 0x37, 0x30, 0x31}                               // "070701"
	eofHeader    = []byte{0x54, 0x52, 0x41, 0x49, 0x4C, 0x45, 0x52, 0x21, 0x21, 0x21, 0x00} // "TRAILER!!!\x00"
)

// A Header represents a single header in a CPIO archive.
type Header struct {
	DeviceID int
	Inode    int
	Mode     os.FileMode
	Owner    int
	Group    int
	Links    int
	ModTime  time.Time
	Size     int64
	Name     string
}

func (h *Header) IsDir() bool {
	if h.Mode&040000 != 0 {
		return true
	}

	return false
}

// ReadHeader creates a new Header, reading from r.
func readHeader(r io.Reader) (*Header, error) {
	// currently only v2 ASCII format supported
	return readASCIIV2Header(r)
}

func writeHeader(w io.Writer, hdr *Header) (n int, err error) {
	// currently only v2 ASCII format supported
	return writeASCIIV2Header(w, hdr)
}

func readASCIIV2Header(r io.Reader) (*Header, error) {
	var buf [110]byte
	if _, err := r.Read(buf[:]); err != nil {
		if err == io.EOF {
			return nil, err
		}

		return nil, fmt.Errorf("error reading file header: %v", err)
	}

	// check magic
	if !bytes.Equal(asciiV2Magic, buf[:6]) {
		return nil, fmt.Errorf("error reading file header: invalid magic number: %0X", buf[:6])
	}

	h := &Header{}

	// read file name
	var nameSize int64
	if _, err := fmt.Sscanf(string(buf[94:102]), "%X", &nameSize); err != nil {
		return nil, fmt.Errorf("error reading name length in file header: %v", err)
	}

	name := make([]byte, nameSize)
	if _, err := io.ReadFull(r, name); err != nil {
		return nil, fmt.Errorf("error reading name in file header: %v", err)
	}
	h.Name = string(name[:nameSize-1]) // trim null terminator

	// check for EOF string
	if bytes.Equal(name, eofHeader) {
		return nil, io.EOF
	}

	// parse remaining fields
	if _, err := fmt.Sscanf(string(buf[6:14]), "%X", &h.Inode); err != nil {
		return nil, fmt.Errorf("error reading inode in file header: %v", err)
	}
	if _, err := fmt.Sscanf(string(buf[14:22]), "%X", &h.Mode); err != nil {
		return nil, fmt.Errorf("error reading mode in file header: %v", err)
	}
	if _, err := fmt.Sscanf(string(buf[22:30]), "%X", &h.Owner); err != nil {
		return nil, fmt.Errorf("error reading Owner in file header: %v", err)
	}
	if _, err := fmt.Sscanf(string(buf[30:38]), "%X", &h.Group); err != nil {
		return nil, fmt.Errorf("error reading device Group in file header: %v", err)
	}
	if _, err := fmt.Sscanf(string(buf[38:46]), "%X", &h.Links); err != nil {
		return nil, fmt.Errorf("error reading link count in file header: %v", err)
	}
	if _, err := fmt.Sscanf(string(buf[54:62]), "%X", &h.Size); err != nil {
		return nil, fmt.Errorf("error reading file size in file header: %v", err)
	}

	var unixTime int64
	if _, err := fmt.Sscanf(string(buf[46:54]), "%X", &unixTime); err != nil {
		return nil, fmt.Errorf("error reading modified time in file header: %v", err)
	}
	h.ModTime = time.Unix(unixTime, 0)

	// skip to end of header - padding to a multiple of 4
	pad := (4 - (len(buf)+len(name))%4) % 4
	if pad > 0 {
		if _, err := io.ReadFull(r, buf[:pad]); err != nil {
			return nil, fmt.Errorf("error reading to end of header: %v", err)
		}
	}

	return h, nil
}

func writeASCIIV2Header(w io.Writer, hdr *Header) (n int, err error) {
	n, err = w.Write(asciiV2Magic)
	if err != nil {
		return
	}

	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Inode)))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Mode)))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Owner)))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Group)))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Links)))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.ModTime.Unix())))
	if err != nil {
		return
	}
	n, err = w.Write([]byte(fmt.Sprintf("%08X", hdr.Size)))
	if err != nil {
		return
	}

	// dev/rdev major/minor
	n, err = w.Write([]byte("00000000000000000000000000000000"))
	if err != nil {
		return
	}

	n, err = w.Write([]byte(fmt.Sprintf("%08X", len(hdr.Name)+1)))
	if err != nil {
		return
	}

	// nil check
	n, err = w.Write([]byte("00000000"))
	if err != nil {
		return
	}

	n, err = w.Write([]byte(hdr.Name))
	if err != nil {
		return
	}

	// pad to multiple of 4
	// 111 is the length of the header plus the null-terminator for the name
	pad := (4 - ((111 + len(hdr.Name)) % 4)) % 4
	n, err = w.Write(zeroBlock[:pad])
	if err != nil {
		return
	}

	return
}
