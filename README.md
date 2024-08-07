# templatebox

⚠️ **This package is experimental .**

templatebox provides a robust abstraction layer over Go HTML templates, allowing developers to manage and render a collection of template sets with ease. By enabling users to pass a map of names to template set, along with optional functions for rendering, templatebox simplifies the process of loading and managing templates.

Features:

- Simplify the process of rendering pages dynamically
- Load multiple template sets into memory
- Map template names to their corresponding sets
- Attach optional custom functions for rendering

## Installation

To install templatebox, use the following command:

```bash
go get github.com/andyfusniak/templatebox
```

## Usage

### Creating a Box

A `Box` struct represents a collection of template sets. To use templatebox, import the package and create a new instance of the `Box` struct using `NewBoxFromFSDir` or `NewBoxFromOSDir`. `NewBoxFromFSDir` reads templates from a directory in an embedded filesystem, while `NewBoxFromOSDir` reads templates from a directory on the host filesystem. Both functions return a pointer to a `Box` struct.

```go
import (
    "github.com/andyfusniak/templatebox"
)

const templateDir = "templates"

box, err := templatebox.NewBoxFromOSDir(templateDir, nil)
if err != nil {
    log.Fatalf("error creating template box: %v", err)
}
```

`NewBoxFromOSDir` accepts a templateDir string that specifies the root directory containing the templates. The second argument is an optional `Config` object that allows you to enable debug mode.

The `Config` object has one field:
- **Debug**: a boolean value that enables debug mode. When debug mode is enabled, the box will reload templates from the filesystem on every render. This is useful for development but should be disabled in production. The default value is false. It will have no effect if the box is created with `NewBoxFromFSDir` (since the embedded filesystem is read only).

Here is an example of creating a box with debug mode enabled:

```go
import (
    "github.com/andyfusniak/templatebox"
)

const templateDir = "templates"

box, err := templatebox.NewBoxFromOSDir(templateDir, templatebox.Config{
    Debug: true,
})
if err != nil {
    log.Fatalf("error creating template box: %v", err)
}
```

### Adding Templates

To add templates to the box, use the `AddTemplates`❶ method. This method takes a map of names to `FileSet`. The `FileSet`❷ struct contains two fields. The `Filenames` field is a slice of filenames, and the `FuncMap` field is an optional `FuncMap` object. The `Filenames`❸ field specifies the template files relative to the Box templateDir. The first file in the slice is the main template file that references the other templates. The `FuncMap`❹ object allows you to attach custom functions to the template set.


In the example code below, the `AddTemplate` method adds a template set named "mypage" to the box. The set consists of two files: "layout.html" and "hello.html". The `FuncMap` object is not used in this example.
```go
err = box.AddTemplate❶("mypage", templatebox.FileSet❷{
    Filenames❸: []string{"layout.html", "hello.html"},
})
if err != nil {
    log.Fatalf("error adding template: %v", err)
}
```

Here is an example with a `FuncMap` object. The `uppr` function is added to the template set. This function converts a string to uppercase. templatebox.FuncMap type has the same underlying type as the `template.FuncMap` type from the `html/template` package.

```go
err = box.AddTemplate❶("mypage", templatebox.FileSet❷{
    Filenames❸: []string{"layout.html", "hello.html"},
    FuncMap❹: templatebox.FuncMap{
        "uppr": strings.ToUpper,
    },
})
```

The `layout.html` file might look like this:
```html
<!DOCTYPE html>
<html lang="en">
...

<body>
  {{ template "hello.html" . }}
</body>
```

The `hello.html` file might look like this:
```html
<h1>Hello, {{ .Name | uppr }}</h1>
```

In those files, the `layout.html` file references the `hello.html` file using the `template` action. The `hello.html` file uses the `Name` field from the data passed to the template. The `uppr` function converts the `Name` field to uppercase. These files are Go templates and are not modified by templatebox.


### Rendering Templates

The `RenderHTML(w io.Writer, name string, data any)` method accepts an `io.Writer`, the name of the template to render, and data to pass to the template. This method renders the HTML template to the writer.

In the example code below, the `RenderHTML` method writes to standard output.
```go
data := map[string]interface{}{
    "Name": "World!",
}
err := box.RenderHTML(os.Stdout, "mypage", data);
if err != nil {
    log.Fatalf("error rendering template: %v", err)
}
```

Typically, you would use the `http.ResponseWriter` as the writer argument to `RenderHTML` in an HTTP handler.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Name": "World!",
    }
    err := box.RenderHTML(w, "mypage", data);
    if err != nil {
        log.Fatalf("error rendering template: %v", err)
        return
    }
}
```

### Thread Safety

The `Box` struct is safe for concurrent use. The `Box` struct is immutable after creation, so you can safely use it across multiple goroutines without any issues.
Internally, templatebox uses a `sync.RWMutex` to protect the box from concurrent reads and writes.
