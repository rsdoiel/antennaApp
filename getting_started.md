
# Getting Started

The Antenna App is a program desided to allow you to easily create and curate websites, blogs and personal news sites with the following knowledge a knowledge of Markdown, the software that comes with your computer and its web browser. Here's the software you should already have.

- Text Editor
- Web Browser
- Terminal

You'll also need the [Antenna App installed](https://rsdoiel.github.io/antennaApp/INSTALL.html). In my examples I'm using a Raspberry Pi computer. The instructions should work on macOS or Windows with minor adjustments.

The things we'll be doing inside the Terminal application include the following.

- list files
- create a directory
- show the name of the current working directory
- change your working directory
- edit files in the working directory using your text editor
- run the `antenna` command in the terminal
- how to start your web browser and navigate to the preview URL

## Terminal Basics

On a Raspberry Pi computer running the graphical desktop you launch the terminal application by clicking on the Raspberry icon at the left of the top menu bar. Click on the "Accessories" menu, then find and click on "Terminal". Once the Terminal application has started, click into the main window. You're now ready to start.

NOTE: If you are not running the graphical desktop you are already in a terminal when you log into your Pi. NOTE: Firefox and Chrome don't work without the graphical desktop installed but `lynx` and `elinks` work well in a terminal[^1].

[^1]: To install lynx and elinks on Raspberry Pi OS, type the following into the Terminal, `sudo apt install -y lynx elinks`. macOS and Windows differ on how you install software.

### Prep Work

In the Terminal application we need to create we need to create directory where we'll hold our website. A website is a collection of documents so let's create a project directory inside the Documents directory setup with your Raspberry Pi. I'm going to call our first website "project1" and create it in the Documents directory but before that I want to show you how to find out what your current working directory is.

~~~shell
pwd
~~~

On my Raspberry Pi computer this shows something like this.

~~~
/home/pi
~~~

In this case "pi" is the username. If my username had been "robert" the path returned by `pwd` would look like this.

~~~
/home/robert
~~~

When you startup Terminal the working directory is usually set similarly to the above. This is referred to as your HOME directory. You can always go back to your HOME directory by using the "cd" command without any additional parameters, e.g. type in `cd` press enter. Use the `pwd` command to confirm your working directory.

~~~shell
cd
pwd
~~~

Now let's change into our Documents directory, then we'll create a "project1" directory for this tutorial. Here's the sequence of comands I want you to type in the terminal.

~~~shell
cd
cd Documents
pwd
~~~

The `pwd` should show a directory something like,  `/home/pi/Documents`. Now we're ready to create our "project1" directory using the "mkdir" command before changing into it and verifying we're in the right place.

~~~shell
mkdir project1
cd project1
pwd
~~~

You should see that you're in a directory like "/home/pi/Documents/project1"[^2].

[^2]: On Raspberry Pi OS paths (like your  current working directory) are case sensitive. This means you need to pay attention to capitilization when you're creating a new directory as well as changing directory.

Congratulations you've completed the prep work! This setup is only done the first time you start an Antenna App project[^3].  You can get back to this working directory by using the the  "cd" command to return to your HOME directory, change to the project1 directory. You can confirm you're in the right place using the "pwd" command.

[^3]: The directory will now exist on your computer until you remove the directory and its contents

~~~shell
cd
cd Documents/project1
~~~

### Starting up Antenna App

