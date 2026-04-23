package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/term"
)

func normalize_path(path string) string {
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, `\ `, " ")
	return filepath.Clean(path)
}

func resolvePath(cwd, input string) string {
	input = normalize_path(input)
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
	hh           oo      oo         iii       
	hh           oo      oo vv   vv     zzzzz 
	hhhhhh   oooooo  oooooo  vv vv  iii   zz  
	hh   hh oo   oo oo   oo   vvv   iii  zz   
	hh   hh  oooooo  oooooo    v    iii zzzzz 
	
	hooviz!!`)
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

func longestCommonPrefix(items []string) string {
	if len(items) == 0 {
		return ""
	}
	prefix := items[0]
	for _, s := range items[1:] {
		for !strings.HasPrefix(s, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func commandCompletions(prefix string) []string {
	var replCommands = []string{"help", "pwd", "ls", "inspect", "cd", "quit", "exit"}
	out := make([]string, 0, len(replCommands))
	for _, c := range replCommands {
		if strings.HasPrefix(c, prefix) {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

func pathCompletions(rawArg, cwd string, folderMap map[string]*DirNode) []string {
	typed := strings.TrimSpace(rawArg)
	typedNorm := normalize_path(typed)

	var parentAbs, namePrefix string
	if typed == "" || strings.HasSuffix(typed, string(os.PathSeparator)) {
		parentAbs = resolvePath(cwd, typedNorm)
		namePrefix = ""
	} else {
		parentPart := filepath.Dir(typedNorm)
		if parentPart == "." {
			parentPart = ""
		}
		parentAbs = resolvePath(cwd, parentPart)
		namePrefix = filepath.Base(typedNorm)
	}

	out := make([]string, 0, 16)
	seen := make(map[string]struct{})
	for p := range folderMap {
		if p == parentAbs {
			continue
		}
		if filepath.Dir(p) != parentAbs {
			continue
		}
		base := filepath.Base(p)
		if !strings.HasPrefix(base, namePrefix) {
			continue
		}

		candidate := p
		if !filepath.IsAbs(typed) {
			if rel, err := filepath.Rel(cwd, p); err == nil {
				candidate = rel
			}
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	sort.Strings(out)
	return out
}

func applyTabCompletion(line []rune, cursor int, cwd string, folderMap map[string]*DirNode) ([]rune, int) {
	left := string(line[:cursor])
	right := string(line[cursor:])

	if strings.TrimSpace(left) == "" {
		return line, cursor
	}

	if !strings.Contains(left, " ") {
		prefix := strings.TrimSpace(left)
		matches := commandCompletions(prefix)
		if len(matches) == 0 {
			return line, cursor
		}
		if len(matches) == 1 {
			newLeft := matches[0]
			return []rune(newLeft + right), len([]rune(newLeft))
		}
		lcp := longestCommonPrefix(matches)
		if len(lcp) > len(prefix) {
			newLeft := lcp
			return []rune(newLeft + right), len([]rune(newLeft))
		}
		fmt.Print("\r\n")
		fmt.Println(strings.Join(matches, "  "))
		return line, cursor
	}

	// Complete path for cd/inspect.
	cmd, rawArg, _ := strings.Cut(left, " ")
	cmd = strings.TrimSpace(cmd)
	if cmd != "cd" && cmd != "inspect" {
		return line, cursor
	}

	matches := pathCompletions(rawArg, cwd, folderMap)
	if len(matches) == 0 {
		return line, cursor
	}

	typed := strings.TrimSpace(rawArg)
	if len(matches) == 1 {
		completedArg := matches[0]
		if cmd == "cd" && !strings.HasSuffix(completedArg, string(os.PathSeparator)) {
			completedArg += string(os.PathSeparator)
		}
		newLeft := cmd + " " + completedArg
		return []rune(newLeft + right), len([]rune(newLeft))
	}

	lcp := longestCommonPrefix(matches)
	if len(lcp) > len(typed) {
		newLeft := cmd + " " + lcp
		return []rune(newLeft + right), len([]rune(newLeft))
	}

	fmt.Print("\r\n")
	for _, m := range matches {
		fmt.Println(m)
	}
	return line, cursor
}

func redrawPrompt(prompt string, line []rune, cursor int) {
	fmt.Printf("\r\033[2K%s%s", prompt, string(line))
	if back := len(line) - cursor; back > 0 {
		fmt.Printf("\033[%dD", back)
	}
}

func executeCommand(line, cwd string, folderMap map[string]*DirNode) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return cwd, false
	}

	cmd, rawArg, _ := strings.Cut(line, " ")
	cmd = strings.TrimSpace(cmd)
	rawArg = strings.TrimSpace(rawArg)

	switch cmd {
	case "help":
		printHelp()
	case "quit", "exit":
		fmt.Println("quitting hddviz...")
		return cwd, true
	case "pwd":
		fmt.Println(cwd)
	case "ls":
		printNode(cwd, folderMap)
	case "inspect":
		if rawArg == "" {
			fmt.Println("usage: inspect <path>")
			return cwd, false
		}
		target := resolvePath(cwd, rawArg)
		if _, ok := folderMap[target]; !ok {
			fmt.Println("directory not found in scan:", target)
			return cwd, false
		}
		printNode(target, folderMap)
	case "cd":
		if rawArg == "" {
			fmt.Println("usage: cd <path>")
			return cwd, false
		}
		target := resolvePath(cwd, rawArg)
		if _, ok := folderMap[target]; !ok {
			fmt.Println("directory not found in scan:", target)
			return cwd, false
		}
		cwd = target
	default:
		fmt.Println("unknown command:", cmd)
	}

	return cwd, false
}

func runREPLLineMode(folderMap map[string]*DirNode, cwd string) {
	scanner := bufio.NewScanner(os.Stdin)
	prompt := "hddviz> "

	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			fmt.Println()
			return
		}
		var quit bool
		cwd, quit = executeCommand(scanner.Text(), cwd, folderMap)
		if quit {
			return
		}
	}
}

func stdinFDInt() (int, error) {
	fd := os.Stdin.Fd()
	maxInt := int(^uint(0) >> 1)
	if fd > uintptr(maxInt) {
		return 0, fmt.Errorf("stdin fd out of int range: %d", fd)
	}
	return int(fd), nil
}

func runREPL(folderMap map[string]*DirNode, cwd string) {
	fmt.Println()
	printWelcome()
	fmt.Println("Entering interactive mode. Type 'help' for commands.")

	fd, err := stdinFDInt()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: invalid stdin fd, fallback to line mode:", err)
		runREPLLineMode(folderMap, cwd)
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: raw mode unavailable, fallback to line mode:", err)
		runREPLLineMode(folderMap, cwd)
		return
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	reader := bufio.NewReader(os.Stdin)
	prompt := "hddviz> "
	line := make([]rune, 0, 128)
	cursor := 0

	fmt.Print(prompt)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			fmt.Print("\r\n")
			return
		}

		switch r {
		case '\r', '\n':
			fmt.Print("\r\n")
			input := string(line)
			line = line[:0]
			cursor = 0

			_ = term.Restore(fd, oldState)
			var quit bool
			cwd, quit = executeCommand(input, cwd, folderMap)
			if quit {
				return
			}
			if _, err := term.MakeRaw(fd); err != nil {
				fmt.Fprintln(os.Stderr, "failed to re-enter raw mode:", err)
				return
			}
			fmt.Print(prompt)

		case '\t':
			line, cursor = applyTabCompletion(line, cursor, cwd, folderMap)
			redrawPrompt(prompt, line, cursor)

		case 127, 8: // backspace
			if cursor > 0 {
				line = append(line[:cursor-1], line[cursor:]...)
				cursor--
				redrawPrompt(prompt, line, cursor)
			}

		case 27: // escape sequences (arrows/delete)
			b1, err := reader.ReadByte()
			if err != nil || b1 != '[' {
				continue
			}
			b2, err := reader.ReadByte()
			if err != nil {
				continue
			}
			switch b2 {
			case 'D': // left
				if cursor > 0 {
					cursor--
					redrawPrompt(prompt, line, cursor)
				}
			case 'C': // right
				if cursor < len(line) {
					cursor++
					redrawPrompt(prompt, line, cursor)
				}
			case '3': // delete: ESC [ 3 ~
				_, _ = reader.ReadByte()
				if cursor < len(line) {
					line = append(line[:cursor], line[cursor+1:]...)
					redrawPrompt(prompt, line, cursor)
				}
			}

		case 3: // Ctrl+C
			fmt.Print("^C\r\n")
			return

		case 4: // Ctrl+D
			if len(line) == 0 {
				fmt.Print("\r\n")
				return
			}

		default:
			if r < 32 {
				continue
			}
			line = append(line[:cursor], append([]rune{r}, line[cursor:]...)...)
			cursor++
			redrawPrompt(prompt, line, cursor)
		}
	}
}
