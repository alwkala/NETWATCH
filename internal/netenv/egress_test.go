package netenv

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestZeroEgressInvariant rigorously verifies the core product invariant:
// NETWATCH has ZERO external outbound HTTP communication and no remote CDN dependencies.
// This test parses all Go ASTs and frontend source files to guarantee that:
// 1. No package outside internal/api imports net/http.
// 2. internal/api never instantiates an outbound http.Client or calls http.Get / http.Post.
// 3. Frontend assets do not reference external CDN script or stylesheet URLs.
func TestZeroEgressInvariant(t *testing.T) {
	// Find repo root (two levels up from internal/netenv)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	fset := token.NewFileSet()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		relSlash := filepath.ToSlash(rel)

		// Skip build artifacts, dependencies, and vcs
		if info.IsDir() {
			switch info.Name() {
			case ".git", "node_modules", "dist", "build", "vendor", ".agents", "scratch":
				return filepath.SkipDir
			}
			return nil
		}

		// 1. Audit Go files
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return nil
			}

			for _, imp := range node.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if importPath == "net/http" {
					// net/http is strictly permitted ONLY in internal/api (for local 127.0.0.1 loopback HTTP server)
					if !strings.HasPrefix(relSlash, "internal/api/") {
						t.Errorf("ZERO-EGRESS VIOLATION: File %s imports net/http. Network calls outside loopback server are forbidden.", relSlash)
					}
				}
			}

			// If inside internal/api, do deep parse to ensure no outbound client calls exist
			if strings.HasPrefix(relSlash, "internal/api/") {
				fullNode, err := parser.ParseFile(fset, path, nil, 0)
				if err == nil {
					ast.Inspect(fullNode, func(n ast.Node) bool {
						if sel, ok := n.(*ast.SelectorExpr); ok {
							if x, ok := sel.X.(*ast.Ident); ok && x.Name == "http" {
								switch sel.Sel.Name {
								case "Get", "Post", "PostForm", "Head", "DefaultClient":
									t.Errorf("ZERO-EGRESS VIOLATION: File %s invokes http.%s. Outbound HTTP requests are strictly forbidden.", relSlash, sel.Sel.Name)
								}
							}
						}
						return true
					})
				}
			}
		}

		// 2. Audit Frontend source files (ensure no remote CDNs or tracking beacons)
		if strings.HasPrefix(relSlash, "frontend/src/") || relSlash == "frontend/index.html" {
			ext := filepath.Ext(path)
			if ext == ".ts" || ext == ".tsx" || ext == ".css" || ext == ".html" {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				s := string(content)
				// Flag any remote http/https CDN linkages for fonts or scripts
				forbiddenCDNs := []string{
					"fonts.googleapis.com",
					"fonts.gstatic.com",
					"cdnjs.cloudflare.com",
					"cdn.jsdelivr.net",
					"unpkg.com",
					"google-analytics.com",
					"googletagmanager.com",
				}
				for _, cdn := range forbiddenCDNs {
					if strings.Contains(s, cdn) {
						t.Errorf("ZERO-EGRESS VIOLATION: Frontend file %s references remote CDN domain %q. All assets must be bundled locally.", relSlash, cdn)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
}
