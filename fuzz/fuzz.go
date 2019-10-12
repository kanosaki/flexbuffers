package fuzz

import "flexbuffers"

func Fuzz(data []byte) int {
	r := flexbuffers.Raw(data)
	r.Root()
	return 0
}
