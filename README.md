# hddviz CLI

## Motivation
Unlike Linux CLI, macOS and windows do not have a convinient command line tool to visualize hard disk usage. This project is my first golang pet project to create a cross-platform CLI tool to visualize hard disk usage. 

hddviz is a file-system agnostic CLI tool that provides a simple test-based listing of disk usage, making it easier for users to identify large files and directories. It uses a combination of heap data structures and depth-first search (DFS) algorithms to efficiently scan and visualize disk usage.
You know, this project is sort of peak because it uses heaps and dfs. 

## Why is this project not cool?
- Reinventing the wheel: There are already existing tools like ncdu, WizTree, and DaisyDisk that provide similar functionality with more features and better performance.
- Not a GUI tool
- Not a file-system based scanning system. 

## Why is this project cool?
- Cross-platform: Works on both Windows and macOS.
- No third-party dependenciees: Pure Go implementation with built-in libraries.
- Interoperability: Provides a simple text-based listing.
- Performance: Uses heaps and dfs to scan and visualize disk usage. Potentially adding a goroitune for concurrent scanning of directories to further improve performance.
- Ease of use: Provides a REPL interface for interactive exploration of disk usage.

Why is it good for me?
- Nice to learn golang
- Motivate me to buy a Linux machine so i don't have to maintain this and just run ncdu.



## How to use this cli.

## TODOs

- Read GHA documentation quickstart
- Read goreleaser documentation quickstart
- Read two Dev articles on CI/CD for setting up


- Setup CI/CD pipeline 
  - Dependabot
  - goreleaser
  - Github actions for testing and building
  - Linting

- Config.json instead of yaml for configuration
- Add goroutine for concurrent scanning of directories to improve performance

## After turning into Public
- Add a license
- Add a PR config, protect main branch. 

Code quality:
- Better error handling and logging
  - log vs fprinln vs fmt.errorf vs errors package
  - Good guide: https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/error_technique/

- Add some unit tests

New features:
- A fancier welcome screen.
- Relative path support for cd, inspect and ls.
- Tab completion for path.

Deployment
- Releases
- support for windows


## Future extensions
- File-system based scanning system to improve performance. For example, WizTree uses the Master File Table (MFT) to quickly scan NTFS file systems, which can be significantly faster than traditional scanning methods.