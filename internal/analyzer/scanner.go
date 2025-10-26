package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectScanner scans and analyzes a C/C++ project
type ProjectScanner struct {
	RootPath     string
	SourceFiles  []string
	HeaderFiles  []string
	BuildTargets []BuildTarget
	ExternalLibs []ExternalLibrary
	VendoredLibs []VendoredLibrary
	IncludeMap   map[string][]string // file -> includes
}

// BuildTarget represents a buildable target (executable)
type BuildTarget struct {
	Name         string
	EntryPoint   string
	SourceFiles  []string
	Dependencies []string
	IncludePaths []string
	Type         string // "executable", "library"
	Directory    string // Subdirectory if any
}

// ExternalLibrary represents a system library dependency
type ExternalLibrary struct {
	Name       string
	HeaderName string
	LinkerFlag string
	PkgConfig  string
	Platforms  map[string]PlatformPackage
}

// PlatformPackage contains platform-specific package info
type PlatformPackage struct {
	PackageName string
	IncludePath string
	LibPath     string
}

// VendoredLibrary represents a vendored/bundled library
type VendoredLibrary struct {
	Name        string
	Path        string
	SourceFiles []string
	HeaderFiles []string
}

// NewProjectScanner creates a new project scanner
func NewProjectScanner(rootPath string) *ProjectScanner {
	return &ProjectScanner{
		RootPath:   rootPath,
		IncludeMap: make(map[string][]string),
	}
}

// ScanProject performs a full project scan
func (ps *ProjectScanner) ScanProject() error {
	// Scan for source and header files
	if err := ps.scanFiles(); err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Parse includes from all files
	if err := ps.parseIncludes(); err != nil {
		return fmt.Errorf("failed to parse includes: %w", err)
	}

	// Detect build targets (files with main())
	if err := ps.detectBuildTargets(); err != nil {
		return fmt.Errorf("failed to detect build targets: %w", err)
	}

	// Detect vendored libraries
	if err := ps.detectVendoredLibraries(); err != nil {
		return fmt.Errorf("failed to detect vendored libraries: %w", err)
	}

	// Detect external libraries
	if err := ps.detectExternalLibraries(); err != nil {
		return fmt.Errorf("failed to detect external libraries: %w", err)
	}

	return nil
}

// scanFiles recursively scans for C/C++ source and header files
func (ps *ProjectScanner) scanFiles() error {
	return filepath.Walk(ps.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common build/dependency directories
		if info.IsDir() {
			name := filepath.Base(path)
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "build" || name == "dist" || name == "node_modules" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		relPath, _ := filepath.Rel(ps.RootPath, path)

		// Collect source files
		if ext == ".c" || ext == ".cpp" || ext == ".cc" || ext == ".cxx" {
			ps.SourceFiles = append(ps.SourceFiles, relPath)
		}

		// Collect header files
		if ext == ".h" || ext == ".hpp" || ext == ".hh" || ext == ".hxx" {
			ps.HeaderFiles = append(ps.HeaderFiles, relPath)
		}

		return nil
	})
}

// parseIncludes extracts #include statements from all files
func (ps *ProjectScanner) parseIncludes() error {
	includeRegex := regexp.MustCompile(`^\s*#include\s+["<]([^">]+)[">]`)

	allFiles := append(ps.SourceFiles, ps.HeaderFiles...)

	for _, file := range allFiles {
		fullPath := filepath.Join(ps.RootPath, file)
		f, err := os.Open(fullPath)
		if err != nil {
			continue // Skip files we can't open
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var includes []string

		for scanner.Scan() {
			line := scanner.Text()
			if matches := includeRegex.FindStringSubmatch(line); matches != nil {
				includes = append(includes, matches[1])
			}
		}

		if len(includes) > 0 {
			ps.IncludeMap[file] = includes
		}
	}

	return nil
}

// detectBuildTargets finds files with main() functions
func (ps *ProjectScanner) detectBuildTargets() error {
	mainRegex := regexp.MustCompile(`\bint\s+main\s*\(`)

	for _, sourceFile := range ps.SourceFiles {
		fullPath := filepath.Join(ps.RootPath, sourceFile)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if mainRegex.Match(content) {
			// Found a main() function - this is a build target
			target := BuildTarget{
				Name:       ps.deriveTargetName(sourceFile),
				EntryPoint: sourceFile,
				Type:       "executable",
				Directory:  filepath.Dir(sourceFile),
			}

			// Collect related source files
			target.SourceFiles = ps.collectRelatedSources(sourceFile)

			ps.BuildTargets = append(ps.BuildTargets, target)
		}
	}

	return nil
}

// deriveTargetName derives a target name from the source file
func (ps *ProjectScanner) deriveTargetName(sourceFile string) string {
	// Remove extension
	name := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))

	// If in a subdirectory, use the directory name
	dir := filepath.Dir(sourceFile)
	if dir != "." && dir != "" {
		// Use the last directory component
		return filepath.Base(dir)
	}

	// Use the base filename
	return filepath.Base(name)
}

