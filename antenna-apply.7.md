apply — apply a theme

SYNOPSIS
  antenna apply THEME_PATH [YAML_FILE_PATH]

DESCRIPTION
  Applies the theme at THEME_PATH to the page generator YAML file at
  YAML_FILE_PATH. If YAML_FILE_PATH is omitted, the default generator YAML
  (page.yaml) is replaced.

  A theme directory contains Markdown and YAML files whose names map to
  generator YAML attributes: header.md, nav.md, footer.md, head.yaml, etc.
  Run 'antenna help themes' for the full theme directory layout.

PARAMETERS
  THEME_PATH      path to the theme directory
  YAML_FILE_PATH  (optional) path to the generator YAML to update

EXAMPLE
  antenna apply theme/my-theme
  antenna apply theme/my-theme collection/feeds.yaml

