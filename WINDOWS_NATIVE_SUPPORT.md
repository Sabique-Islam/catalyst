# Native Windows Support for Catalyst

Catalyst now supports native Windows development without requiring WSL2 or MSYS2! Here's what's new:

## Supported Windows Compilers

Catalyst automatically detects and uses the best available compiler:

1. **Microsoft Visual C++ (cl.exe)** - Recommended for Windows
   - Part of Visual Studio Build Tools
   - Best performance and Windows integration
   - Automatic MSVC flag conversion

2. **LLVM/Clang** - Modern, cross-platform
   - Excellent standards compliance
   - Good performance and diagnostics
   - Works well with vcpkg

3. **GCC variants**:
   - TDM-GCC: Easy to install, good compatibility
   - MinGW-w64: Full featured GCC port
   - W64DevKit: Portable development kit

## Installation Options

### Using winget (Recommended)
```bash
# Install Visual Studio Build Tools (includes cl.exe)
winget install Microsoft.VisualStudio.2022.BuildTools

# Or install Clang/LLVM
winget install LLVM.LLVM

# Or install TDM-GCC
winget install TDM-GCC.TDM-GCC

# For library management
winget install Microsoft.vcpkg
```

### Using Chocolatey
```bash
# Install Visual Studio Build Tools
choco install visualstudio2022buildtools -y

# Or install LLVM
choco install llvm -y

# Or install MinGW
choco install mingw -y
```

## Features

### Automatic Compiler Detection
- No manual configuration required
- Falls back through compiler priorities
- Helpful error messages with installation instructions

### Native Flag Conversion
- Automatically converts GCC flags to MSVC equivalents
- Example: `-fopenmp` → `/openmp`, `-O2` → `/O2`
- Handles include paths, defines, and library linking

### Package Manager Integration
- **winget**: Modern Windows package manager
- **Chocolatey**: Popular community package manager  
- **Scoop**: Developer-focused package manager
- **vcpkg**: Microsoft's C++ library manager

### Library Support
Libraries are automatically mapped to Windows equivalents:
- `openmp` → LLVM OpenMP or MinGW OpenMP
- `curl` → Windows curl
- `sqlite3` → SQLite for Windows
- `jansson` → Native Windows builds via vcpkg

### vcpkg Integration
If `VCPKG_ROOT` environment variable is set:
- Automatically adds include/lib paths
- Seamless library integration
- Works with GCC and Clang compilers

## Getting Started

1. **Install a compiler** (see options above)
2. **Create your project**:
   ```bash
   catalyst init
   ```
3. **Build your project**:
   ```bash
   catalyst build
   ```

Catalyst will automatically:
- Detect your compiler
- Install required libraries
- Add appropriate compilation flags
- Generate the Windows executable

## OpenMP Example

```c
#include <stdio.h>
#include <omp.h>

int main() {
    printf("Native Windows OpenMP!\n");
    
    #pragma omp parallel
    {
        int id = omp_get_thread_num();
        printf("Thread %d\n", id);
    }
    
    return 0;
}
```

This will compile with the correct OpenMP flags automatically!

## No More WSL2 or MSYS2 Required!

- ✅ Native Windows executables
- ✅ Full Windows API access
- ✅ Better performance
- ✅ Simpler setup
- ✅ IDE integration
- ✅ Windows-specific optimizations

Ready to build native Windows C/C++ applications with ease!