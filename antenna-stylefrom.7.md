%antenna(7) user manual | version 0.1.2 24f4114
% R. S. Doiel
% 2026-11-23

# NAME

stylefrom

# SYNOPSIS

antenna stylefrom INPUT_FILE [OUTPUT_PATH]

# DESCRIPTION

Extracts the embedded CSS from a LibreOffice Writer HTML export
(INPUT_FILE, .html or .htm) and writes it to OUTPUT_PATH. OUTPUT_PATH
defaults to theme/style.css; the directory is created if needed.

Use this action to seed a theme stylesheet from a styled Writer document.

# PARAMETERS

INPUT_FILE
: path to the LibreOffice-exported HTML file

OUTPUT_PATH
: (optional) output CSS path (default: theme/style.css)

# EXAMPLES

antenna stylefrom my-doc.html
antenna stylefrom my-doc.html css/libreoffice.css


