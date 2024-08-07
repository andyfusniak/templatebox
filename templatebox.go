package templatebox

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FuncMap is a map of functions that can be added to a template.
type FuncMap map[string]any

// Box is a collection of templates and a global FuncMap that can be used to
// render templates loaded from the filesystem or embed.FS.
type Box struct {
	cfg           *Config
	fs            *embed.FS
	templateDir   string
	globalFuncMap FuncMap

	mu   sync.RWMutex
	html map[string]*template.Template

	// set of name to template map to be used for rebuilding the template
	// upon every request
	muHTMLRerender        sync.RWMutex
	rerenderTemplatesHTML map[string]FileSet
}

// Config is a configuration struct for creating a new Box. The Debug field
// is used to enable debug mode. In debug mode the Box will check
// if the template needs to be rebuilt before rendering it. This is useful
// during development to avoid restarting the application when a template
// changes.
type Config struct {
	Debug bool
}

// default config
var defaultConfig = &Config{
	Debug: false,
}

// NewBoxFromFS creates a new Box with the given embed.FS. The templateDir
// is the directory within the embed.FS where the templates are located. The
// Box will use the embed.FS to read the templates. If the embed.FS is nil
// then an error is returned. The Box will use the default configuration if
// cfg is nil. The default configuration has Debug set to false.
func NewBoxFromFSDir(fs *embed.FS, templateDir string, cfg *Config) (*Box, error) {
	if cfg == nil {
		cfg = defaultConfig
	}

	if fs == nil {
		return nil, fmt.Errorf("embed.FS cannot be nil")
	}
	box := Box{
		cfg:         cfg,
		fs:          fs,
		templateDir: templateDir,
		html:        make(map[string]*template.Template),
	}
	if cfg.Debug {
		box.rerenderTemplatesHTML = make(map[string]FileSet)
	}
	return &box, nil
}

// NewBoxFromOSDir creates a new Box for the OS filesystem at the given
// templateDir.
func NewBoxFromOSDir(templateDir string, cfg *Config) (*Box, error) {
	if cfg == nil {
		cfg = defaultConfig
	}

	// ensure the templateDir exists
	_, err := os.Stat(templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template directory %s does not exist", templateDir)
		}
		return nil, fmt.Errorf("os.Stat failed: %w", err)
	}

	box := Box{
		cfg:         cfg,
		templateDir: templateDir,
		html:        make(map[string]*template.Template),
	}
	if cfg.Debug {
		box.rerenderTemplatesHTML = make(map[string]FileSet)
	}
	return &box, nil
}

// FileSet is a set of template files and a FuncMap. The FuncMap is used to
// add functions to that template.
type FileSet struct {
	Filenames []string
	FuncMap   FuncMap
}

// TemplateSet is a set of template strings and a FuncMap. The FuncMap is used to
// add functions to that template.
type TemplateSet struct {
	Templates []string
	FuncMap   FuncMap
}

// SetGlobalFuncMap sets the global FuncMap available to all templates.
func (b *Box) SetGlobalFuncMap(g FuncMap) {
	b.globalFuncMap = g
}

// AddTemplateMap accepts a map of template names to FileSets and adds the
// templates to the Box. The map key is the name of the template and the value
// is the FileSet. The FileSet must contain at least one filename. The first
// filename in the FileSet is used as the name of the template. The filenames
// in the FileSet must be relative to the templateDir. The FuncMap in the
// FileSet is added to the template.
func (b *Box) AddTemplateMap(m map[string]FileSet) error {
	for k, v := range m {
		if err := b.AddTemplate(k, v); err != nil {
			return err
		}
	}
	return nil
}

