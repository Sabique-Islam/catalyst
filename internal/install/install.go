package install

import (
	_ "embed"
	"encoding/json"
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

//go:embed windows_issues.json
var windowsIssuesJSON []byte

// WindowsIssuesDatabase represents the JSON structure
type WindowsIssuesDatabase struct {
	Version     string                         `json:"version"`
	LastUpdated string                         `json:"last_updated"`
	Description string                         `json:"description"`
	Issues      map[string]WindowsPackageIssue `json:"issues"`
}

// WindowsPackageIssue represents a known issue with a package on Windows
type WindowsPackageIssue struct {
	PackageName  string `json:"package_name"`
	DisplayName  string `json:"display_name"`
	Issue        string `json:"issue"`
	Alternative  string `json:"alternative"`
	WorkaroundURL string `json:"workaround_url"`
}

var issuesDB *WindowsIssuesDatabase

// loadWindowsIssuesDB loads the Windows issues database from embedded JSON
func loadWindowsIssuesDB() (*WindowsIssuesDatabase, error) {
	if issuesDB != nil {
		return issuesDB, nil
	}

	var db WindowsIssuesDatabase
	if err := json.Unmarshal(windowsIssuesJSON, &db); err != nil {
		return nil, fmt.Errorf("failed to parse windows_issues.json: %w", err)
	}

	issuesDB = &db
	return issuesDB, nil
}

// getWindowsPackageIssue retrieves issue information for a package (case-insensitive)
func getWindowsPackageIssue(packageName string) (*WindowsPackageIssue, bool) {
	db, err := loadWindowsIssuesDB()
	if err != nil {
		fmt.Printf("Warning: Failed to load Windows issues database: %v\n", err)
		return nil, false
	}

	pkgLower := strings.ToLower(packageName)
	for key, issue := range db.Issues {
		if strings.ToLower(key) == pkgLower {
			return &issue, true
		}
	}
	return nil, false
}

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
			fmt.Println()
			var lastErr error
			successCount := 0
			hasMSYS2 := false
			msys2Packages := []string{}

			// First pass: install base packages via winget, collect MSYS2 packages
			for _, dep := range dependencies {
				winPkg := mapToWindowsPackage(dep, "winget")

				// Check for Windows compatibility issues
				checkWindowsPackageCompatibility(dep)

				// Check if this is a package that should be installed via MSYS2 pacman
				if shouldUseMSYS2Pacman(dep) {
					msys2Packages = append(msys2Packages, dep)
					continue
				}

				fmt.Printf("Installing %s", dep)
				if winPkg != dep {
					fmt.Printf(" (package: %s)", winPkg)
				}
				fmt.Println("...")

				if winPkg == "MSYS2.MSYS2" {
					hasMSYS2 = true
				}

				err = runWingetInstall(winPkg)
				if err != nil {
					// For winget, check if it's an "already installed" or "no applicable installer" error
					if isWingetNonCriticalError(err) {
						fmt.Printf("  → Skipped: Package may already be installed or installation was interrupted\n")
						if winPkg == "MSYS2.MSYS2" {
							hasMSYS2 = true // Still mark as available for pacman use
							fmt.Printf("     MSYS2 appears to be already installed\n")
						}
						fmt.Println()
						continue // Continue with other packages
					}
					fmt.Printf("  → Failed to install %s\n\n", dep)
					lastErr = err
					// Continue trying other packages instead of stopping
					continue
				}
				fmt.Printf("  → Successfully installed %s\n\n", dep)
				successCount++
			}

			// Second pass: install development libraries via MSYS2 pacman if available
			if len(msys2Packages) > 0 {
				if hasMSYS2 || isMSYS2Installed() {
					fmt.Printf("\nInstalling development libraries via MSYS2 pacman: %v\n", msys2Packages)
					if err := installViaMSYS2Pacman(msys2Packages); err != nil {
						fmt.Printf("Warning: Failed to install some packages via MSYS2: %v\n", err)
						fmt.Printf("You may need to manually install these packages:\n")
						for _, pkg := range msys2Packages {
							msys2Pkg := mapToMSYS2Package(pkg)
							fmt.Printf("  pacman -S %s\n", msys2Pkg)
						}
					} else {
						successCount += len(msys2Packages)
					}
				} else {
					fmt.Printf("\nWarning: The following packages require MSYS2 but it's not installed: %v\n", msys2Packages)
					fmt.Printf("Please install MSYS2 from https://www.msys2.org/ and then run:\n")
					for _, pkg := range msys2Packages {
						msys2Pkg := mapToMSYS2Package(pkg)
						fmt.Printf("  pacman -S %s\n", msys2Pkg)
					}
				}
			}

			// Only return error if all packages failed and none were skipped
			if successCount == 0 && lastErr != nil {
				err = lastErr
			} else {
				err = nil
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

	// Install each package
	for _, pkg := range deps {
		if err := installPackage(pkg); err != nil {
			return nil, fmt.Errorf("failed to install package %s: %w", pkg, err)
		}
	}

	// Generate comprehensive linking flags
	libFlags := generateLinkingFlags(deps)
	if len(libFlags) > 0 {
		fmt.Printf("Adding linking flags: %s\n", strings.Join(libFlags, " "))
	}
	return libFlags, nil
}

// generateLinkingFlags generates linking flags based on detected dependencies
func generateLinkingFlags(dependencies []string) []string {
	var linkFlags []string

	// Common library mappings for linking
	linkMap := map[string]string{
		// Math library
		"math": "m",

		// Threading
		"pthread": "pthread",

		// Networking
		"curl":                 "curl",
		"libcurl":              "curl",
		"libcurl4-openssl-dev": "curl",

		// JSON libraries
		"jansson":        "jansson",
		"libjansson-dev": "jansson",
		"json-c":         "json-c",
		"cjson":          "cjson",

		// Terminal libraries
		"ncurses":        "ncurses",
		"libncurses-dev": "ncurses",
		"termcap":        "termcap",

		// Database libraries
		"sqlite":         "sqlite3",
		"sqlite3":        "sqlite3",
		"libsqlite3-dev": "sqlite3",

		// SSL/Crypto
		"openssl":    "ssl",
		"libssl-dev": "ssl",
		"ssl":        "ssl",
		"crypto":     "crypto",

		// Compression
		"zlib":       "z",
		"zlib1g-dev": "z",

		// Linear algebra
		"blas":     "blas",
		"lapack":   "lapack",
		"openblas": "openblas",

		// GLib
		"glib":     "glib-2.0",
		"glib-2.0": "glib-2.0",
	}

	// Always add math library for C projects
	linkFlags = append(linkFlags, "-lm")

	// Process dependencies and add linking flags
	for _, dep := range dependencies {
		// Normalize dependency name
		depLower := strings.ToLower(dep)

		if linkLib, found := linkMap[depLower]; found {
			linkFlag := "-l" + linkLib
			// Avoid duplicates
			isDuplicate := false
			for _, existing := range linkFlags {
				if existing == linkFlag {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				linkFlags = append(linkFlags, linkFlag)
			}
		}
	}

	return linkFlags
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
		// Check for Windows compatibility issues before installation
		checkWindowsPackageCompatibility(pkg)

		// Windows Package Manager - check if package should use MSYS2 pacman instead
		if shouldUseMSYS2Pacman(pkg) {
			if isMSYS2Installed() {
				fmt.Printf("Installing %s via MSYS2 pacman...\n", pkg)
				return installViaMSYS2Pacman([]string{pkg})
			} else {
				fmt.Printf("Warning: %s requires MSYS2 but it's not installed\n", pkg)
				fmt.Printf("Please install MSYS2 from https://www.msys2.org/ and run: pacman -S %s\n", mapToMSYS2Package(pkg))
				return nil // Don't fail, just warn
			}
		}

		// For winget packages
		winPkg := mapToWindowsPackage(pkg, "winget")
		fmt.Printf("Installing %s with %s...\n", pkg, pkgManager)
		err := runWingetInstall(winPkg)
		if err != nil {
			if isWingetNonCriticalError(err) {
				fmt.Printf("  Note: %s may already be installed or unavailable via winget\n", winPkg)
				return nil // Treat as success
			}
			return fmt.Errorf("failed installing %s with winget: %w", pkg, err)
		}
		return nil
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
			"msys2":                "MSYS2.MSYS2",
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

// WindowsPackageIssue represents known issues with packages on Windows
type WindowsPackageIssue struct {
	PackageName  string
	Issue        string
	Alternative  string
	WorkaroundURL string
}
// NOTE: Package compatibility information is now loaded from the embedded
// JSON file `windows_issues.json`. See loadWindowsIssuesDB() and
// getWindowsPackageIssue() above for the loader and access helpers.

// checkWindowsPackageCompatibility checks if a package has known Windows issues and warns the user
func checkWindowsPackageCompatibility(pkg string) {
	if runtime.GOOS != "windows" {
		return
	}
	// Use the embedded JSON database for package issues
	issue, found := getWindowsPackageIssue(pkg)
	if !found {
		// try a few normalized variants (e.g., lib*-dev -> core name)
		// strip common prefixes/suffixes
		normalized := strings.ToLower(pkg)
		normalized = strings.TrimPrefix(normalized, "lib")
		normalized = strings.TrimSuffix(normalized, "-dev")
		if normalized != strings.ToLower(pkg) {
			if issue2, ok2 := getWindowsPackageIssue(normalized); ok2 {
				issue = issue2
				found = true
			}
		}
	}

	if !found {
		return
	}

	fmt.Printf("\n⚠️  WARNING: Windows Compatibility Issue Detected\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if issue.DisplayName != "" {
		fmt.Printf("Package: %s (%s)\n", issue.PackageName, issue.DisplayName)
	} else {
		fmt.Printf("Package: %s\n", issue.PackageName)
	}
	fmt.Printf("Issue: %s\n\n", issue.Issue)
	fmt.Printf("💡 Suggestion:\n")
	fmt.Printf("   %s\n\n", issue.Alternative)
	if issue.WorkaroundURL != "" {
		fmt.Printf("📖 More Info:\n")
		fmt.Printf("   %s\n", issue.WorkaroundURL)
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}

// shouldUseMSYS2Pacman checks if a package should be installed via MSYS2 pacman instead of winget
func shouldUseMSYS2Pacman(pkg string) bool {
	// Packages that are development libraries and not available via winget
	msys2OnlyPackages := []string{
		"curl",
		"jansson",
		"sqlite3",
		"libjansson-dev",
		"libcurl4-openssl-dev",
		"libssl-dev",
		"libsqlite3-dev",
		"ncurses",
		"libncurses-dev",
	}

	pkgLower := strings.ToLower(pkg)
	for _, msys2Pkg := range msys2OnlyPackages {
		if pkgLower == msys2Pkg {
			return true
		}
	}
	return false
}

// isMSYS2Installed checks if MSYS2 is installed on the system
func isMSYS2Installed() bool {
	// Check common MSYS2 installation paths
	commonPaths := []string{
		"C:\\msys64\\usr\\bin\\bash.exe",
		"C:\\msys32\\usr\\bin\\bash.exe",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// getMSYS2BashPath returns the path to MSYS2 bash executable
func getMSYS2BashPath() (string, error) {
	commonPaths := []string{
		"C:\\msys64\\usr\\bin\\bash.exe",
		"C:\\msys32\\usr\\bin\\bash.exe",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("MSYS2 bash not found in common locations")
}

// mapToMSYS2Package maps a generic package name to MSYS2 UCRT64 package name
func mapToMSYS2Package(pkg string) string {
	// Map to mingw-w64-ucrt-x86_64-* packages for UCRT64 environment
	msys2Map := map[string]string{
		"jansson":              "mingw-w64-ucrt-x86_64-jansson",
		"libjansson-dev":       "mingw-w64-ucrt-x86_64-jansson",
		"curl":                 "mingw-w64-ucrt-x86_64-curl",
		"libcurl4-openssl-dev": "mingw-w64-ucrt-x86_64-curl",
		"sqlite3":              "mingw-w64-ucrt-x86_64-sqlite3",
		"libsqlite3-dev":       "mingw-w64-ucrt-x86_64-sqlite3",
		"openssl":              "mingw-w64-ucrt-x86_64-openssl",
		"libssl-dev":           "mingw-w64-ucrt-x86_64-openssl",
		"ncurses":              "mingw-w64-ucrt-x86_64-ncurses",
		"libncurses-dev":       "mingw-w64-ucrt-x86_64-ncurses",
	}

	if msys2Pkg, exists := msys2Map[pkg]; exists {
		return msys2Pkg
	}

	// If not in map, try adding the prefix
	return "mingw-w64-ucrt-x86_64-" + pkg
}

// installViaMSYS2Pacman installs packages using MSYS2's pacman
func installViaMSYS2Pacman(packages []string) error {
	bashPath, err := getMSYS2BashPath()
	if err != nil {
		return err
	}

	// Map packages to MSYS2 names
	msys2Packages := []string{}
	for _, pkg := range packages {
		msys2Packages = append(msys2Packages, mapToMSYS2Package(pkg))
	}

	// Build pacman command
	pacmanCmd := "pacman -S --noconfirm " + strings.Join(msys2Packages, " ")

	fmt.Printf("\nRunning MSYS2 pacman: %s\n", pacmanCmd)

	// Execute via bash -lc to get proper environment
	cmd := exec.Command(bashPath, "-lc", pacmanCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runCommand executes a command with arguments
func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// runWingetInstall runs winget install with better error handling
func runWingetInstall(packageID string) error {
	cmd := exec.Command("winget", "install", "--id", packageID, "--accept-package-agreements", "--accept-source-agreements")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		// Check for specific winget exit codes
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Common winget exit codes (hex values):
			// 0x8a15000f: Package already installed
			// 0x8a150014: No applicable installer
			// 0x8a150011: Package install already in progress
			// 0x8a150006: Installer error (may need manual install or already installed)
			// 0x8a150005: Installer download error
			// 0x8a15002b: No upgrade available (package already installed)
			// Treat these as non-critical - continue installation
			nonCriticalCodesHex := []uint32{0x8a15000f, 0x8a150014, 0x8a150011, 0x8a150006, 0x8a150005, 0x8a15002b}
			for _, code := range nonCriticalCodesHex {
				if uint32(exitCode) == code {
					return &wingetNonCriticalError{
						exitCode:  exitCode,
						output:    "",
						packageID: packageID,
					}
				}
			}
		}
		return err
	}
	return nil
}

// wingetNonCriticalError represents non-critical winget errors (already installed, etc.)
type wingetNonCriticalError struct {
	exitCode  int
	output    string
	packageID string
}

func (e *wingetNonCriticalError) Error() string {
	return fmt.Sprintf("winget non-critical error (exit code: %d, package: %s)", e.exitCode, e.packageID)
}

// isWingetNonCriticalError checks if an error is a non-critical winget error
func isWingetNonCriticalError(err error) bool {
	_, ok := err.(*wingetNonCriticalError)
	return ok
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
