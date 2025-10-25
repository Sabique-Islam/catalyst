package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

	fmt.Printf("✅ Compilation successful: %s\n", output)
	return nil
}

// BuildProject handles the complete build process including dependency installation and compilation
func BuildProject(args []string) error {
	// 1️⃣ Install dependencies and get linker flags
	linkerFlags, err := install.InstallDependenciesAndGetLinkerFlags()
	if err != nil {
		return err
	}

	// 2️⃣ Separate source files from compiler flags
	sourceFiles := []string{}
	flags := []string{}
	for _, arg := range args {
		if len(arg) > 0 && arg[0] == '-' {
			flags = append(flags, arg)
		} else {
			sourceFiles = append(sourceFiles, arg)
		}
	}

	// 3️⃣ Add linker flags to compilation flags
	flags = append(flags, linkerFlags...)

	// 4️⃣ Determine output binary
	output := "bin/project"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	// 5️⃣ Compile the C/C++ sources with linker flags
	if err := CompileC(sourceFiles, output, flags); err != nil {
		return err
	}

	fmt.Println("✅ Build complete")
	return nil
}

// RunProject executes the compiled binary, building it first if necessary
func RunProject(args []string) error {
	// 1️⃣ Build the project first if binary doesn't exist or sources are newer
	if len(args) > 0 {
		if err := BuildProject(args); err != nil {
			return err
		}
	} else {
		// Try to find a default binary to run
		output := "bin/project"
		if runtime.GOOS == "windows" {
			output += ".exe"
		}

		if _, err := os.Stat(output); os.IsNotExist(err) {
			return fmt.Errorf("no binary found at %s. Please build the project first with 'catalyst build <source files>'", output)
		}
	}

	// 2️⃣ Determine the binary path
	output := "bin/project"
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	// 3️⃣ Execute the binary
	fmt.Printf("Running: %s\n", output)
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

	// Remove bin directory
	binDir := "bin"
	if _, err := os.Stat(binDir); err == nil {
		if err := os.RemoveAll(binDir); err != nil {
			return fmt.Errorf("failed to remove bin directory: %w", err)
		}
		fmt.Println("✅ Removed bin/ directory")
	}

	// Remove common executable names
	commonExecs := []string{"project", "project.exe", "a.out", "a.exe"}
	removed := 0
	for _, exec := range commonExecs {
		if _, err := os.Stat(exec); err == nil {
			if err := os.Remove(exec); err != nil {
				fmt.Printf("⚠️  Failed to remove %s: %v\n", exec, err)
			} else {
				fmt.Printf("✅ Removed %s\n", exec)
				removed++
			}
		}
	}

	if removed == 0 {
		fmt.Println("✅ No build artifacts found to clean")
	} else {
		fmt.Printf("✅ Cleaned %d build artifact(s)\n", removed)
	}

	return nil
}
