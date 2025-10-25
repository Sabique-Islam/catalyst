package platform

import (
	"fmt"
	"os/exec"
	"runtime"
)

// DetectOS detects the host operating system and returns a normalized string
// Returns one of: "linux", "macos", or "windows"
func DetectOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

// DetectPackageManager detects the available package manager for the given OS
// It checks for package managers in order of preference and returns the first one found
func DetectPackageManager(os string) (string, error) {
	switch os {
	case "linux":
		// Check for apt (Debian/Ubuntu)
		if _, err := exec.LookPath("apt"); err == nil {
			return "apt", nil
		}
		// Check for dnf (Fedora/RHEL)
		if _, err := exec.LookPath("dnf"); err == nil {
			return "dnf", nil
		}
		// Check for pacman (Arch Linux)
		if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman", nil
		}
		return "", fmt.Errorf("no supported package manager found on Linux (checked: apt, dnf, pacman)")

	case "macos":
		// Check for brew (Homebrew)
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew", nil
		}
		return "", fmt.Errorf("homebrew not found on macOS")

	case "windows":
		// Check for vcpkg
		if _, err := exec.LookPath("vcpkg"); err == nil {
			return "vcpkg", nil
		}
		// Check for choco (Chocolatey)
		if _, err := exec.LookPath("choco"); err == nil {
			return "choco", nil
		}
		return "", fmt.Errorf("no supported package manager found on Windows (checked: vcpkg, choco)")

	default:
		return "", fmt.Errorf("unsupported operating system: %s", os)
	}
}
