package cpio

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	svr4MaxNameSize = 4096 // MAX_PATH
	svr4MaxFileSize = 4294967295
)

func readHex(s string) int64 {
	// errors are ignored and 0 returned
	i, _ := strconv.ParseInt(s, 16, 64)
	return i
}

func readSVR4Header(r io.Reader) (*Header, error) {
	var buf [110]byte
	if _, err := r.Read(buf[:]); err != nil {
		return nil, err
	}

	// TODO: check endianness

	// check magic
	hasCRC := false
	if !bytes.HasPrefix(buf[:], []byte{0x30, 0x37, 0x30, 0x37, 0x30}) { // 07070
		return nil, ErrHeader
	}
	if buf[5] == 0x32 { // '2'
		hasCRC = true
	} else if buf[5] != 0x31 { // '1'
		return nil, ErrHeader
	}

	// create invariant that all bytes are valid hex values in ascii
	for i := 0; i < len(buf); i++ {
		if buf[i] < 0x30 || buf[i] > 0x46 { // ASCII '0' to 'F'
			return nil, ErrHeader
		}
	}

	asc := string(buf[:])
	hdr := &Header{}

	hdr.Inode = readHex(asc[6:14])
	hdr.Mode = os.FileMode(readHex(asc[14:22]))
	hdr.UID = int(readHex(asc[22:30]))
	hdr.GID = int(readHex(asc[30:38]))
	hdr.Links = int(readHex(asc[38:46]))
	hdr.ModTime = time.Unix(readHex(asc[46:54]), 0)
	hdr.Size = readHex(asc[54:62])
	if hdr.Size > svr4MaxFileSize {
		return nil, ErrHeader
	}
	nameSize := readHex(asc[94:102])
	if nameSize > svr4MaxNameSize {
		return nil, ErrHeader
	}
	hdr.Checksum = Checksum(readHex(asc[102:110]))
	if !hasCRC && hdr.Checksum != 0 {
		return nil, ErrHeader
	}

	name := make([]byte, nameSize)
	if _, err := io.ReadFull(r, name); err != nil {
		return nil, err
	}
	if bytes.Equal(name, headerEOF) {
		return nil, io.EOF
	}
	hdr.Name = string(name[:nameSize-1])

	// skip to end of header - padding to a multiple of 4
	pad := (4 - (len(buf)+len(name))%4) % 4
	if pad > 0 {
		if _, err := io.ReadFull(r, buf[:pad]); err != nil {
			return nil, nil
		}
	}

	return hdr, nil
}
