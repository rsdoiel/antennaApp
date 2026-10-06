package antennaApp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every verb in the completion table must be a help topic, so the table
// cannot drift from the commands the user is told about.
func TestCompletionVerbsAreHelpTopics(t *testing.T) {
	for _, c := range completionCommands {
		if c.Name == "help" { // a verb that introduces topics, not a topic itself
			continue
		}
		var buf bytes.Buffer
		if !PrintHelpTopic(&buf, c.Name, "antenna", "v", "d", "h") {
			t.Errorf("completion verb %q is not a help topic", c.Name)
		}
		if !strings.Contains(HelpTopicsText(), "  "+c.Name+" ") {
			t.Errorf("completion verb %q missing from HelpTopicsText", c.Name)
		}
	}
}

func TestCompletionBash(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCompletion(&buf, "antenna", "bash"); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{"-F _antenna antenna", "generate", "harvest", "stylefrom", "-config", "powershell", "topics"} {
		if !strings.Contains(s, want) {
			t.Errorf("bash completion missing %q", want)
		}
	}
}

func TestCompletionPowerShell(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCompletion(&buf, "antenna", "powershell"); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{"Register-ArgumentCompleter -Native -CommandName antenna", "generate", "harvest", "stylefrom", "-config", "CompletionResult"} {
		if !strings.Contains(s, want) {
			t.Errorf("powershell completion missing %q", want)
		}
	}
}

func TestCompletionUsesAppName(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCompletion(&buf, "ant.exe", "powershell"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "-CommandName ant,ant.exe") {
		t.Errorf("expected both ant and ant.exe registered, got:\n%s", buf.String())
	}
}

func TestCompletionUnknownShell(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCompletion(&buf, "antenna", "zsh"); err == nil {
		t.Error("expected error for unsupported shell")
	}
	if buf.Len() != 0 {
		t.Error("nothing should be written for an unsupported shell")
	}
}

func TestRunCompletionDispatch(t *testing.T) {
	app := NewAntennaApp("antenna")
	var out, eout bytes.Buffer
	if err := app.Run(nil, &out, &eout, "antenna.yaml", "completion", []string{"bash"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "-F _") {
		t.Error("completion bash produced no script")
	}
	if err := app.Run(nil, &out, &eout, "antenna.yaml", "completion", nil); err == nil {
		t.Error("expected usage error with no shell argument")
	}
}

func TestInstallCompletionBash(t *testing.T) {
	home := t.TempDir()
	path, err := InstallCompletion("antenna", "bash", home, "", "linux")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".local", "share", "bash-completion", "completions", "antenna")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "-F _") {
		t.Error("installed file is not a completion script")
	}
	// Installing again updates our own file without error.
	if _, err := InstallCompletion("antenna", "bash", home, "", "linux"); err != nil {
		t.Errorf("reinstall: %s", err)
	}
}

func TestInstallCompletionBashXDG(t *testing.T) {
	home, xdg := t.TempDir(), t.TempDir()
	path, err := InstallCompletion("antenna", "bash", home, xdg, "linux")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, xdg) {
		t.Errorf("path %q should be under XDG_DATA_HOME %q", path, xdg)
	}
}

func TestInstallCompletionBashRefusesForeignFile(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(dir, "antenna")
	if err := os.WriteFile(foreign, []byte("something else\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallCompletion("antenna", "bash", home, "", "linux"); err == nil {
		t.Error("expected refusal to overwrite a file antenna did not write")
	}
	got, _ := os.ReadFile(foreign)
	if string(got) != "something else\n" {
		t.Error("foreign file was modified")
	}
}

func TestInstallCompletionPowerShell(t *testing.T) {
	home := t.TempDir()
	path, err := InstallCompletion("antenna", "powershell", home, "", "linux")
	if err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	prof, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prof), path) {
		t.Errorf("profile does not source %s:\n%s", path, prof)
	}
	// Idempotent: the profile line appears once however often we install.
	if _, err := InstallCompletion("antenna", "powershell", home, "", "linux"); err != nil {
		t.Fatal(err)
	}
	prof, _ = os.ReadFile(profile)
	if n := strings.Count(string(prof), path); n != 1 {
		t.Errorf("profile sources the script %d times, want 1", n)
	}
}

func TestInstallCompletionPowerShellKeepsProfile(t *testing.T) {
	home := t.TempDir()
	profile := filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	os.MkdirAll(filepath.Dir(profile), 0o755)
	os.WriteFile(profile, []byte("Set-Alias ll ls\n"), 0o644)
	if _, err := InstallCompletion("antenna", "powershell", home, "", "linux"); err != nil {
		t.Fatal(err)
	}
	prof, _ := os.ReadFile(profile)
	if !strings.HasPrefix(string(prof), "Set-Alias ll ls\n") {
		t.Error("existing profile content was lost")
	}
}

func TestInstallCompletionPowerShellWindows(t *testing.T) {
	home := t.TempDir()
	if _, err := InstallCompletion("antenna.exe", "powershell", home, "", "windows"); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if _, err := os.Stat(profile); err != nil {
		t.Errorf("windows profile not written: %s", err)
	}
}

func TestInstallCompletionUnknownShell(t *testing.T) {
	home := t.TempDir()
	if _, err := InstallCompletion("antenna", "zsh", home, "", "linux"); err == nil {
		t.Error("expected error for unsupported shell")
	}
	entries, _ := os.ReadDir(home)
	if len(entries) != 0 {
		t.Error("nothing should be created for an unsupported shell")
	}
}

func TestRunCompletionInstallDispatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	app := NewAntennaApp("antenna")
	for _, args := range [][]string{{"bash", "-install"}, {"--install", "bash"}} {
		var out, eout bytes.Buffer
		if err := app.Run(nil, &out, &eout, "antenna.yaml", "completion", args); err != nil {
			t.Fatalf("%v: %s", args, err)
		}
		if !strings.Contains(out.String(), "bash-completion") {
			t.Errorf("%v: output should name the installed file, got %q", args, out.String())
		}
	}
	if err := app.Run(nil, &bytes.Buffer{}, &bytes.Buffer{}, "antenna.yaml", "completion", []string{"-install"}); err == nil {
		t.Error("expected usage error: -install without a shell")
	}
}
