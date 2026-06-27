css — generate a default CSS stylesheet

SYNOPSIS
  antenna css [CSS_PATH]

DESCRIPTION
  Writes a comprehensive starter stylesheet to CSS_PATH within the htdocs
  directory configured in antenna.yaml. If CSS_PATH is omitted it defaults to
  css/site.css. Directory levels are created automatically.

  If a stylesheet already exists at the target path it is backed up to
  CSS_PATH.bak before being overwritten.

  After writing the CSS, antenna patches the generator YAML (page.yaml) to
  add a <link rel="stylesheet"> entry pointing to the new file. If page.yaml
  already has a link: section, antenna prints instructions for adding the
  entry by hand instead of modifying the file automatically.

  The generated stylesheet includes:
    • CSS custom properties for colors, fonts, and layout (easy to override)
    • Dark-mode overrides via @media (prefers-color-scheme: dark)
    • Skip-navigation link (WCAG 2.4.1 — required by default HTML output)
    • Navigation bar, article cards, standalone pages, and site footer
    • Typography for headings, code blocks, blockquotes, and tables

PARAMETERS
  CSS_PATH  (optional) path relative to htdocs (default: css/site.css)

EXAMPLE
  antenna css
  antenna css css/custom/theme.css

SEE ALSO
  antenna help accessibility
  antenna help themes

