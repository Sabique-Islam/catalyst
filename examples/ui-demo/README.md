# Catalyst UI Demo

## Features

### 1. Main Menu (`RunMainMenu()`)
- Interactive selection menu
- Navigate with arrow keys
- Press Enter to select
- Press Ctrl+C to cancel

### 2. Init Wizard (`RunInitWizard()`)
- Step-by-step project configuration
- Collects:
  - Project name
  - Source files (C source files)
  - Linux dependencies (pthread, m)
  - Windows dependencies (ws2_32.lib)
  - Resource URLs and paths
- Generates a YAML configuration

## How to Run

From this directory:

```bash
go run main.go
```

Or from the project root:

```bash
go run ./examples/ui-demo/main.go
```

## What to Expect

1. **Main Menu**: You'll see a menu with options to Build, Run, Clean, Init, or Exit
2. **If you select "Init"**: The wizard will guide you through creating a `catalyst.yml` configuration
3. **Interactive Prompts**: 
   - Use arrow keys to navigate menus
   - Type text and press Enter for input fields
   - Leave inputs empty (just press Enter) to skip/finish loops
4. **Output**: After the wizard, you'll see the generated YAML and can optionally save it

## Example Workflow

1. Select "Init (Create catalyst.yml)" from the main menu
2. Enter a project name (e.g., "my-c-project")
3. Add source files (e.g., "src/main.c", "src/utils.c")
4. Press Enter on an empty line to finish adding sources
5. Choose which Linux libraries to include
6. Choose which Windows libraries to include
7. Optionally add resource URLs
8. View the generated YAML configuration
9. Choose whether to save it to `catalyst.yml`

## Tips

- Press **Ctrl+C** at any time to cancel
- Leave fields **empty** (just press Enter) to skip or finish loops
- Use **arrow keys** to navigate Yes/No selections
- The wizard handles errors gracefully and provides clear feedback


## Run :

- go run main.go

note: in this directory