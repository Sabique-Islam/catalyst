import os
import platform
import subprocess
import sys

# Path to your C program
c_program = "program.c"
output_exe = "program.exe" if platform.system() == "Windows" else "program"

# Path to MSYS2 UCRT64 on Windows
msys2_path = r"F:\msys2\ucrt64.exe"

def run_command(cmd, use_msys=False):
    if use_msys:
        # Run the command inside MSYS2 UCRT64
        full_cmd = f'{msys2_path} -c "{cmd}"'
    else:
        full_cmd = cmd
    print(f"Running: {full_cmd}")
    result = subprocess.run(full_cmd, shell=True)
    if result.returncode != 0:
        print(f"Command failed: {cmd}")
        sys.exit(1)

def install_sqlite():
    system = platform.system()
    if system == "Windows":
        print("Detected Windows, using MSYS2 UCRT64")
        # Install sqlite3 using pacman inside MSYS2
        run_command("pacman -S --needed mingw-w64-ucrt-x86_64-sqlite3", use_msys=True)
    elif system == "Linux":
        print("Detected Linux")
        run_command("sudo apt update && sudo apt install -y libsqlite3-dev")
    elif system == "Darwin":
        print("Detected macOS")
        run_command("brew install sqlite")
    else:
        print(f"Unsupported OS: {system}")
        sys.exit(1)

def build_program():
    system = platform.system()
    if system == "Windows":
        run_command(f"gcc {c_program} -o {output_exe} -lsqlite3", use_msys=True)
    else:
        run_command(f"gcc {c_program} -o {output_exe} -lsqlite3")

def launch_program():
    print(f"Launching {output_exe}...")
    if platform.system() == "Windows":
        run_command(f"./{output_exe}", use_msys=True)
    else:
        run_command(f"./{output_exe}")

if __name__ == "__main__":
    install_sqlite()
    build_program()
    launch_program()
