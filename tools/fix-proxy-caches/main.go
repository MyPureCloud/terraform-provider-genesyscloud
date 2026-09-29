// fix-proxy-caches wraps proxy struct cache assignments with rc.CacheForProxy.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	cacheAssignPattern = regexp.MustCompile(`(?m)^(\t+)(\w+Cache):\s+(\w+Cache),$`)
	providerImport       = `"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"`
)

func main() {
	root := "genesyscloud"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	var changed int
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_proxy.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(content, []byte("rc.CacheForProxy(")) {
			return nil
		}
		if !cacheAssignPattern.Match(content) {
			return nil
		}

		updated := cacheAssignPattern.ReplaceAllStringFunc(string(content), func(match string) string {
			submatches := cacheAssignPattern.FindStringSubmatch(match)
			if len(submatches) != 4 {
				return match
			}
			indent := submatches[1]
			fieldName := submatches[2]
			cacheName := submatches[3]
			if fieldName == "createdClientCache" {
				return match
			}
			return fmt.Sprintf("%s%s: rc.CacheForProxy(%s),", indent, fieldName, cacheName)
		})

		if updated == string(content) {
			return nil
		}

		if !strings.Contains(updated, providerImport) {
			updated = addProviderImport(updated)
		}

		formatted, err := format.Source([]byte(updated))
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}

		if err := os.WriteFile(path, formatted, 0o644); err != nil {
			return err
		}
		changed++
		fmt.Println(path)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("updated %d proxy files\n", changed)
}

func addProviderImport(content string) string {
	const importBlock = "import ("
	idx := strings.Index(content, importBlock)
	if idx == -1 {
		return content
	}

	insertAt := idx + len(importBlock)
	return content[:insertAt] + "\n\t" + providerImport + content[insertAt:]
}
