package nap

import "io"

type BodyProvider interface {
	ContentType() string
	Body() (io.Reader, error)
}
