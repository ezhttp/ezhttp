package version

import "fmt"

// Build Information
const BuildVersion string = "0.0.3"

var BuildDate string
var BuildGoVersion string
var BuildGitHash string

// PrintVersion prints the full version information
func PrintVersion() {
	fmt.Printf("v%s\nDate: %s\nGo Version: %s\nGit Hash: %s\n", BuildVersion, BuildDate, BuildGoVersion, BuildGitHash)
}

// PrintVersionShort prints the short version information
func PrintVersionShort() {
	fmt.Printf("### EZhttp v%s - Date: %s - Go Version: %s\n", BuildVersion, BuildDate, BuildGoVersion)
}
