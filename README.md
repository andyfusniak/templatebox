# templatebox

⚠️ **This package is not yet tested and is not ready for use.**

templatebox provides a robust abstraction layer over Go templates, allowing developers to manage and render a collection of template sets with ease. By enabling users to pass a map of names to template sets, along with optional functions for rendering, templatebox simplifies the process of loading and managing templates.

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

To use templatebox, import the package and create a new instance of the `Box` struct using `NewBoxFromFSDir` or `NewBoxFromOSDir`. `NewBoxFromFSDir` reads templates from a directory in of an embedded filesystem, while `NewBoxFromOSDir` reads templates from a directory in the host filesystem. Both functions return a pointer to a `Box` struct.

To add templates to the box, use the `AddTemplates` method. This method takes a map of names to template sets, where each template set is a `templatebox.FileSet` object.

```go
err = box.AddTemplate("mypage", templatebox.FileSet{
    Filenames: []string{"layout.html", "hello.html"},
})
if err != nil {
    log.Fatalf("error adding template: %v", err)
}
```

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
