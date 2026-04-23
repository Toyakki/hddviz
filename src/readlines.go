// Borrowed the code from https://github.com/chzyer/readline
package main

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

var runes = Runes{}
var TabWidth = 4

type Runes struct{}

func (Runes) EqualRune(a, b rune, fold bool) bool {
	if a == b {
		return true
	}
	if !fold {
		return false
	}
	if a > b {
		a, b = b, a
	}
	if b < utf8.RuneSelf && 'A' <= a && a <= 'Z' {
		if b == a+'a'-'A' {
			return true
		}
	}
	return false
}

func (r Runes) EqualRuneFold(a, b rune) bool {
	return r.EqualRune(a, b, true)
}

func (r Runes) EqualFold(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if r.EqualRuneFold(a[i], b[i]) {
			continue
		}
		return false
	}

	return true
}

func (Runes) Equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (rs Runes) IndexAllBckEx(r, sub []rune, fold bool) int {
	for i := len(r) - len(sub); i >= 0; i-- {
		found := true
		for j := 0; j < len(sub); j++ {
			if !rs.EqualRune(r[i+j], sub[j], fold) {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

// Search in runes from end to front
func (rs Runes) IndexAllBck(r, sub []rune) int {
	return rs.IndexAllBckEx(r, sub, false)
}

// Search in runes from front to end
func (rs Runes) IndexAll(r, sub []rune) int {
	return rs.IndexAllEx(r, sub, false)
}

func (rs Runes) IndexAllEx(r, sub []rune, fold bool) int {
	for i := 0; i < len(r); i++ {
		found := true
		if len(r[i:]) < len(sub) {
			return -1
		}
		for j := 0; j < len(sub); j++ {
			if !rs.EqualRune(r[i+j], sub[j], fold) {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

func (Runes) Index(r rune, rs []rune) int {
	for i := 0; i < len(rs); i++ {
		if rs[i] == r {
			return i
		}
	}
	return -1
}

func (Runes) ColorFilter(r []rune) []rune {
	newr := make([]rune, 0, len(r))
	for pos := 0; pos < len(r); pos++ {
		if r[pos] == '\033' && r[pos+1] == '[' {
			idx := runes.Index('m', r[pos+2:])
			if idx == -1 {
				continue
			}
			pos += idx + 2
			continue
		}
		newr = append(newr, r[pos])
	}
	return newr
}

var zeroWidth = []*unicode.RangeTable{
	unicode.Mn,
	unicode.Me,
	unicode.Cc,
	unicode.Cf,
}

var doubleWidth = []*unicode.RangeTable{
	unicode.Han,
	unicode.Hangul,
	unicode.Hiragana,
	unicode.Katakana,
}

func (Runes) Width(r rune) int {
	if r == '\t' {
		return TabWidth
	}
	if unicode.IsOneOf(zeroWidth, r) {
		return 0
	}
	if unicode.IsOneOf(doubleWidth, r) {
		return 2
	}
	return 1
}

func (Runes) WidthAll(r []rune) (length int) {
	for i := 0; i < len(r); i++ {
		length += runes.Width(r[i])
	}
	return
}

func (Runes) Backspace(r []rune) []byte {
	return bytes.Repeat([]byte{'\b'}, runes.WidthAll(r))
}

func (Runes) Copy(r []rune) []rune {
	n := make([]rune, len(r))
	copy(n, r)
	return n
}

func (Runes) HasPrefixFold(r, prefix []rune) bool {
	if len(r) < len(prefix) {
		return false
	}
	return runes.EqualFold(r[:len(prefix)], prefix)
}

func (Runes) HasPrefix(r, prefix []rune) bool {
	if len(r) < len(prefix) {
		return false
	}
	return runes.Equal(r[:len(prefix)], prefix)
}

func (Runes) Aggregate(candicate [][]rune) (same []rune, size int) {
	for i := 0; i < len(candicate[0]); i++ {
		for j := 0; j < len(candicate)-1; j++ {
			if i >= len(candicate[j]) || i >= len(candicate[j+1]) {
				goto aggregate
			}
			if candicate[j][i] != candicate[j+1][i] {
				goto aggregate
			}
		}
		size = i + 1
	}
aggregate:
	if size > 0 {
		same = runes.Copy(candicate[0][:size])
		for i := 0; i < len(candicate); i++ {
			n := runes.Copy(candicate[i])
			copy(n, n[size:])
			candicate[i] = n[:len(n)-size]
		}
	}
	return
}

func (Runes) TrimSpaceLeft(in []rune) []rune {
	firstIndex := len(in)
	for i, r := range in {
		if !unicode.IsSpace(r) {
			firstIndex = i
			break
		}
	}
	return in[firstIndex:]
}

// Caller type for dynamic completion
type DynamicCompleteFunc func(string) []string

type PrefixCompleterInterface interface {
	Print(prefix string, level int, buf *bytes.Buffer)
	Do(line []rune, pos int) (newLine [][]rune, length int)
	GetName() []rune
	GetChildren() []PrefixCompleterInterface
	SetChildren(children []PrefixCompleterInterface)
}

type DynamicPrefixCompleterInterface interface {
	PrefixCompleterInterface
	IsDynamic() bool
	GetDynamicNames(line []rune) [][]rune
}

type PrefixCompleter struct {
	Name     []rune
	Dynamic  bool
	Callback DynamicCompleteFunc
	Children []PrefixCompleterInterface
}

func (p *PrefixCompleter) Tree(prefix string) string {
	buf := bytes.NewBuffer(nil)
	p.Print(prefix, 0, buf)
	return buf.String()
}

func Print(p PrefixCompleterInterface, prefix string, level int, buf *bytes.Buffer) {
	if strings.TrimSpace(string(p.GetName())) != "" {
		buf.WriteString(prefix)
		if level > 0 {
			buf.WriteString("├")
			buf.WriteString(strings.Repeat("─", (level*4)-2))
			buf.WriteString(" ")
		}
		buf.WriteString(string(p.GetName()) + "\n")
		level++
	}
	for _, ch := range p.GetChildren() {
		ch.Print(prefix, level, buf)
	}
}

func (p *PrefixCompleter) Print(prefix string, level int, buf *bytes.Buffer) {
	Print(p, prefix, level, buf)
}

func (p *PrefixCompleter) IsDynamic() bool {
	return p.Dynamic
}

func (p *PrefixCompleter) GetName() []rune {
	return p.Name
}

func (p *PrefixCompleter) GetDynamicNames(line []rune) [][]rune {
	var names = [][]rune{}
	for _, name := range p.Callback(string(line)) {
		names = append(names, []rune(name+" "))
	}
	return names
}

func (p *PrefixCompleter) GetChildren() []PrefixCompleterInterface {
	return p.Children
}

func (p *PrefixCompleter) SetChildren(children []PrefixCompleterInterface) {
	p.Children = children
}

func NewPrefixCompleter(pc ...PrefixCompleterInterface) *PrefixCompleter {
	return PcItem("", pc...)
}

func PcItem(name string, pc ...PrefixCompleterInterface) *PrefixCompleter {
	name += " "
	return &PrefixCompleter{
		Name:     []rune(name),
		Dynamic:  false,
		Children: pc,
	}
}

func PcItemDynamic(callback DynamicCompleteFunc, pc ...PrefixCompleterInterface) *PrefixCompleter {
	return &PrefixCompleter{
		Callback: callback,
		Dynamic:  true,
		Children: pc,
	}
}

func (p *PrefixCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	return doInternal(p, line, pos, line)
}

func Do(p PrefixCompleterInterface, line []rune, pos int) (newLine [][]rune, offset int) {
	return doInternal(p, line, pos, line)
}

func doInternal(p PrefixCompleterInterface, line []rune, pos int, origLine []rune) (newLine [][]rune, offset int) {
	line = runes.TrimSpaceLeft(line[:pos])
	goNext := false
	var lineCompleter PrefixCompleterInterface
	for _, child := range p.GetChildren() {
		childNames := make([][]rune, 1)

		childDynamic, ok := child.(DynamicPrefixCompleterInterface)
		if ok && childDynamic.IsDynamic() {
			childNames = childDynamic.GetDynamicNames(origLine)
		} else {
			childNames[0] = child.GetName()
		}

		for _, childName := range childNames {
			if len(line) >= len(childName) {
				if runes.HasPrefix(line, childName) {
					if len(line) == len(childName) {
						newLine = append(newLine, []rune{' '})
					} else {
						newLine = append(newLine, childName)
					}
					offset = len(childName)
					lineCompleter = child
					goNext = true
				}
			} else {
				if runes.HasPrefix(childName, line) {
					newLine = append(newLine, childName[len(line):])
					offset = len(line)
					lineCompleter = child
				}
			}
		}
	}

	if len(newLine) != 1 {
		return
	}

	tmpLine := make([]rune, 0, len(line))
	for i := offset; i < len(line); i++ {
		if line[i] == ' ' {
			continue
		}

		tmpLine = append(tmpLine, line[i:]...)
		return doInternal(lineCompleter, tmpLine, len(tmpLine), origLine)
	}

	if goNext {
		return doInternal(lineCompleter, nil, 0, origLine)
	}
	return
}
