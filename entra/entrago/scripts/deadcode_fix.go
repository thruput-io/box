// Package main provides a tool to automatically remove dead code reported by the deadcode tool.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"time"
)

// Package represents a package as reported by deadcode -json.
//
//nolint:tagliatelle // deadcode -json uses PascalCase
type Package struct {
	Name  string     `json:"Name"` // deadcode -json uses PascalCase
	Path  string     `json:"Path"`
	Funcs []Function `json:"Funcs"`
}

// Function represents a dead function.
//
//nolint:tagliatelle // deadcode -json uses PascalCase
type Function struct {
	Name     string   `json:"Name"` // deadcode -json uses PascalCase
	Position Position `json:"Position"`
}

// Position represents a source position.
//
//nolint:tagliatelle // deadcode -json uses PascalCase
type Position struct {
	File string `json:"File"` // deadcode -json uses PascalCase
	Line int    `json:"Line"`
}

const (
	timeoutSeconds = 60
	exitCodeError  = 1
	zero           = 0
)

func main() {
	err := run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(exitCodeError)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	out, err := runDeadcode(ctx)
	if err != nil {
		return err
	}

	if len(out) == zero {
		return nil
	}

	pkgs, err := parseDeadcodeOutput(out)
	if err != nil {
		return err
	}

	processPackages(pkgs)

	return nil
}

func runDeadcode(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "deadcode", "-json", "-test", "./...")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run deadcode: %w", err)
	}

	return out, nil
}

func parseDeadcodeOutput(out []byte) ([]Package, error) {
	var pkgs []Package

	err := json.Unmarshal(out, &pkgs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return pkgs, nil
}

func processPackages(pkgs []Package) {
	fileFuncs := make(map[string][]int)

	for _, pkg := range pkgs {
		for _, fn := range pkg.Funcs {
			fileFuncs[fn.Position.File] = append(fileFuncs[fn.Position.File], fn.Position.Line)
		}
	}

	for file, lines := range fileFuncs {
		_, _ = fmt.Fprintf(os.Stdout, "Processing %s: removing functions at lines %v\n", file, lines)

		err := removeFuncs(file, lines)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
		}
	}
}

func removeFuncs(file string, lines []int) error {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", file, err)
	}

	newDecls, modified := filterDecls(node.Decls, lines, fset)

	if modified {
		node.Decls = newDecls

		err = writeNodeToFile(file, fset, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func filterDecls(decls []ast.Decl, lines []int, fset *token.FileSet) ([]ast.Decl, bool) {
	lineToRemove := make(map[int]bool)

	for _, l := range lines {
		lineToRemove[l] = true
	}

	newDecls := make([]ast.Decl, zero, len(decls))
	modified := false

	for _, decl := range decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			pos := fset.Position(fd.Pos())
			if lineToRemove[pos.Line] {
				modified = true

				continue
			}
		}

		newDecls = append(newDecls, decl)
	}

	return newDecls, modified
}

func writeNodeToFile(file string, fset *token.FileSet, node *ast.File) error {
	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", file, err)
	}

	err = format.Node(outFile, fset, node)
	if err != nil {
		_ = outFile.Close()

		return fmt.Errorf("failed to format node for file %s: %w", file, err)
	}

	err = outFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s: %w", file, err)
	}

	return nil
}
