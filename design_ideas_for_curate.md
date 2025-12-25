

# Curate Concept

Curate is the interactive text interface for Antenna App. It is the default interface designed for novice and casual use.  When completed it'll include support for all the actions provided in the scriptable command line interface. It'll also provide a more fluid means of managing metadata than hand coded YAML.

## Origin idea

I want an easy way to perform the actions supported by Antenna App when using the command line isn't the right choice. Something interactive but still in the terminal. In an ideal world all the YAML configuration and front matter would be automatically generated and editable without having to know YAML. An interactive command could move the Antenna App in that direction.

## the command

The command I'm describing is `antenna curate` or simply invoking `antenna` without any options. If an antenna.yaml file is not in the current working directory you'll be prompted to initialize an Antenna managed website in the current directory. If initialization has already occurred then you'll be shown a list of currently defined collections.

## Curating collections

There are several possible actions at the site level. These commands can be run on all collections or specific ones. It also provides a means of curating  a specific collection. The  following actions are planned or being implemented.

- initialize a project
- list available collections
- add a collection
- del a collection
- harvest all or a single collection
- generate HTML, RSS, OPML, Sitemaps for all or a single collection
- curate items in a collection
- action recorder (on FILENAME | off) 
  - saves the the cli commands representing you interactive actions taken
  - generate shell and PowerShell scripts of action sequence with comments
  - action player that will execute the shell or PowerShell script generated

### Curate items in a collection

When you enter this menu items are listed. You can review help to see the implemented actions. The follow actions are either implemented or planned.

- list items
- set items status
  - to published, review or clear status
- list pages in a collection
- render pages in a collection
- choose and apply a theme to collection
- render posts as HTML
- apply the collection's default SQL filters

## Curate pages in a collection

- list pages
- render pages or page as HTML

## Curate a themes

TBD

- list themes
- create/add a theme
- edit elements in a theme
- apply theme to collection
- remove theme from collection (e.g. set back to default theme)

## Curate default and named filters

TBD

### Curate the antenna App configuration

TBD

### someday, maybe ideas

I think most writers like to use their favorite writing tool so when I say "edit" I really meaning envoking a writer's favorite editor and passing content to it be edited. Antenna App shouldn't privide an editor but should use the one provided by the system.

- named custom filters
- edit collection Markdown front matter
- edit collection Markdown text
- edit item Markdown front matter
- edit item Markdown text
- blogit will import a Markdown document, setting up appropriate front matter
- post adds an item to a collection's feed from a local Markdown document
- unpost removes a local Markdown document from feed
- pages lists the pages so you can curate them
- posts lists just the post items so you can curate them
- antenna.yaml manager, allow easy configuration without knowing YAML
- a tool for create common SQL filters, sort of a Q & A dialog that results in generate SQLite3 SQL for the collection



# Notes

One of my goals with antenna command line tool is to make the actions highly scriptable. This is helpful for when your own site finds it rythem of updates. The terminal interface should map to the command line version commands such that you can use the interactive interface to construct a script of actions you can run automatically.

I think the URL should display the command line version of the comments being presented interactively. This could be used to construct automations without resorting to direct shell programming. 

I was playing with the editor, [aretext](https://aretext.org), and the way it handles the UI would be interesting to explore for curate action. Take a closer look at the find and open method, how it displays files in the current directory and lets you scroll around it. It uses tcell module for UI. I need to decide if that's the direction I want to go with curate action or if I want something simpler.

