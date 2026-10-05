completion — generate a Bash or PowerShell completion script

SYNOPSIS
  antenna completion SHELL

DESCRIPTION
  Writes a shell completion script to standard output. SHELL is bash or
  powershell. The script completes actions, the top level options, help
  topics after "help", "new" after "themes", shell names after "completion",
  and file names after -config and for other actions.

PARAMETERS
  SHELL  bash or powershell

EXAMPLE
  source <(antenna completion bash)
  antenna completion bash > ~/.local/share/bash-completion/completions/antenna
  antenna completion powershell | Out-String | Invoke-Expression
