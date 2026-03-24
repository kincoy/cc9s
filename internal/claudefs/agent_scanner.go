package claudefs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ScanAgents scans the local Claude Code agent roots and returns discovered agents.
func ScanAgents() AgentScanResult {
	start := time.Now()
	result := AgentScanResult{
		Agents: make([]AgentResource, 0),
	}

	roots, err := getAgentDiscoveryRoots()
	if err != nil {
		result.Err = err
		return result
	}

	agents, err := scanAgentRoots(roots)
	if err != nil {
		result.Err = err
		return result
	}

	reconciled, err := reconcileAgentRecognition(agents)
	if err != nil {
		result.Err = err
		return result
	}

	for _, agent := range reconciled {
		if agent.Status == AgentStatusReady {
			result.ReadyCount++
		} else {
			result.InvalidCount++
		}
	}

	result.Agents = reconciled
	result.ScanDuration = time.Since(start)
	return result
}

func getAgentDiscoveryRoots() ([]AgentDiscoveryRoot, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	roots := make([]AgentDiscoveryRoot, 0, 4)
	roots = append(roots, AgentDiscoveryRoot{
		Path:   filepath.Join(homeDir, ".claude", "agents"),
		Source: AgentSourceUser,
	})

	projects := ScanProjects()
	if projects.Err != nil {
		return nil, projects.Err
	}

	for _, project := range projects.Projects {
		if project.Path == "" {
			continue
		}
		roots = append(roots, AgentDiscoveryRoot{
			Path:        filepath.Join(project.Path, ".claude", "agents"),
			Source:      AgentSourceProject,
			ProjectName: project.Name,
			ProjectPath: project.Path,
		})
	}

	pluginRoots, err := getInstalledAgentDiscoveryRoots(homeDir)
	if err != nil {
		return nil, err
	}
	roots = append(roots, pluginRoots...)

	return uniqueAgentDiscoveryRoots(roots), nil
}

func getInstalledAgentDiscoveryRoots(homeDir string) ([]AgentDiscoveryRoot, error) {
	installedPath := filepath.Join(homeDir, ".claude", "plugins", "installed_plugins.json")
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

	roots := make([]AgentDiscoveryRoot, 0)
	for pluginKey, installs := range installed.Plugins {
		pluginName := installedPluginName(pluginKey)
		for _, install := range installs {
			if strings.TrimSpace(install.InstallPath) == "" {
				continue
			}
			roots = append(roots, pluginAgentInstallRoots(pluginName, install)...)
		}
	}

	return roots, nil
}

func pluginAgentInstallRoots(pluginName string, install installedPluginEntry) []AgentDiscoveryRoot {
	projectPath := strings.TrimSpace(install.ProjectPath)
	projectName := ""
	if projectPath != "" {
		projectName = ExtractProjectName(projectPath)
	}

	candidates := []string{
		"agents",
		filepath.Join(".claude", "agents"),
	}

	roots := make([]AgentDiscoveryRoot, 0, len(candidates))
	for _, relPath := range candidates {
		roots = append(roots, AgentDiscoveryRoot{
			Path:              filepath.Join(install.InstallPath, relPath),
			Source:            AgentSourcePlugin,
			ProjectName:       projectName,
			ProjectPath:       projectPath,
			PluginName:        pluginName,
			PluginInstallMode: strings.TrimSpace(install.Scope),
		})
	}

	return roots
}

func uniqueAgentDiscoveryRoots(roots []AgentDiscoveryRoot) []AgentDiscoveryRoot {
	unique := make([]AgentDiscoveryRoot, 0, len(roots))
	seen := make(map[string]struct{}, len(roots))

	for _, root := range roots {
		cleanPath := filepath.Clean(root.Path)
		if _, exists := seen[cleanPath]; exists {
			continue
		}
		seen[cleanPath] = struct{}{}
		root.Path = cleanPath
		unique = append(unique, root)
	}

	return unique
}

