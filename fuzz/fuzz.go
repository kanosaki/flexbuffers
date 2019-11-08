package fuzz

import "flexbuffers"

func Fuzz(data []byte) int {
	r := flexbuffers.Raw(data)
	if err := r.Validate(); err != nil {
		return -1
	} else {
		return 1
	}
}
