// +build dev

package data

import (
	"go/build"
	"log"
	"net/http"
	"path/filepath"
)

// Assets contains project static assets, like image and CSS-files.
var Assets http.FileSystem = http.Dir(filepath.Join(importPathToDir("github.com/boutros/ulvemelk/data"), "assets"))

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}