// collectRelatedSources collects all sources related to an entry point
func (ps *ProjectScanner) collectRelatedSources(entryPoint string) []string {
	sources := []string{entryPoint}
	entryDir := filepath.Dir(entryPoint)

	// If entry point is in a subdirectory, include all sources from that directory
	if entryDir != "." && entryDir != "" {
		for _, src := range ps.SourceFiles {
			srcDir := filepath.Dir(src)
			if srcDir == entryDir || strings.HasPrefix(srcDir, entryDir+"/") {
				if src != entryPoint {
					sources = append(sources, src)
				}
			}
		}
	} else {
		// Entry point is in root - include sources from root and common directories
		for _, src := range ps.SourceFiles {
			if src == entryPoint {
				continue
			}
			srcDir := filepath.Dir(src)
			// Include from src/ directory or root
			if srcDir == "." || srcDir == "src" || strings.HasPrefix(srcDir, "src/") {
				sources = append(sources, src)
			}
		}
	}

	return sources
}

// detectVendoredLibraries finds vendored/bundled libraries
func (ps *ProjectScanner) detectVendoredLibraries() error {
	// Common vendored library directory patterns
	vendorPatterns := []string{"vendor", "third_party", "external", "lib", "libs", "deps"}

	for _, pattern := range vendorPatterns {
		vendorPath := filepath.Join(ps.RootPath, pattern)
		if _, err := os.Stat(vendorPath); err == nil {
			// Directory exists, scan it
			ps.scanVendorDirectory(pattern)
		}
	}

	// Also check for self-contained library directories (e.g., cjson/)
	ps.detectSelfContainedLibraries()

	return nil
}

// scanVendorDirectory scans a vendor directory for libraries
func (ps *ProjectScanner) scanVendorDirectory(vendorDir string) {
	filepath.Walk(filepath.Join(ps.RootPath, vendorDir), func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(ps.RootPath, path)
		if relPath == vendorDir {
			return nil // Skip the vendor directory itself
		}

		// Check if this directory contains both .c and .h files
		var sources, headers []string
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			if entry.IsDir() {
				return nil
			}
			ext := filepath.Ext(entry.Name())
			if ext == ".c" || ext == ".cpp" {
				sources = append(sources, filepath.Join(relPath, entry.Name()))
			}
			if ext == ".h" || ext == ".hpp" {
				headers = append(headers, filepath.Join(relPath, entry.Name()))
			}
		}

		if len(sources) > 0 && len(headers) > 0 {
			ps.VendoredLibs = append(ps.VendoredLibs, VendoredLibrary{
				Name:        filepath.Base(relPath),
				Path:        relPath,
				SourceFiles: sources,
				HeaderFiles: headers,
			})
		}

		return filepath.SkipDir // Don't recurse deeper
	})
}

// detectSelfContainedLibraries finds self-contained library directories
func (ps *ProjectScanner) detectSelfContainedLibraries() {
	// Look for directories with both source and header files that might be libraries
	dirFiles := make(map[string][]string)

	for _, src := range ps.SourceFiles {
		dir := filepath.Dir(src)
		dirFiles[dir] = append(dirFiles[dir], src)
	}

	for dir, files := range dirFiles {
		if dir == "." || dir == "src" || strings.HasPrefix(dir, "src/") {
			continue // Skip main source directories
		}

		// Check if this looks like a library (e.g., cjson/, json-c/)
		dirName := filepath.Base(dir)
		if ps.looksLikeVendoredLib(dirName, files) {
			var headers []string
			for _, h := range ps.HeaderFiles {
				if filepath.Dir(h) == dir {
					headers = append(headers, h)
				}
			}

			if len(headers) > 0 {
				ps.VendoredLibs = append(ps.VendoredLibs, VendoredLibrary{
					Name:        dirName,
					Path:        dir,
					SourceFiles: files,
					HeaderFiles: headers,
				})
			}
		}
	}
}

