package cpio

import (
	"os"
	"path"
	"time"
)

// headerFileInfo implements os.FileInfo.
type headerFileInfo struct {
	h *Header
}

// Name returns the base name of the file.
func (fi headerFileInfo) Name() string {
	if fi.IsDir() {
		return path.Base(path.Clean(fi.h.Name))
	}
	return path.Base(fi.h.Name)
}

func (fi headerFileInfo) Size() int64        { return fi.h.Size }
func (fi headerFileInfo) Mode() os.FileMode  { return fi.h.Mode }
func (fi headerFileInfo) IsDir() bool        { return fi.h.Mode.IsDir() }
func (fi headerFileInfo) ModTime() time.Time { return fi.h.ModTime }
func (fi headerFileInfo) Sys() interface{}   { return fi.h }
