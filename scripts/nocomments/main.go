package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var generatedMarker = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

var directivePrefixes = []string{"//go:", "//nolint", "//line ", "//export ", "// +build"}

func allowed(text string) bool {
	if generatedMarker.MatchString(text) {
		return true
	}
	for _, p := range directivePrefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

func skipDir(root, path, name string) bool {
	if path == root {
		return false
	}
	return strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata"
}

func check(root string, report func(pos token.Position, text string)) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(root, path, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, group := range f.Comments {
			for _, c := range group.List {
				if !allowed(c.Text) {
					report(fset.Position(c.Pos()), c.Text)
				}
			}
		}
		return nil
	})
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	found := 0
	err := check(root, func(pos token.Position, text string) {
		line, _, _ := strings.Cut(text, "\n")
		fmt.Fprintf(os.Stderr, "%s: comment not allowed: %s\n", pos, line)
		found++
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if found > 0 {
		fmt.Fprintf(os.Stderr, "%d comment(s) found; the codebase carries no comments, rename instead\n", found)
		os.Exit(1)
	}
}
