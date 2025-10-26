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
	"ssl": {
		"apt":    "libssl-dev",
		"dnf":    "openssl-devel",
		"pacman": "openssl",
		"brew":   "openssl",
		"vcpkg":  "openssl",
		"choco":  "openssl",
	},
	"crypto": {
		"apt":    "libssl-dev",
		"dnf":    "openssl-devel",
		"pacman": "openssl",
		"brew":   "openssl",
		"vcpkg":  "openssl",
		"choco":  "openssl",
	},
	"curl": {
		"apt":    "libcurl4-openssl-dev",
		"dnf":    "libcurl-devel",
		"pacman": "curl",
		"brew":   "curl",
		"vcpkg":  "curl",
		"choco":  "curl",
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
	"sqlite3": {
		"apt":    "libsqlite3-dev",
		"dnf":    "sqlite-devel",
		"pacman": "sqlite",
		"brew":   "sqlite",
		"vcpkg":  "sqlite3",
		"choco":  "sqlite",
	},
	"sqlite": {
		"apt":    "libsqlite3-dev",
		"dnf":    "sqlite-devel",
		"pacman": "sqlite",
		"brew":   "sqlite",
		"vcpkg":  "sqlite3",
		"choco":  "sqlite",
	},
	"pthread": {
		"apt":    "", // Built into glibc on Linux
		"dnf":    "", // Built into glibc on Linux
		"pacman": "", // Built into glibc on Linux
		"brew":   "", // Built into darwin
		"vcpkg":  "pthreads",
		"choco":  "pthreads",
	},
	"omp": {
		"apt":    "libomp-dev",
		"dnf":    "libomp-devel",
		"pacman": "openmp",
		"brew":   "libomp",
		"vcpkg":  "", // OpenMP included with gcc on Windows
		"choco":  "", // OpenMP included with mingw/gcc
	},
	"jansson": {
		"apt":    "libjansson-dev",
		"dnf":    "jansson-devel",
		"pacman": "jansson",
		"brew":   "jansson",
		"vcpkg":  "jansson",
		"choco":  "jansson",
	},
	"readline": {
		"apt":    "libreadline-dev",
		"dnf":    "readline-devel",
		"pacman": "readline",
		"brew":   "readline",
		"vcpkg":  "readline",
		"choco":  "readline",
	},
	"ncurses": {
		"apt":    "libncurses-dev",
		"dnf":    "ncurses-devel",
		"pacman": "ncurses",
		"brew":   "ncurses",
		"vcpkg":  "ncurses",
		"choco":  "ncurses",
	},
	"pcre": {
		"apt":    "libpcre3-dev",
		"dnf":    "pcre-devel",
		"pacman": "pcre",
		"brew":   "pcre",
		"vcpkg":  "pcre",
		"choco":  "pcre",
	},
	"json": {
		"apt":    "libjansson-dev",
		"dnf":    "jansson-devel",
		"pacman": "jansson",
		"brew":   "jansson",
		"vcpkg":  "jansson",
		"choco":  "jansson",
	},
	// Standard library headers - no package needed
	"stdio": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"stdlib": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"string": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"math": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"time": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"ctype": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"assert": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"errno": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"signal": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"stdarg": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"stdbool": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"stdint": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"unistd": {
		"apt":    "",
		"dnf":    "",
		"pacman": "",
		"brew":   "",
		"vcpkg":  "",
		"choco":  "",
	},
	"fcntl": {
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

// TranslateWithSearch attempts static translation first, then falls back to dynamic search
func TranslateWithSearch(abstractName, pkgManager string) (string, bool) {
	// First try static translation
	if realName, found := Translate(abstractName, pkgManager); found {
		return realName, true
	}

	// If not found in static database, try dynamic search
	searchResults, err := DynamicSearch(abstractName, pkgManager)
	if err != nil {
		return "", false
	}

	// Get the best match from search results
	bestMatch, found := GetBestMatch(searchResults)
	return bestMatch, found
}
