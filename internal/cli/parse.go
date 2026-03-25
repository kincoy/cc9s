package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// resourceAliases maps alias strings to canonical resource commands.
var resourceAliases = map[string]TopLevelCommand{
	"projects": CmdProjects,
	"project":  CmdProjects,
	"proj":     CmdProjects,
	"sessions": CmdSessions,
	"session":  CmdSessions,
	"ss":       CmdSessions,
	"skills":   CmdSkills,
	"skill":    CmdSkills,
	"sk":       CmdSkills,
	"agents":   CmdAgents,
	"agent":    CmdAgents,
	"ag":       CmdAgents,
}

// Parse converts raw CLI arguments into a Command.
func Parse(args []string) (*Command, error) {
	cmd := &Command{Args: args}

	if len(args) == 0 {
		return nil, nil // no args means TUI
	}

	// Check for --json flag anywhere in args
	jsonIdx := -1
	for i, a := range args {
		if a == "--json" {
			jsonIdx = i
			cmd.Output = OutputJSON
			break
		}
	}

	// Short flags: -h, -v, -?
	for _, a := range args {
		switch a {
		case "-h", "-?", "--help":
			cmd.TopLevel = CmdHelp
			return cmd, nil
		case "-v", "--version":
			cmd.TopLevel = CmdVersion
			return cmd, nil
		}
	}

	// Extract positional args (excluding --json and flags starting with --)
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if i == jsonIdx {
			continue
		}
		if strings.HasPrefix(a, "--") {
			switch {
			case a == "--limit":
				i++ // skip value
				continue
			case a == "--project":
				i++ // skip value
				continue
			case a == "--state":
				i++ // skip value
				continue
			case a == "--older-than":
				i++ // skip value
				continue
			case a == "--dry-run":
				cmd.DryRun = true
				continue
			case a == "--scope", a == "--type":
				i++ // skip value
				continue
			case a == "--sort":
				i++ // skip value
				continue
			case a == "--json":
				continue
			default:
				return nil, CLIError{Message: fmt.Sprintf("unknown flag: %s", a)}
			}
		}
		positional = append(positional, a)
	}

	if len(positional) == 0 {
		return nil, CLIError{Message: "expected a command"}
	}

	// Resolve first positional
	first := strings.ToLower(positional[0])

	if first == "help" {
		cmd.TopLevel = CmdHelp
		return cmd, nil
	}

	if first == "version" {
		cmd.TopLevel = CmdVersion
		return cmd, nil
	}

	if first == "status" {
		cmd.TopLevel = CmdStatus
		if err := parseStatusFlags(cmd, args); err != nil {
			return nil, err
		}
		return cmd, nil
	}

	// Check resource aliases
	if topLevel, ok := resourceAliases[first]; ok {
		cmd.TopLevel = topLevel
		switch topLevel {
		case CmdProjects:
			cmd.Resource = ResourceProject
		case CmdSessions:
			cmd.Resource = ResourceSession
		case CmdSkills:
			cmd.Resource = ResourceSkill
		case CmdAgents:
			cmd.Resource = ResourceAgent
		}
		if err := parseResourceVerb(cmd, positional[1:], args); err != nil {
			return nil, err
		}
		return cmd, nil
	}

	return nil, CLIError{Message: fmt.Sprintf("unknown command: %s", first)}
}

// parseResourceVerb handles the verb and target for resource commands.
func parseResourceVerb(cmd *Command, rest []string, rawArgs []string) error {
	if len(rest) == 0 {
		// Default verb is list
		cmd.Verb = VerbList
		return parseResourceFlags(cmd, rawArgs)
	}

	verb := strings.ToLower(rest[0])
	switch verb {
	case "list":
		cmd.Verb = VerbList
		return parseResourceFlags(cmd, rawArgs)
	case "inspect":
		cmd.Verb = VerbInspect
		if len(rest) < 2 {
			return CLIError{Message: fmt.Sprintf("inspect requires a target name or ID")}
		}
		cmd.Target = rest[1]
		return parseResourceFlags(cmd, rawArgs)
	case "cleanup":
		cmd.Verb = VerbCleanup
		if cmd.Resource != ResourceSession {
			return CLIError{Message: "cleanup is only supported for sessions"}
		}
		if !cmd.DryRun {
			return CLIError{Message: "--dry-run is required for cleanup (preview-only in v1)"}
		}
		return parseResourceFlags(cmd, rawArgs)
	default:
		// Could be an implicit inspect if it's a session ID or project name
		if cmd.TopLevel == CmdSessions {
			// Treat unknown verb as session ID for convenience
			cmd.Verb = VerbInspect
			cmd.Target = verb
			return parseResourceFlags(cmd, rawArgs)
		}
		return CLIError{Message: fmt.Sprintf("unknown verb: %s (expected list, inspect, or cleanup)", verb)}
	}
}

