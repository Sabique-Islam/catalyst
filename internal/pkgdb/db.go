package pkgdb

// PackageDB is a translation database that maps abstract package names
// (as found by the dependency scanner) to real, installable package names
// for different system package managers.
//
// Format: [AbstractName][PackageManager] -> RealPackageName
var PackageDB = map[string]map[string]string{
	"openssl": {
		"apt":    "libssl-dev",
		"dnf":    "openssl-devel",
		"pacman": "openssl",
		"brew":   "openssl",
		"vcpkg":  "openssl",
		"choco":  "openssl",
	},
	"png": {
		"apt":    "libpng-dev",
		"dnf":    "libpng-devel",
		"pacman": "libpng",
		"brew":   "libpng",
		"vcpkg":  "libpng",
		"choco":  "libpng",
	},
	"zlib": {
		"apt":    "zlib1g-dev",
		"dnf":    "zlib-devel",
		"pacman": "zlib",
		"brew":   "zlib",
		"vcpkg":  "zlib",
		"choco":  "zlib",
	},
	"stdio": {
		// stdio is part of the C standard library, no separate package needed
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"stdlib": {
		// stdlib is part of the C standard library, no separate package needed
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"string": {
		// string.h is part of the C standard library, no separate package needed
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"math": {
		// math.h is part of the C standard library, but may need -lm flag
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
}

// Translate converts an abstract package name to the real package name
// for a specific package manager.
//
// Parameters:
//   - abstractName: The abstract package name (e.g., "openssl", "png")
//   - pkgManager: The package manager (e.g., "apt", "brew", "dnf")
//
// Returns:
//   - string: The real package name for the given package manager
//   - bool: true if a translation was found, false otherwise
//
// If the abstract name is not in the database, or if the package manager
// is not supported for that package, it returns ("", false).
// An empty string with true means the package is part of the standard library.
func Translate(abstractName, pkgManager string) (string, bool) {
	// Check if the abstract name exists in the database
	pkgMap, exists := PackageDB[abstractName]
	if !exists {
		return "", false
	}

	// Check if the package manager is supported for this package
	realName, exists := pkgMap[pkgManager]
	if !exists {
		return "", false
	}

	return realName, true
}
