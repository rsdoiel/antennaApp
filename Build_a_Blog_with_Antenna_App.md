---
title: Build a Blog with Antenna
dateCreated: "2025-09-05"
dateModified: "2025-10-08"
datePublished: "2025-09-05"
author: R. S. Doiel
postPath: "blog/2025/09/05/Build_a_Blog_with_Antenna.html"
---

# Build a Blog with Antenna

By R. S. Doiel, 2025-09-05

Antenna is a feed oriented content management tool. In this tutorial it'll be used to build a very simple blog. I'll cover using the **antenna** command, describe how you can use Markdown for configuring the blog and posting content to the blog. Antenna is configured using YAML configuration files and HTML pages and posts are assembled using a YAML configuration.

## Setting up

The first thing we need is a directory to hold our website. That can be done from a terminal on Linux, macOS or Windows using the following command.

~~~shell
mkdir myblog
~~~

We want to change into that directory.

~~~shell
cd myblog
~~~

Now we're ready to begin.

## Initializing the blog

The Antenna application, `antenna` has an "initialization" action. This creates the configuration file needed for your blog. Here's what you need to type in the terminal (same for Linux, macOS and Windows).

~~~shell
antenna init
~~~

This will result in a file called "antenna.yaml" being created. This is the main configuration file for Antenna. The "init" action will also create a "page.yaml" file[^1]. This is a page generator description in YAML. It contains the information to take a post or list of feed items and render a web page. If you are using macOS or Linux you can type the following and see them.

[^1]: The "page.yaml" can be edited to define the HTML you want to include in pages and posts. It can also be easily themed using the "apply" action. See [theme.5.md](theme.5.md) for details.

~~~shell
ls -l
~~~

On Windows you would use the `dir` command instead.

~~~shell
dir
~~~

## Defining our blog collection

A blog is built from three parts

- file system path(s) holding the blog posts[^1]
- One or more HTML pages showing the available posts in reverse chronological order
- An RSS feed of the recent blog posts

[^1]: The file paths to posts can be whatever you want. Antenna doesn't impose an structure. Traditionally a structure broken down for year, two digit month and another for two digit days is common. So posts are contained in a director called "blog" it'll have a path broken to by year, month and day. The day directory holds the blog post. Example I have a post called "updates.md" for the date August 5th, 2025. That might be held in `blog/2025/08/05/updates.md` Markdown file.

In terms of Antenna these are contained in a collection which needs to be defined. Then
we use the post action to add Markdown documents, with front matter, as blog posts. The post
action also generates an HTML version of the post based on the settings in the front matter. Finally
the RSS feed and HTML aggregation page are render by Antenna's generate action. That's pretty much it.

Let's first create our collection. I am going to call it "index.md". The reason I call it "index.md" is that it will result in an HTML page called "index.html" as well as an RSS file called index.xml and an OPML file called "index.opml". 


~~~markdown
---
title: My blog
---

# Welcome to My Blog

This is My blog where I use the Antenna app to curate a simple blog.
The Markdown forms the definition of the "index.md" collection. The blog
will be managed in the "index.db" SQLite3 database. It can be configured by
modifying the "page.yaml" file generated when this collection is added to the
Antenna configuration using the "add" action.
~~~

That's all that is needed, save this Markdown document as "index.md". Now let's add this to our Antenna collection. We only need to do this once.

~~~shell
antenna add index.md
~~~

If you list the directory you should see "index.db" and "page.yaml". You can modify the "page.yaml" to set the various HTML elements that will host either the list of blog posts or the individual blog post content.

~~~shell
ls -1 index.* page.*
~~~

or for Windows.

~~~shell
dir index.* page.*
~~~

## Adding the first blog post

I am going to assume the first blog post is called "welcome.md". I am also going to assume you'll using a blog oriented directory structure. For this example today's date is September 5th, 2025. I need to create a place to hold my blog post "welcome.md". I will first create a directory to hold it.

~~~shell
mkdir -p blog/2025/09/05
~~~

On Windows

~~~shell
New-Item -ItemType Directory -Force -Path blog\2025\09\05
~~~

Now you need to open the a new file in that directory called "welcome.md". If you have VS Code or Codium installed as your editor you can do the following.

~~~shell
code blog/2025/09/05/welcome.md
~~~

Or for Windows.

~~~shell
code blog\2025\09\05\welcome.md
~~~

In that file create our welcome post. We need to include the following attributes in our front matter, "postPath", "link", "pubDate". Here's my version of "welcome.md" Markdown.

~~~markdown
---
title: Welcome
postPath: "blog/2025/09/05/welcome.html"
pubDate: "2025-09-05"
---

# Welcome

This is a demonstration of Blogging with Antenna.

~~~

NOTE: If you are on Windows, you'll want to use the version of postPath that looks like this, `postPath: blog\2025\09\05\welcome.html`. You can now add the the post using the post action.

~~~shell
antenna post index.md blog/2025/09/05/welcome.md
~~~

Or on Windows

~~~shell
antenna post index.md blog\2025\09\05\welcome.md
~~~

We're are almost done. You should see the version of welcome.md you created and a new "welcome.html" generated when you use the post action. We need to generate the index.html and index.xml files with the updated post.

~~~shell
antenna generate
~~~

You can preview your new blog post at `http://localhost:8000` using the preview action and pointing your web browser at that URL.

~~~shell
antenna preview
~~~

## Updating a post

Any time you run the post command on your Markdown file the post with the matching link and postPath gets updated. Below I open and update the "welcome.md" file. Then I post it again to regenerate the HTML page.

~~~shell
code blog/2025/09/05/welcome.md
antenna post index.md blog/2025/09/05/welcome.md
~~~

That's it you now can add and update posts for your blog. Antenna will manage the index.html and index.xml documents for you when you run the `antenna generate` command again. You then use the preview action to view it in your web browser.

NOTE: If this was your blog you'd change the value I used for the link element to match how your website is structured and use it's URL.  I used a localhost URL with port number just to keep things simple and to allow us to test using the `antenna preview`.

## Enhancing you blog 

You will likely want some navigation and other text on your blog pages. This is accomplished by updating the "index.yaml" file. In this file you can set the path to your custom CSS, to any JavaScript script you might want to include. You can also set your site header, footer and nav elements. Finally you can even include HTML elements before and after the generated content. Here's an example of a customized "page.yaml" file.


~~~yaml
link:
  - rel: stylesheet
    href: css/site.css
header: |
  <h1>Welcome to My Blog</h1>

nav: |
  <ul>
    <li><a href="/">Home</a></li>
  </ul>

footer: |
  <!-- your custom footer inner HTML goes here -->
~~~

Now re-post your welcome.md then generate followed by preview to see the changes.

~~~shell
antenna post index.md blog/2025/09/05/welcome.md
antenna generate
antenna preview
~~~

You're blog is staged you can now publish on the Internet using the tools provided by your host. If you're using GitHub that might mean committing and pushing to a specific branch setup for your website (see GitHub pages documentation). 

Happy blogging!
