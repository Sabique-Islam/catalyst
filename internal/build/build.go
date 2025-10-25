package build

import (
    "fmt"
    "os/exec"
    "runtime"
		"runtime"
		"strings"
)

// detectLinuxPackageManager tries to find a supported package manager on Linux.
func detectLinuxPackageManager() (string, error) {
    candidates := []string{"apt-get", "dnf", "yum", "pacman", "zypper"}
    for _, c := range candidates {
        if _, err := exec.LookPath(c); err == nil {
            return c, nil
        }
    }
    return "", errors.New("no supported linux package manager found (supported: apt-get, dnf, yum, pacman, zypper)")
}

func Install(dependencies map[string][]string) error {
    osType := runtime.GOOS
    deps, ok := dependencies[osType]
    if !ok {
        return fmt.Errorf("no dependencies listed for OS: %s", osType)
    }

    fmt.Printf("Installing dependencies for %s: %v\n", osType, deps)

    var installCmd *exec.Cmd

    switch osType {
    case "linux":
        args := append([]string{"install", "-y"}, deps...)
        installCmd = exec.Command("sudo", append([]string{"apt-get"}, args...)...)
    case "darwin":
        args := append([]string{"install"}, deps...)
        installCmd = exec.Command("brew", args...)
    case "windows":
        args := append([]string{"install"}, deps...)
        installCmd = exec.Command("choco", args...)
    default:
        return fmt.Errorf("unsupported OS: %s", osType)
    }

    installCmd.Stdout = nil
    installCmd.Stderr = nil

    if err := installCmd.Run(); err != nil {
        return fmt.Errorf("failed to install dependencies: %w", err)
    }

    fmt.Println("✅ All dependencies installed successfully.")
    return nil
}
