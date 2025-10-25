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
  
  # Windows dependencies (installed via Chocolatey)
  windows:
    - "mingw"
    - "curl"
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
- **Cross-Platform Package Management**: Automatically detects and uses the appropriate package manager (apt, brew, pacman, dnf, yum, zypper, chocolatey)
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
    dependencies: ["mingw", "sdl2"]
    resources:
      - url: "https://example.com/windows-renderer.dll"
        path: "bin/renderer.dll"
```
