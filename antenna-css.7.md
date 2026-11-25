%antenna(7) user manual | version 0.2.1 8278d2c
% R. S. Doiel
% 2026-11-25

# NAME

css

# SYNOPSIS

antenna css [CSS_PATH]

# DESCRIPTION

Writes a comprehensive starter stylesheet to CSS_PATH within the htdocs
directory configured in antenna.yaml. If CSS_PATH is omitted it defaults to
css/site.css. Directory levels are created automatically.

If a stylesheet already exists at the target path it is backed up to
CSS_PATH.bak before being overwritten.

After writing the CSS, antenna patches the generator YAML (page.yaml) to
add a link entry referencing the new file. If page.yaml already has a link
section, antenna prints instructions for adding the entry by hand.

The generated stylesheet includes CSS custom properties for colors, fonts,
and layout, dark-mode support, skip-navigation link, navigation bar, article
cards, standalone pages, site footer, and typography.

# PARAMETERS

CSS_PATH
: (optional) path relative to htdocs (default: css/site.css)

# SEE ALSO

antenna help accessibility
antenna help themes


