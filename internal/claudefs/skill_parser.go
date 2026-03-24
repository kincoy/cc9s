package claudefs

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func parseDirectorySkill(root SkillDiscoveryRoot, dirPath string) SkillResource {
	name := filepath.Base(dirPath)
	entryFile := filepath.Join(dirPath, "SKILL.md")
	resource := SkillResource{
		Name:              name,
		Path:              dirPath,
		Source:            root.Source,
		Scope:             sourceScope(root.Source),
		Kind:              root.Kind,
		ProjectName:       root.ProjectName,
		ProjectPath:       root.ProjectPath,
		PluginName:        root.PluginName,
		PluginInstallMode: root.PluginInstallMode,
		Status:            SkillStatusReady,
		EntryFile:         entryFile,
		Shape:             SkillShapeDirectory,
	}

	var reasons []string
	info, err := os.Stat(entryFile)
	if err != nil {
		resource.Status = SkillStatusInvalid
		reasons = append(reasons, "Missing entry file SKILL.md")
	} else if info.IsDir() {
		resource.Status = SkillStatusInvalid
		reasons = append(reasons, "Entry file path points to a directory")
	}

	if resource.Status == SkillStatusReady {
		summary, summaryReasons := extractSkillSummary(entryFile)
		resource.Summary = summary
		if len(summaryReasons) > 0 {
			reasons = append(reasons, summaryReasons...)
		}
		if summary == "" {
			reasons = append(reasons, "No readable summary text found")
		} else {
			reasons = append(reasons, fmt.Sprintf("Recognized directory %s entry", strings.ToLower(string(root.Kind))))
		}
	}

	resource.AssociatedFiles = collectAssociatedFiles(dirPath, entryFile)
	resource.ValidationReasons = normalizeSkillReasons(resource.Status, reasons)
	return resource
}

func parseSingleFileSkill(root SkillDiscoveryRoot, filePath string) SkillResource {
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	resource := SkillResource{
		Name:              name,
		Path:              filePath,
		Source:            root.Source,
		Scope:             sourceScope(root.Source),
		Kind:              root.Kind,
		ProjectName:       root.ProjectName,
		ProjectPath:       root.ProjectPath,
		PluginName:        root.PluginName,
		PluginInstallMode: root.PluginInstallMode,
		Status:            SkillStatusReady,
		EntryFile:         filePath,
		Shape:             SkillShapeSingleFile,
	}

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		resource.Status = SkillStatusInvalid
		resource.ValidationReasons = []string{"Skill file is not readable"}
		return resource
	}

	summary, reasons := extractSkillSummary(filePath)
	resource.Summary = summary
	if summary == "" {
		resource.Status = SkillStatusInvalid
		reasons = append(reasons, "No readable summary text found")
	} else {
		reasons = append(reasons, fmt.Sprintf("Recognized standalone %s file", strings.ToLower(string(root.Kind))))
	}

	resource.ValidationReasons = normalizeSkillReasons(resource.Status, reasons)
	return resource
}

func sourceScope(source SkillSource) SkillScope {
	if source == SkillSourceProject {
		return SkillScopeProject
	}
	if source == SkillSourcePlugin {
		return SkillScopePlugin
	}
	return SkillScopeUser
}

func normalizeSkillReasons(status SkillStatus, reasons []string) []string {
	clean := make([]string, 0, len(reasons))
	seen := make(map[string]struct{}, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(reason)
		if reason == "" {
			continue
		}
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		clean = append(clean, reason)
	}

	if len(clean) == 0 {
		if status == SkillStatusReady {
			return []string{"Skill structure is ready"}
		}
		return []string{"Skill structure is invalid"}
	}

	return clean
}

func collectAssociatedFiles(dirPath, entryFile string) []string {
	files := make([]string, 0)
	_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if path == entryFile {
			return nil
		}
		rel, relErr := filepath.Rel(dirPath, path)
		if relErr != nil {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	slices.Sort(files)
	return files
}

func extractSkillSummary(path string) (string, []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", []string{fmt.Sprintf("Failed to read entry file: %v", err)}
	}

	content := string(data)
	if description := extractFrontmatterDescription(content); description != "" {
		return description, []string{"Summary derived from description metadata"}
	}

	if paragraph := extractFirstReadableParagraph(content); paragraph != "" {
		return paragraph, []string{"Summary derived from markdown description"}
	}

	return "", nil
}

func extractFrontmatterDescription(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			break
		}
		if !strings.HasPrefix(strings.ToLower(line), "description:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, line[:len("description:")]))
		if idx := strings.Index(line, ":"); idx >= 0 {
			value = strings.TrimSpace(line[idx+1:])
		}
		value = strings.Trim(value, `"'`)
		return value
	}

	return ""
}

func extractFirstReadableParagraph(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inFrontmatter := false
	inCodeFence := false
	var paragraph []string
	lineNo := 0

	flush := func() string {
		if len(paragraph) == 0 {
			return ""
		}
		return strings.Join(paragraph, " ")
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if lineNo == 0 && trimmed == "---" {
			inFrontmatter = true
			lineNo++
			continue
		}
		if inFrontmatter {
			if trimmed == "---" {
				inFrontmatter = false
			}
			lineNo++
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			inCodeFence = !inCodeFence
			lineNo++
			continue
		}
		if inCodeFence {
			lineNo++
			continue
		}

		if trimmed == "" {
			if summary := flush(); summary != "" {
				return summary
			}
			paragraph = nil
			lineNo++
			continue
		}

		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "|") ||
			strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "```") {
			if summary := flush(); summary != "" {
				return summary
			}
			paragraph = nil
			lineNo++
			continue
		}

		paragraph = append(paragraph, trimmed)
		lineNo++
	}

	return flush()
}
