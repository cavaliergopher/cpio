package cpio

import (
	"fmt"
	"io"
	"os"
)

func Example() {
	f, err := os.Open("testdata/test_svr4_crc.cpio")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening cpio archive: %v\n", err)
		return
	}
	defer f.Close()

	r := NewReader(f)
	for {
		hdr, err := r.Next()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error reading cpio header: %v\n", err)
			}
			return
		}

		if hdr.Mode.IsRegular() {
			fmt.Printf("=== %v ===\n", hdr.Name)
			io.Copy(os.Stdout, r)
		}
	}

	// Output:
	// === gophers.txt ===
	// Gopher names:
	// George
	// Geoffrey
	// Gonzo
	// === readme.txt ===
	// This archive contains some text files.
	// === todo.txt ===
	// Get animal handling license.
}
