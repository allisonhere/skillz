package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Target is one place a skill can be installed. It mirrors install-skill's
// TARGET_NAMES/target_base/target_mode tables exactly, so status shown here
// always matches what the shell installer would do.
type Target struct {
	Name string // agents | codex | claude | hermes
	Base string // e.g. ~/.agents/skills
	Copy bool   // true for hermes: cp -a instead of symlink, category-namespaced
}

func loadTargets() []Target {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return []Target{
		{Name: "agents", Base: filepath.Join(home, ".agents", "skills")},
		{Name: "codex", Base: filepath.Join(home, ".codex", "skills")},
		{Name: "claude", Base: filepath.Join(home, ".claude", "skills")},
		{Name: "hermes", Base: filepath.Join(home, ".hermes", "skills"), Copy: true},
	}
}

// Status is a skill's install state at one target.
type Status int

const (
	StatusNotInstalled Status = iota
	StatusInstalled
	StatusConflict     // something else is already at the destination
	StatusAgentMissing // the agent's own root dir isn't on this machine
)

func (s Status) String() string {
	switch s {
	case StatusInstalled:
		return "installed"
	case StatusConflict:
		return "conflict"
	case StatusAgentMissing:
		return "agent not found"
	default:
		return "not installed"
	}
}

// dest returns the on-disk destination path for skill s at target t, and the
// hermes category used (empty for non-hermes targets).
func (t Target) dest(s Skill) (path string, category string) {
	if !t.Copy {
		return filepath.Join(t.Base, s.Name), ""
	}
	cat := s.HermesCategory
	if cat == "" {
		cat = "general"
	}
	return filepath.Join(t.Base, cat, s.Name), cat
}

// status reports where a skill stands at a target, matching install-skill's
// own detection but treating an existing hermes copy as "installed" rather
// than a bare "not a symlink" conflict.
func (t Target) status(s Skill) (Status, string) {
	agentRoot := filepath.Dir(t.Base)
	if _, err := os.Stat(agentRoot); err != nil {
		return StatusAgentMissing, agentRoot
	}

	dest, _ := t.dest(s)
	info, err := os.Lstat(dest)
	if err != nil {
		return StatusNotInstalled, dest
	}

	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(dest)
		if err == nil && resolved == s.Dir {
			return StatusInstalled, dest
		}
		return StatusConflict, dest
	}

	// Exists and isn't a symlink.
	if t.Copy {
		return StatusInstalled, dest // the expected shape for a hermes copy
	}
	return StatusConflict, dest
}

// install symlinks (or, for hermes, copies) skillDir to this target. It
// refuses to touch a destination it doesn't already know is safe — callers
// should only call this when status() reported NotInstalled.
func (t Target) install(s Skill) error {
	dest, _ := t.dest(s)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(dest), err)
	}
	if t.Copy {
		cmd := exec.Command("cp", "-a", s.Dir, dest)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cp -a: %w: %s", err, out)
		}
		return nil
	}
	return os.Symlink(s.Dir, dest)
}

// uninstall removes a skill from this target. For symlink targets it only
// ever removes a symlink that resolves back to this exact skill directory.
// For the hermes copy target it removes the skill's own namespaced
// directory, which is always a controlled path (base/category/skill) and
// never a path supplied by the user.
func (t Target) uninstall(s Skill) error {
	status, dest := t.status(s)
	if status != StatusInstalled {
		return fmt.Errorf("not installed at %s (%s)", t.Name, status)
	}
	if t.Copy {
		return os.RemoveAll(dest)
	}
	resolved, err := filepath.EvalSymlinks(dest)
	if err != nil || resolved != s.Dir {
		return fmt.Errorf("refusing to remove %s: no longer a symlink to this skill", dest)
	}
	return os.Remove(dest)
}
