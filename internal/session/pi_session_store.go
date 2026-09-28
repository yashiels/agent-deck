package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type piSessionLocation struct {
	Dir    string
	Suffix string
}

func (l piSessionLocation) owns(name string) bool {
	return strings.HasSuffix(name, l.Suffix)
}

func (l piSessionLocation) ownsPath(path string) bool {
	return piPathInInstanceDir(path, l.Dir) && l.owns(filepath.Base(path))
}

func (l piSessionLocation) files() []string {
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && l.owns(entry.Name()) {
			paths = append(paths, filepath.Join(l.Dir, entry.Name()))
		}
	}
	sort.Strings(paths)
	return paths
}

func piAgentDir(home string) string {
	if dir := strings.TrimSpace(os.Getenv("PI_CODING_AGENT_DIR")); dir != "" {
		return ExpandPath(dir)
	}
	return filepath.Join(home, ".pi", "agent")
}

func piProjectSessionDirName(cwd string) string {
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}
	cwd = strings.TrimLeft(filepath.Clean(cwd), `/\`)
	return "--" + strings.NewReplacer("/", "-", `\`, "-", ":", "-").Replace(cwd) + "--"
}

func piInstanceDirHasSession(dir string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	return len(matches) > 0
}

func piInstanceSessionLocation(inst *Instance) (piSessionLocation, error) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return piSessionLocation{}, fmt.Errorf("resolve Pi home: %w", err)
	}
	instanceDir := filepath.Join(home, ".pi", "agent-deck", inst.ID)
	if piInstanceDirHasSession(instanceDir) {
		return piSessionLocation{Dir: instanceDir, Suffix: ".jsonl"}, nil
	}
	storeDir := filepath.Join(piAgentDir(home), "sessions", piProjectSessionDirName(inst.EffectiveWorkingDir()))
	return piSessionLocation{Dir: storeDir, Suffix: "_" + inst.ID + ".jsonl"}, nil
}

func piSafeInstanceSessionLocation(inst *Instance) (piSessionLocation, error) {
	loc, err := piInstanceSessionLocation(inst)
	if err != nil {
		return piSessionLocation{}, err
	}
	if err := ensureNoSymlinkPath(loc.Dir); err != nil {
		return piSessionLocation{}, fmt.Errorf("unsafe Pi instance session directory: %w", err)
	}
	info, err := os.Lstat(loc.Dir)
	if err != nil {
		return piSessionLocation{}, err
	}
	if !info.IsDir() {
		return piSessionLocation{}, fmt.Errorf("Pi instance session directory is not a real directory")
	}
	return loc, nil
}
