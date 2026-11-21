%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

init

# SYNOPSIS

antenna init

# DESCRIPTION

Creates the two configuration files antenna needs if they do not exist:

  antenna.yaml  main configuration (htdocs path, port, collections list)
  page.yaml     page generator (link, meta, nav, header, footer, scripts)

Also creates a default pages.md collection and pages.db database.

After running init, run 'antenna css' to generate a starter stylesheet
and automatically link it in page.yaml.

# EXAMPLES

mkdir myblog && cd myblog
antenna init
antenna css