// looksLikeVendoredLib checks if a directory looks like a vendored library
func (ps *ProjectScanner) looksLikeVendoredLib(dirName string, files []string) bool {
	// Common library name patterns
	libPatterns := []string{"json", "xml", "yaml", "http", "crypto", "ssl", "sqlite", "curl"}

	dirLower := strings.ToLower(dirName)
	for _, pattern := range libPatterns {
		if strings.Contains(dirLower, pattern) {
			return true
		}
	}

	// Check if files are self-contained (all from same directory)
	if len(files) > 0 {
		firstDir := filepath.Dir(files[0])
		for _, f := range files {
			if filepath.Dir(f) != firstDir {
				return false
			}
		}
		return true
	}

	return false
}

// detectExternalLibraries detects system library dependencies
func (ps *ProjectScanner) detectExternalLibraries() error {
	// Collect all includes
	allIncludes := make(map[string]bool)
	for _, includes := range ps.IncludeMap {
		for _, inc := range includes {
			allIncludes[inc] = true
		}
	}

	// Check against known external libraries
	knownLibs := getKnownLibraries()

	for include := range allIncludes {
		// Skip standard library headers
		if isStandardHeader(include) {
			continue
		}

		// Skip project headers
		if ps.isProjectHeader(include) {
			continue
		}

		// Check if it matches a known external library
		for _, lib := range knownLibs {
			if include == lib.HeaderName || strings.Contains(include, lib.HeaderName) {
				ps.ExternalLibs = append(ps.ExternalLibs, lib)
				break
			}
		}
	}

	return nil
}

// isProjectHeader checks if a header is part of the project
func (ps *ProjectScanner) isProjectHeader(include string) bool {
	// Check if the include matches any project header
	for _, header := range ps.HeaderFiles {
		if filepath.Base(header) == include || header == include {
			return true
		}
	}
	return false
}

// isStandardHeader checks if a header is a standard C/C++ library header
func isStandardHeader(header string) bool {
	standardHeaders := []string{
		"stdio.h", "stdlib.h", "string.h", "math.h", "time.h",
		"ctype.h", "errno.h", "assert.h", "stddef.h", "stdint.h",
		"stdbool.h", "limits.h", "float.h", "signal.h", "setjmp.h",
		"stdarg.h", "locale.h", "wchar.h", "wctype.h", "unistd.h",
		"pthread.h", "sys/types.h", "sys/stat.h", "fcntl.h",
		// C++ headers
		"iostream", "vector", "string", "map", "algorithm",
		"memory", "functional", "thread", "mutex", "atomic",
	}

	for _, std := range standardHeaders {
		if header == std {
			return true
		}
	}
	return false
}

// GetSummary returns a summary of the scan results
func (ps *ProjectScanner) GetSummary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Project Analysis Summary\n"))
	sb.WriteString(fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))
	sb.WriteString(fmt.Sprintf("Source Files: %d\n", len(ps.SourceFiles)))
	sb.WriteString(fmt.Sprintf("Header Files: %d\n", len(ps.HeaderFiles)))
	sb.WriteString(fmt.Sprintf("Build Targets: %d\n\n", len(ps.BuildTargets)))

	if len(ps.BuildTargets) > 0 {
		sb.WriteString("Build Targets:\n")
		for i, target := range ps.BuildTargets {
			sb.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, target.Name, target.Type))
			sb.WriteString(fmt.Sprintf("     Entry: %s\n", target.EntryPoint))
			sb.WriteString(fmt.Sprintf("     Sources: %d files\n", len(target.SourceFiles)))
			if target.Directory != "." && target.Directory != "" {
				sb.WriteString(fmt.Sprintf("     Directory: %s/\n", target.Directory))
			}
		}
		sb.WriteString("\n")
	}

	if len(ps.VendoredLibs) > 0 {
		sb.WriteString(fmt.Sprintf("Vendored Libraries: %d\n", len(ps.VendoredLibs)))
		for _, lib := range ps.VendoredLibs {
			sb.WriteString(fmt.Sprintf("  • %s (%s/)\n", lib.Name, lib.Path))
		}
		sb.WriteString("\n")
	}

	if len(ps.ExternalLibs) > 0 {
		sb.WriteString(fmt.Sprintf("External Dependencies: %d\n", len(ps.ExternalLibs)))
		for _, lib := range ps.ExternalLibs {
			sb.WriteString(fmt.Sprintf("  • %s\n", lib.Name))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
