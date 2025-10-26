package platform

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

// IsPackageInstalled checks if a package is installed using the specified package manager
// Returns true if the package is installed, false otherwise
func IsPackageInstalled(pkgName string, pkgManager string) bool {
	switch pkgManager {
	case "apt":
		return isInstalledApt(pkgName)
	case "dnf":
		return isInstalledDnf(pkgName)
	case "pacman":
		return isInstalledPacman(pkgName)
	case "brew":
		return isInstalledBrew(pkgName)
	case "vcpkg":
		return isInstalledVcpkg(pkgName)
	case "choco":
		return isInstalledChoco(pkgName)
	case "winget":
		return isInstalledWinget(pkgName)
	default:
		return false
	}
}

// isInstalledApt checks if a package is installed using apt (Debian/Ubuntu)
// Uses: dpkg -s <pkgName>
func isInstalledApt(pkgName string) bool {
	cmd := exec.Command("dpkg", "-s", pkgName)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// isInstalledDnf checks if a package is installed using dnf (Fedora/RHEL)
// Uses: dnf list installed <pkgName>
func isInstalledDnf(pkgName string) bool {
	cmd := exec.Command("dnf", "list", "installed", pkgName)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// isInstalledPacman checks if a package is installed using pacman (Arch Linux)
// Uses: pacman -Q <pkgName>
func isInstalledPacman(pkgName string) bool {
	cmd := exec.Command("pacman", "-Q", pkgName)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

// isInstalledBrew checks if a package is installed using brew (darwin Homebrew)
// Uses: brew list --formula | grep -q <pkgName>
func isInstalledBrew(pkgName string) bool {
	// First get the list of installed formulas
	listCmd := exec.Command("brew", "list", "--formula")
	var out bytes.Buffer
	listCmd.Stdout = &out
	listCmd.Stderr = io.Discard

	if err := listCmd.Run(); err != nil {
		return false
	}

	// Check if the package name is in the output
	installedPackages := strings.Split(out.String(), "\n")
	for _, pkg := range installedPackages {
		if strings.TrimSpace(pkg) == pkgName {
			return true
		}
	}
	return false
}

// isInstalledVcpkg checks if a package is installed using vcpkg (Windows)
// Uses: vcpkg list <pkgName>
func isInstalledVcpkg(pkgName string) bool {
	cmd := exec.Command("vcpkg", "list", pkgName)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return false
	}

	// Check if the output contains the package name
	return strings.Contains(out.String(), pkgName)
}

// isInstalledChoco checks if a package is installed using choco (Windows Chocolatey)
// Uses: choco list --local-only <pkgName>
func isInstalledChoco(pkgName string) bool {
	cmd := exec.Command("choco", "list", "--local-only", pkgName)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Check if the package name appears in the output
	output := strings.ToLower(out.String())
	pkgLower := strings.ToLower(pkgName)
	
	// Look for the package name in the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Chocolatey lists packages in format: "packagename version"
		if strings.HasPrefix(line, pkgLower+" ") || strings.HasPrefix(line, pkgLower+"\t") || line == pkgLower {
			return true
		}
	}
	return false
}

// isInstalledWinget checks if a package is installed using winget (Windows Package Manager)
// Uses: winget list --id <pkgName>
func isInstalledWinget(pkgName string) bool {
	cmd := exec.Command("winget", "list", "--id", pkgName)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	
	if err := cmd.Run(); err != nil {
		return false
	}
	
	// Check if the package ID appears in the output
	output := out.String()
	
	// winget list returns the package if it's installed
	// The output contains the package ID if installed
	return strings.Contains(output, pkgName)
}
