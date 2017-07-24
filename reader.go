package cpio

import (
	"io"
	"io/ioutil"
)

// A Reader provides sequential access to the contents of a CPIO archive. A CPIO
// archive consists of a sequence of files. The Next method advances to the next
// file in the archive (including the first), and then it can be treated as an
// io.Reader to access the file's data.
type Reader struct {
	r   io.Reader // underlying file reader
	hdr *Header   // current Header
	eof int64     // bytes until the end of the current file
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

	// skip ahead
	// TODO: padding is version specific. Should be determined from header
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
	var err error
	r.eof = 0
	r.hdr, err = readHeader(r.r)
	if err == nil {
		r.eof = r.hdr.Size
	}
	return r.hdr, err
}
