# Catalyst


## Problem Staement :

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
