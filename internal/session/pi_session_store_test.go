package session

import (
	"os"
	"path/filepath"
	"testing"
)

func writePiSessionFixture(t *testing.T, path, id, cwd string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	header := `{"type":"session","version":3,"id":"` + id + `","cwd":"` + cwd + `"}` + "\n"
	if err := os.WriteFile(path, []byte(header), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPiInstanceSessionLocation_ResolvesOwnFileInPiStore(t *testing.T) {
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PI_CODING_AGENT_DIR", "")
	project := filepath.Join(home, "work", "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	inst := &Instance{ID: "abc123-1790000000", Tool: "pi", ProjectPath: project}

	storeDir := filepath.Join(home, ".pi", "agent", "sessions", piProjectSessionDirName(project))
	own := filepath.Join(storeDir, "2026-09-28T10-00-00-000Z_"+inst.ID+".jsonl")
	writePiSessionFixture(t, own, inst.ID, project)
	writePiSessionFixture(t, filepath.Join(storeDir, "2026-09-28T11-00-00-000Z_someone-else.jsonl"), "someone-else", project)

	loc, err := piInstanceSessionLocation(inst)
	if err != nil {
		t.Fatalf("piInstanceSessionLocation() error = %v", err)
	}
	if files := loc.files(); len(files) != 1 || files[0] != own {
		t.Fatalf("owned files = %v, want only %s", files, own)
	}
	artifact, err := resolveExactPiSessionArtifact(inst)
	if err != nil {
		t.Fatalf("resolveExactPiSessionArtifact() error = %v", err)
	}
	if artifact.Path != own || artifact.SessionID != inst.ID {
		t.Fatalf("artifact = %+v, want %s / %s", artifact, own, inst.ID)
	}
	if got := recallInstanceTranscript(inst); got != own {
		t.Fatalf("recallInstanceTranscript() = %q, want %q", got, own)
	}
	if !inst.hasLocalPiSessionFile() {
		t.Fatal("hasLocalPiSessionFile() = false for a store session")
	}
}

func TestPiInstanceSessionLocation_PrefersLegacyInstanceDir(t *testing.T) {
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	inst := &Instance{ID: "legacy-1790000000", Tool: "pi", ProjectPath: home}
	legacy := filepath.Join(home, ".pi", "agent-deck", inst.ID, "2026-09-28T10-00-00-000Z_0197-uuid.jsonl")
	writePiSessionFixture(t, legacy, "0197-uuid", home)

	loc, err := piInstanceSessionLocation(inst)
	if err != nil {
		t.Fatalf("piInstanceSessionLocation() error = %v", err)
	}
	if files := loc.files(); len(files) != 1 || files[0] != legacy {
		t.Fatalf("owned files = %v, want legacy %s", files, legacy)
	}
}

func TestPiProjectSessionDirName_MatchesPiEncoding(t *testing.T) {
	if got, want := piProjectSessionDirName("/nonexistent/Users/me/Developer"), "--nonexistent-Users-me-Developer--"; got != want {
		t.Fatalf("piProjectSessionDirName() = %q, want %q", got, want)
	}
}
