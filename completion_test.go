package antennaApp

import (
	"bytes"
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
