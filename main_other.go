//go:build !windows

// The desktop shell is Windows-only (WebView2). On other platforms run the
// headless engine instead: go run ./cmd/netwatchd
package main

import "fmt"

func main() {
	fmt.Println("NETWATCH desktop targets Windows. For development run: go run ./cmd/netwatchd")
}
