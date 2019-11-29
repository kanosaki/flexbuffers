package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"flexbuffers"
)

func main() {
	out := os.Stdout
	in := bufio.NewReader(os.Stdin)

	for {
		var l uint32
		if err := binary.Read(in, binary.BigEndian, &l); err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		buf := make([]byte, l)
		_, err := io.ReadFull(in, buf)
		if err != nil {
			panic(err)
		}
		root, err := flexbuffers.Raw(buf).Root()
		if err != nil {
			panic(err)
		}
		if err := root.WriteAsJson(out); err != nil {
			panic(err)
		}
		if _, err := fmt.Fprintln(out); err != nil {
			panic(err)
		}
	}
}
