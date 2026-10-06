/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
*/
package antennaApp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/** CompletionCommand describes one command line verb for shell completion.
 *
 * Fields:
 *   Name    (string)   — the verb as typed
 *   Aliases ([]string) — alternate spellings accepted by Run
 *
 * Example:
 *   c := CompletionCommand{Name: "generate", Aliases: []string{"build"}}
 */
type CompletionCommand struct {
	Name    string
	Aliases []string
}

/** completionCommands is the verb table both completion generators read.
 * Keep it in step with the cases in Run; TestCompletionVerbsAreHelpTopics
 * checks each entry against the help topics.
 */
var completionCommands = []CompletionCommand{
	{Name: "add"},
	{Name: "apply"},
	{Name: "blogit"},
	{Name: "completion"},
	{Name: "css"},
	{Name: "del"},
	{Name: "generate", Aliases: []string{"build"}},
	{Name: "harvest", Aliases: []string{"fetch"}},
	{Name: "help"},
	{Name: "init"},
	{Name: "interactive", Aliases: []string{"tui"}},
	{Name: "items"},
	{Name: "list"},
	{Name: "page"},
	{Name: "pages"},
	{Name: "post"},
	{Name: "posts"},
	{Name: "preview"},
	{Name: "quote", Aliases: []string{"reply"}},
	{Name: "rss"},
	{Name: "sitemap"},
	{Name: "stylefrom"},
	{Name: "themes"},
	{Name: "unpage"},
	{Name: "unpost"},
}

// completionFlags are the top level options, before the verb.
var completionFlags = []string{"-config", "-help", "-license", "-version"}

// completionShells are the shells WriteCompletion can generate for.
var completionShells = []string{"bash", "powershell"}

// completionTopics are help topics that are not verbs.
var completionTopics = []string{"accessibility", "configuration", "metadata", "topics"}

func completionVerbs() []string {
	var verbs []string
	for _, c := range completionCommands {
		verbs = append(verbs, c.Name)
		verbs = append(verbs, c.Aliases...)
	}
	return verbs
}

func completionHelpTopics() []string {
	return append(completionVerbs(), completionTopics...)
}

var nonIdent = regexp.MustCompile(`[^A-Za-z0-9_]`)

/** WriteCompletion writes a shell completion script for the named shell.
 * Nothing is written when the shell is not supported.
 *
 * Parameters:
 *   w       (io.Writer) — destination for the script
 *   appName (string)    — binary name to register; a .exe suffix is ignored
 *   shell   (string)    — "bash" or "powershell" (case-insensitive; "pwsh" is accepted)
 *
 * Returns:
 *   error — non-nil for an unsupported shell or a failed write
 *
 * Example:
 *   err := WriteCompletion(os.Stdout, "antenna", "bash")
 */
func WriteCompletion(w io.Writer, appName, shell string) error {
	name := strings.TrimSuffix(appName, ".exe")
	var script string
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "bash":
		script = bashCompletion
	case "powershell", "pwsh":
		script = powershellCompletion
	default:
		return usageErrorf("unsupported shell %q — use one of: %s", shell, strings.Join(completionShells, ", "))
	}
	r := strings.NewReplacer(
		"{app_name}", name,
		"{app_func}", nonIdent.ReplaceAllString(name, "_"),
		"{verbs}", strings.Join(completionVerbs(), " "),
		"{flags}", strings.Join(completionFlags, " "),
		"{topics}", strings.Join(completionHelpTopics(), " "),
		"{shells}", strings.Join(completionShells, " "),
		"{ps_verbs}", psList(completionVerbs()),
		"{ps_flags}", psList(completionFlags),
		"{ps_topics}", psList(completionHelpTopics()),
		"{ps_shells}", psList(completionShells),
	)
	_, err := io.WriteString(w, r.Replace(script))
	return err
}

func psList(items []string) string {
	q := make([]string, len(items))
	for i, s := range items {
		q[i] = "'" + s + "'"
	}
	return strings.Join(q, ", ")
}

const bashCompletion = `# bash completion for {app_name}
# Load with:  source <({app_name} completion bash)
_{app_func}() {
    local cur prev verb i w
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    case "$prev" in
        -config|--config)
            COMPREPLY=( $(compgen -f -- "$cur") )
            return 0 ;;
    esac
    # Find the verb: first word after the command that is not an option
    # and not the value of -config.
    verb=""
    for (( i=1; i<COMP_CWORD; i++ )); do
        w="${COMP_WORDS[i]}"
        case "$w" in
            -config|--config) (( i++ )) ;;
            -*) ;;
            *) verb="$w"; break ;;
        esac
    done
    if [[ -z "$verb" ]]; then
        if [[ "$cur" == -* ]]; then
            COMPREPLY=( $(compgen -W "{flags}" -- "$cur") )
        else
            COMPREPLY=( $(compgen -W "{verbs}" -- "$cur") )
        fi
        return 0
    fi
    case "$verb" in
        help)       COMPREPLY=( $(compgen -W "{topics}" -- "$cur") ) ;;
        themes)     COMPREPLY=( $(compgen -W "new" -- "$cur") ) ;;
        completion) COMPREPLY=( $(compgen -W "{shells}" -- "$cur") ) ;;
        *)          COMPREPLY=( $(compgen -f -- "$cur") ) ;;
    esac
    return 0
}
complete -o default -F _{app_func} {app_name}
`

