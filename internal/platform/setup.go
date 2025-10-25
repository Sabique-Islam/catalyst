package platform

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// SetupPackageManager ensures the package manager and required tools are available
func SetupPackageManager(pkgManager string) error {
	switch pkgManager {
	case "apt":
		return setupApt()
	case "dnf":
		return setupDnf()
	case "pacman":
		return setupPacman()
	case "brew":
		return setupBrew()
	case "vcpkg":
		return setupVcpkg()
	case "choco":
		return setupChoco()
	default:
		return fmt.Errorf("unsupported package manager: %s", pkgManager)
	}
}

// setupApt ensures apt and apt-file are available
func setupApt() error {
	// Check if apt-file is available for better header searching
	if _, err := exec.LookPath("apt-file"); err != nil {
		fmt.Println("Note: apt-file not found. Install it for better header file resolution:")
		fmt.Println("  sudo apt update && sudo apt install apt-file")
		fmt.Println("  sudo apt-file update")
		return nil // Not a critical error
	}

	// Check if apt-file database is up to date
	output, err := exec.Command("apt-file", "search", "stdio.h").Output()
	if err != nil || len(output) == 0 {
		fmt.Println("Note: apt-file database may be outdated. Update it with:")
		fmt.Println("  sudo apt-file update")
	}

	return nil
}

// setupDnf ensures dnf is properly configured
func setupDnf() error {
	// Check if dnf is available
	if _, err := exec.LookPath("dnf"); err != nil {
		return fmt.Errorf("dnf not found - install with: sudo yum install dnf")
	}
	return nil
}

// setupPacman ensures pacman is properly configured
func setupPacman() error {
	// Check if pacman is available
	if _, err := exec.LookPath("pacman"); err != nil {
		return fmt.Errorf("pacman not found")
	}
	return nil
}

// setupBrew ensures Homebrew is properly installed and updated
func setupBrew() error {
	// Check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Install from: https://brew.sh/")
	}

	// Check if brew needs updating (optional)
	output, err := exec.Command("brew", "--version").Output()
	if err != nil {
		return fmt.Errorf("brew command failed: %v", err)
	}

	if strings.Contains(string(output), "Homebrew") {
		fmt.Println("Homebrew detected. Consider running 'brew update' for latest package info.")
	}

	return nil
}

// setupVcpkg checks if vcpkg is available and properly configured
func setupVcpkg() error {
	if _, err := exec.LookPath("vcpkg"); err != nil {
		return fmt.Errorf("vcpkg not found. Install from: https://github.com/Microsoft/vcpkg")
	}

	// Check if vcpkg is integrated
	output, err := exec.Command("vcpkg", "list").Output()
	if err != nil {
		fmt.Println("Note: vcpkg may need integration. Run: vcpkg integrate install")
	}

	if len(output) == 0 {
		fmt.Println("Note: No vcpkg packages installed yet.")
	}

	return nil
}

// setupChoco checks if Chocolatey is available
func setupChoco() error {
	if _, err := exec.LookPath("choco"); err != nil {
		return fmt.Errorf("Chocolatey not found. Install from: https://chocolatey.org/install")
	}
	return nil
}

// GetPackageManagerSetupAdvice returns setup advice for the current platform
func GetPackageManagerSetupAdvice() string {
	osName := runtime.GOOS
	
	switch osName {
	case "linux":
		return `
Package Manager Setup (Linux):
  Ubuntu/Debian: apt is pre-installed
    • Install apt-file: sudo apt install apt-file && sudo apt-file update
  Fedora/RHEL: dnf is pre-installed  
  Arch Linux: pacman is pre-installed
`
	case "darwin":
		return `
Package Manager Setup (macOS):
  Install Homebrew: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  Update packages: brew update
`
	case "windows":
		return `
Package Manager Setup (Windows):
  vcpkg: 
    • Clone: git clone https://github.com/Microsoft/vcpkg.git
    • Build: .\vcpkg\bootstrap-vcpkg.bat
    • Integrate: .\vcpkg\vcpkg integrate install
  
  Chocolatey:
    • PowerShell (Admin): Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
`
	default:
		return "Package manager setup varies by platform. Check your distribution's documentation."
	}
}