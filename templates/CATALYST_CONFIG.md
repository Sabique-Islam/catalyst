# Catalyst Configuration Guide

## catalyst.yml File Structure

The `catalyst.yml` file is the heart of your Catalyst project. It defines how your C/C++ project should be built and what dependencies it needs across different platforms.

## Basic Structure

```yaml
project_name: "your-project-name"
description: "Project description" 
author: "Your Name <email@example.com>"

sources:
  - "src/main.c"
  - "src/utils.c"

dependencies:
  linux: ["gcc", "libssl-dev"]
  darwin: ["gcc", "openssl"]  
  windows: ["mingw", "openssl"]

resources: []
env: {}
```

## Configuration Fields

### Required Fields

- **`project_name`**: Name of your project
- **`sources`**: List of C/C++ source files to compile
- **`dependencies`**: OS-specific package dependencies

### Optional Fields

- **`description`**: Project description
- **`author`**: Author information
- **`resources`**: External files to download
- **`env`**: Environment variables
- **`platforms`**: Platform-specific overrides
- **`created_at`**: Auto-generated timestamp

## Dependencies by Platform

### Linux (installed via system package manager)
```yaml
dependencies:
  linux:
    - "gcc"                    # GNU Compiler Collection
    - "make"                   # Build tool
    - "libcurl4-openssl-dev"   # HTTP client library
    - "libssl-dev"             # SSL/TLS library
    - "libsqlite3-dev"         # SQLite database
    - "pthread"                # POSIX threads
    - "m"                      # Math library
```

### macOS (installed via Homebrew)
```yaml
dependencies:
  darwin:
    - "gcc"        # GNU Compiler Collection
    - "curl"       # HTTP client library  
    - "openssl"    # SSL/TLS library
    - "sqlite"     # SQLite database
```

### Windows (installed via Chocolatey)
```yaml
dependencies:
  windows:
    - "mingw"      # MinGW compiler
    - "curl"       # HTTP client library
    - "sqlite"     # SQLite database
    - "ws2_32.lib" # Windows Sockets library
```

## External Resources

Download files before building:

```yaml
resources:
  - url: "https://example.com/data.json"
    path: "data/data.json"
  - url: "https://github.com/user/repo/releases/download/v1.0/lib.a"  
    path: "lib/libexample.a"
```

## Environment Variables

Set environment variables during build:

```yaml
env:
  DEBUG: "1"
  LOG_LEVEL: "info" 
  DATA_DIR: "./data"
```

## Platform-Specific Overrides

Override settings for specific platforms:

```yaml
platforms:
  windows:
    dependencies:
      - "mingw"
      - "ws2_32.lib"
    resources:
      - url: "https://example.com/windows.dll"
        path: "bin/windows.dll"
        
  linux:
    dependencies:
      - "gcc"
      - "libcurl4-openssl-dev"
```

## Common Use Cases

### Simple Hello World
```yaml
project_name: "hello-world"
sources: ["src/main.c"]
dependencies:
  linux: ["gcc"]
  darwin: ["gcc"]
  windows: ["mingw"]
resources: []
```

### Networking Application
```yaml
project_name: "http-client"
sources: ["src/main.c", "src/http.c"]
dependencies:
  linux: ["gcc", "libcurl4-openssl-dev"]
  darwin: ["gcc", "curl"]
  windows: ["mingw", "curl", "ws2_32.lib"]
```

### Database Application  
```yaml
project_name: "todo-app"
sources: ["src/main.c", "src/db.c"]
dependencies:
  linux: ["gcc", "libsqlite3-dev"]
  darwin: ["gcc", "sqlite"]
  windows: ["mingw", "sqlite"]
```

## Template Files

Use these template files as starting points:

- **`catalyst.yml.template`** - Comprehensive template with all features
- **`catalyst.minimal.yml`** - Minimal template for simple projects  
- **`catalyst.example.yml`** - Working example for Hello World

## Usage Commands

```bash
# Install dependencies only
catalyst install

# Build project  
catalyst build src/main.c src/utils.c

# Build and run
catalyst run src/main.c src/utils.c

# Clean build artifacts
catalyst clean

# Initialize new project (interactive)
catalyst init
```