

# Curate Concept

I want an easy way to perform the actions supported by Antenna App when using the command line isn't the right choice. Something interactive but still in the terminal. In an ideal world all the YAML configuration and front matter would be automatically generated and editable without having to know YAML. An interactive command could move the Antenna App in that direction.

## the command

The command I'm describing is `antenna curate`. The goal of the command is to provide an interactive terminal interface for curating our Antenna App collection(s).

The bare command `antenna curate` should display the list of collections discovered in the current directory. It should display their name, title and possibly description. There needs to be a means of selecting one of the collections. Below the list should be a list of menu options:

- init
- add COLLECTION
- edit COLLECTION (this will lead to a new menu for working with the contents of a collection or editing it's metadata)
- del COLLECTION
- harvest COLLECTION
- generate COLLECTION
- command history (on FILENAME | off) 
  - saves the commands composed into a text file

### Curate a collection

Command `antenna curate COLLECTION_NAME`. This drops you into your main work area for a specific collection.

- post/unpost/edit post
- page/unpage/edit page
- filters (default means of batch curating posts in a collection)
- command history (show the commands executed in this session so you can copy them into a a script)

### Curate a theme and page elements

TBD

### Curate the antenna App configuration

TBD

# Notes

One of my goals with antenna command line tool is to make the actions highly scriptable. This is helpful for when your own site finds it rythem of updates. The terminal interface should map to the command line version commands such that you can use the interactive interface to construct a script of actions you can run automatically.

I think the URL should display the command line version of the comments being presented interactively. This could be used to construct automations without resorting to direct shell programming. 

I was playing with the editor, [aretext](https://aretext.org), and the way it handles the UI would be interesting to explore for curate action. Take a closer look at the find and open method, how it displays files in the current directory and lets you scroll around it. It uses tcell module for UI. I need to decide if that's the direction I want to go with curate action or if I want something simpler.

