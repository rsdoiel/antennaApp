%antenna(7) user manual | version 0.2.1 8278d2c
% R. S. Doiel
% 2026-11-25

# NAME

init

# SYNOPSIS

antenna init

# DESCRIPTION

Creates the two configuration files antenna needs if they do not exist:

  antenna.yaml  main configuration (htdocs path, port, collections list)
  page.yaml     page generator (link, meta, nav, header, footer, scripts)

Also creates a default pages.md collection and pages.db database.

A new antenna.yaml has max_items_per_feed: 1000, a cap on how many of its newest
items one feed shows (see antenna help configuration). An antenna.yaml that
already exists is left as it is and gets no cap.

After running init, run 'antenna css' to generate a starter stylesheet
and automatically link it in page.yaml.

# EXAMPLES

mkdir myblog && cd myblog
antenna init
antenna css


