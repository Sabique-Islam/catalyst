
---

#Catalyst

> Team New Folder(1)


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

### Windows Users - MSYS2 Auto-Installation

**New Feature!** Catalyst now automatically manages MSYS2 development libraries on Windows.

When you run `catalyst build` or `catalyst install` on Windows:
1. **Automatic MSYS2 Setup**: Installs MSYS2 via winget if not already present
2. **Smart Package Detection**: Identifies which packages need MSYS2 (development libraries) vs winget (applications)
3. **Seamless Installation**: Automatically installs development packages like `sqlite3`, `curl`, `jansson` via MSYS2's pacman
4. **No Manual Commands**: No need to manually run `pacman -S` commands

**Example**: When building a project with `curl` and `jansson` dependencies:
```bash
catalyst build
# Automatically installs MSYS2, then runs:
# pacman -S mingw-w64-ucrt-x86_64-curl mingw-w64-ucrt-x86_64-jansson
```

**Supported MSYS2 Packages**: `sqlite3`, `curl`, `jansson`, `ncurses`, `openssl`, and other development libraries

**Windows Compatibility Warnings**: Catalyst automatically detects packages with known Windows compatibility issues (like `ncurses`, `X11`, `GTK`, `ALSA`) and provides helpful warnings with alternative suggestions. This helps you avoid spending time on libraries that won't work properly on Windows.

**Note**: Compiled binaries need MSYS2 DLLs in PATH to run. Add `C:\msys64\ucrt64\bin` to your PATH or use the provided run scripts.

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
    - "msys2"         # Base MSYS2 installation (via winget)
    - "gcc"           # -> mingw (choco) or MSYS2.MSYS2 (winget) or gcc (scoop)
    - "make"          # -> make (choco) or GnuWin32.Make (winget) or make (scoop)  
    - "curl"          # -> Installed via MSYS2 pacman (development library)
    - "jansson"       # -> Installed via MSYS2 pacman (development library)
    - "sqlite3"       # -> Installed via MSYS2 pacman (development library)
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

flags:
  - -lcurl
  - -ljansson
  - -lsqlite3

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

**Note**: Development libraries (`curl`, `jansson`, `sqlite3`, etc.) are automatically installed via MSYS2's pacman with the correct `mingw-w64-ucrt-x86_64-*` package names. You don't need to specify the full MSYS2 package names!

### Windows Package Manager Priority

Catalyst automatically detects Windows package managers in this priority order:
1. **Windows Package Manager (winget)** - Modern, built-in Windows 10/11 package manager
2. **Chocolatey (choco)** - Popular third-party package manager  
3. **Scoop** - Lightweight package manager for developers

### Windows Development Libraries via MSYS2

For development libraries (headers + libraries), Catalyst uses **MSYS2's pacman** package manager:

**Automatically Handled Packages**:
- `curl`, `libcurl4-openssl-dev` → `mingw-w64-ucrt-x86_64-curl`
- `jansson`, `libjansson-dev` → `mingw-w64-ucrt-x86_64-jansson`
- `sqlite3`, `libsqlite3-dev` → `mingw-w64-ucrt-x86_64-sqlite3`
- `ncurses`, `libncurses-dev` → `mingw-w64-ucrt-x86_64-ncurses`
- `openssl`, `libssl-dev` → `mingw-w64-ucrt-x86_64-openssl`

**How It Works**:
1. Catalyst installs MSYS2 base via winget
2. Detects which packages are development libraries
3. Automatically runs `pacman -S` with the correct UCRT64 package names
4. Sets up proper include and library paths

**Running Compiled Programs**:
Programs compiled with MSYS2 libraries need the DLLs in PATH:
```powershell
# Add MSYS2 UCRT64 bin directory to PATH
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"
.\build\your-program.exe
```

Or create a wrapper script (e.g., `run.bat`):
```batch
@echo off
set PATH=C:\msys64\ucrt64\bin;%PATH%
build\your-program.exe %*
```

### Windows Compatibility Warnings

Catalyst includes an intelligent compatibility detection system that warns you about packages with known Windows issues:

**Automatically Detected Issues**:
- **ncurses**: Limited Windows support, suggests PDCurses or WSL
- **X11**: Not available on Windows, suggests Win32 API, SDL2, GLFW, or Qt
- **GTK**: Limited Windows support, suggests Qt or wxWidgets
- **ALSA/PulseAudio**: Linux-specific audio, suggests PortAudio or Windows Audio APIs

**How It Works**:
When you try to install a problematic package on Windows, Catalyst displays:
```
⚠️  WARNING: Windows Compatibility Issue Detected
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Package: ncurses
Issue: ncurses has limited Windows support. The MSYS2 port 
       has incomplete symbol exports and may cause linking errors.

💡 Suggestion:
   PDCurses (Public Domain Curses) - a Windows-compatible 
   curses implementation

📖 More Info:
   Consider using PDCurses or running your application in 
   WSL (Windows Subsystem for Linux) for full ncurses support.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

This **doesn't prevent** the installation but provides valuable guidance to help you make informed decisions about cross-platform compatibility.
