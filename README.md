# Catalyst


## Problem Statement :

C programs rely on platform-specific tools, external libraries, and configurations, making them difficult to run across different operating systems.
Porting from macOS/Linux to Windows (and vice versa) often needs manual setup, dependency fixes, and code changes.
This process is highly time-consuming, prone to errors, and discourages developers from maintaining cross-platform compatibility.
Our aim is to create a solution that automates environment setup and dependency management, thus improving software portability and developer efficiency.


## Solution Description :

We are building “Catalyst”, a portability framework for execution of C programs across operating systems.

- Automatic environment setup: configures compiler, paths, and dependencies for the detected OS.
- Unified build translator: Adapts Makefiles, linking rules, and scripts dynamically.
- Lightweight sandbox: Ensures build conditions without Docker or VMs.
- Resource fetcher: Automatically downloads required external assets (for example big file like GloVe embeddings).

## Installation & Usage

### Installing Dependencies and External Resources

Catalyst can automatically install system dependencies and download external files defined in your `catalyst.yml` configuration file.

#### Install Everything (Default)
```bash
# Installs both system dependencies and downloads external resources
catalyst install
```

#### Install Only System Dependencies
```bash
# Installs only system package dependencies (apt, brew, pacman, etc.)
catalyst install --deps-only
```

#### Download Only External Resources
```bash
# Downloads only external files/resources (skips system dependencies)
catalyst install --resources-only
```

### Configuration Format

#### System Dependencies

Define OS-specific system dependencies that will be installed using the system package manager:

```yaml
dependencies:
  # Linux dependencies (installed via apt/dnf/pacman/zypper)
  linux:
    - "gcc"
    - "make"
    - "libcurl4-openssl-dev"
    - "libssl-dev"
  
  # macOS dependencies (installed via Homebrew)
  darwin:
    - "gcc"
    - "curl"
    - "openssl"
  
  # Windows dependencies (supports multiple package managers)
  windows:
    - "gcc"           # -> mingw (choco) or MSYS2.MSYS2 (winget) or gcc (scoop)
    - "make"          # -> make (choco) or GnuWin32.Make (winget) or make (scoop)  
    - "curl"          # -> curl (choco) or cURL.cURL (winget) or curl (scoop)
    - "git"           # -> git (choco) or Git.Git (winget) or git (scoop)
    - "ws2_32.lib"    # System libraries are automatically skipped
    - "user32.lib"    # System libraries are automatically skipped
```

#### External Resources

Define external files to be downloaded before building:

```yaml
# Global resources (downloaded for all platforms)
resources:
  - url: "https://example.com/data/config.json"
    path: "data/config.json"
  - url: "https://github.com/user/repo/releases/download/v1.0/library.a"
    path: "lib/library.a"
  - url: "https://nlp.stanford.edu/data/glove.6B.50d.txt"
    path: "embeddings/glove.6B.50d.txt"
```

#### Platform-Specific Overrides

Override dependencies and resources for specific platforms:

```yaml
platforms:
  # Linux-specific configuration  
  linux:
    dependencies:
      - "gcc"
      - "libcurl4-openssl-dev"
    resources:
      - url: "https://example.com/linux-specific-file.so"
        path: "lib/linux-specific.so"
    
  # macOS-specific configuration
  darwin:
    dependencies:
      - "gcc"
      - "curl"
    resources:
      - url: "https://example.com/macos-framework.framework"
        path: "frameworks/macos-framework.framework"
  
  # Windows-specific configuration
  windows:
    dependencies:
      - "mingw"
      - "ws2_32.lib"
    resources:
      - url: "https://example.com/windows-library.dll"
        path: "bin/windows-library.dll"
```

### Features

- **Smart Resource Management**: Files are only downloaded if they don't already exist locally
- **Cross-Platform Package Management**: Automatically detects and uses the appropriate package manager
  - **Linux**: apt-get, dnf, yum, pacman, zypper
  - **macOS**: Homebrew (brew)  
  - **Windows**: Windows Package Manager (winget), Chocolatey (choco), Scoop
- **Platform-Specific Overrides**: Different dependencies and resources for different operating systems
- **Progress Tracking**: Shows download progress and file information
- **Error Handling**: Comprehensive error reporting with recovery suggestions
- **Directory Creation**: Automatically creates necessary directories for downloaded files

### Examples

#### Basic C Project with External Libraries
```yaml
project_name: "network-app"
sources: ["src/main.c", "src/network.c"]

dependencies:
  linux: ["gcc", "libcurl4-openssl-dev"]
  darwin: ["gcc", "curl"]

resources:
  - url: "https://curl.se/ca/cacert.pem"
    path: "certs/ca-bundle.crt"
```

#### Machine Learning Project with Large Datasets
```yaml
project_name: "ml-classifier"
sources: ["src/main.c", "src/ml.c"]

dependencies:
  linux: ["gcc", "libblas-dev"]
  darwin: ["gcc", "openblas"]

resources:
  - url: "https://nlp.stanford.edu/data/glove.6B.100d.txt"
    path: "data/embeddings.txt"
  - url: "https://example.com/trained-model.bin"
    path: "models/classifier.bin"
```

#### Game Development with Platform-Specific Assets
```yaml
project_name: "cross-platform-game"

platforms:
  linux:
    dependencies: ["gcc", "libsdl2-dev"]
    resources:
      - url: "https://example.com/linux-renderer.so"
        path: "lib/renderer.so"
  
  windows:
    dependencies: ["gcc", "make", "ws2_32.lib"]  # gcc->mingw, system libs skipped
    resources:
      - url: "https://example.com/windows-renderer.dll"
        path: "bin\\renderer.dll"  # Windows path separators supported
```

#### Windows Development Example
```yaml
project_name: "windows-native-app"
sources: ["src/main.c", "src/windows_gui.c"]

dependencies:
  windows:
    - "gcc"           # Maps to mingw (choco) or MSYS2.MSYS2 (winget)
    - "make"          # Maps to make (choco) or GnuWin32.Make (winget)
    - "curl"          # Maps to curl (choco) or cURL.cURL (winget)
    - "user32.lib"    # System library - automatically skipped
    - "gdi32.lib"     # System library - automatically skipped

resources:
  - url: "https://example.com/app-config.json"
    path: "config\\app.json"        # Windows-style paths
  - url: "https://example.com/assets.zip"
    path: "assets/resources.zip"    # Mixed separators work too

platforms:
  windows:
    resources:
      - url: "https://example.com/windows-specific.dll"
        path: "bin\\native.dll"
```

### Windows Package Manager Priority

Catalyst automatically detects Windows package managers in this priority order:
1. **Windows Package Manager (winget)** - Modern, built-in Windows 10/11 package manager
2. **Chocolatey (choco)** - Popular third-party package manager  
3. **Scoop** - Lightweight package manager for developers
