package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	config "github.com/Sabique-Islam/catalyst/internal/config"
	install "github.com/Sabique-Islam/catalyst/internal/install"
)

// CompilerInfo holds information about a detected compiler
type CompilerInfo struct {
	Name       string
	Executable string
	Flags      []string
}

// detectCompiler finds the best available compiler for the current platform
func detectCompiler() (*CompilerInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return detectWindowsCompiler()
	default:
		return detectUnixCompiler()
	}
}

// detectWindowsCompiler tries to find a suitable C compiler on Windows
func detectWindowsCompiler() (*CompilerInfo, error) {
	// Priority order: cl.exe (MSVC) > clang > gcc
	compilers := []struct {
		name       string
		executable string
		flags      []string
	}{
		{"MSVC", "cl", []string{"/nologo"}},
		{"Clang", "clang", []string{}},
		{"GCC", "gcc", []string{}},
		{"MinGW-GCC", "x86_64-w64-mingw32-gcc", []string{}},
		{"TDM-GCC", "tdm-gcc", []string{}},
		{"W64DevKit-GCC", "w64devkit-gcc", []string{}}, // Portable GCC toolchain
	}

	// Check for vcpkg and add vcpkg-specific flags if found
	vcpkgFlags := []string{}
	if vcpkgRoot := os.Getenv("VCPKG_ROOT"); vcpkgRoot != "" {
		vcpkgFlags = append(vcpkgFlags, "-I"+filepath.Join(vcpkgRoot, "installed", "x64-windows", "include"))
		vcpkgFlags = append(vcpkgFlags, "-L"+filepath.Join(vcpkgRoot, "installed", "x64-windows", "lib"))
		fmt.Printf("Found vcpkg installation at: %s\n", vcpkgRoot)
	}

	for _, compiler := range compilers {
		if _, err := exec.LookPath(compiler.executable); err == nil {
			fmt.Printf("Found %s compiler: %s\n", compiler.name, compiler.executable)
			
			// Add vcpkg flags if available and not MSVC (MSVC uses vcpkg integration differently)
			flags := compiler.flags
			if len(vcpkgFlags) > 0 && compiler.name != "MSVC" {
				flags = append(flags, vcpkgFlags...)
			}
			
			return &CompilerInfo{
				Name:       compiler.name,
				Executable: compiler.executable,
				Flags:      flags,
			}, nil
		}
	}

	return nil, fmt.Errorf(`no C compiler found on Windows. Please install one of:
  • Visual Studio Build Tools (includes cl.exe): https://visualstudio.microsoft.com/downloads/
  • LLVM/Clang: https://releases.llvm.org/download.html
  • MinGW-w64: https://www.mingw-w64.org/downloads/
  • TDM-GCC: https://jmeubank.github.io/tdm-gcc/
  • W64DevKit: https://github.com/skeeto/w64devkit (portable)

Or install via package manager:
  winget install Microsoft.VisualStudio.2022.BuildTools
  winget install LLVM.LLVM
  winget install TDM-GCC.TDM-GCC
  choco install llvm
  choco install mingw
  choco install visualstudio2022buildtools

For library management, consider installing vcpkg:
  winget install Microsoft.vcpkg`)
}

// detectUnixCompiler tries to find a suitable C compiler on Unix-like systems
func detectUnixCompiler() (*CompilerInfo, error) {
	// Priority order: gcc > clang > cc
	compilers := []struct {
		name       string
		executable string
		flags      []string
	}{
		{"GCC", "gcc", []string{}},
		{"Clang", "clang", []string{}},
		{"CC", "cc", []string{}},
	}

	for _, compiler := range compilers {
		if _, err := exec.LookPath(compiler.executable); err == nil {
			return &CompilerInfo{
				Name:       compiler.name,
				Executable: compiler.executable,
				Flags:      compiler.flags,
			}, nil
		}
	}

	return nil, fmt.Errorf("no C compiler found, install gcc or clang using your package manager")
}

