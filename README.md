# hddviz CLI

## Demo video
Coming soon!

## Motivation
Unlike Linux CLI, macOS does not have a convenient command line tool to visualize hard disk usage. This tool aims to visualize hard disk usage from your command line everywhere, without the need to install a GUI tool. 

hddviz is a file-system agnostic CLI tool that provides a simple test-based listing of disk usage, making it easier for users to identify large files and directories. It uses a combination of heap and recursive algorithm to efficiently scan and visualize disk usage.
Currently, it is only tested on macOS, but I plan to add support for Windows in the future, if I have time.

## Why is this project not cool?
- Reinventing the wheel: There are already existing tools like ncdu, WizTree, and DaisyDisk that provide similar functionality with more features and better performance.
- Not a GUI tool
- Not a file-system based scanning system. 

## Why is this project cool?
- Cross-platform: Works on both Windows and macOS.
- No third-party dependenciees: Pure Go implementation with built-in libraries.
- Interoperability: Provides a simple text-based listing.
- Performance: Uses heaps and dfs to scan and visualize disk usage. Potentially adding a goroitune for concurrent scanning of directories to further improve performance. (.idea/)
- Ease of use: Provides a REPL interface for interactive exploration of disk usage.

## Why is it good for me?
- Nice to learn golang
- Motivate me to buy a Linux machine so i don't have to maintain this and just run ncdu.

## How to use this cli.
There are two ways of using this CLI tool. Run locally or install it using homebrew.
Run the following command to install it using homebrew.
```bash
brew tap Toyakki/hddviz
brew install --cask hddviz
```
After installation, run:
```bash
hddviz
```
If hddviz is not run, verify where it resolves from with:
```bash
which hddviz
ls -l "$(brew --prefix)/bin/hddviz"
```
If the symlink is not pointing back into the staged cask under $(brew --caskroom), please try to resolve the symlink issue.



Reference: https://go.dev/doc/tutorial/compile-install

## TODOs
### Public deployment
- [] Write a doc for easier cli command installation and usage. The one that uses go install . 
- [] Review .goreleaser.yml for brew public release.


### Development:

#### Code quality:
- [x] Better error handling and logging
  - [x] log vs fprinln vs fmt.errorf vs errors. 
- [x] Add some unit tests for serial scanning.

#### Features:
Scanning features:
- [x] Create goroutines for concurrent scanning of directories to improve performance.
  - [x] Store the prototyped version in .idea folder 
  - [ ] Write unit tests for them.
  - [ ] Add a command line flag to enable/disable concurrent scanning.
  - [ ] Add a command line flag to set the maximum number of goroutines to use for scanning.
  - [ ] Add a fallback mechanism to sequential scanning if the number of goroutines exceeds a certain threshold to prevent overwhelming the system.

REPL features:
- [ ] A fancier welcome screen.
- [ ] Tab completion for path.
- [ ] Add a bar graph viz of disk usage for each directory.

## Potential extensions
- File-system based scanning system to improve performance. For example, WizTree uses the Master File Table (MFT) to quickly scan NTFS file systems, which can be significantly faster than traditional scanning methods.
