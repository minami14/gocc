package asm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/minami14/gocc/link"
)

var DefaultAssembler = new(Assembler)

var nextInt int32

func randName() string {
	name := fmt.Sprintf("minami14_gocc_asm_%v", atomic.AddInt32(&nextInt, 1))
	return filepath.Join(os.TempDir(), name)
}

type Assembly struct {
	io.Reader
}

type Assembler struct {
	Option Option
}

type Option struct{}

// TODO: Implement assembler in pure Go.
func (a *Assembler) Assemble(src *Assembly) (*link.Object, error) {
	name := randName()
	assembly := name + ".s"
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(assembly, data, 0644); err != nil {
		return nil, err
	}

	command := exec.Command("cc", "-o", name, assembly)
	if err := command.Run(); err != nil {
		return nil, err
	}

	out, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return &link.Object{Reader: bytes.NewBuffer(out)}, nil
}

func Assemble(src *Assembly) (*link.Object, error) {
	return DefaultAssembler.Assemble(src)
}
