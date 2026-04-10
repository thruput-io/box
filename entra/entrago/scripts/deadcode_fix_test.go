package main

import (
	"go/parser"
	"go/token"
	"os"
	"testing"
)

func TestFilterDecls(t *testing.T) {
	t.Parallel()

	src := `package main
func Keep() {}
func Remove() {}
`

	const removeLine = 3

	const expectedDecls = 1

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, "test.go", src, zero)
	if err != nil {
		t.Fatal(err)
	}

	// Remove second function (starts at line 3)
	lines := []int{removeLine}
	newDecls, modified := filterDecls(node.Decls, lines, fset)

	if !modified {
		t.Error("expected modification")
	}

	if len(newDecls) != expectedDecls {
		t.Errorf("expected %d decl, got %d", expectedDecls, len(newDecls))
	}
}

func TestProcessPackages(t *testing.T) {
	t.Parallel()

	// Create a temporary file to test removeFuncs as part of processPackages
	content := `package main
func Dead() {}
`

	const deadLine = 2

	tmpDir := t.TempDir()

	tmpfile, err := os.CreateTemp(tmpDir, "deadcode_test_*.go")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	_, err = tmpfile.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}

	_ = tmpfile.Close()

	pkgs := []Package{
		{
			Name: "main",
			Path: "main",
			Funcs: []Function{
				{
					Name: "Dead",
					Position: Position{
						File: tmpfile.Name(),
						Line: deadLine,
					},
				},
			},
		},
	}

	processPackages(pkgs)

	// Check if Dead was removed
	newContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(newContent) == content {
		t.Error("expected content to change")
	}
}

func TestRun_Empty(t *testing.T) {
	t.Parallel()
}

func TestParseDeadcodeOutput(t *testing.T) {
	t.Parallel()

	data := `[{"Name":"pkg","Path":"path","Funcs":[{"Name":"fn","Position":{"File":"f.go","Line":1}}]}]`

	const expectedCount = 1

	pkgs, err := parseDeadcodeOutput([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != expectedCount || pkgs[zero].Name != "pkg" {
		t.Errorf("unexpected output: %v", pkgs)
	}
}
