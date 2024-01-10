package thea

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

type Config struct {
	FS fs.FS
}

func New(cfg *Config) (*ViewManager, error) {
	vm := ViewManager{
		cfg:           cfg,
		templateCache: make(map[string]*template.Template),
		funcMap:       nil,
	}
	err := vm.parse()
	if err != nil {
		err = fmt.Errorf("unable to parse templates: %w", err)
		return nil, err
	}
	return &vm, nil
}

type ViewManager struct {
	templateCache map[string]*template.Template
	cfg           *Config
	funcMap       template.FuncMap
}

func (vm ViewManager) Render(w http.ResponseWriter, view string, data interface{}) {
	vm.templateCache[view].Execute(w, data)
}

func (vm ViewManager) parse() error {
	layouts, err := fs.Glob(vm.cfg.FS, "layouts/*.layout.html")
	if err != nil {
		err = fmt.Errorf("unable to fs.Glob the layouts/*.layout.html pattern: %w", err)
		return err
	}
	partials, err := fs.Glob(vm.cfg.FS, "partials/*.partial.html")
	if err != nil {
		err = fmt.Errorf("unable to fs.Glob the partials/*.partial.html pattern: %w", err)
		return err
	}
	otherFiles := append(layouts, partials...)
	pages, err := fs.Glob(vm.cfg.FS, "pages/*.page.html")
	if err != nil {
		err = fmt.Errorf("unable to fs.Glob the pages/*.page.html pattern: %w", err)
		return err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(vm.funcMap).ParseFS(vm.cfg.FS, page)
		if err != nil {
			err = fmt.Errorf("unable to parse the page: %w", err)
			return err
		}
		ts, err = ts.ParseFS(vm.cfg.FS, otherFiles...)
		if err != nil {
			err = fmt.Errorf("unable to parse the layouts and paritals: %w", err)
			return err
		}
		vm.templateCache[name] = ts
	}
	return nil
}