I need to have previously installed [Antenna App](https://rsdoiel.github.io/antennaApp) to complete this section of the tutorial. Here's what  I'll be covering in this section.

1. Confirm you are in the "Documents/project1" directory
2. Run the `antenna` command to setup the project and pages collection
3. Create an initial Markdown document called "index.md", this is our "home page" for the site
4. Add the page to our pages collection
5. Render the website and preview it in our web browser

In the Terminal, type the "pwd" command to confirm you're still in the "Documents/project1" directory.

~~~shell
pwd
~~~

If you don't see are result similar to "/home/pi/Documents/project", change into it now. You can use the "cd" command to do that. Type of the following if needed.

~~~shell
cd
cd Documents/project1
~~~

Now let's run the `antenna` command. This will startup an interactive text program for creating and curating your "project1" website. The first time you run the command in this directory it'll prompt you to intialize the project. Type "yes" and press enter when your read the prompt. Initialization will create several files in this directory. Type the following in the terminal.

~~~shell
antenna
~~~

As mentioned type "yes" and press enter if you haven't done so. The way the interactive program works is it presents you with a list of things you can act on. After the  project is initialized the first time you'll see a list of collections to manage. The main collection, the one automatically created, is called "pages.md".  For a  new project that should look something like this.

~~~
Curate Collections

	 1: pages.md, An Antenna Website

(1/1, [h]elp or [q]uit): 
~~~

If you type "h" and press enter you'll be taken to the help page for managing collections. That should look something like this.

~~~
Curate collection(s). Command syntax.

  NUMBER ENTER_KEY
  ACTION [NAME] ENTER_KEY

Actions:

NUMBER
: curate collection NUMBER

+NUMBER or -NUMBER
: Page by NUMBER of items through list

[a]dd
: Add a new collection. You'll be prompted for a Markdown filename, a title
and description. If Markdown file already exists it'll be used along with any front matter providing title and description.

[d]elete NAME|NUMBER
: Remove collection with NAME or collection NUMBER from configuration. It does not remove the files from disk.

[H]arvest [NAME|NUMBER]
: Harvest all collections or one specified by NAME
or NUMBER

[g]enerate [NAME|NUMBER]
: Generate all collections or one specified by NAME
or NUMBER

[h]elp
: Display this help

[q]uit
: To quit

(NOTE: Pressing enter without an action will page through results)

Press enter to exit help.
~~~

As more actions become available in newer versions of Antenna App this page will get longer. This is help is an example from an early release of Antenna App (v0.0.20). Press enter again to exit the help.

Now type "q" and press enter. You should now be back at the Terminal. You can use the "ls" command to list the  files in the project1 directory and see what was setup.

~~~shell
ls
~~~

You should see something like this.

~~~
antenna.yaml  pages.db  pages.md  page.yaml
~~~

These files are managed by Antenna App. Here's a run down of the role they play.

antenna.yaml
: This is the main configuration file. It describes the website in this directory and where to find the files needed to run.

pages.db
: This is a SQLite3 database holding the pages collection. A collection will hold the metadata and content used in your website.

pages.md
: This is a collections definition file. Open it in your text editor and take a look.  It's just a simple Markdown document. If you were creating an 

pages.yaml
: This describes how pages are assembled.

You don't need to worry about editing these directly but there not terribly mysterious either. The collection Markdown file is just a Markdown file containing  some text alongside YAML front matter that holds metadata about the collection. The to files ending in ".yaml" are YAML formatted files.  YAML is a simple notation to express structured data, again metadata used to manage your Antenna App project.

## Customizing project1

The initial version of "pages.md" looks something like what follows.

~~~markdown
---
title: An Antenna Website
description: This is the default websites created by the antenna init action.
---

# Welcome to your Antenna

~~~

The section that starts with "---" and ends with "---" is sets is called "front matter". It expresses two pieces of metadata, a title and description. The ": " seperates the metadata name from it's value. After the ": " is a line of text. This is the value associated with the title or description metadata elements.

Let's change the line "title: An Antenna Website" to "title: Project 1 Website" and change the line "description: This is the default websites created by the antenna init action." to "description: Our first Antenna App Project.".  Below  the second "---" is the body of the Markdown document. The line with "# Welcome to your Antenna" is a page title. When the Markdown is converted into HTML it will transformed into a section heading element[^4]. Basically it functions like a page title. Let's change this to "# Project No. 1". Save your changes.


[^4]: In HTML this is called an "h1" element.

Your updated file should  look like the following.

~~~markdown
---
title: Project 1 Website
description: Our first Antenna App Project.
---

# Project No. 1

~~~

You can add any additional Markdown content to the body you like. If we want to aggregate feeds from other sites, the Markdown body is used to define the feeds meaning aggregated with the "pages.md" collection. I often just put notes about the website project in the body.

## Adding a home page

Using your text editor, create a file called "index.md" in the "project1" directory. We'll start with a simple Markdown file like the one below.

~~~markdown
Hello from Project No. 1
~~~

We can add to our website using the `antenna` application by adding it to our collection. 







