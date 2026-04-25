package claudefs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ScanSkills scans the local Claude Code skill roots and returns discovered skills.
func ScanSkills() SkillScanResult {
	start := time.Now()
	result := SkillScanResult{
		Skills: make([]SkillResource, 0),
	}

	roots, err := getSkillDiscoveryRoots()
	if err != nil {
		result.Err = err
		return result
	}

	skills, err := scanSkillRoots(roots)
	if err != nil {
		result.Err = err
		return result
	}

	for _, skill := range skills {
		if skill.Status == SkillStatusReady {
			result.ReadyCount++
		} else {
			result.InvalidCount++
		}
	}

	result.Skills = skills
	result.ScanDuration = time.Since(start)
	return result
}

func getSkillDiscoveryRoots() ([]SkillDiscoveryRoot, error) {
	roots := make([]SkillDiscoveryRoot, 0, 4)
	roots = append(roots, SkillDiscoveryRoot{
		Path:   SkillsDir(),
		Source: SkillSourceUser,
		Kind:   SkillKindSkill,
	})
	roots = append(roots, SkillDiscoveryRoot{
		Path:   CommandsDir(),
		Source: SkillSourceUser,
		Kind:   SkillKindCommand,
	})

	projects := ScanProjects()
	if projects.Err != nil {
		return nil, projects.Err
	}

	for _, project := range projects.Projects {
		if project.Path == "" {
			continue
		}
		roots = append(roots, SkillDiscoveryRoot{
			Path:        filepath.Join(project.Path, ".claude", "skills"),
			Source:      SkillSourceProject,
			Kind:        SkillKindSkill,
			ProjectName: project.Name,
			ProjectPath: project.Path,
		})
		roots = append(roots, SkillDiscoveryRoot{
			Path:        filepath.Join(project.Path, ".claude", "commands"),
			Source:      SkillSourceProject,
			Kind:        SkillKindCommand,
			ProjectName: project.Name,
			ProjectPath: project.Path,
		})
	}

	pluginRoots, err := getInstalledPluginDiscoveryRoots()
	if err != nil {
		return nil, err
	}
	roots = append(roots, pluginRoots...)

	return uniqueSkillDiscoveryRoots(roots), nil
}

func uniqueSkillDiscoveryRoots(roots []SkillDiscoveryRoot) []SkillDiscoveryRoot {
	unique := make([]SkillDiscoveryRoot, 0, len(roots))
	seenRoots := make(map[string]struct{}, len(roots))

	for _, root := range roots {
		cleanPath := filepath.Clean(root.Path)
		if _, exists := seenRoots[cleanPath]; exists {
			continue
		}
		seenRoots[cleanPath] = struct{}{}
		root.Path = cleanPath
		unique = append(unique, root)
	}

	return unique
}

func scanSkillRoots(roots []SkillDiscoveryRoot) ([]SkillResource, error) {
	skills := make([]SkillResource, 0)

	for _, root := range roots {
		entries, err := os.ReadDir(root.Path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, entry := range entries {
			path := filepath.Join(root.Path, entry.Name())
			if entry.IsDir() && root.Kind == SkillKindSkill {
				skills = append(skills, parseDirectorySkill(root, path))
				continue
			}

			if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				skills = append(skills, parseSingleFileSkill(root, path))
			}
		}
	}

	sort.SliceStable(skills, func(i, j int) bool {
		if sourcePriority(skills[i].Source) != sourcePriority(skills[j].Source) {
			return sourcePriority(skills[i].Source) < sourcePriority(skills[j].Source)
		}
		if ownerLabel(skills[i]) != ownerLabel(skills[j]) {
			return ownerLabel(skills[i]) < ownerLabel(skills[j])
		}
		if skills[i].Kind != skills[j].Kind {
			return skills[i].Kind < skills[j].Kind
		}
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})

	return skills, nil
}

type installedPluginsFile struct {
	Plugins map[string][]installedPluginEntry `json:"plugins"`
}

type installedPluginEntry struct {
	Scope       string `json:"scope"`
	InstallPath string `json:"installPath"`
	ProjectPath string `json:"projectPath"`
}

func getInstalledPluginDiscoveryRoots() ([]SkillDiscoveryRoot, error) {
	installedPath := PluginsInstalledPath()
	data, err := os.ReadFile(installedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var installed installedPluginsFile
	if err := json.Unmarshal(data, &installed); err != nil {
		return nil, err
	}

	roots := make([]SkillDiscoveryRoot, 0)
	for pluginKey, installs := range installed.Plugins {
		pluginName := installedPluginName(pluginKey)
		for _, install := range installs {
			if strings.TrimSpace(install.InstallPath) == "" {
				continue
			}
			roots = append(roots, pluginInstallRoots(pluginName, install)...)
		}
	}

	return roots, nil
}

func pluginInstallRoots(pluginName string, install installedPluginEntry) []SkillDiscoveryRoot {
	projectName := ""
	projectPath := strings.TrimSpace(install.ProjectPath)
	if projectPath != "" {
		projectName = ExtractProjectName(projectPath)
	}

	candidates := []struct {
		relPath string
		kind    SkillKind
	}{
		{"skills", SkillKindSkill},
		{"commands", SkillKindCommand},
		{filepath.Join(".claude", "skills"), SkillKindSkill},
		{filepath.Join(".claude", "commands"), SkillKindCommand},
	}

	roots := make([]SkillDiscoveryRoot, 0, len(candidates))
	for _, candidate := range candidates {
		roots = append(roots, SkillDiscoveryRoot{
			Path:              filepath.Join(install.InstallPath, candidate.relPath),
			Source:            SkillSourcePlugin,
			Kind:              candidate.kind,
			ProjectName:       projectName,
			ProjectPath:       projectPath,
			PluginName:        pluginName,
			PluginInstallMode: strings.TrimSpace(install.Scope),
		})
	}
	return roots
}

func installedPluginName(pluginKey string) string {
	if idx := strings.Index(pluginKey, "@"); idx >= 0 {
		return pluginKey[:idx]
	}
	return pluginKey
}

func sourcePriority(source SkillSource) int {
	switch source {
	case SkillSourceProject:
		return 0
	case SkillSourceUser:
		return 1
	case SkillSourcePlugin:
		return 2
	default:
		return 3
	}
}

func ownerLabel(skill SkillResource) string {
	switch skill.Source {
	case SkillSourceProject:
		return strings.ToLower(skill.ProjectName)
	case SkillSourcePlugin:
		return strings.ToLower(skill.PluginName)
	default:
		return ""
	}
}
