package test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/minami14/gocc/asm"
	"github.com/minami14/gocc/cc"
	"github.com/minami14/gocc/link"
)

var outName = filepath.Join(os.TempDir(), "test.out")

func compileTest() error {
	cSrc, err := os.Open("src/test.c")
	if err != nil {
		return err
	}
	defer cSrc.Close()

	assembly, err := cc.Compile(&cc.Source{Reader: cSrc})
	if err != nil {
		return err
	}

	obj, err := asm.Assemble(assembly)
	if err != nil {
		return err
	}

	r, err := link.Link(obj)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outName, data, 0744); err != nil {
		return err
	}

	return nil
}

func TestCompile(t *testing.T) {
	if err := compileTest(); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(outName)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		t.Log(err)
	}

	if command.ProcessState.ExitCode() != -10+12-2/2*(5+1)+14 {
		t.Error(command.ProcessState.ExitCode())
	}
}
