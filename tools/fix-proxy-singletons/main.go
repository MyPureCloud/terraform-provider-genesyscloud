// fix-proxy-singletons rewrites get*Proxy functions to use provider.SingletonOrFresh.
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
	proxyGetterPattern = regexp.MustCompile(`(?ms)^func ((?:get|Get)\w+Proxy)\((\w+) \*platformclientv2\.Configuration\) \*(\w+) \{\n\tif internalProxy == nil \{\n\t\tinternalProxy = (new\w+\(\w+\))\n\t\}\n\n?\treturn internalProxy\n\}`)
	providerImport     = `"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"`
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
		if !bytes.Contains(content, []byte("if internalProxy == nil")) {
			return nil
		}

		updated := proxyGetterPattern.ReplaceAllStringFunc(string(content), func(match string) string {
			submatches := proxyGetterPattern.FindStringSubmatch(match)
			if len(submatches) != 5 {
				return match
			}
			getterName := submatches[1]
			configParam := submatches[2]
			returnType := submatches[3]
			factoryName := strings.TrimSuffix(submatches[4], "("+configParam+")")
			return fmt.Sprintf(
				"func %s(%s *platformclientv2.Configuration) *%s {\n\treturn provider.SingletonOrFresh(&internalProxy, %s, %s)\n}",
				getterName,
				configParam,
				returnType,
				configParam,
				factoryName,
			)
		})

		if updated == string(content) {
			fmt.Fprintf(os.Stderr, "warning: no rewrite applied for %s\n", path)
			return nil
		}

		if !strings.Contains(updated, providerImport) {
			updated = addProviderImport(updated)
		}

		updated = updateProxyComment(updated)

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

func addProviderImport(source string) string {
	const importNeedle = `"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"`
	idx := strings.Index(source, importNeedle)
	if idx < 0 {
		return source
	}
	insertAt := idx + len(importNeedle)
	return source[:insertAt] + "\n\t" + providerImport + source[insertAt:]
}

func updateProxyComment(source string) string {
	old := "\t3.  A get private constructor function that the classes in the package can be used to retrieve\n\t    the proxy.  This proxy should check to see if the package level proxy instance is nil and\n\t    should initialize it, otherwise it should return the instance"
	newComment := "\t3.  A get private constructor function that the classes in the package can be used to retrieve\n\t    the proxy. When MRMO standalone mode is active, each call returns a fresh proxy bound to\n\t    the supplied client config; otherwise the legacy package-level singleton is used."
	if strings.Contains(source, old) {
		return strings.Replace(source, old, newComment, 1)
	}
	return source
}
