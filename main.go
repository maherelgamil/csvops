/*
Copyright © 2025 Maher El Gamil
*/
package main

import "github.com/maherelgamil/csvops/cmd"

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
