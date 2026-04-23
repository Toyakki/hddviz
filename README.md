# hddviz CLI

## Demo video
https://github.com/user-attachments/assets/d022c779-ca6d-41db-8f7b-a9597c6529ef

## Motivation
Unlike Linux CLI, macOS does not have a convenient command line tool to visualize hard disk usage. This tool aims to visualize hard disk usage from your command line everywhere. It also provides a REPL interface where you can enter commands to navigate directories and visualize folder-based disk usage.

hddviz is a file-system agnostic CLI tool that provides a simple test-based listing of disk usage, making it easier for users to identify large files and directories. This tool scans all the files under the Home directory and lets you enter the REPL interface to navigate through the directories. The following commands are so far supported in the REPL interface.

## How to install (macOS only so far.)
1. Download the tar.gz file from the releases page.
2. Extract it:
```bash
tar -xzf hddviz_*_darwin_*.tar.gz
```

3. There are multiple ways to add the hddviz binary as your CLI tool. 

a. If you have a homebrew, then try running
```bash
# Apple Silicon
install -m 0700 hddviz /opt/homebrew/bin/hddviz

# Intel
install -m 0700 hddviz /usr/local/bin/hddviz
```
b. If you don't have homebrew, you can move the hddviz binary to a directory in your CLI tool path. You need to add it to PATH first:
```bash
mkdir -p "$HOME/.local/bin"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
source "$HOME/.zshrc"
install -m 0700 hddviz "$HOME/.local/bin/hddviz"
```

c. Install from source code:

Go to src/ folder and compile the code using `make build`.
Run
```bash
go install .
```
This installs `hddviz` into:

- `$GOBIN` (if set), or
- `$(go env GOPATH)/bin` (default: `$HOME/go/bin`)

If `hddviz` is not found, add Go bin to your PATH:

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> "$HOME/.zshrc"
source "$HOME/.zshrc"
```

4. Verify:
```bash
hddviz --help
```

That's it. You can just run hddviz from your terminal. 

## Supported commands in REPL:
You can see the following commands by running help in the REPL interface. But here is a more detailed description of the supported commands in the REPL interface:
- `help`: Display the list of available commands and their descriptions.
- `ls`: List the top-K largest subdirectories under your current path. 
- `inspect <path>`: Show summary for a path without changing the current directory. <path> can be a relative path (including '../') or an absolute path.
- `cd <path>`: Change the current directory to the specified folder.
- `pwd`: Print the current directory path.
- `quit`: Exit the REPL interface.


## Technical parts:
- [x] Uses a heap for top-K largest subdirs listing
- [x] Uses recursion to scan all directories.
- [x] Supports concurrent scanning (sometimes overcounts file usage, so it is only used for estimation.)
- [x] REPL interface.
- [x] Added both blackbox and whitebox tests for scanning logic.

## Selfish reasons for me to build this:
- [x] Nice to learn golang for me.
- [x] Motivate me to buy a Linux machine so I don't have to maintain this project and just run ncdu. 

## TODOs for me and any devs
### Public deployment
- [x] Write a doc for easier cli command installation and usage. The one that uses go install . 
- [x] Review .goreleaser.yml for brew public release.

### Development:

#### Code quality:
- [x] Better error handling and logging
  - [x] log vs fprinln vs fmt.errorf vs errors. 
- [x] Add some unit tests for serial scanning.

#### Features:
Scanning features:
- [x] Create goroutines for concurrent scanning of directories to improve performance.
  - [x] Store the prototyped version in .idea folder 
  - [x] Write unit tests for them.
  - [x] Add a command line flag to enable/disable concurrent scanning.


REPL features:
- [x] A fancier welcome screen.
- [x] Tab completion for path.
- [x] I also need to make sure that both cd and ls support the folderName with space in it. Use both forward and backward slashes like 'cd Application\ Support'.

## Potential extensions
- Support for Windows?
- File-system based scanning system to improve performance. For example, WizTree uses the Master File Table (MFT) to quickly scan NTFS file systems, which can be significantly faster than traditional scanning methods.