// convertToMSVCFlag converts GCC/Clang-style flags to MSVC equivalents
func convertToMSVCFlag(gccFlag string) string {
	flagMap := map[string]string{
		// Optimization levels
		"-O0": "/Od", // No optimization
		"-O1": "/O1", // Minimize size
		"-O2": "/O2", // Maximize speed
		"-O3": "/Ox", // Full optimization
		"-Os": "/O1", // Optimize for size

		// Debug information
		"-g":  "/Zi", // Debug information
		"-gg": "/Z7", // Debug info in object files

		// Warnings
		"-Wall":   "/Wall",
		"-Werror": "/WX",

		// Defines
		"-DNDEBUG": "/DNDEBUG",

		// Threading/OpenMP
		"-fopenmp": "/openmp",
		"-pthread": "", // MSVC handles threading differently

		// Position independent code (not applicable to MSVC in the same way)
		"-fPIC": "",

		// Math optimizations
		"-ffast-math": "/fp:fast",

		// Security
		"-fstack-protector-strong": "/GS",
	}

	// Handle -D defines
	if strings.HasPrefix(gccFlag, "-D") {
		return "/" + gccFlag[1:] // Convert -DFOO to /DFOO
	}

	// Handle -I includes
	if strings.HasPrefix(gccFlag, "-I") {
		return "/" + gccFlag[1:] // Convert -Ipath to /Ipath
	}

	// Handle -l libraries (convert to .lib files)
	if strings.HasPrefix(gccFlag, "-l") {
		libName := gccFlag[2:]
		// Common library mappings
		libMap := map[string]string{
			"m":       "", // Math library is built-in on Windows
			"pthread": "", // Threading handled differently
			"gomp":    "", // OpenMP handled by /openmp flag
			"omp":     "", // OpenMP handled by /openmp flag
		}
		if msvcLib, ok := libMap[libName]; ok {
			return msvcLib
		}
		return libName + ".lib"
	}

	// Look up direct mapping
	if msvcFlag, ok := flagMap[gccFlag]; ok {
		return msvcFlag
	}

	// Return empty string for unsupported flags
	return ""
}

