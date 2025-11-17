package cli

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func isPath(s string) bool {
	if s == "" {
		return false
	}

	if s == "." {
		return true
	}

	if strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") {
		return true
	}

	if filepath.IsAbs(s) {
		return true
	}

	return false
}

func isNamedGroup(name string) bool {
	return strings.HasPrefix(name, "[")
}

func isTemplateDir(name string) bool {
	return !isNamedGroup(name)
}

func searchInNonNamedGroups(fsys fs.FS, baseDir, template string) string {
	templatePath := filepath.Join(baseDir, template)
	if info, err := fs.Stat(fsys, templatePath); err == nil && info.IsDir() {
		return templatePath
	}

	entries, err := fs.ReadDir(fsys, baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && isTemplateDir(entry.Name()) {
			groupPath := filepath.Join(baseDir, entry.Name())
			if found := searchInNonNamedGroups(fsys, groupPath, template); found != "" {
				return found
			}
		}
	}

	return ""
}

func searchInParenthesesGroups(fsys fs.FS, baseDir, template string) string {
	entries, err := fs.ReadDir(fsys, baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "(") {
			groupPath := filepath.Join(baseDir, entry.Name())
			templatePath := filepath.Join(groupPath, template)

			if info, err := fs.Stat(fsys, templatePath); err == nil && info.IsDir() {
				return templatePath
			}

			if found := searchInParenthesesGroups(fsys, groupPath, template); found != "" {
				return found
			}
		}
	}

	return ""
}

func searchInParenthesesForNamedGroups(fsys fs.FS, baseDir string, groups []string, template string) string {
	entries, err := fs.ReadDir(fsys, baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "(") {
			parenthesesGroupPath := filepath.Join(baseDir, entry.Name())
			if found := searchInNamedGroups(fsys, parenthesesGroupPath, groups, template); found != "" {
				return found
			}
		}
	}

	return ""
}

func searchInNamedGroups(fsys fs.FS, baseDir string, groups []string, template string) string {
	pathParts := []string{baseDir}
	hasMixedGroups := false
	for _, group := range groups {
		if strings.HasPrefix(group, "[") || strings.HasPrefix(group, "(") {
			pathParts = append(pathParts, group)
		} else {
			hasMixedGroups = true
			pathParts = append(pathParts, "["+group+"]")
		}
	}
	pathParts = append(pathParts, template)
	fullPath := filepath.Join(pathParts...)
	if info, err := fs.Stat(fsys, fullPath); err == nil && info.IsDir() {
		return fullPath
	}

	if hasMixedGroups && len(groups) >= 1 {
		firstGroup := groups[0]
		if !strings.HasPrefix(firstGroup, "[") && !strings.HasPrefix(firstGroup, "(") {
			mixedPath := filepath.Join(baseDir, "("+firstGroup+")", "["+firstGroup+"]", template)
			if info, err := fs.Stat(fsys, mixedPath); err == nil && info.IsDir() {
				return mixedPath
			}

			if len(groups) > 1 {
				mixedPathParts := []string{baseDir, "(" + firstGroup + ")"}
				for i := 1; i < len(groups); i++ {
					group := groups[i]
					if strings.HasPrefix(group, "[") || strings.HasPrefix(group, "(") {
						mixedPathParts = append(mixedPathParts, group)
					} else {
						mixedPathParts = append(mixedPathParts, "["+group+"]")
					}
				}
				mixedPathParts = append(mixedPathParts, template)
				mixedPath = filepath.Join(mixedPathParts...)
				if info, err := fs.Stat(fsys, mixedPath); err == nil && info.IsDir() {
					return mixedPath
				}
			}
		}
	}

	return ""
}

func SearchTemplate(fsys fs.FS, baseDir string, nameOrPath string) string {
	if isPath(nameOrPath) {
		return nameOrPath
	}

	parts := strings.Split(nameOrPath, "/")
	if len(parts) == 0 {
		return ""
	}

	template := parts[len(parts)-1]
	groups := parts[:len(parts)-1]

	currentDir := baseDir
	for {
		vieDir := filepath.Join(currentDir, ".vie")
		if info, err := fs.Stat(fsys, vieDir); err == nil && info.IsDir() {
			if len(groups) > 0 {
				if found := searchInNamedGroups(fsys, vieDir, groups, template); found != "" {
					return found
				}

				if found := searchInParenthesesForNamedGroups(fsys, vieDir, groups, template); found != "" {
					return found
				}
			} else {
				if found := searchInNonNamedGroups(fsys, vieDir, template); found != "" {
					return found
				}

				if found := searchInParenthesesGroups(fsys, vieDir, template); found != "" {
					return found
				}
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir || parentDir == "." {
			break
		}
		currentDir = parentDir
	}

	return ""
}
