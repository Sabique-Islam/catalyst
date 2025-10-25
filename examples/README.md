# Catalyst Examples - Testing Guide

## Overview

This directory contains example projects demonstrating Catalyst's cross-platform C development capabilities. All examples are configured to automatically install dependencies and compile on Windows, Linux, and macOS.

## Windows Users - MSYS2 Integration

**Good news!** Catalyst now automatically installs development libraries via MSYS2 on Windows. You don't need to manually run pacman commands anymore.

### What Happens Automatically:
1. When you run `catalyst build` or `catalyst install`, Catalyst will:
   - Install MSYS2 via winget (if not already installed)
   - Automatically install required libraries (sqlite3, curl, jansson, etc.) via MSYS2's pacman
   - Compile your program with proper linking

### First Time Setup:
- Ensure you have **winget** (Windows Package Manager) installed - it comes with Windows 11 or via the Microsoft Store on Windows 10
- Run any example's build command, and Catalyst will handle the rest!

### Windows Compatibility Notes:
- ✅ **arc**, **menu**, **Research Assistant**: Fully working on Windows
- ⚠️ **BMI Tracker**: Uses ncurses which has limited Windows support. Best experienced on Linux/macOS or WSL.

### Example Output:
```
Installing dependencies for windows: [msys2 sqlite3]
Installing msys2 with winget...
  → Skipped: Package may already be installed
     MSYS2 appears to be already installed

Installing development libraries via MSYS2 pacman: [sqlite3]
Running MSYS2 pacman: pacman -S --noconfirm mingw-w64-ucrt-x86_64-sqlite3
...
Dependencies installed successfully!
```

## Quick Reference

For testing the /arc ->
### Windows PowerShell Commands

From the `examples/arc` directory:

| Action | Command |
|--------|---------|
| View config | `Get-Content catalyst.yml` |
| Install deps | `..\..\catalyst.exe install --deps-only` |
| Build | `..\..\catalyst.exe build` |
| Run | `.\arc.exe` |
| Clean | `..\..\catalyst.exe clean` |

### Linux/macOS Commands

From the `examples/arc` directory:

| Action | Command |
|--------|---------|
| View config | `cat catalyst.yml` |
| Install deps | `../../catalyst install --deps-only` |
| Build | `../../catalyst build` |
| Run | `./arc` |
| Clean | `../../catalyst clean` |

