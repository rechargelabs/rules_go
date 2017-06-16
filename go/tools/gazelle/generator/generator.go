/* Copyright 2016 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package generator provides core functionality of
// BUILD file generation in gazelle.
package generator

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/rechargelabs/rules_go/go/tools/gazelle/packages"
	"github.com/rechargelabs/rules_go/go/tools/gazelle/rules"
)

const (
	// goRulesBzl is the label of the Skylark file which provides Go rules
	goRulesBzl = "@io_bazel_rules_go//go:def.bzl"
)

// Generator generates BUILD files for a Go repository.
type Generator struct {
	repoRoot      string
	goPrefix      string
	buildFileName string
	buildTags     map[string]bool
	platforms     packages.PlatformConstraints
	g             rules.Generator
}

// New returns a new Generator which is responsible for a Go repository.
//
// "repoRoot" is a path to the root directory of the repository.
// "goPrefix" is the go_prefix corresponding to the repository root directory.
// See also https://github.com/rechargelabs/rules_go#go_prefix.
// "buildFileName" is the name of the BUILD file (BUILD or BUILD.bazel).
// "buildTags" is set of build tags that are true on all platforms. Some
// additional tags will be added to this. May be nil.
// "external" is how external packages should be resolved.
func New(repoRoot, goPrefix, buildFileName string, buildTags map[string]bool, external rules.ExternalResolver) (*Generator, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}

	if buildTags == nil {
		buildTags = make(map[string]bool)
	}
	platforms := packages.DefaultPlatformConstraints
	packages.PreprocessTags(buildTags, platforms)

	return &Generator{
		repoRoot:      repoRoot,
		goPrefix:      goPrefix,
		buildFileName: buildFileName,
		buildTags:     buildTags,
		platforms:     platforms,
		g:             rules.NewGenerator(repoRoot, goPrefix, external),
	}, nil
}

// Generate generates a BUILD file for each Go package found under
// the given directory.
// The directory must be the repository root directory the caller
// passed to New, or its subdirectory.
// Errors will be logged. BUILD files may or may not be returned for directories
// that have errors, depending on the severity of the error.
func (g *Generator) Generate(dir string) []*bzl.File {
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Print(err)
		return nil
	}
	if !isDescendingDir(dir, g.repoRoot) {
		log.Printf("dir %s is not under the repository root %s", dir, g.repoRoot)
		return nil
	}

	var files []*bzl.File
	packages.Walk(g.buildTags, g.platforms, g.repoRoot, g.goPrefix, dir, func(pkg *packages.Package) {
		rel, err := filepath.Rel(g.repoRoot, pkg.Dir)
		if err != nil {
			log.Print(err)
			return
		}
		if rel == "." {
			rel = ""
		}
		if len(files) == 0 && rel != "" {
			// "dir" was not a buildable Go package but still need a BUILD file
			// for go_prefix.
			files = append(files, g.emptyToplevel())
		}

		files = append(files, g.generateOne(rel, pkg))
	})
	return files
}

func (g *Generator) emptyToplevel() *bzl.File {
	return &bzl.File{
		Path: g.buildFileName,
		Stmt: []bzl.Expr{
			loadExpr("go_prefix"),
			&bzl.CallExpr{
				X: &bzl.LiteralExpr{Token: "go_prefix"},
				List: []bzl.Expr{
					&bzl.StringExpr{Value: g.goPrefix},
				},
			},
		},
	}
}

func (g *Generator) generateOne(rel string, pkg *packages.Package) *bzl.File {
	rs := g.g.Generate(filepath.ToSlash(rel), pkg)
	file := &bzl.File{Path: filepath.Join(rel, g.buildFileName)}
	for _, r := range rs {
		file.Stmt = append(file.Stmt, r.Call)
	}
	if load := g.generateLoad(file); load != nil {
		file.Stmt = append([]bzl.Expr{load}, file.Stmt...)
	}
	return file
}

func (g *Generator) generateLoad(f *bzl.File) bzl.Expr {
	var list []string
	for _, kind := range []string{
		"go_prefix",
		"go_library",
		"go_binary",
		"go_test",
		"cgo_library",
	} {
		if len(f.Rules(kind)) > 0 {
			list = append(list, kind)
		}
	}
	if len(list) == 0 {
		return nil
	}
	return loadExpr(list...)
}

func loadExpr(rules ...string) *bzl.CallExpr {
	sort.Strings(rules)

	list := []bzl.Expr{
		&bzl.StringExpr{Value: goRulesBzl},
	}
	for _, r := range rules {
		list = append(list, &bzl.StringExpr{Value: r})
	}

	return &bzl.CallExpr{
		X:            &bzl.LiteralExpr{Token: "load"},
		List:         list,
		ForceCompact: true,
	}
}

func isDescendingDir(dir, root string) bool {
	if dir == root {
		return true
	}
	return strings.HasPrefix(dir, fmt.Sprintf("%s%c", root, filepath.Separator))
}
