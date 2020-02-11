package update

import "flexbuffers/process"

type UpdateOperation struct {
	Path []string
	Op   Operation
}

type Operation interface {
	process.DocumentWriter
}
