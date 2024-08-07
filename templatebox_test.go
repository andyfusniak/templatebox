package templatebox_test

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andyfusniak/templatebox"
)

//go:embed testdata/templates/*.html
var templateFS embed.FS

// TestBoxOSDirBasic tests the basic functionality of the Box type. It
// creates a new Box from the OS filesystem and adds a template to it.
// It then renders the template to a buffer and compares the output to
// the expected output. The Box is created with the default configuration.
func TestBoxOSDirBasic(t *testing.T) {
	box, err := templatebox.NewBoxFromOSDir("testdata/templates", nil)
	if err != nil {
		t.Fatalf("NewBoxFromOSDir failed: %v", err)
	}

	if box == nil {
		t.Fatalf("NewBoxFromOSDir returned nil box")
	}

	if box.TemplateDir() != "testdata/templates" {
		t.Fatalf("TemplateDir() returned %s, expected %s", box.TemplateDir(), "testdata/templates")
	}

	if box.Config().Debug {
		t.Fatalf("Config().Debug returned true, expected false")
	}

	err = box.AddTemplate("a", templatebox.FileSet{
		Filenames: []string{"layout.html", "a.html"},
	})
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	var buf bytes.Buffer

	err = box.RenderHTML(&buf, "a", nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	expected := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>Page A</h1>
</body>
</html>
`

	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}
}

func TestBoxOSDirBasicWithDebug(t *testing.T) {
	path, err := os.MkdirTemp("", "templatebox-test-*")
	if err != nil {
		t.Fatalf("os.MkdirTemp failed: %v", err)
	}

	// create two templates in the path directory
	err = os.WriteFile(filepath.Join(path, "layout.html"), []byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  {{ template "content" . }}
</body>
</html>`), 0644)
	if err != nil {
		t.Fatalf("os.WriteFile failed: %v", err)
	}

	err = os.WriteFile(filepath.Join(path, "a.html"), []byte(`{{ define "content" }}<h1>Page A unchanged</h1>{{ end }}`), 0644)
	if err != nil {
		t.Fatalf("os.WriteFile failed: %v", err)
	}

	// create a new Box from the path directory
	box, err := templatebox.NewBoxFromOSDir(path, &templatebox.Config{
		Debug: true,
	})
	if err != nil {
		t.Fatalf("NewBoxFromOSDir failed: %v", err)
	}

	if box == nil {
		t.Fatalf("NewBoxFromOSDir returned nil box")
	}

	if box.TemplateDir() != path {
		t.Fatalf("TemplateDir() returned %s, expected %s", box.TemplateDir(), "testdata/templates")
	}

	if !box.Config().Debug {
		t.Fatalf("Config().Debug returned false, expected true")
	}

	err = box.AddTemplate("a", templatebox.FileSet{
		Filenames: []string{"layout.html", "a.html"},
	})
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	var buf bytes.Buffer

	// render the first time
	err = box.RenderHTML(&buf, "a", nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	expected := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>Page A unchanged</h1>
</body>
</html>`

	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}

	// modify the template
	err = os.WriteFile(filepath.Join(path, "a.html"), []byte(`{{ define "content" }}<h1>Page A changed</h1>{{ end }}`), 0644)
	if err != nil {
		t.Fatalf("os.WriteFile failed: %v", err)
	}

	// render the second time
	buf.Reset()
	err = box.RenderHTML(&buf, "a", nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	expected = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>Page A changed</h1>
</body>
</html>`

	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}

	// cleanup
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatalf("os.RemoveAll failed: %v", err)
	}
}

func TestBoxOSDirBasicWithLocalFuncMap(t *testing.T) {
	box, err := templatebox.NewBoxFromOSDir("testdata/templates", nil)
	if err != nil {
		t.Fatalf("NewBoxFromOSDir failed: %v", err)
	}

	if box == nil {
		t.Fatalf("NewBoxFromOSDir returned nil box")
	}

	if box.TemplateDir() != "testdata/templates" {
		t.Fatalf("TemplateDir() returned %s, expected %s", box.TemplateDir(), "testdata/templates")
	}

	if box.Config().Debug {
		t.Fatalf("Config().Debug returned true, expected false")
	}

	err = box.AddTemplate("a", templatebox.FileSet{
		Filenames: []string{"layout.html", "d.html"},
		FuncMap: map[string]any{
			"uppr": strings.ToUpper,
		},
	})
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	var buf bytes.Buffer
	data := struct {
		Name string
	}{
		Name: "test content",
	}
	err = box.RenderHTML(&buf, "a", data)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	expected := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>TEST CONTENT</h1>
</body>
</html>
`
	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}
}

// TestBoxFSDir tests the basic functionality of the Box type. It creates
// a new Box from the embed.FS filesystem and adds a template to it. It
// then renders the template to a buffer and compares the output to the
// expected output. The Box is created with the default configuration.
func TestBoxFSDir(t *testing.T) {
	box, err := templatebox.NewBoxFromFSDir(&templateFS, "testdata/templates", nil)
	if err != nil {
		t.Fatalf("NewBoxFromFSDir failed: %v", err)
	}

	if box == nil {
		t.Fatalf("NewBoxFromFSDir returned nil box")
	}

	if box.TemplateDir() != "testdata/templates" {
		t.Fatalf("TemplateDir() returned %s, expected %s", box.TemplateDir(), "testdata/templates")
	}

	if box.Config().Debug {
		t.Fatalf("Config().Debug returned true, expected false")
	}

	err = box.AddTemplate("a", templatebox.FileSet{
		Filenames: []string{"layout.html", "a.html"},
	})
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	var buf bytes.Buffer

	err = box.RenderHTML(&buf, "a", nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	expected := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>Page A</h1>
</body>
</html>
`

	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}
}

func TestBoxOSDirBasicWithRawTemplates(t *testing.T) {
	box, err := templatebox.NewBoxFromFSDir(&templateFS, "testdata/templates", nil)
	if err != nil {
		t.Fatalf("NewBoxFromFSDir failed: %v", err)
	}

	if box == nil {
		t.Fatalf("NewBoxFromFSDir returned nil box")
	}

	if box.TemplateDir() != "testdata/templates" {
		t.Fatalf("TemplateDir() returned %s, expected %s", box.TemplateDir(), "testdata/templates")
	}

	if box.Config().Debug {
		t.Fatalf("Config().Debug returned true, expected false")
	}

	err = box.AddTemplateRaw("t1", templatebox.TemplateSet{
		Templates: []string{`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  {{ template "content" . }}
</body>
</html>`, `{{ define "content" }}<h1>Hi from content template</h1>{{ end }}`},
	})
	if err != nil {
		t.Fatalf("AddTemplateRaw failed: %v", err)
	}

	expected := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Document</title>
</head>
<body>
  <h1>Hi from content template</h1>
</body>
</html>`

	var buf bytes.Buffer

	err = box.RenderHTML(&buf, "t1", nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if buf.String() != expected {
		t.Fatalf("RenderHTML returned %s, expected %s", buf.String(), expected)
	}
}
