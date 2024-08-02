package main

import (
	"fmt"
	"os"

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
	box, err := templatebox.NewBoxFromOSDir("./testdata/templates")
	if err != nil {
		return fmt.Errorf("templatebox.NewBoxFromOSDir failed: %w", err)
	}
	fmt.Printf("%#v\n", box)

	// Add a template to the Box.
	err = box.AddTemplate("hello", templatebox.FileSet{
		Filenames: []string{"layout.html", "hello.html"},
	})
	if err != nil {
		return fmt.Errorf("box.AddTemplate failed: %w", err)
	}

	// Render the template.
	data := map[string]interface{}{
		"Name": "World!",
	}
	if err := box.RenderHTML(os.Stdout, "hello", data); err != nil {
		return fmt.Errorf("box.RenderHTML failed: %w", err)
	}

	return nil
}