func scanAgentRoots(roots []AgentDiscoveryRoot) ([]AgentResource, error) {
	agents := make([]AgentResource, 0)

	for _, root := range roots {
		entries, err := os.ReadDir(root.Path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				continue
			}

			agents = append(agents, parseAgentFile(root, filepath.Join(root.Path, entry.Name())))
		}
	}

	sort.SliceStable(agents, func(i, j int) bool {
		if agentSourcePriority(agents[i].Source) != agentSourcePriority(agents[j].Source) {
			return agentSourcePriority(agents[i].Source) < agentSourcePriority(agents[j].Source)
		}
		if agentOwnerLabel(agents[i]) != agentOwnerLabel(agents[j]) {
			return agentOwnerLabel(agents[i]) < agentOwnerLabel(agents[j])
		}
		return strings.ToLower(agents[i].Name) < strings.ToLower(agents[j].Name)
	})

	return agents, nil
}

func reconcileAgentRecognition(agents []AgentResource) ([]AgentResource, error) {
	if len(agents) == 0 {
		return agents, nil
	}

	needGlobal := false
	projectsNeedingLookup := make(map[string]struct{})
	for _, agent := range agents {
		switch agent.Source {
		case AgentSourceUser:
			needGlobal = true
		case AgentSourcePlugin:
			if agent.ProjectPath == "" {
				needGlobal = true
			} else {
				projectsNeedingLookup[agent.ProjectPath] = struct{}{}
			}
		case AgentSourceProject:
			if agent.ProjectPath != "" {
				projectsNeedingLookup[agent.ProjectPath] = struct{}{}
			}
		}
	}

	globalSnapshot := AgentRecognitionSnapshot{}
	var err error
	if needGlobal {
		globalSnapshot, err = lookupAgentRecognition("", "user")
		if err != nil {
			return nil, err
		}
	}

	projectSnapshots := make(map[string]AgentRecognitionSnapshot, len(projectsNeedingLookup))
	projectPaths := make([]string, 0, len(projectsNeedingLookup))
	for projectPath := range projectsNeedingLookup {
		projectPaths = append(projectPaths, projectPath)
	}
	sort.Strings(projectPaths)
	for _, projectPath := range projectPaths {
		snapshot, lookupErr := lookupAgentRecognition(projectPath, "project,local")
		if lookupErr != nil {
			return nil, lookupErr
		}
		projectSnapshots[projectPath] = snapshot
	}

	reconciled := make([]AgentResource, 0, len(agents))
	for _, agent := range agents {
		updated := agent
		reasons := append([]string(nil), agent.ValidationReasons...)

		snapshot, ok := snapshotForAgent(agent, globalSnapshot, projectSnapshots)
		recognized := ok && agentRecognizedBySnapshot(agent, snapshot)
		if recognized {
			updated.Status = AgentStatusReady
			reasons = []string{"Agent is recognized by Claude Code"}
		} else {
			updated.Status = AgentStatusInvalid
			reasons = append(reasons, "Agent file is not recognized by Claude Code")
		}

		updated.ValidationReasons = normalizedAgentReasons(updated.Status, reasons)
		updated.Availability = AgentAvailabilityResult{
			Status:  updated.Status,
			Reasons: updated.ValidationReasons,
		}
		reconciled = append(reconciled, updated)
	}

	return reconciled, nil
}

func snapshotForAgent(agent AgentResource, global AgentRecognitionSnapshot, projects map[string]AgentRecognitionSnapshot) (AgentRecognitionSnapshot, bool) {
	switch agent.Source {
	case AgentSourceUser:
		return global, true
	case AgentSourcePlugin:
		if agent.ProjectPath == "" {
			return global, true
		}
		snapshot, ok := projects[agent.ProjectPath]
		return snapshot, ok
	case AgentSourceProject:
		snapshot, ok := projects[agent.ProjectPath]
		return snapshot, ok
	default:
		return AgentRecognitionSnapshot{}, false
	}
}

func agentSourcePriority(source AgentSource) int {
	switch source {
	case AgentSourceProject:
		return 0
	case AgentSourceUser:
		return 1
	case AgentSourcePlugin:
		return 2
	default:
		return 3
	}
}

func agentOwnerLabel(agent AgentResource) string {
	switch agent.Source {
	case AgentSourceProject:
		return strings.ToLower(agent.ProjectName)
	case AgentSourcePlugin:
		return strings.ToLower(agent.PluginName)
	default:
		return ""
	}
}