// parseResourceFlags extracts filter flags from raw args.
func parseResourceFlags(cmd *Command, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if err := validateFlagAllowed(cmd, "--limit"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--limit requires a value"}
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return CLIError{Message: fmt.Sprintf("invalid --limit value: %s", args[i])}
			}
			cmd.Limit = val
		case "--sort":
			if err := validateFlagAllowed(cmd, "--sort"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--sort requires a value"}
			}
			i++
			cmd.Sort = args[i]
			if err := validateSortField(cmd); err != nil {
				return err
			}
		case "--project":
			if err := validateFlagAllowed(cmd, "--project"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--project requires a value"}
			}
			i++
			cmd.ProjectFilter = args[i]
		case "--state":
			if err := validateFlagAllowed(cmd, "--state"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--state requires a value"}
			}
			i++
			cmd.StateFilter = args[i]
		case "--scope":
			if err := validateFlagAllowed(cmd, "--scope"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--scope requires a value"}
			}
			i++
			cmd.ScopeFilter = args[i]
			if err := validateScopeFilter(cmd); err != nil {
				return err
			}
		case "--type":
			if err := validateFlagAllowed(cmd, "--type"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--type requires a value"}
			}
			i++
			cmd.TypeFilter = args[i]
			if err := validateTypeFilter(cmd); err != nil {
				return err
			}
		case "--older-than":
			if err := validateFlagAllowed(cmd, "--older-than"); err != nil {
				return err
			}
			if i+1 >= len(args) {
				return CLIError{Message: "--older-than requires a value"}
			}
			i++
			cmd.OlderThan = args[i]
			if _, err := parseOlderThanDuration(cmd.OlderThan); err != nil {
				return CLIError{Message: fmt.Sprintf("invalid --older-than duration: %s", cmd.OlderThan)}
			}
		case "--json":
			// Already handled in Parse
		case "--dry-run":
			if err := validateFlagAllowed(cmd, "--dry-run"); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseStatusFlags extracts flags specific to the status command.
func parseStatusFlags(cmd *Command, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			// Already handled
		default:
			if strings.HasPrefix(args[i], "--") {
				return CLIError{Message: fmt.Sprintf("unknown flag for status: %s", args[i])}
			}
		}
	}
	return nil
}

// validateScopeFilter checks that --scope is only used on skills/agents list.
func validateScopeFilter(cmd *Command) error {
	if cmd.TopLevel != CmdSkills && cmd.TopLevel != CmdAgents {
		return CLIError{Message: "--scope is only supported for skills and agents"}
	}
	return nil
}

// validateTypeFilter checks that --type is only used on skills list.
func validateTypeFilter(cmd *Command) error {
	if cmd.TopLevel != CmdSkills {
		return CLIError{Message: "--type is only supported for skills"}
	}
	return nil
}

func validateFlagAllowed(cmd *Command, flag string) error {
	if isFlagAllowed(cmd, flag) {
		return nil
	}

	resource := resourceLabel(cmd.TopLevel)
	verb := verbLabel(cmd.Verb)
	if resource == "" {
		return CLIError{Message: fmt.Sprintf("%s is not supported here", flag)}
	}
	if verb == "" {
		return CLIError{Message: fmt.Sprintf("%s is not supported for %s", flag, resource)}
	}
	return CLIError{Message: fmt.Sprintf("%s is not supported for %s %s", flag, resource, verb)}
}

func isFlagAllowed(cmd *Command, flag string) bool {
	switch flag {
	case "--json":
		return true
	case "--limit":
		return (cmd.TopLevel == CmdProjects || cmd.TopLevel == CmdSessions) && cmd.Verb == VerbList
	case "--sort":
		return (cmd.TopLevel == CmdProjects || cmd.TopLevel == CmdSessions) && cmd.Verb == VerbList
	case "--project":
		return (cmd.TopLevel == CmdSessions && (cmd.Verb == VerbList || cmd.Verb == VerbCleanup)) ||
			((cmd.TopLevel == CmdSkills || cmd.TopLevel == CmdAgents) && cmd.Verb == VerbList)
	case "--state":
		return cmd.TopLevel == CmdSessions && (cmd.Verb == VerbList || cmd.Verb == VerbCleanup)
	case "--scope":
		return (cmd.TopLevel == CmdSkills || cmd.TopLevel == CmdAgents) && cmd.Verb == VerbList
	case "--type":
		return cmd.TopLevel == CmdSkills && cmd.Verb == VerbList
	case "--older-than":
		return cmd.TopLevel == CmdSessions && cmd.Verb == VerbCleanup
	case "--dry-run":
		return cmd.TopLevel == CmdSessions && cmd.Verb == VerbCleanup
	default:
		return false
	}
}

func resourceLabel(topLevel TopLevelCommand) string {
	switch topLevel {
	case CmdProjects:
		return "projects"
	case CmdSessions:
		return "sessions"
	case CmdSkills:
		return "skills"
	case CmdAgents:
		return "agents"
	default:
		return ""
	}
}

func verbLabel(verb Verb) string {
	switch verb {
	case VerbList:
		return "list"
	case VerbInspect:
		return "inspect"
	case VerbCleanup:
		return "cleanup"
	default:
		return ""
	}
}

// validateSortField checks that --sort is only used on projects/sessions list with valid fields.
func validateSortField(cmd *Command) error {
	if cmd.TopLevel != CmdProjects && cmd.TopLevel != CmdSessions {
		return CLIError{Message: "--sort is only supported for projects and sessions"}
	}
	validFields := map[string]bool{}
	switch cmd.TopLevel {
	case CmdProjects:
		validFields = map[string]bool{"name": true, "sessions": true}
	case CmdSessions:
		validFields = map[string]bool{"updated": true, "state": true, "project": true}
	}
	if !validFields[cmd.Sort] {
		// Build a helpful error listing valid fields
		fields := make([]string, 0, len(validFields))
		for f := range validFields {
			fields = append(fields, f)
		}
		return CLIError{Message: fmt.Sprintf("invalid --sort field %q (valid: %s)", cmd.Sort, strings.Join(fields, ", "))}
	}
	return nil
}

// parseOlderThanDuration parses a duration string that may contain "d" (days) suffix.
// For example: "7d" -> 168h, "30d" -> 720h. Other suffixes (h, m, s) are passed
// through to time.ParseDuration as-is.
func parseOlderThanDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		daysStr := s[:len(s)-1]
		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 0 {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