// CompileC compiles a C/C++ source file or project into a binary
func CompileC(sourceFiles []string, output string, flags []string) error {
	if len(sourceFiles) == 0 {
		return fmt.Errorf("no source files provided for compilation")
	}

	// Ensure output directory exists
	outDir := filepath.Dir(output)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine compiler based on platform and availability
	compilerInfo, err := detectCompiler()
	if err != nil {
		return err
	}

	// Build command arguments based on compiler type
	var args []string
	if compilerInfo.Name == "MSVC" {
		// MSVC uses different syntax: cl /Fe<output> <sources> [flags]
		args = append(compilerInfo.Flags, "/Fe"+output)
		args = append(args, sourceFiles...)
		// Convert GCC-style flags to MSVC equivalents
		for _, flag := range flags {
			msvcFlag := convertToMSVCFlag(flag)
			if msvcFlag != "" {
				args = append(args, msvcFlag)
			}
		}
	} else {
		// GCC/Clang style: compiler -o output sources [flags]
		args = append(compilerInfo.Flags, "-o", output)
		args = append(args, sourceFiles...)
		args = append(args, flags...)
	}

	cmd := exec.Command(compilerInfo.Executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Compiling with %s: %s %s\n", compilerInfo.Name, compilerInfo.Executable, strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	fmt.Printf("Compilation successful: %s\n", output)
	return nil
}

// ensureCompilerAvailable checks if a compiler is available and offers to install one if not
func ensureCompilerAvailable() error {
	_, err := detectCompiler()
	if err == nil {
		return nil // Compiler already available
	}

	if runtime.GOOS == "windows" {
		fmt.Println("No C compiler found. Catalyst can help install one for you.")
		fmt.Println("\nRecommended options for Windows:")
		fmt.Println("  1. Microsoft Visual C++ Build Tools (cl.exe) - Best for Windows")
		fmt.Println("  2. LLVM/Clang - Cross-platform, modern")
		fmt.Println("  3. TDM-GCC - Traditional GCC for Windows")
		fmt.Println()

		// Try to auto-install the best option
		fmt.Println("Attempting to install Microsoft Visual C++ Build Tools...")
		if installCmd := getInstallCommand("build-essential"); installCmd != nil {
			fmt.Printf("Running: %s\n", strings.Join(installCmd, " "))
			cmd := exec.Command(installCmd[0], installCmd[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to install build tools automatically: %v\n", err)
				fmt.Println("Please install a C compiler manually. See the error message above for guidance.")
				return err
			}
			fmt.Println("Build tools installation completed. Please restart your terminal and try again.")
		}
	}

	return err // Return original error if we couldn't fix it
}

// getInstallCommand returns the command to install a package using available package manager
func getInstallCommand(pkg string) []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Try winget first (most modern and reliable)
	if _, err := exec.LookPath("winget"); err == nil {
		if pkg == "build-essential" {
			return []string{"winget", "install", "--id", "Microsoft.VisualStudio.2022.BuildTools", "--accept-package-agreements", "--accept-source-agreements"}
		}
	}

	// Fall back to chocolatey
	if _, err := exec.LookPath("choco"); err == nil {
		if pkg == "build-essential" {
			return []string{"choco", "install", "visualstudio2022buildtools", "-y"}
		}
	}

	return nil
}

// BuildProject handles the complete build process including dependency installation and compilation
func BuildProject(args []string) error {
	// First ensure a compiler is available
	if err := ensureCompilerAvailable(); err != nil {
		return fmt.Errorf("compiler not available: %w", err)
	}

	var sourceFiles []string
	var flags []string
	var output string

	// Check if catalyst.yml exists
	if _, err := os.Stat("catalyst.yml"); err == nil {
		// Load configuration from catalyst.yml
		cfg, err := config.LoadConfig("catalyst.yml")
		if err != nil {
			return fmt.Errorf("failed to load catalyst.yml: %w", err)
		}

		// Use sources from config if no args provided
		if len(args) == 0 {
			if len(cfg.Sources) == 0 {
				return fmt.Errorf("no source files specified in catalyst.yml or command line")
			}
			sourceFiles = cfg.Sources
			fmt.Printf("Building from catalyst.yml: %s\n", cfg.ProjectName)
			fmt.Printf("Source files: %v\n", sourceFiles)

			// Use flags from config
			if len(cfg.Flags) > 0 {
				flags = append(flags, cfg.Flags...)
			}

			// Use output name from config
			if cfg.Output != "" {
				output = cfg.Output
			} else {
				output = cfg.ProjectName
			}
		} else {
			// Use command-line args
			for _, arg := range args {
				if len(arg) > 0 && arg[0] == '-' {
					flags = append(flags, arg)
				} else {
					sourceFiles = append(sourceFiles, arg)
				}
			}
		}

		// Install dependencies and get compiler and linker flags
		fmt.Println()
		fmt.Println("Installing dependencies...")
		compilerFlags, linkerFlags, err := install.InstallDependenciesAndGetFlags()
		if err != nil {
			return err
		}

		// Add compiler and linker flags to compilation flags
		flags = append(flags, compilerFlags...)
		flags = append(flags, linkerFlags...)
	} else {
		// No catalyst.yml, require command-line args
		if len(args) == 0 {
			return fmt.Errorf("no catalyst.yml found and no source files provided\n\nUsage:\n  catalyst build <source files>\n  or create catalyst.yml with 'catalyst init'")
		}

		// Separate source files from compiler flags
		for _, arg := range args {
			if len(arg) > 0 && arg[0] == '-' {
				flags = append(flags, arg)
			} else {
				sourceFiles = append(sourceFiles, arg)
			}
		}
	}

	// Determine output binary path (always in build/ directory)
	if output == "" {
		output = "project"
	}
	outputPath := filepath.Join("build", output)
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	// Compile the C/C++ sources with linker flags
	fmt.Println()
	fmt.Println("Compiling project...")
	if err := CompileC(sourceFiles, outputPath, flags); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Build complete!")
	fmt.Printf("Binary: %s\n", outputPath)
	return nil
}

// RunProject executes the compiled binary, building it first if necessary
func RunProject(args []string) error {
	// Determine the binary path from config or default
	output := "project"

	// Try to load config to get output name
	if _, err := os.Stat("catalyst.yml"); err == nil {
		cfg, err := config.LoadConfig("catalyst.yml")
		if err == nil {
			if cfg.Output != "" {
				output = cfg.Output
			} else if cfg.ProjectName != "" {
				output = cfg.ProjectName
			}
		}
	}

	outputPath := filepath.Join("build", output)
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	// Build the project first if binary doesn't exist or sources are provided
	if len(args) > 0 {
		if err := BuildProject(args); err != nil {
			return err
		}
	} else {
		// Check if binary exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			// Try to build from catalyst.yml
			fmt.Println("Binary not found, building from catalyst.yml...")
			if err := BuildProject([]string{}); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}
		}
	}

	// Execute the binary
	fmt.Println()
	fmt.Println("Running project...")
	fmt.Println("==============================================")
	fmt.Println()

	cmd := exec.Command("./" + outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

// CleanProject removes build artifacts and compiled binaries
func CleanProject() error {
	fmt.Println("Cleaning build artifacts...")

	// Remove build directory
	buildDir := "build"
	if _, err := os.Stat(buildDir); err == nil {
		if err := os.RemoveAll(buildDir); err != nil {
			return fmt.Errorf("failed to remove build directory: %w", err)
		}
		fmt.Println("Removed build/ directory")
	}

	// Remove bin directory (legacy)
	binDir := "bin"
	if _, err := os.Stat(binDir); err == nil {
		if err := os.RemoveAll(binDir); err != nil {
			return fmt.Errorf("failed to remove bin directory: %w", err)
		}
		fmt.Println("Removed bin/ directory")
	}

	// Remove common executable names
	commonExecs := []string{"project", "project.exe", "a.out", "a.exe"}
	removed := 0
	for _, exec := range commonExecs {
		if _, err := os.Stat(exec); err == nil {
			if err := os.Remove(exec); err != nil {
				fmt.Printf("Warning: Failed to remove %s: %v\n", exec, err)
			} else {
				fmt.Printf("Removed %s\n", exec)
				removed++
			}
		}
	}

	if removed == 0 && buildDir != "" {
		fmt.Println("Clean complete - no additional artifacts found")
	} else {
		fmt.Printf("Cleaned %d build artifact(s)\n", removed)
	}

	return nil
}
