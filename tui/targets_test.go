package main

import (
	"os"
	"path/filepath"
	"testing"
)

// setupFakeSkill creates a minimal on-disk skill directory under a temp repo
// so install/uninstall can be exercised without touching the real machine.
func setupFakeSkill(t *testing.T) Skill {
	t.Helper()
	repo := t.TempDir()
	skillDir := filepath.Join(repo, "demo-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo-skill\ndescription: demo\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return Skill{Name: "demo-skill", Description: "demo", Dir: skillDir}
}

func TestSymlinkTargetLifecycle(t *testing.T) {
	s := setupFakeSkill(t)
	agentRoot := t.TempDir() // stands in for e.g. ~/.agents
	target := Target{Name: "agents", Base: filepath.Join(agentRoot, "skills")}

	if status, _ := target.status(s); status != StatusNotInstalled {
		t.Fatalf("want NotInstalled before install, got %v", status)
	}

	if err := target.install(s); err != nil {
		t.Fatalf("install: %v", err)
	}
	status, dest := target.status(s)
	if status != StatusInstalled {
		t.Fatalf("want Installed after install, got %v", status)
	}
	if info, err := os.Lstat(dest); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected a symlink at %s", dest)
	}

	if err := target.uninstall(s); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if status, _ := target.status(s); status != StatusNotInstalled {
		t.Fatalf("want NotInstalled after uninstall, got %v", status)
	}
}

func TestSymlinkTargetRefusesToRemoveConflict(t *testing.T) {
	s := setupFakeSkill(t)
	agentRoot := t.TempDir()
	target := Target{Name: "agents", Base: filepath.Join(agentRoot, "skills")}
	dest := filepath.Join(target.Base, s.Name)

	if err := os.MkdirAll(target.Base, 0o755); err != nil {
		t.Fatal(err)
	}
	// Something unrelated already occupies the destination.
	if err := os.WriteFile(dest, []byte("not a skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	if status, _ := target.status(s); status != StatusConflict {
		t.Fatalf("want Conflict, got %v", status)
	}
	if err := target.install(s); err == nil {
		// install() on a conflicting non-symlink path would clobber the file
		// via os.Symlink failing with "file exists" - confirm it does NOT
		// silently succeed.
		t.Fatalf("install must not silently succeed over a conflicting path")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("conflicting file must survive a failed install: %v", err)
	}
}

func TestHermesTargetCopiesAndUninstalls(t *testing.T) {
	s := setupFakeSkill(t)
	hermesRoot := t.TempDir()
	target := Target{Name: "hermes", Base: filepath.Join(hermesRoot, "skills"), Copy: true}

	if status, _ := target.status(s); status != StatusNotInstalled {
		t.Fatalf("want NotInstalled before install, got %v", status)
	}
	if err := target.install(s); err != nil {
		t.Fatalf("install: %v", err)
	}
	status, dest := target.status(s)
	if status != StatusInstalled {
		t.Fatalf("want Installed after copy, got %v", status)
	}
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Fatalf("copied skill should contain SKILL.md: %v", err)
	}
	if info, _ := os.Lstat(dest); info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("hermes install must be a real copy, not a symlink")
	}

	if err := target.uninstall(s); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be gone after uninstall", dest)
	}
}

func TestAgentMissingWhenRootAbsent(t *testing.T) {
	s := setupFakeSkill(t)
	target := Target{Name: "agents", Base: filepath.Join(t.TempDir(), "does-not-exist", "skills")}
	if status, _ := target.status(s); status != StatusAgentMissing {
		t.Fatalf("want AgentMissing, got %v", status)
	}
}
