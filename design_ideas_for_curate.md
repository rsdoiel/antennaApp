

# Curate Concept

I want an easy way to perform the actions supported by Antenna App when using the command line isn't the right choice. Something interactive but still in the terminal. In an ideal world all the YAML configuration and front matter would be automatically generated and editable without having to know YAML. An interactive command could move the Antenna App in that direction.

## the command

The command I'm describing is `antenna curate`. The goal of the command is to provide an interactive terminal interface for curating our Antenna App collection(s).

### Curate collections

The bare command `antenna curate` should display the list of collections discovered in the current directory. If the directory you are in does not have an Antenna App configuration file (antenna.yaml), it will prompt you to initialize the project. Otherwise you are dropped into the collection level curation options. The following are planned or being implemented.

- initialize a project
- list available collections
- add COLLECTION
- del COLLECTION
- harvest COLLECTION
- generate COLLECTION
- curate items in a collection
- command history (on FILENAME | off) 
  - saves the commands composed into a text file

### Curate items in a collection

When you enter this menu items are listed. You can review help to see the implemented actions. The follow actions are either implemented or planned.

- set items to published status
- set items to review status
- clear items status
- page through items
- apply the collection's default SQL filters
- if command history is active the commands you execute will be saved in the command history file

### Curate a theme and page elements

TBD

### Curate the antenna App configuration

TBD

### someday, maybe ideas

I think most writers like to use their favorite writing tool so when I say "edit" I really meaning envoking a writer's favorite editor and passing content to it be edited. Antenna App shouldn't privide an editor but should use the one provided by the system.

- edit collection Markdown front matter
- edit collection Markdown text
- edit item Markdown front matter
- edit item Markdown text
- blogit will import a Markdown document, setting up appropriate front matter
- post adds an item to a collection's feed from a local Markdown document
- unpost removes a local Markdown document from feed
- pages lists the pages so you can curate them
- posts lists just the post items so you can curate them
- themes lists the themes available
  - create, add, edit, apply or remove themes from this view
- antenna.yaml manager, allow easy configuration without knowing YAML
- a tool for create common SQL filters, sort of a Q & A dialog that results in generate SQLite3 SQL for the collection



# Notes

One of my goals with antenna command line tool is to make the actions highly scriptable. This is helpful for when your own site finds it rythem of updates. The terminal interface should map to the command line version commands such that you can use the interactive interface to construct a script of actions you can run automatically.

I think the URL should display the command line version of the comments being presented interactively. This could be used to construct automations without resorting to direct shell programming. 

I was playing with the editor, [aretext](https://aretext.org), and the way it handles the UI would be interesting to explore for curate action. Take a closer look at the find and open method, how it displays files in the current directory and lets you scroll around it. It uses tcell module for UI. I need to decide if that's the direction I want to go with curate action or if I want something simpler.

