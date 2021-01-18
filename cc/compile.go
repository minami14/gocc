package cc

import (
	"bytes"
	"io"

	"github.com/minami14/gocc/asm"
	"github.com/minami14/gocc/ast"
)

var DefaultCompiler = new(Compiler)

type Source struct {
	io.Reader
}

type Compiler struct {
	Option Option
}

type Option struct{}

func (c *Compiler) Compile(src *Source) (*asm.Assembly, error) {
	tok, err := ast.Tokenize(src)
	if err != nil {
		return nil, err
	}

	node, err := ast.Expr(tok)
	if err != nil {
		return nil, err
	}

	gen, err := node.Gen()
	if err != nil {
		return nil, err
	}

	buf := []byte(".intel_syntax noprefix\n.globl main\nmain:\n")
	buf = append(buf, gen...)
	buf = append(buf, "  pop rax\n  ret\n"...)

	return &asm.Assembly{Reader: bytes.NewBuffer(buf)}, nil
}

func Compile(src *Source) (*asm.Assembly, error) {
	return DefaultCompiler.Compile(src)
}
