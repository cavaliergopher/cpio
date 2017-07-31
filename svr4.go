package cpio

import (
	"bytes"
	"io"
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
	if _, err := io.ReadFull(r, buf[:]); err != nil {
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

	asc := string(buf[:])
	hdr := &Header{}

	hdr.Inode = readHex(asc[6:14])
	hdr.Mode = FileMode(readHex(asc[14:22]))
	hdr.UID = int(readHex(asc[22:30]))
	hdr.GID = int(readHex(asc[30:38]))
	hdr.Links = int(readHex(asc[38:46]))
	hdr.ModTime = time.Unix(readHex(asc[46:54]), 0)
	hdr.Size = readHex(asc[54:62])
	if hdr.Size > svr4MaxFileSize {
		return nil, ErrHeader
	}
	nameSize := readHex(asc[94:102])
	if nameSize < 1 || nameSize > svr4MaxNameSize {
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

	// padding between end of file and next header
	hdr.pad = (4 - (hdr.Size % 4)) % 4

	// skip to end of header/start of file
	pad := (4 - (len(buf)+len(name))%4) % 4
	if pad > 0 {
		if _, err := io.ReadFull(r, buf[:pad]); err != nil {
			return nil, err
		}
	}

	// read link name
	if hdr.Mode&^ModePerm == ModeSymlink {
		if hdr.Size < 1 || hdr.Size > svr4MaxNameSize {
			return nil, ErrHeader
		}
		b := make([]byte, hdr.Size)
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}
		hdr.Linkname = string(b)
		hdr.Size = 0
	}

	return hdr, nil
}
