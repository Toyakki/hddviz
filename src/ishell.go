package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolvePath(cwd, input string) string {
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	return filepath.Clean(filepath.Join(cwd, input))
}

func printNode(path string, folderMap map[string]*DirNode) {
	node := folderMap[path]
	fmt.Println("path:", path)
	fmt.Println("name:", node.FolderName)
	fmt.Println("size:", sizeify(ByteSize(node.Size)))
	fmt.Println("top children:")
	for _, child := range node.TopKChildren {
		fmt.Println(" -", child, "(", sizeify(ByteSize(folderMap[child].Size)), ")")
	}
}

func printWelcome() {
	fmt.Println(`
	hh           dd      dd         iii       
	hh           dd      dd vv   vv     zzzzz 
	hhhhhh   dddddd  dddddd  vv vv  iii   zz  
	hh   hh dd   dd dd   dd   vvv   iii  zz   
	hh   hh  dddddd  dddddd    v    iii zzzzz 
	`)
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("help: Show this help")
	fmt.Println("pwd: Show current directory in REPL")
	fmt.Println("ls: Show current directory summary")
	fmt.Println("inspect <path> Show summary for a path")
	fmt.Println("cd <path> Change REPL current directory")
	fmt.Println("quit | exit Leave interactive mode")
}

func runREPL(folderMap map[string]*DirNode, cwd string) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println()
	printWelcome()
	fmt.Println("Entering interactive mode. Type 'help' for commands.")

	for {
		fmt.Printf("hddviz> ")
		if !scanner.Scan() {
			fmt.Println()
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		cmd := parts[0]
		switch cmd {
		case "help":
			printHelp()

		case "quit", "exit":
			fmt.Println("quitting hddviz...")
			return
		case "pwd":
			fmt.Println(cwd)
		case "ls":
			printNode(cwd, folderMap)
		case "inspect":
			if len(parts) < 2 {
				fmt.Println("usage: inspect <SUBFOLDER_NAME>")
				continue
			}
			target := resolvePath(cwd, parts[1])
			if _, ok := folderMap[target]; !ok {
				fmt.Println("directory not found in scan: ", target)
				continue
			}
			printNode(target, folderMap)
		case "cd":
			if len(parts) < 2 {
				fmt.Println("usage: cd <abs path or relative path >")
				continue
			}
			target := resolvePath(cwd, parts[1])
			if _, ok := folderMap[target]; !ok {
				fmt.Println("directory not found in scan: ", target)
				continue
			}
			cwd = target
		default:
			fmt.Println("unknown command: ", cmd)
		}
	}
}
