//go:build ignore
// +build ignore

// This is an example of how to use the templatebox package.
//
// The example demonstrates how to create a new Box from the OS filesystem
// and render a template.
//
// The example uses the following directory structure:
//
//	testdata/
//	└── templates/
//	    ├── layout.html
//	    ├── hello.html
//	    └── apples.html
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/andyfusniak/templatebox"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create a new Box from the OS filesystem.
	box, err := templatebox.NewBoxFromOSDir("./testdata/templates", nil)
	if err != nil {
		return fmt.Errorf("templatebox.NewBoxFromOSDir failed: %w", err)
	}

	// Add a template to the Box.
	// err = box.AddTemplate("hello", templatebox.FileSet{
	// 	Filenames: []string{"layout.html", "hello.html"},
	// })
	// if err != nil {
	// 	return fmt.Errorf("box.AddTemplate failed: %w", err)
	// }
	box.SetGlobalFuncMap(templatebox.FuncMap{
		"upper": strings.ToUpper,
	})

	err = box.AddTemplateMap(map[string]templatebox.FileSet{
		"apples": {
			Filenames: []string{"layout.html", "apples.html"},
			// FuncMap: templatebox.FuncMap{
			// 	"upper": strings.ToUpper,
			// },
		},
		"hello": {
			Filenames: []string{"layout.html", "hello.html"},
		},
	})
	if err != nil {
		return fmt.Errorf("box.AddTemplateMap failed: %w", err)
	}
	// Render the template.
	data := map[string]interface{}{
		"Name": "World!",
	}
	if err := box.RenderHTML(os.Stdout, "hello", data); err != nil {
		return fmt.Errorf("box.RenderHTML failed: %w", err)
	}
	if err := box.RenderHTML(os.Stdout, "apples", data); err != nil {
		return fmt.Errorf("box.RenderHTML failed: %w", err)
	}

	err = box.AddTemplateRaw("greetings", templatebox.TemplateSet{
		Templates: []string{
			`<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta http-equiv="X-UA-Compatible" content="IE=edge">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Greetings</title>
			</head>
			<body>
				{{ template "content" . }}
			</body>
			</html>
		`, `{{ define "content" }}hello{{ end }}{{ define "apples" }}apples{{ end }}`},
	})
	if err != nil {
		return fmt.Errorf("box.AddTemplateRaw failed: %w", err)
	}

	if err := box.RenderHTML(os.Stdout, "greetings", nil); err != nil {
		return fmt.Errorf("box.RenderHTML failed: %w", err)
	}

	return nil
}