// AddTemplate accepts either a FileSet or StringSet and adds the template to
// the Box.
func (b *Box) AddTemplate(name string, s FileSet) error {
	if len(s.Filenames) == 0 {
		return fmt.Errorf("no filenames provided")
	}

	// the first filename in the FileSet is used as the name of the template
	// although RenderHTML will call Execute without a name so the name is
	// not strictly necessary but it is useful for debugging.
	t := template.New(s.Filenames[0])
	if b.globalFuncMap != nil {
		t = t.Funcs(template.FuncMap(b.globalFuncMap))
	}
	if s.FuncMap != nil {
		t = t.Funcs(template.FuncMap(s.FuncMap))
	}

	// all templates filenames within the FileSet must be relative to the
	// templateDir
	var filenames []string
	if b.templateDir != "" {
		filenames = make([]string, len(s.Filenames))
		for i, filename := range s.Filenames {
			filenames[i] = filepath.Join(b.templateDir, filename)
		}
	} else {
		filenames = s.Filenames
	}

	// if b.fs is nil then we are using the OS filesystem
	// and we need to read the template files from the OS filesystem
	// otherwise we are using the embed.FS and we need to read the
	// template files from the embed.FS.
	var err error
	if b.fs == nil {
		t, err = t.ParseFiles(filenames...)
	} else if len(s.Filenames) > 0 {
		t, err = t.ParseFS(b.fs, filenames...)
	}
	if err != nil {
		return fmt.Errorf("add template failed: %w", err)
	}

	b.mu.Lock()
	b.html[name] = t
	b.mu.Unlock()

	// keep a copy of the FileSet to be used for rebuilding the template
	// upon every call to RenderHTML
	if b.cfg.Debug {
		b.muHTMLRerender.Lock()
		b.rerenderTemplatesHTML[name] = s
		b.muHTMLRerender.Unlock()
	}
	return nil
}

// AddTemplateRaw accepts a name and a TemplateSet and adds the template
// to the Box. The name is the key used to add the template to the Box. The
// TemplateSet must contain at least one template string. The first template
// string in the TemplateSet is used as the name of the template. The FuncMap
// in the TemplateSet is added to the template. The template is parsed using
// the html/template package.
func (b *Box) AddTemplateRaw(name string, s TemplateSet) error {
	if len(s.Templates) == 0 {
		return fmt.Errorf("no templates provided")
	}

	// initialise the template with the first template string in the TemplateSet
	t := template.New(name)
	if b.globalFuncMap != nil {
		t = t.Funcs(template.FuncMap(b.globalFuncMap))
	}
	if s.FuncMap != nil {
		t = t.Funcs(template.FuncMap(s.FuncMap))
	}

	for i, tmplStr := range s.Templates {
		var err error
		t, err = t.Parse(tmplStr)
		if err != nil {
			return fmt.Errorf("failed to parse template %s at index %d: %w\nTemplate content:\n%s",
				name, i, err, tmplStr)
		}
	}

	b.mu.Lock()
	b.html[name] = t
	b.mu.Unlock()

	return nil
}

// Config returns the Box configuration.
func (b *Box) Config() *Config {
	return b.cfg
}

// TemplateDir returns the template directory.
func (b *Box) TemplateDir() string {
	return b.templateDir
}

// RenderHTML renders the named template to the given io.Writer with the given
// data. The data is passed to the template for rendering. No structure is
// enforced on the data so it should match the structure expected by the
// template. The template must have been added to the Box using AddTemplate
// otherwise an error is returned. The name of the template is the key used to
// add the template to the Box.
func (b *Box) RenderHTML(w io.Writer, name string, data any) error {
	if b.cfg.Debug {
		// check if the template needs to be rebuilt
		b.muHTMLRerender.RLock()
		s1, ok := b.rerenderTemplatesHTML[name]
		b.muHTMLRerender.RUnlock()

		// only rebuild from OS filesystem (embed.FS is read-only)
		if ok && b.fs == nil {
			if err := b.AddTemplate(name, s1); err != nil {
				return fmt.Errorf("rebuild HTML template failed: %w", err)
			}
		}
	}

	b.mu.RLock()
	t, ok := b.html[name]
	b.mu.RUnlock()
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	return t.Execute(w, data)
}
