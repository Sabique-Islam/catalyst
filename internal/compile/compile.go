package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
