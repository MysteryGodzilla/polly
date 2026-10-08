package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// interpreters run script tools on Windows, which can't execute a script file itself.
var interpreters = map[string]string{".py": "python"}

// ScriptCommand returns the command that runs a shell tool with args: the file itself, or on
// Windows its interpreter with the file as the first argument.
func ScriptCommand(ctx context.Context, path string, args ...string) *exec.Cmd {
	if interp, ok := interpreter(runtime.GOOS, path); ok {
		return exec.CommandContext(ctx, interp, append([]string{path}, args...)...)
	}
	return exec.CommandContext(ctx, path, args...)
}

func interpreter(goos, path string) (string, bool) {
	if goos != "windows" {
		return "", false
	}
	interp, ok := interpreters[strings.ToLower(filepath.Ext(path))]
	return interp, ok
}

// runnable reports whether a file can be run as a shell tool: executable, or a script with an
// interpreter on Windows, where files carry no executable bit.
func runnable(goos, path string, info os.FileInfo) bool {
	_, ok := interpreter(goos, path)
	return ok || info.Mode()&0111 != 0
}
