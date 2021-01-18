package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	"github.com/minami14/gocc/asm"
	"github.com/minami14/gocc/cc"
	"github.com/minami14/gocc/link"
)

var out string

func main() {
	app := cli.NewApp()
	app.Name = "gocc"
	app.Usage = "c compiler"
	app.UsageText = "gocc [option] [source file]"
	app.Action = action
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "o",
			Usage:       "output file name",
			Destination: &out,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {
	if !c.Args().Present() {
		return errors.New("arguments required")
	}

	name := c.Args().Get(0)
	ext := filepath.Ext(name)
	if ext != ".c" {
		return fmt.Errorf("invalid extension: %v", name)
	}

	src, err := os.Open(name)
	if err != nil {
		return err
	}
	defer src.Close()

	var (
		assembly *asm.Assembly
		obj      *link.Object
		reader   io.Reader
	)

	assembly, err = cc.Compile(&cc.Source{Reader: src})
	if err != nil {
		return err
	}

	if out == "" {
		out = strings.TrimSuffix(name, ext)
	}

	outExt := filepath.Ext(out)
	if outExt == ".s" {
		reader = assembly
		goto End
	}

	obj, err = asm.Assemble(assembly)
	if err != nil {
		return err
	}

	if outExt == ".o" {
		reader = assembly
		goto End
	}

	reader, err = link.Link(obj)
	if err != nil {
		return err
	}

End:
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(out, data, 0744); err != nil {
		return err
	}

	return nil
}
