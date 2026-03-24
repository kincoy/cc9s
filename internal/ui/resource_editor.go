package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func preferredEditor() string {
	if visual := strings.TrimSpace(os.Getenv("VISUAL")); visual != "" {
		return visual
	}
	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return editor
	}
	return "vim"
}

func buildFileEditorCommand(targetPath, workDir string) (*exec.Cmd, error) {
	editor := preferredEditor()
	parts, err := splitCommandLine(editor)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("no editor configured")
	}

	cmd := exec.Command(parts[0], append(parts[1:], targetPath)...)
	cmd.Dir = workDir
	return cmd, nil
}

func splitCommandLine(input string) ([]string, error) {
	var (
		args         []string
		current      strings.Builder
		inSingle     bool
		inDouble     bool
		escapeNext   bool
		sawCharacter bool
	)

	flush := func() {
		if !sawCharacter {
			return
		}
		args = append(args, current.String())
		current.Reset()
		sawCharacter = false
	}

	for _, r := range input {
		switch {
		case escapeNext:
			current.WriteRune(r)
			sawCharacter = true
			escapeNext = false
		case r == '\\' && !inSingle:
			escapeNext = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			sawCharacter = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
			sawCharacter = true
		case unicode.IsSpace(r) && !inSingle && !inDouble:
			flush()
		default:
			current.WriteRune(r)
			sawCharacter = true
		}
	}

	if escapeNext {
		current.WriteRune('\\')
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote in editor command")
	}

	flush()
	return args, nil
}
