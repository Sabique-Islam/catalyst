package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	config "github.com/Sabique-Islam/catalyst/internal/config"
	install "github.com/Sabique-Islam/catalyst/internal/install"
)

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

	// Determine compiler
	compiler := "gcc" // default for C
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("gcc"); err != nil {
			return fmt.Errorf("gcc not found in PATH")
		}
	} else {
		if _, err := exec.LookPath("gcc"); err != nil {
			return fmt.Errorf("gcc not found, install it using your package manager")
		}
	}

	// Build command arguments
	args := append(flags, "-o", output)
	args = append(args, sourceFiles...)

	cmd := exec.Command(compiler, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Compiling with: %s %s\n", compiler, args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	fmt.Printf("Compilation successful: %s\n", output)
	return nil
}

// BuildProject handles the complete build process including dependency installation and compilation
func BuildProject(args []string) error {
	var sourceFiles []string
	var flags []string

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

		// Install dependencies and get linker flags
		fmt.Println()
		fmt.Println("Installing dependencies...")
		linkerFlags, err := install.InstallDependenciesAndGetLinkerFlags()
		if err != nil {
			return err
		}

		// Add linker flags to compilation flags
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

	// Determine output binary
	output := "build/project"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	// Compile the C/C++ sources with linker flags
	fmt.Println()
	fmt.Println("Compiling project...")
	if err := CompileC(sourceFiles, output, flags); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Build complete!")
	fmt.Printf("Binary: %s\n", output)
	return nil
}

// RunProject executes the compiled binary, building it first if necessary
func RunProject(args []string) error {
	// Determine the binary path
	output := "build/project"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	// Build the project first if binary doesn't exist or sources are provided
	if len(args) > 0 {
		if err := BuildProject(args); err != nil {
			return err
		}
	} else {
		// Check if binary exists
		if _, err := os.Stat(output); os.IsNotExist(err) {
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

	cmd := exec.Command("./" + output)
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
