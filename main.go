package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed templates/*
var templates embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()
	cmd := Goinit{dir: flag.Arg(0)}
	return cmd.Run()
}

type Goinit struct {
	dir string
}

func (g *Goinit) Run() error {
	steps := []func() error{
		func() error { return os.Mkdir(g.dir, 0755) },
		func() error { return g.writeTemplate("main.go") },
		func() error { return g.cmd("go", "mod", "init") },
		func() error { return g.cmd("go", "get", "github.com/stretchr/testify") },
	}

	if inRepo, err := g.insideRepo(); err != nil {
		return err
	} else if !inRepo {
		steps = append(steps,
			func() error { return g.cmd("git", "init") },
			func() error { return g.cmd("git", "add", ".") },
			func() error { return g.cmd("git", "commit", "-m", "goinit") },
		)
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Goinit) insideRepo() (bool, error) {
	if info, err := os.Stat(filepath.Join(filepath.Dir(g.dir), ".git")); err == nil {
		return info.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (g *Goinit) writeTemplate(name string) error {
	data, err := templates.ReadFile(filepath.Join("templates/", name))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(g.dir, name), data, 0644)
}

func (g *Goinit) cmd(name string, args ...string) error {
	var out bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Dir = g.dir
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}
