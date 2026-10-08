package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScriptInterpreter(t *testing.T) {
	cases := []struct {
		goos, path, want string
		ok               bool
	}{
		{"windows", `C:\plugins\websearch.py`, "python", true},
		{"windows", "plugins/WebFetch.PY", "python", true},
		{"windows", "plugins/datetime.sh", "", false},
		{"windows", "tool.exe", "", false},
		{"linux", "plugins/websearch.py", "", false},
		{"darwin", "plugins/websearch.py", "", false},
	}
	for _, c := range cases {
		if got, ok := interpreter(c.goos, c.path); got != c.want || ok != c.ok {
			t.Errorf("interpreter(%q, %q) = %q, %v; want %q, %v", c.goos, c.path, got, ok, c.want, c.ok)
		}
	}
}

func TestRunnable(t *testing.T) {
	dir := t.TempDir()
	py := filepath.Join(dir, "tool.py")
	sh := filepath.Join(dir, "tool.sh")
	for _, p := range []string{py, sh} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	info := func(p string) os.FileInfo {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		return fi
	}
	if !runnable("windows", py, info(py)) {
		t.Error("a .py file on Windows should run through python")
	}
	if runnable("windows", sh, info(sh)) {
		t.Error("a .sh file without an executable bit should not run on Windows")
	}
	if runnable("linux", py, info(py)) {
		t.Error("a non-executable .py file should not run on Linux")
	}
}

func TestScriptCommand(t *testing.T) {
	cmd := ScriptCommand(context.Background(), "plugins/tool.sh", "--schema")
	if got := cmd.Args; len(got) != 2 || got[0] != "plugins/tool.sh" || got[1] != "--schema" {
		t.Errorf("args = %q", got)
	}
}

// On Windows a .py tool's output with non-Latin text reaches the bot intact.
func TestScriptCommandPythonUTF8(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows runs .py tools through python")
	}
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not installed")
	}
	script := filepath.Join(t.TempDir(), "tool.py")
	if err := os.WriteFile(script, []byte("print('\\u304a\\u306f\\u3088\\u3046 \\U0001f98b')\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := ScriptCommand(context.Background(), script, "--execute", "{}").Output()
	if err != nil {
		t.Fatalf("tool failed: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "おはよう \U0001f98b" {
		t.Errorf("output = %q", got)
	}
}
