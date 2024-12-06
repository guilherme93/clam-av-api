package dto

import (
	"io"
)

type Request struct {
	FileName   string
	FileReader io.Reader
}

type Response struct {
	HasVirus  bool
	VirusText string
}

type APIErr struct {
	Err error
}
