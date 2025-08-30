
antennaApp provides a simple application supporting feed reading and link blogging. I based it on my experimental
[Antenna](https://github.com/rsdoiel/antenna). In creating the antennaApp I hope that others will experiment with
their own "Antenna" and help foster a more open, diverse, web.

The goal of my original Antenna experiment was to have a local news site based on feeds I could find covering my
geographic region. Over the course of the experiment the number of feeds I followed grew. Eventually they became
collections of feeds. The current experiment is running at <https://rsdoiel.github.io/antenna>. It where I go first
to read things on the web, followed by <https:/news.scripting.com> (the site that inspired by experiment) then to 
a short list of other sites I enjoy reading. 

In the process of creating and running my experiment I noticed I was avoiding much of the toxic wasteland that
is social media today but not feeling like I was missing out.  All it took was a basic understanding of how the web
works, even today. 

antennaApp is my attempt to create a simple program to help others do the same.  I've attempt to keep the functionality
simple but flexible. You do need to have some know. I presume have a text editor and know how to install programs
on your computer. I assume you can type commands into your computer's terminal application. 
I assume you can lookup how to create simple Markdown documents if you don't know Markdown or CommonMark already.
with that knowledge the antennaApp can help you create your own link blog running on your own computer for your
enjoyment. For those who are more adventuresome and familiar with static website publicatiojn, antennaApp can be used
with that to create public facing link blogs.

Here's a bit more detail after you've installed antennaApp and can run it in your terminal.

Step 1, create a text file called "example.md" using the following Markdown text.

~~~Markdown
---
title: Example of a antennaApp feed file
---

# Example #1

This is an example of a feed list expressed as Markdown. All text is ignored except
the list below.

- [R. S. Doiel blog](https://rsdoiel.github.io/rss.xml "Robert's blog RSS feed")
- [Scripting News](https://scripting.com/http://scripting.com/rss.xml "Script News RSS feed")
~~~

(NOTE: It's actually a fancy Markdown document. It includes some front matter in YAML.
you can get away with just the last two bottom lines that start with a dash and spaec "- ")

This Markdown document defines two feeds. All the extra text is just there to show it's a "normal" Markdown
file. As you grow the number of feeds listed in the document you may find it helpful to keep notes in the
same file. It also gives you an option of including this Markdown document in your link blog site.

Step 2. Setting up the antennaApp

Setting up a new Antenna is done by using the antennaApp program called "antenna".  The first time you create a new
Antenna site you want to initialize it. Those is fancy words for "set things up". In a terminal type the following
command (assumes antennaApp is already installed).

~~~shell
antenna init
~~~

This will create a file called "antenna.yaml". It holds the set up of your Antenna site. Now we need to "add" the
"index.md" that will define our first collection of feeds.

~~~shell
antennna add index.md
~~~

Your Antenna site is ready to recieve content.

Step 3, retrieve the feed content described in your Markdown file.

~~~
antenna update
~~~

Step 4, render the website.

~~~shell
antenna generate
~~~

You can preview the website with

~~~shell
antenna preview
~~~

There you have it, you a minimal link blog on your local machine. In the terminal window
you will see a URL that you can use in your web browser to see the preview site.

This is just the simple example. You can use antennaApp to add items to a collection and generate new
feeds with it too. antennaApp let's you curate both the feeds you read and the feeds you render as a local
website. 

The configuration files that are used by antennaApp are rewritten in a YAML notation. This is the same notation
commonly used for "front matter" in Markdown and CommonMark documents. By customizing the YAML files you can,
with a little knowledge of HTML, CSS and JavaScript, fully customize your link blog. You can even integrate other
tools like [PageFind](https://pagefind.app) to provide search across you Link Blog. The content of the antennaApp
are staged in a directory called "htdocs", this can be used to publish a public website on one of the many static
website hostign services.

The antennaApp also provides an opportuntity to "manage" and curate the contents of feeds, choosing what is
published and what is ignored. You can use it like a microblog by adding your own Markdown/CommonMark documents to a
localhost feed you manage. It all works because of Markdown and RSS. 

While I publish my antenna to my public site you don't have to. It is very useful as a "localhost" website. That is actually
how it started. Eventually I wanted to read it on my phone when I was out and about so I put it on my public website.

If you can run a command line program, have some knowledge of HTML and can create and edit documents in CommonMark or Markdown.