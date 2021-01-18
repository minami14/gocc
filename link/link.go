package link

import "io"

var DefaultLinker = new(Linker)

type Object struct {
	io.Reader
}

type Linker struct {
	Option Option
}

type Option struct{}

func (l *Linker) Link(obj *Object) (io.Reader, error) {
	return obj, nil
}

func Link(obj *Object) (io.Reader, error) {
	return DefaultLinker.Link(obj)
}
