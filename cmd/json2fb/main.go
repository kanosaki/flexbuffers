package main

import (
	"bufio"
	"encoding/binary"
	"os"

	"flexbuffers/convert"
)

func main() {
	out := os.Stdout
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		raw, err := convert.FromJson(scanner.Bytes())
		if err != nil {
			panic(err)
		}
		if err := binary.Write(out, binary.BigEndian, uint32(len(raw))); err != nil {
			panic(err)
		}
		if _, err := out.Write(raw); err != nil {
			panic(err)
		}
	}
	if scanner.Err() != nil {
		panic(scanner.Err())
	}
}
