package build

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
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

	switch osType {
	case "linux":
		pkgMgr, err := detectLinuxPackageManager()
		if err != nil {
			return err
		}

		var args []string
		switch pkgMgr {
		case "apt-get":
			args = append([]string{"install", "-y"}, deps...)
			err = runCommand("sudo", append([]string{"apt-get"}, args...)...)
		case "dnf", "yum":
			args = append([]string{"install", "-y"}, deps...)
			err = runCommand("sudo", append([]string{pkgMgr}, args...)...)
		case "pacman":
			args = append([]string{"-S", "--noconfirm"}, deps...)
			err = runCommand("sudo", append([]string{"pacman"}, args...)...)
		case "zypper":
			args = append([]string{"install", "-y"}, deps...)
			err = runCommand("sudo", append([]string{"zypper"}, args...)...)
		}

		if err != nil {
			return fmt.Errorf("failed installing with %s: %w", pkgMgr, err)
		}

	case "darwin":
		if _, err := exec.LookPath("brew"); err != nil {
			return errors.New("homebrew not found — install it from https://brew.sh/")
		}
		args := append([]string{"install"}, deps...)
		if err := runCommand("brew", args...); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}

	case "windows":
		if _, err := exec.LookPath("choco"); err != nil {
			return errors.New("chocolatey not found — install it from https://chocolatey.org/install")
		}
		args := append([]string{"install", "-y"}, deps...)
		if err := runCommand("choco", args...); err != nil {
			return fmt.Errorf("choco install failed: %w", err)
		}

	default:
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	fmt.Println("✅ All dependencies installed successfully.")
	return nil
}

// runCommand is a helper function to execute a command with arguments
func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
