package install

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Sabique-Islam/catalyst/internal/platform"
)

// InstallationResult represents the result of a dependency installation
type InstallationResult struct {
	Package  string
	Success  bool
	Error    error
	Skipped  bool
	Reason   string
	Commands []string
}

// DependencyInstaller handles cross-platform dependency installation
type DependencyInstaller struct {
	OS         string
	PkgManager string
	DryRun     bool
	Verbose    bool
}

// NewDependencyInstaller creates a new installer for the current platform
func NewDependencyInstaller(dryRun, verbose bool) (*DependencyInstaller, error) {
	osName := platform.DetectOS()
	pkgManager, err := platform.DetectPackageManager(osName)
	if err != nil {
		return nil, fmt.Errorf("could not detect package manager: %w", err)
	}

	return &DependencyInstaller{
		OS:         osName,
		PkgManager: pkgManager,
		DryRun:     dryRun,
		Verbose:    verbose,
	}, nil
}

// InstallDependencies installs a list of packages
func (d *DependencyInstaller) InstallDependencies(packages []string) ([]InstallationResult, error) {
	var results []InstallationResult

	if len(packages) == 0 {
		return results, nil
	}

	// Setup package manager if needed
	if err := platform.SetupPackageManager(d.PkgManager); err != nil {
		if d.Verbose {
			fmt.Printf("Warning: Package manager setup: %v\n", err)
		}
	}

	// Update package manager database first
	if err := d.updatePackageDatabase(); err != nil {
		if d.Verbose {
			fmt.Printf("Warning: Failed to update package database: %v\n", err)
		}
	}

	// Install each package
	for _, pkg := range packages {
		result := d.installPackage(pkg)
		results = append(results, result)
	}

	return results, nil
}

// updatePackageDatabase updates the package manager's database
func (d *DependencyInstaller) updatePackageDatabase() error {
	var cmd *exec.Cmd

	switch d.PkgManager {
	case "apt":
		cmd = exec.Command("sudo", "apt", "update")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "makecache")
	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-Sy")
	case "brew":
		cmd = exec.Command("brew", "update")
	case "vcpkg":
		// vcpkg doesn't need database updates
		return nil
	case "choco":
		// Chocolatey updates automatically
		return nil
	default:
		return fmt.Errorf("unsupported package manager: %s", d.PkgManager)
	}

	if d.DryRun {
		if d.Verbose {
			fmt.Printf("DRY RUN: Would execute: %s\n", strings.Join(cmd.Args, " "))
		}
		return nil
	}

	if d.Verbose {
		fmt.Printf("Updating package database: %s\n", strings.Join(cmd.Args, " "))
	}

	return cmd.Run()
}

// installPackage installs a single package
func (d *DependencyInstaller) installPackage(pkg string) InstallationResult {
	result := InstallationResult{
		Package: pkg,
	}

	// Skip empty packages (standard library)
	if pkg == "" {
		result.Skipped = true
		result.Reason = "Standard library (no package needed)"
		return result
	}

	// Check if already installed
	if platform.IsPackageInstalled(pkg, d.PkgManager) {
		result.Skipped = true
		result.Reason = "Already installed"
		return result
	}

	// Generate install command
	cmd, err := d.getInstallCommand(pkg)
	if err != nil {
		result.Error = err
		return result
	}

	result.Commands = cmd.Args

	// Execute or simulate
	if d.DryRun {
		if d.Verbose {
			fmt.Printf("DRY RUN: Would execute: %s\n", strings.Join(cmd.Args, " "))
		}
		result.Success = true
		result.Reason = "Dry run - would install"
		return result
	}

	// Execute installation
	if d.Verbose {
		fmt.Printf("Installing %s: %s\n", pkg, strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("installation failed: %w\nOutput: %s", err, string(output))
		return result
	}

	// Verify installation
	if platform.IsPackageInstalled(pkg, d.PkgManager) {
		result.Success = true
		result.Reason = "Successfully installed"
	} else {
		result.Error = fmt.Errorf("installation reported success but package not found")
	}

	return result
}

// getInstallCommand generates the appropriate install command for the package
func (d *DependencyInstaller) getInstallCommand(pkg string) (*exec.Cmd, error) {
	switch d.PkgManager {
	case "apt":
		return exec.Command("sudo", "apt", "install", "-y", pkg), nil
	case "dnf":
		return exec.Command("sudo", "dnf", "install", "-y", pkg), nil
	case "pacman":
		return exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg), nil
	case "brew":
		return exec.Command("brew", "install", pkg), nil
	case "vcpkg":
		return exec.Command("vcpkg", "install", pkg), nil
	case "choco":
		return exec.Command("choco", "install", pkg, "-y"), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", d.PkgManager)
	}
}

