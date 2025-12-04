%antenna(5) user manual | version 0.0.17 f32fc59
% R. S. Doiel
% 2025-12-03

# NAME

antenna themes

# SYNOPSIS

A directory with page elements in Markdown or YAML

# DESCRIPTION

A directory with files that can be used to generate an antenna page generator
description. The antenna uses a page generator description YAML file to render
HTML pages. The YAML structure is organized around those elements that are in the
HTML head element as well as the body elements of HTML pages.

A theme is held in a directory. The directory name is used as the theme's name.
Inside the directory are zero or more files where their names map the YAML attribute
names in a page generator YAML file. Here is an example of a theme called "theme"
that can be applied to generate a generator YAML file.

~~~
theme\header.md
theme\nav.md
theme\top_content.md
theme\bottom_content.md
theme\footer.md
theme\head.yaml
theme\style.css
~~~

The following Markdown documents are used to express their related attributes in the
page generator YAML files. Markdown is used to express the HTML values that will be
used in the page generator file for these attributes. The elements describe form
the innerHTML of the body element in an HTML document. They are rendered in the
order presented if they are present.

header.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a header HTML element.

nav.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a nav HTML element.

top_content.md
: (optional, used when present) This Markdown document contains a Markdown
expressing the HTML that will appear after the nav element and before a section
element if present.

bottom_content.md
: (optional, used when present) This Markdown document contains a Markdown
expressing the innerHTML that will appear after section element and before
the footer element.

footer.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a footer HTML element. It is rendered before
closing the body element.

The head element's content may also be included in a theme. It is expressed as a
YAML file called "head.yaml". YAML is used because there 
is not a direct relationship between the element attributes and how they could be expressed
using Markdown. Most of the time the head.yaml isn't necessary in the theme because 
antenna generates most of the head elements' content automatically.  There are times when
my wish to enhance the generated content (e.g. include link elements pointing to files or
include script elements JavaScript). The head element's innerHTML is populated in the order of
meta elements, link elements and script elements if they are defined in the YAML as the 
attributes meta, link and script. Each of these top level YAML elements are list and the
individual items in the list express the attribute names and values that form that element.

title
: (optional, used when present) A page title represented as a string.

meta
: (optional, used when present) A list of objects expressing a sequence of meta 
HTML elements attributes. Each item in the list is formed from the attribute names
and values that are define in a meta element. See 
<https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/meta>

link
: (optional, used when present) A list of objects expressing a sequence of link 
HTML elements attributes. Each item in the list is formed from the attribute names
and values that are defined in a link element. See
<https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/link>

script
: (optional, used when present) A list of objects expressing a sequence of script 
HTML elements attributes.  Each item in the list is formed from the attribute names
and values that are defined in a script element. See
https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/script

style
: (optional, used when present) A string holding CSS to be injected as the last
element of the head when rendering HTML.

Here is an example "head.yaml"

~~~yaml
title: My theme based title
meta:
  - charset: utf-8
  - name: language
    content: en-US
link:
  - rel: alternate
    type: application/rss+xml
	href: archive.xml
  - rel: stylesheet
    href: /css/site.css
script:
  - type: module
    src: modules/myscript.js
style: |+
  /* This CSS will turn headings vertical */
  h1 {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    text-orientation: mixed;
  }

~~~

NOTE: In this example the last style element will override the H1 definitions
previously included in the CSS files using the link attributes.

# Also 

- [antenna (1)](antenna.1.md)


