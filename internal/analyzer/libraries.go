package analyzer

// getKnownLibraries returns a database of known external libraries
func getKnownLibraries() []ExternalLibrary {
	return []ExternalLibrary{
		{
			Name:       "libmicrohttpd",
			HeaderName: "microhttpd.h",
			LinkerFlag: "-lmicrohttpd",
			PkgConfig:  "libmicrohttpd",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "libmicrohttpd",
					IncludePath: "/opt/homebrew/opt/libmicrohttpd/include",
					LibPath:     "/opt/homebrew/opt/libmicrohttpd/lib",
				},
				"linux": {
					PackageName: "libmicrohttpd-dev",
					IncludePath: "/usr/include",
					LibPath:     "/usr/lib",
				},
				"windows": {
					PackageName: "libmicrohttpd",
				},
			},
		},
		{
			Name:       "libcurl",
			HeaderName: "curl/curl.h",
			LinkerFlag: "-lcurl",
			PkgConfig:  "libcurl",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "curl",
					IncludePath: "/opt/homebrew/opt/curl/include",
					LibPath:     "/opt/homebrew/opt/curl/lib",
				},
				"linux": {
					PackageName: "libcurl4-openssl-dev",
					IncludePath: "/usr/include",
					LibPath:     "/usr/lib",
				},
				"windows": {
					PackageName: "curl",
				},
			},
		},
		{
			Name:       "sqlite3",
			HeaderName: "sqlite3.h",
			LinkerFlag: "-lsqlite3",
			PkgConfig:  "sqlite3",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "sqlite",
					IncludePath: "/opt/homebrew/opt/sqlite/include",
					LibPath:     "/opt/homebrew/opt/sqlite/lib",
				},
				"linux": {
					PackageName: "libsqlite3-dev",
					IncludePath: "/usr/include",
					LibPath:     "/usr/lib",
				},
				"windows": {
					PackageName: "sqlite",
				},
			},
		},
		{
			Name:       "openssl",
			HeaderName: "openssl/ssl.h",
			LinkerFlag: "-lssl -lcrypto",
			PkgConfig:  "openssl",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "openssl",
					IncludePath: "/opt/homebrew/opt/openssl/include",
					LibPath:     "/opt/homebrew/opt/openssl/lib",
				},
				"linux": {
					PackageName: "libssl-dev",
					IncludePath: "/usr/include",
					LibPath:     "/usr/lib",
				},
				"windows": {
					PackageName: "openssl",
				},
			},
		},
		{
			Name:       "zlib",
			HeaderName: "zlib.h",
			LinkerFlag: "-lz",
			PkgConfig:  "zlib",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "zlib",
				},
				"linux": {
					PackageName: "zlib1g-dev",
				},
				"windows": {
					PackageName: "zlib",
				},
			},
		},
		{
			Name:       "pcre",
			HeaderName: "pcre.h",
			LinkerFlag: "-lpcre",
			PkgConfig:  "libpcre",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "pcre",
				},
				"linux": {
					PackageName: "libpcre3-dev",
				},
				"windows": {
					PackageName: "pcre",
				},
			},
		},
		{
			Name:       "pthread",
			HeaderName: "pthread.h",
			LinkerFlag: "-pthread",
			Platforms: map[string]PlatformPackage{
				"darwin":  {PackageName: ""},
				"linux":   {PackageName: ""},
				"windows": {PackageName: "pthreads-win32"},
			},
		},
		{
			Name:       "libxml2",
			HeaderName: "libxml/parser.h",
			LinkerFlag: "-lxml2",
			PkgConfig:  "libxml-2.0",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "libxml2",
				},
				"linux": {
					PackageName: "libxml2-dev",
				},
				"windows": {
					PackageName: "libxml2",
				},
			},
		},
		{
			Name:       "libuv",
			HeaderName: "uv.h",
			LinkerFlag: "-luv",
			PkgConfig:  "libuv",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "libuv",
				},
				"linux": {
					PackageName: "libuv1-dev",
				},
				"windows": {
					PackageName: "libuv",
				},
			},
		},
		{
			Name:       "libpng",
			HeaderName: "png.h",
			LinkerFlag: "-lpng",
			PkgConfig:  "libpng",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "libpng",
				},
				"linux": {
					PackageName: "libpng-dev",
				},
				"windows": {
					PackageName: "libpng",
				},
			},
		},
		{
			Name:       "SDL2",
			HeaderName: "SDL2/SDL.h",
			LinkerFlag: "-lSDL2",
			PkgConfig:  "sdl2",
			Platforms: map[string]PlatformPackage{
				"darwin": {
					PackageName: "sdl2",
				},
				"linux": {
					PackageName: "libsdl2-dev",
				},
				"windows": {
					PackageName: "sdl2",
				},
			},
		},
	}
}
