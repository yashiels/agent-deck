package session

import (
	"testing"
)

const sharedNewestHermesSessionID = "20261012_150000_ffffff"

func withFakeHermesSharingNewestSession(t *testing.T) {
	t.Helper()
	withFakeHermes(t, `if [ "$1" = sessions ]; then
  printf 'Title Workspace LastActive ID\n--- --- --- ---\nNewest home 1m `+sharedNewestHermesSessionID+`\n'
  exit 0
fi
exec sleep 30`)
}

func TestHermesRestartResumesEachInstanceFromItsOwnAnchor(t *testing.T) {
	skipIfNoTmuxBinary(t)
	withFakeHermesSharingNewestSession(t)
	sharedDir := t.TempDir()

	anchors := map[string]string{
		"hermes-anchor-a": "20261012_140000_aaaaaa",
		"hermes-anchor-b": "20261012_140500_bbbbbb",
	}
	for id, anchor := range anchors {
		inst := NewInstanceWithTool(id, sharedDir, "hermes")
		inst.ID = id
		WriteHookSessionAnchor(id, anchor)
		if err := inst.Restart(); err != nil {
			t.Fatalf("%s restart: %v", id, err)
		}
		t.Cleanup(func() { _ = inst.Kill() })
		if inst.HermesSessionID != anchor {
			t.Errorf("%s resumed %q, want its own anchor %q", id, inst.HermesSessionID, anchor)
		}
	}
}

func TestHermesRestartFallsBackToNewestWithoutAnchor(t *testing.T) {
	skipIfNoTmuxBinary(t)
	withFakeHermesSharingNewestSession(t)

	inst := NewInstanceWithTool("hermes-no-anchor", t.TempDir(), "hermes")
	inst.ID = "hermes-no-anchor"
	ClearHookSessionAnchor(inst.ID)
	if err := inst.Restart(); err != nil {
		t.Fatalf("restart: %v", err)
	}
	t.Cleanup(func() { _ = inst.Kill() })
	if inst.HermesSessionID != sharedNewestHermesSessionID {
		t.Errorf("resumed %q, want newest-in-folder fallback %q", inst.HermesSessionID, sharedNewestHermesSessionID)
	}
}

func TestHermesFreshStartClearsAnchor(t *testing.T) {
	inst := &Instance{ID: "hermes-fresh-anchor", Tool: "hermes", HermesSessionID: "20261012_140000_aaaaaa"}
	WriteHookSessionAnchor(inst.ID, inst.HermesSessionID)
	inst.clearSessionBindingForFreshStart()
	if inst.HermesSessionID != "" {
		t.Errorf("HermesSessionID = %q after fresh start, want empty", inst.HermesSessionID)
	}
	if got := ReadHookSessionAnchor(inst.ID); got != "" {
		t.Errorf("anchor = %q after fresh start, want cleared so a later restart cannot resume it", got)
	}
}
