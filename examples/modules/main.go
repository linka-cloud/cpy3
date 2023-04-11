package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	python3 "github.com/go-python/cpy3"
)

//go:embed all:src
var src embed.FS

func main() {
	python3.Py_Initialize()
	defer python3.Py_Finalize()

	fmt.Println("Loading modules...")

	modules, err := fs.Sub(src, "src")
	if err != nil {
		panic(err)
	}
	if err := loadModules(modules, "lib"); err != nil {
		panic(err)
	}
	main, err := src.ReadFile("src/main.py")
	if err != nil {
		panic(err)
	}
	if err := python3.PySys_SetArgv([]string{"main.py", "some-args"}); err != nil {
		panic(err)
	}
	fmt.Printf("running main.py...\n\n")
	if e := python3.PyRun_SimpleString(string(main)); e != 0 {
		panic(fmt.Errorf("PyRun_SimpleString failed with %d", e))
	}
}

func loadModules(src fs.FS, base string) error {
	return fs.WalkDir(src, base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".py" {
			return nil
		}
		b, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return loadModule(path, b)
	})
}

func loadModule(path string, content []byte) error {
	name := strings.TrimSuffix(strings.ReplaceAll(strings.TrimSuffix(path, ".py"), "/", "."), ".__init__")
	fmt.Printf("loading module %s (%s)\n", name, path)

	binPath := filepath.Join("/tmp/cpy3-modules", path+".bin")
	var c *python3.PyObject
	if _, err := os.Stat(binPath); err != nil {
		c = python3.Py_CompileString(string(content), path)
		if c == nil {
			python3.PyErr_Print()
			python3.PyErr_Clear()
			return fmt.Errorf("Py_CompileString failed for %s", path)
		}
		defer c.DecRef()
		if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
			return err
		}
		fmt.Printf("saving compiled python code to %s\n", binPath)
		if err := os.WriteFile(binPath, c.Marshal(), os.ModePerm); err != nil {
			return err
		}
	} else {
		fmt.Printf("loading compiled python code from %s\n", binPath)
		b, err := os.ReadFile(binPath)
		if err != nil {
			return err
		}
		c = python3.Unmarshal(b)
		if c == nil {
			python3.PyErr_Print()
			python3.PyErr_Clear()
			return fmt.Errorf("ObjectFromString failed for %s", path)
		}
		defer c.DecRef()
	}

	e := python3.PyImport_ExecCodeModule(name, c)
	if e == nil {
		python3.PyErr_Print()
		python3.PyErr_Clear()
		return fmt.Errorf("PyImport_ExecCodeModule failed for %s", path)
	}
	defer e.DecRef()

	return nil
}