package commands

import (
	"fmt"
	"io"
	"strings"
)

// CommandInfo describes one command to the shell: the word offered after
// "stackdrift", and the words offered after the command itself.
type CommandInfo struct {
	Name    string
	Help    string
	Options []OptionInfo
}

// OptionInfo is a word that can follow a command, either a flag such as --yes
// or a fixed argument such as a shell name.
type OptionInfo struct {
	Name string
	Help string
}

func CompletionShells() []string {
	return []string{"bash", "zsh", "fish", "powershell"}
}

func Completion(out io.Writer, args []string) error {
	shells := strings.Join(CompletionShells(), "|")
	if len(args) == 0 {
		return fmt.Errorf("usage: stackdrift completion <%s>", shells)
	}

	scripts := map[string]string{
		"bash":       bashCompletion,
		"zsh":        zshCompletion,
		"fish":       fishCompletion,
		"powershell": powershellCompletion,
	}

	script, ok := scripts[strings.ToLower(args[0])]
	if !ok {
		return fmt.Errorf("unsupported shell: %s (choose %s)", args[0], shells)
	}

	_, err := io.WriteString(out, script)
	return err
}

// CompleteLine answers the shell's question of what could follow the command
// line typed so far. The whole line arrives as a single argument, which keeps
// each shell stub to a few lines and sidesteps their differing rules for
// passing an empty word. Candidates are printed one per line, with an optional
// description after a tab.
func CompleteLine(out io.Writer, commands []CommandInfo, line string) error {
	words := strings.Fields(line)

	// A trailing space puts the cursor on a fresh word, so nothing is half
	// typed and everything valid at that position is offered.
	prefix := ""
	if !endsInSpace(line) && len(words) > 0 {
		prefix = words[len(words)-1]
		words = words[:len(words)-1]
	}
	if len(words) > 0 {
		words = words[1:]
	}

	if len(words) == 0 {
		for _, c := range commands {
			suggest(out, prefix, c.Name, c.Help)
		}
		return nil
	}

	for _, c := range commands {
		if c.Name != words[0] {
			continue
		}
		for _, option := range c.Options {
			if alreadyTyped(words[1:], option.Name) {
				continue
			}
			suggest(out, prefix, option.Name, option.Help)
		}
	}
	return nil
}

func endsInSpace(line string) bool {
	return line != strings.TrimRight(line, " \t")
}

func alreadyTyped(words []string, name string) bool {
	for _, word := range words {
		if word == name {
			return true
		}
	}
	return false
}

func suggest(out io.Writer, prefix, name, help string) {
	if !strings.HasPrefix(name, prefix) {
		return
	}
	if help == "" {
		fmt.Fprintln(out, name)
		return
	}
	fmt.Fprintf(out, "%s\t%s\n", name, help)
}

// Every script asks the binary what to offer rather than listing the commands
// itself, so an installed completion keeps working after the CLI gains a
// command and is never left describing an older release.

const bashCompletion = `_stackdrift() {
    local line="${COMP_LINE:0:${COMP_POINT}}"
    local IFS=$'\n'
    local candidate
    COMPREPLY=()
    for candidate in $(stackdrift __complete "$line" 2>/dev/null); do
        COMPREPLY+=("${candidate%%$'\t'*}")
    done
}
complete -F _stackdrift stackdrift
`

const zshCompletion = `#compdef stackdrift

_stackdrift() {
    local -a typed candidates
    local candidate
    typed=("${(@)words[1,CURRENT]}")
    for candidate in ${(f)"$(stackdrift __complete "${(j: :)typed}" 2>/dev/null)"}; do
        candidates+=("${candidate/$'\t'/:}")
    done
    _describe stackdrift candidates
}

if [ "$funcstack[1]" = "_stackdrift" ]; then
    _stackdrift "$@"
else
    compdef _stackdrift stackdrift
fi
`

const fishCompletion = `function __stackdrift_complete
    set -l line (commandline --cut-at-cursor --current-process)
    stackdrift __complete "$line" 2>/dev/null
end

complete -c stackdrift -f -a "(__stackdrift_complete)"
`

const powershellCompletion = `Register-ArgumentCompleter -Native -CommandName stackdrift -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # The syntax tree stops at the last word, so an empty word being completed
    # is what tells us the cursor has moved on to a new one.
    $line = $commandAst.ToString()
    if (-not $wordToComplete) { $line = "$line " }

    stackdrift __complete $line 2>$null | ForEach-Object {
        $parts = $_ -split "\t", 2
        $tip = if ($parts.Count -gt 1) { $parts[1] } else { $parts[0] }
        [System.Management.Automation.CompletionResult]::new($parts[0], $parts[0], 'ParameterValue', $tip)
    }
}
`
