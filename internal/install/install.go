package install

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	config "github.com/Sabique-Islam/catalyst/internal/config"
)

// detectLinuxPackageManager tries to find a supported package manager on Linux.
func detectLinuxPackageManager() (string, error) {
	candidates := []string{"apt-get", "dnf", "yum", "pacman", "zypper"}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c, nil
		}
	}
	return "", errors.New("no supported linux package manager found (supported: apt-get, dnf, yum, pacman, zypper)")
}

// Install installs the given dependencies (already OS-specific)
func Install(dependencies []string) error {
	if len(dependencies) == 0 {
		fmt.Println("No dependencies to install.")
		return nil
	}

	osType := runtime.GOOS

	switch osType {
	case "linux":
		pkgMgr, err := detectLinuxPackageManager()
		if err != nil {
			return err
		}

		var args []string
		switch pkgMgr {
		case "apt-get":
			args = append([]string{"install", "-y"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("sudo", append([]string{"apt-get"}, args...)...)
		case "dnf", "yum":
			args = append([]string{"install", "-y"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("sudo", append([]string{pkgMgr}, args...)...)
		case "pacman":
			args = append([]string{"-S", "--noconfirm"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("sudo", append([]string{"pacman"}, args...)...)
		case "zypper":
			args = append([]string{"install", "-y"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("sudo", append([]string{"zypper"}, args...)...)
		}

		if err != nil {
			return fmt.Errorf("failed installing with %s: %w", pkgMgr, err)
		}

	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			return errors.New("homebrew not found - install it from https://brew.sh/")
		}
		fmt.Println("Using package manager: brew")
		args := append([]string{"install"}, dependencies...)
		if err := runCommand("brew", args...); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}

	case "windows":
		pkgMgr := getPackageManager()
		if pkgMgr == "unknown" {
			return errors.New("no Windows package manager found. Please install winget, chocolatey (https://chocolatey.org/install), or scoop (https://scoop.sh)")
		}

		var args []string
		var err error
		switch pkgMgr {
		case "choco":
			args = append([]string{"install", "-y"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("choco", args...)
		case "winget":
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			for _, dep := range dependencies {
				winPkg := mapToWindowsPackage(dep, "winget")
				err = runCommand("winget", "install", "--id", winPkg, "--accept-package-agreements", "--accept-source-agreements")
				if err != nil {
					break
				}
			}
		case "scoop":
			args = append([]string{"install"}, dependencies...)
			fmt.Printf("Using package manager: %s\n", pkgMgr)
			err = runCommand("scoop", args...)
		default:
			return fmt.Errorf("unsupported Windows package manager: %s", pkgMgr)
		}

		if err != nil {
			return fmt.Errorf("failed installing with %s: %w", pkgMgr, err)
		}

	default:
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	return nil
}

// InstallDependencies loads the config, gets OS-specific dependencies, and installs them
// Also downloads external resources (files) specified in the config
func InstallDependencies() error {
	// Load catalyst.yml
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Install system dependencies
	deps := cfg.GetDependencies() // returns []string
	if len(deps) > 0 {
		fmt.Printf("Installing system dependencies for %s: %v\n", runtime.GOOS, deps)
		fmt.Println()

		if err := Install(deps); err != nil {
			return fmt.Errorf("system dependency installation failed: %w", err)
		}

		fmt.Println()
		fmt.Println("System dependencies installed successfully!")
		fmt.Println()
	} else {
		fmt.Println("No system dependencies to install for this OS.")
		fmt.Println()
	}

	// Install external resources (download files)
	if err := InstallResources(cfg); err != nil {
		return fmt.Errorf("external resource installation failed: %w", err)
	}

	return nil
}

// InstallExternalResourcesOnly downloads only external resources without installing system dependencies
func InstallExternalResourcesOnly() error {
	// Load catalyst.yml
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Install only external resources
	return InstallResources(cfg)
}

// InstallSystemDependenciesOnly installs only system dependencies without downloading external resources
func InstallSystemDependenciesOnly() error {
	// Load catalyst.yml
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Install only system dependencies
	deps := cfg.GetDependencies()
	if len(deps) == 0 {
		fmt.Println("No system dependencies to install for this OS.")
		return nil
	}

	fmt.Printf("Installing system dependencies for %s: %v\n", runtime.GOOS, deps)
	fmt.Println()

	if err := Install(deps); err != nil {
		return fmt.Errorf("system dependency installation failed: %w", err)
	}

	fmt.Println()
	fmt.Println("System dependencies installed successfully!")
	return nil
}

// InstallDependenciesAndGetLinkerFlags installs dependencies and returns linker flags for them
func InstallDependenciesAndGetLinkerFlags() ([]string, error) {
	// Load catalyst.yml
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get dependencies for current OS only
	deps := cfg.GetDependencies() // returns []string
	if len(deps) == 0 {
		fmt.Println("No dependencies to install for this OS.")
		return []string{}, nil
	}

	fmt.Printf("Installing dependencies for %s: %v\n", runtime.GOOS, deps)

	// Install each package and collect linker flags
	libFlags := []string{}
	for _, pkg := range deps {
		if err := installPackage(pkg); err != nil {
			return nil, fmt.Errorf("failed to install package %s: %w", pkg, err)
		}
		// Assuming link name is same as package (for libraries)
		if isLibraryPackage(pkg) {
			libName := extractLibraryName(pkg)
			if libName != "" {
				libFlags = append(libFlags, "-l"+libName)
			}
		}
	}

	fmt.Printf("Dependencies installed with linker flags: %v\n", libFlags)
	return libFlags, nil
}

func getPackageManager() string {
	// Check for different package managers based on OS
	osType := runtime.GOOS

	switch osType {
	case "windows":
		// Priority order for Windows: winget > choco > scoop
		if _, err := exec.LookPath("winget"); err == nil {
			return "winget"
		}
		if _, err := exec.LookPath("choco"); err == nil {
			return "choco"
		}
		if _, err := exec.LookPath("scoop"); err == nil {
			return "scoop"
		}
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
	case "linux":
		// Check for different Linux package managers
		if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman"
		}
		if _, err := exec.LookPath("apt-get"); err == nil {
			return "apt"
		}
		if _, err := exec.LookPath("dnf"); err == nil {
			return "dnf"
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return "yum"
		}
		if _, err := exec.LookPath("zypper"); err == nil {
			return "zypper"
		}
	}

	return "unknown"
}

// installPackage installs a single package
func installPackage(pkg string) error {
	var cmd *exec.Cmd

	// Skip system libraries that don't need installation
	systemLibs := []string{"m", "pthread", "dl", "rt"}
	windowsSystemLibs := []string{"ws2_32.lib", "user32.lib", "kernel32.lib", "advapi32.lib", "shell32.lib", "ole32.lib", "oleaut32.lib", "uuid.lib", "winmm.lib", "gdi32.lib", "comctl32.lib", "comdlg32.lib", "winspool.lib"}

	osType := runtime.GOOS

	// Check Unix/Linux system libraries
	for _, sysLib := range systemLibs {
		if pkg == sysLib {
			fmt.Printf("Skipping installation of system library: %s\n", pkg)
			return nil
		}
	}

	// Check Windows system libraries
	if osType == "windows" {
		for _, sysLib := range windowsSystemLibs {
			if pkg == sysLib || strings.EqualFold(pkg, sysLib) {
				fmt.Printf("Skipping installation of Windows system library: %s\n", pkg)
				return nil
			}
		}
	}

	pkgManager := getPackageManager()

	switch pkgManager {
	case "pacman":
		// Arch Linux package names
		archPkg := mapToArchPackage(pkg)
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", archPkg)
	case "apt":
		cmd = exec.Command("sudo", "apt-get", "install", "-y", pkg)
	case "brew":
		cmd = exec.Command("brew", "install", pkg)
	case "yum":
		cmd = exec.Command("sudo", "yum", "install", "-y", pkg)
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", pkg)
	case "zypper":
		cmd = exec.Command("sudo", "zypper", "install", "-y", pkg)
	case "choco":
		// Chocolatey for Windows
		winPkg := mapToWindowsPackage(pkg, "choco")
		cmd = exec.Command("choco", "install", winPkg, "-y")
	case "winget":
		// Windows Package Manager
		winPkg := mapToWindowsPackage(pkg, "winget")
		cmd = exec.Command("winget", "install", "--id", winPkg, "--accept-package-agreements", "--accept-source-agreements")
	case "scoop":
		// Scoop for Windows
		winPkg := mapToWindowsPackage(pkg, "scoop")
		cmd = exec.Command("scoop", "install", winPkg)
	default:
		osType := runtime.GOOS
		switch osType {
		case "windows":
			return fmt.Errorf("no Windows package manager found. Please install one of: winget (Windows Package Manager), chocolatey (https://chocolatey.org/install), or scoop (https://scoop.sh)")
		case "darwin":
			return fmt.Errorf("homebrew not found. Please install it from https://brew.sh/")
		case "linux":
			return fmt.Errorf("no supported Linux package manager found. Supported: apt-get, dnf, yum, pacman, zypper")
		default:
			return fmt.Errorf("unsupported operating system: %s", osType)
		}
	}

	fmt.Printf("Installing %s with %s...\n", pkg, pkgManager)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed installing with %s: %s\nOutput: %s", pkgManager, err, string(output))
	}
	return nil
}

func mapToArchPackage(pkg string) string {
	// Map common package names to Arch equivalents
	archMap := map[string]string{
		"gcc":                  "gcc",
		"make":                 "make",
		"build-essential":      "base-devel",
		"libcurl4-openssl-dev": "curl",
		"libjansson-dev":       "jansson",
		"libssl-dev":           "openssl",
		"pkg-config":           "pkgconf",
	}

	if archPkg, exists := archMap[pkg]; exists {
		return archPkg
	}
	return pkg // Return original if no mapping found
}

func mapToWindowsPackage(pkg string, pkgManager string) string {
	// Map common package names to Windows equivalents based on package manager
	var pkgMap map[string]string

	switch pkgManager {
	case "choco":
		pkgMap = map[string]string{
			"gcc":                  "mingw",
			"make":                 "make",
			"build-essential":      "mingw",
			"curl":                 "curl",
			"libcurl4-openssl-dev": "curl",
			"libssl-dev":           "openssl",
			"openssl":              "openssl",
			"git":                  "git",
			"cmake":                "cmake",
			"python":               "python",
			"nodejs":               "nodejs",
			"sqlite":               "sqlite",
			"sqlite3":              "sqlite",
			"zlib":                 "zlib",
			"pkg-config":           "pkgconfiglite",
		}
	case "winget":
		pkgMap = map[string]string{
			"gcc":                  "MSYS2.MSYS2",
			"make":                 "GnuWin32.Make",
			"build-essential":      "MSYS2.MSYS2",
			"curl":                 "cURL.cURL",
			"libcurl4-openssl-dev": "cURL.cURL",
			"git":                  "Git.Git",
			"cmake":                "Kitware.CMake",
			"python":               "Python.Python.3.11",
			"nodejs":               "OpenJS.NodeJS",
			"sqlite":               "SQLite.SQLite",
			"sqlite3":              "SQLite.SQLite",
		}
	case "scoop":
		pkgMap = map[string]string{
			"gcc":     "gcc",
			"make":    "make",
			"curl":    "curl",
			"git":     "git",
			"cmake":   "cmake",
			"python":  "python",
			"nodejs":  "nodejs",
			"sqlite":  "sqlite3",
			"sqlite3": "sqlite3",
		}
	default:
		return pkg
	}

	if winPkg, exists := pkgMap[pkg]; exists {
		return winPkg
	}
	return pkg // Return original if no mapping found
}

// isLibraryPackage checks if a package is a library that needs linking
func isLibraryPackage(pkg string) bool {
	// List of known library packages that need linking
	knownLibraries := []string{
		"curl", "jansson", "ssl", "crypto", "sqlite", "sqlite3", "pthread", "m", "z", "dl", "rt",
		"openssl", "zlib", "pcre", "glib", "gtk", "qt", "boost",
	}

	pkgLower := strings.ToLower(pkg)

	// Check direct matches
	for _, lib := range knownLibraries {
		if pkgLower == lib {
			return true
		}
	}

	// Check common library naming patterns
	libraryPatterns := []string{
		"lib", "-dev", ".lib", "-devel",
	}

	for _, pattern := range libraryPatterns {
		if strings.Contains(pkgLower, pattern) {
			return true
		}
	}

	return false
}

// extractLibraryName extracts the library name for linking from package name
func extractLibraryName(pkg string) string {
	// Handle common package name to library name mappings
	libMappings := map[string]string{
		"curl":                 "curl",
		"jansson":              "jansson",
		"sqlite":               "sqlite3",
		"libssl-dev":           "ssl",
		"libcrypto-dev":        "crypto",
		"libcurl4-openssl-dev": "curl",
		"libjansson-dev":       "jansson",
		"libsqlite3-dev":       "sqlite3",
		"sqlite3":              "sqlite3",
		"pthread":              "pthread",
		"m":                    "m",
		"ws2_32.lib":           "ws2_32",
		"user32.lib":           "user32",
		"kernel32.lib":         "kernel32",
		"openssl":              "ssl",
		"zlib":                 "z",
	}

	// Direct mapping
	if libName, exists := libMappings[pkg]; exists {
		return libName
	}

	// Extract from lib*-dev pattern
	if strings.HasPrefix(pkg, "lib") && strings.HasSuffix(pkg, "-dev") {
		return pkg[3 : len(pkg)-4] // Remove "lib" prefix and "-dev" suffix
	}

	// Extract from *.lib pattern
	if strings.HasSuffix(pkg, ".lib") {
		return pkg[:len(pkg)-4] // Remove ".lib" suffix
	}

	// For simple library names, use as-is
	if isSimpleLibrary(pkg) {
		return pkg
	}

	return ""
}

// isSimpleLibrary checks if this is a simple library name that can be used directly
func isSimpleLibrary(pkg string) bool {
	simpleLibs := []string{"pthread", "m", "z", "dl", "ssl", "crypto", "curl", "jansson", "sqlite3"}
	for _, lib := range simpleLibs {
		if pkg == lib {
			return true
		}
	}
	return false
}

// runCommand executes a command with arguments
func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// DownloadResource downloads a file from a URL to a local path
func DownloadResource(url, localPath string) error {
	// Normalize path separators for the current OS
	normalizedPath := filepath.Clean(localPath)

	// Create the directory if it doesn't exist
	dir := filepath.Dir(normalizedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(normalizedPath); err == nil {
		fmt.Printf("Resource already exists: %s (skipping download)\n", normalizedPath)
		return nil
	}

	fmt.Printf("Downloading %s -> %s\n", url, normalizedPath)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the HTTP request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: HTTP %d %s", url, resp.StatusCode, resp.Status)
	}

	// Create the output file
	file, err := os.Create(normalizedPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", normalizedPath, err)
	}
	defer file.Close()

	// Copy the response body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		// Clean up partial file on error
		os.Remove(normalizedPath)
		return fmt.Errorf("failed to write file %s: %w", normalizedPath, err)
	}

	fmt.Printf("Successfully downloaded: %s\n", normalizedPath)
	return nil
}

// InstallResources downloads external resources defined in the config
func InstallResources(cfg *config.Config) error {
	osType := runtime.GOOS

	// Get resources using the config method
	resources := cfg.GetResources()

	if len(resources) == 0 {
		fmt.Println("No external resources to download.")
		return nil
	}

	fmt.Printf("Downloading %d external resources for %s...\n", len(resources), osType)
	fmt.Println()

	// Download each resource
	for i, resource := range resources {
		fmt.Printf("[%d/%d] ", i+1, len(resources))

		if resource.URL == "" {
			fmt.Printf("Skipping resource with empty URL\n")
			continue
		}

		if resource.Path == "" {
			fmt.Printf("Skipping resource %s with empty path\n", resource.URL)
			continue
		}

		if err := DownloadResource(resource.URL, resource.Path); err != nil {
			return fmt.Errorf("failed to download resource %s: %w", resource.URL, err)
		}
	}

	fmt.Println()
	fmt.Println("External resources downloaded successfully!")
	return nil
}