// InstallBatch installs dependencies in batches for better performance
func (d *DependencyInstaller) InstallBatch(packages []string, batchSize int) ([]InstallationResult, error) {
	var allResults []InstallationResult

	if batchSize <= 0 {
		batchSize = 5 // Default batch size
	}

	// Filter out empty packages and already installed ones
	var toInstall []string
	for _, pkg := range packages {
		if pkg == "" {
			allResults = append(allResults, InstallationResult{
				Package: pkg,
				Skipped: true,
				Reason:  "Standard library (no package needed)",
			})
			continue
		}

		if platform.IsPackageInstalled(pkg, d.PkgManager) {
			allResults = append(allResults, InstallationResult{
				Package: pkg,
				Skipped: true,
				Reason:  "Already installed",
			})
			continue
		}

		toInstall = append(toInstall, pkg)
	}

	// Install in batches
	for i := 0; i < len(toInstall); i += batchSize {
		end := i + batchSize
		if end > len(toInstall) {
			end = len(toInstall)
		}

		batch := toInstall[i:end]
		results, err := d.installBatch(batch)
		if err != nil {
			return allResults, err
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// installBatch installs a batch of packages with a single command if supported
func (d *DependencyInstaller) installBatch(packages []string) ([]InstallationResult, error) {
	// Some package managers support batch installation
	if d.supportsBatchInstall() {
		return d.installMultiplePackages(packages)
	}

	// Fall back to individual installation
	var results []InstallationResult
	for _, pkg := range packages {
		result := d.installPackage(pkg)
		results = append(results, result)
	}

	return results, nil
}

// supportsBatchInstall checks if the package manager supports batch installation
func (d *DependencyInstaller) supportsBatchInstall() bool {
	switch d.PkgManager {
	case "apt", "dnf", "pacman", "brew":
		return true
	case "vcpkg", "choco":
		return false // Install one by one for better error handling
	default:
		return false
	}
}

// installMultiplePackages installs multiple packages in a single command
func (d *DependencyInstaller) installMultiplePackages(packages []string) ([]InstallationResult, error) {
	var results []InstallationResult

	// Generate batch install command
	var cmd *exec.Cmd
	switch d.PkgManager {
	case "apt":
		args := append([]string{"apt", "install", "-y"}, packages...)
		cmd = exec.Command("sudo", args...)
	case "dnf":
		args := append([]string{"dnf", "install", "-y"}, packages...)
		cmd = exec.Command("sudo", args...)
	case "pacman":
		args := append([]string{"pacman", "-S", "--noconfirm"}, packages...)
		cmd = exec.Command("sudo", args...)
	case "brew":
		args := append([]string{"install"}, packages...)
		cmd = exec.Command("brew", args...)
	default:
		return nil, fmt.Errorf("batch installation not supported for %s", d.PkgManager)
	}

	// Execute or simulate
	if d.DryRun {
		if d.Verbose {
			fmt.Printf("DRY RUN: Would execute: %s\n", strings.Join(cmd.Args, " "))
		}
		for _, pkg := range packages {
			results = append(results, InstallationResult{
				Package:  pkg,
				Success:  true,
				Reason:   "Dry run - would install",
				Commands: cmd.Args,
			})
		}
		return results, nil
	}

	if d.Verbose {
		fmt.Printf("Installing packages: %s\n", strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()

	// Check results for each package
	for _, pkg := range packages {
		result := InstallationResult{
			Package:  pkg,
			Commands: cmd.Args,
		}

		if err != nil {
			result.Error = fmt.Errorf("batch installation failed: %w\nOutput: %s", err, string(output))
		} else if platform.IsPackageInstalled(pkg, d.PkgManager) {
			result.Success = true
			result.Reason = "Successfully installed"
		} else {
			result.Error = fmt.Errorf("batch installation reported success but package not found")
		}

		results = append(results, result)
	}

	return results, nil
}

// PrintResults prints installation results in a user-friendly format
func PrintResults(results []InstallationResult, verbose bool) {
	if len(results) == 0 {
		fmt.Println("No packages to install.")
		return
	}

	successCount := 0
	skipCount := 0
	errorCount := 0

	fmt.Printf("\nInstallation Results:\n")
	fmt.Printf("====================\n")

	for _, result := range results {
		if result.Success {
			fmt.Printf("SUCCESS: %s - %s\n", result.Package, result.Reason)
			successCount++
		} else if result.Skipped {
			if verbose {
				fmt.Printf("SKIPPED: %s - %s\n", result.Package, result.Reason)
			}
			skipCount++
		} else {
			fmt.Printf("FAILED:  %s - %v\n", result.Package, result.Error)
			if verbose && len(result.Commands) > 0 {
				fmt.Printf("         Command: %s\n", strings.Join(result.Commands, " "))
			}
			errorCount++
		}
	}

	fmt.Printf("\nSummary: %d succeeded, %d skipped, %d failed\n", successCount, skipCount, errorCount)
}
