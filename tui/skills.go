package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill is one installable skill directory in the repo.
type Skill struct {
	Name           string
	Description    string
	HermesCategory string
	Dir            string // absolute path to the skill directory
}

// discoverSkills scans repoDir for top-level directories containing SKILL.md,
// parsing just enough frontmatter to display and route the skill.
func discoverSkills(repoDir string) ([]Skill, error) {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", repoDir, err)
	}

	var skills []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMD := filepath.Join(repoDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); err != nil {
			continue
		}
		fm, err := parseFrontmatter(skillMD)
		if err != nil {
			continue
		}
		desc := fm["description"]
		if desc == "" {
			desc = "(no description)"
		}
		skills = append(skills, Skill{
			Name:           e.Name(),
			Description:    desc,
			HermesCategory: fm["metadata.hermes.category"],
			Dir:            filepath.Join(repoDir, e.Name()),
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

// parseFrontmatter does a minimal, line-oriented read of the YAML frontmatter
// block between the leading "---" markers. It only understands the shapes
// actually used by this repo's SKILL.md files: top-level "key: value" pairs
// and the two-level nested "metadata.hermes.category" path. That's enough to
// avoid pulling in a YAML library for a handful of fields.
func parseFrontmatter(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fm := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if !sc.Scan() || strings.TrimSpace(sc.Text()) != "---" {
		return fm, fmt.Errorf("%s: missing frontmatter", path)
	}

	inMetadata, inHermes := false, false
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		switch {
		case strings.HasPrefix(line, "metadata:"):
			inMetadata, inHermes = true, false
			continue
		case inMetadata && strings.HasPrefix(line, "  hermes:"):
			inHermes = true
			continue
		case inHermes && strings.HasPrefix(line, "    category:"):
			fm["metadata.hermes.category"] = strings.TrimSpace(strings.TrimPrefix(line, "    category:"))
			continue
		case inMetadata && len(line) > 0 && line[0] != ' ':
			inMetadata, inHermes = false, false
		}

		if idx := strings.Index(line, ": "); idx > 0 && (len(line) == 0 || line[0] != ' ') {
			key := line[:idx]
			val := strings.TrimSpace(line[idx+1:])
			if key == "name" || key == "description" || key == "version" || key == "license" {
				fm[key] = val
			}
		}
	}
	return fm, sc.Err()
}
