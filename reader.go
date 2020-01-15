package cpio

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
)

var magic map[string][]byte = map[string][]byte{
	"binary-le": []byte{0xc7, 0x71},                         // 070707
	"svr4":      []byte{0x30, 0x37, 0x30, 0x37, 0x30, 0x31}, // "070701"
	"svr4-crc":  []byte{0x30, 0x37, 0x30, 0x37, 0x30, 0x32}, // "070702"
}

// A Reader provides sequential access to the contents of a CPIO archive. A CPIO
// archive consists of a sequence of files. The Next method advances to the next
// file in the archive (including the first), and then it can be treated as an
// io.Reader to access the file's data.
type Reader struct {
	r      io.Reader // underlying file reader
	hdr    *Header   // current Header
	eof    int64     // bytes until the end of the current file
	robust bool      // try to skip junk bytes when reading an archive
}

// NewReader creates a new Reader reading from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

// Read reads from the current entry in the CPIO archive. It returns 0, io.EOF
// when it reaches the end of that entry, until Next is called to advance to the
// next entry.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.hdr == nil || r.eof == 0 {
		return 0, io.EOF
	}
	rn := len(p)
	if r.eof < int64(rn) {
		rn = int(r.eof)
	}
	n, err = r.r.Read(p[0:rn])
	r.eof -= int64(n)
	return
}

// Next advances to the next entry in the CPIO archive.
// io.EOF is returned at the end of the input.
func (r *Reader) Next() (*Header, error) {
	if r.hdr == nil {
		return r.next()
	}
	skp := r.eof + r.hdr.pad
	if skp > 0 {
		_, err := io.CopyN(ioutil.Discard, r.r, skp)
		if err != nil {
			return nil, err
		}
	}
	return r.next()
}

func (r *Reader) next() (*Header, error) {
	r.eof = 0
	hdr, err := readHeader(r.r, r.robust)
	if err != nil {
		return nil, err
	}
	r.hdr = hdr
	r.eof = hdr.Size
	return hdr, nil
}

// ReadHeader creates a new Header, reading from r.
func readHeader(r io.Reader, robust bool) (*Header, error) {

	var junk = 0
	readOne := func() (*Header, error) {
		// A binary magic is two bytes, an ascii magic is 6.
		var binMagic [2]byte

		if n, err := io.ReadFull(r, binMagic[:]); err != nil {
			junk += n
			return nil, err
		}

		if bytes.Equal(binMagic[:], magic["binary-le"]) {
			// try to read a binary header
			return readBinaryHeader(r)
		}

		// check if this starts an ascii magic
		if bytes.Equal(binMagic[:], magic["svr4"][:2]) {
			var ascMagic [6]byte
			copy(ascMagic[:], binMagic[:])
			if n, err := io.ReadFull(r, ascMagic[2:]); err != nil {
				junk += 2 + n
				return nil, err
			}
			if bytes.Equal(ascMagic[:], magic["svr4"]) {
				return readSVR4Header(r, false)
			} else if bytes.Equal(ascMagic[:], magic["svr4-crc"]) {
				return readSVR4Header(r, true)
			} else {
				junk += 6
				return nil, ErrMagic
			}
		}

		junk += 2
		return nil, ErrMagic
	}

	if robust {
		for {
			hdr, err := readOne()
			if err == ErrMagic {
				continue
			}
			if err != nil {
				return nil, err
			}
			if junk > 0 {
				log.Printf("skipped %d bytes of junk\n", junk)
			}
			return hdr, nil
		}
	} else {
		return readOne()
	}
}
