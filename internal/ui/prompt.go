package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Ask(prompt string) string {
	fmt.Print(prompt)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func Confirm(prompt string, def bool) bool {
	suffix := " [y/N] "
	if def {
		suffix = " [Y/n] "
	}
	answer := strings.ToLower(Ask(prompt + suffix))
	if answer == "" {
		return def
	}
	return answer == "y" || answer == "yes"
}

func AskInt(prompt string, min, max int) (int, bool) {
	answer := Ask(prompt)
	if answer == "" {
		return 0, false
	}
	value, err := strconv.Atoi(answer)
	if err != nil || value < min || value > max {
		return 0, false
	}
	return value, true
}

func Println(args ...any) {
	fmt.Println(args...)
}

func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}
