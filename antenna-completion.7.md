%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

completion

# SYNOPSIS

antenna completion SHELL [-install]

# DESCRIPTION

Writes a shell completion script to standard output. SHELL is bash or
powershell. The script completes actions, the top level options, help
topics after "help", "new" after "themes", shell names after "completion",
and file names after -config and for other actions.

With -install the script is installed for every new session instead of
printed. For bash it is written to
$XDG_DATA_HOME/bash-completion/completions/antenna (default
~/.local/share/...), replacing only a file that antenna wrote. For
PowerShell it is written beside the profile and one line is added to the
profile to load it, once however often you install. The path written is
reported.

# PARAMETERS

SHELL
: bash or powershell

-install
: install the script instead of printing it

# EXAMPLES

Bash, for the current session:

    source <(antenna completion bash)

Bash, for every session:

    antenna completion bash -install

PowerShell, for the current session:

    antenna completion powershell | Out-String | Invoke-Expression

PowerShell, for every session:

    antenna completion powershell -install