const powershellCompletion = `# PowerShell completion for {app_name}
# Load with:  {app_name} completion powershell | Out-String | Invoke-Expression
Register-ArgumentCompleter -Native -CommandName {app_name},{app_name}.exe -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $verbs  = @({ps_verbs})
    $flags  = @({ps_flags})
    $topics = @({ps_topics})
    $shells = @({ps_shells})

    # Words typed after the command name, not counting the word being completed.
    $words = @($commandAst.CommandElements | Select-Object -Skip 1 | ForEach-Object { $_.ToString() })
    if ($wordToComplete -ne '' -and $words.Count -gt 0) {
        $words = @($words | Select-Object -SkipLast 1)
    }

    $prev = if ($words.Count -gt 0) { $words[-1] } else { '' }
    if ($prev -in '-config', '--config') {
        return Get-ChildItem -Path "$wordToComplete*" -ErrorAction SilentlyContinue | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_.Name, $_.Name, 'ProviderItem', $_.Name)
        }
    }

    # The verb is the first word that is not an option or the value of -config.
    $verb = $null
    for ($i = 0; $i -lt $words.Count; $i++) {
        if ($words[$i] -in '-config', '--config') { $i++; continue }
        if ($words[$i].StartsWith('-')) { continue }
        $verb = $words[$i]
        break
    }

    $candidates = if (-not $verb) {
        if ($wordToComplete.StartsWith('-')) { $flags } else { $verbs }
    } else {
        switch ($verb) {
            'help'       { $topics }
            'themes'     { @('new') }
            'completion' { $shells }
            default      { $null }
        }
    }

    if ($null -eq $candidates) {
        # Fall back to file names, as the shell would.
        return Get-ChildItem -Path "$wordToComplete*" -ErrorAction SilentlyContinue | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_.Name, $_.Name, 'ProviderItem', $_.Name)
        }
    }
    $candidates | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
`

// completionMarker tags the profile line InstallCompletion adds, so a second
// install finds it and leaves the profile alone.
const completionMarker = "# antenna completion"

/** InstallCompletion installs the completion script for a shell so it loads
 * in every new session, and returns the path of the script it wrote.
 *
 * For bash the script goes in the bash-completion user directory
 * ($XDG_DATA_HOME/bash-completion/completions/NAME, or the same under
 * ~/.local/share), which bash-completion loads on demand. An existing file
 * there is replaced only if it is an antenna completion script.
 *
 * For PowerShell the script is written beside the profile and the profile
 * gets one line that dot-sources it. The line is added once.
 *
 * Parameters:
 *   appName  (string) — binary name to register; a .exe suffix is ignored
 *   shell    (string) — "bash" or "powershell" ("pwsh" is accepted)
 *   home     (string) — the user's home directory
 *   dataHome (string) — $XDG_DATA_HOME, or "" for the default
 *   goos     (string) — runtime.GOOS, which picks the PowerShell profile location
 *
 * Returns:
 *   (string, error) — the installed script path; an error for an unsupported
 *   shell, a foreign file in the way, or a failed write
 *
 * Example:
 *   path, err := InstallCompletion("antenna", "bash", home, "", "linux")
 */
func InstallCompletion(appName, shell, home, dataHome, goos string) (string, error) {
	name := strings.TrimSuffix(appName, ".exe")
	var script bytes.Buffer
	if err := WriteCompletion(&script, appName, shell); err != nil {
		return "", err
	}
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "bash":
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		path := filepath.Join(dataHome, "bash-completion", "completions", name)
		if old, err := os.ReadFile(path); err == nil {
			if !strings.HasPrefix(string(old), "# bash completion for ") {
				return "", cantCreatef("%s exists and was not written by %s, not overwriting", path, name)
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
		return path, writeFileIn(path, script.Bytes())
	default: // powershell, checked by WriteCompletion
		profileDir := filepath.Join(home, ".config", "powershell")
		if goos == "windows" {
			profileDir = filepath.Join(home, "Documents", "PowerShell")
		}
		path := filepath.Join(profileDir, name+"-completion.ps1")
		if err := writeFileIn(path, script.Bytes()); err != nil {
			return "", err
		}
		profile := filepath.Join(profileDir, "Microsoft.PowerShell_profile.ps1")
		line := fmt.Sprintf(". '%s' %s", path, completionMarker)
		cur, err := os.ReadFile(profile)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if strings.Contains(string(cur), path) {
			return path, nil
		}
		if len(cur) > 0 && !bytes.HasSuffix(cur, []byte("\n")) {
			cur = append(cur, '\n')
		}
		return path, os.WriteFile(profile, append(cur, []byte(line+"\n")...), 0o644)
	}
}

// writeFileIn writes data to path, creating its directory first.
func writeFileIn(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
