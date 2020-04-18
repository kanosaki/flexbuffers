package manipulation

import "flexbuffers/process"

type UpdateOperation struct {
	Path []string
	Op   Operation
}

type Operation interface {
	process.DocumentWriter
}
