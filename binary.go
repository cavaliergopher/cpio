package cpio

import (
	"encoding/binary"
	"io"
	"time"
)

const (
	binaryHeaderLength = 26
)

func readInt16(b []byte) int {
	return int(binary.LittleEndian.Uint16(b))
}

func int64FromInt32(b []byte) int64 {
	// BigEndian order is called out in the cpio spec for the 16 bit words.
	// This hard-codes LittleEndian Machine within the 16-bit words which
	// should actually be parameterized by machine/archive
	t := int64(binary.BigEndian.Uint32(
		[]byte{
			b[1],
			b[0],
			b[3],
			b[2],
		},
	))
	return t
}

func readBinaryHeader(r io.Reader) (*Header, error) {
	// TODO: support binary-be

	var buf [binaryHeaderLength - 2]byte // -2 for magic already read
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, err
	}
	hdr := &Header{}

	hdr.DeviceID = readInt16(buf[:2])
	hdr.Inode = int64(readInt16(buf[2:4]))
	hdr.Mode = FileMode(readInt16(buf[4:6]))
	hdr.UID = readInt16(buf[6:8])
	hdr.GID = readInt16(buf[8:10])
	hdr.Links = readInt16(buf[10:12])
	// skips rdev at buf[12:14]
	hdr.ModTime = time.Unix(int64FromInt32(buf[14:18]), 0)
	nameSize := readInt16(buf[18:20])
	if nameSize < 1 {
		return nil, ErrHeader
	}

	hdr.Size = int64(int64FromInt32(buf[20:])) // :24 is the end

	name := make([]byte, nameSize)
	if _, err := io.ReadFull(r, name); err != nil {
		return nil, err
	}
	hdr.Name = string(name[:nameSize-1])
	if hdr.Name == headerEOF {
		return nil, io.EOF
	}

	// store padding between end of file and next header
	hdr.pad = (2 - (hdr.Size % 2)) % 2

	// skip to end of header/start of file
	// +2 for magic bytes already read
	pad := (2 - (2+len(buf)+len(name))%2) % 2
	if pad > 0 {
		if _, err := io.ReadFull(r, buf[:pad]); err != nil {
			return nil, err
		}
	}

	// read link name
	if hdr.Mode&^ModePerm == ModeSymlink {
		if hdr.Size < 1 {
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
