---
title: Build a Blog with Antenna App
description: |
  A short tutorial on how to build a blog and micro blog with
  Antenna App

author:
  - name: R. S. Doiel
    email: rsdoiel@gmail
keywords:
  - blog
  - Antenna
dateCreated: "2025-08-29"
dateModified: "2025-09-29"
datePublished: "2025-09-29"
---

# Build a Blog with Antenna App

## What is Antenna App?

**antenna** is a command line program you can use as a feed oriented content management system. Antenna App focus on the relationship between pages, posts and feeds.  Markdown is the common markup used to curate a list(s) of feeds you're following. It is used to provide the content of web pages and posts.  This simplifies the work in curating your website. 

Metadata used to curate your posts and feeds is expressed as [Front Matter]() in your Markdown documents. Configuration of the Antenna App is done through a few YAML files, YAML being the way markup used in your Markdown's front matter. The Antenna app uses these to harvest feeds you are following, aggregate them and render HTML pages of the aggregations along with the RSS feeds holding the aggregated content. Similarly if you post your Markdown document into a collection it Antenna App will use your posts to create an aggregated posts page and RSS file that others can use to follow you.

By being feed oriented Antenna allows us to create a social web without relying on a central organization or platform. Focusing on RSS feeds allows us to have a peer to peer web and slower conversations than the walled gardens provide. Slow, being helpful, in encouraging more meaningful discourse than the doom scrolling to X, Meta, Google et el.

## Why Markdown and Front Matter?

Markdown is an easy markup to read, learn and type. Much easier than HTML used by your web browser. Because it was designed to be transformed in to HTML's core features it is easy to rendering the HTML needed by your web browser. Antenna application provides an easy way to turn a list of feeds expressed as Markdown into an aggregation of those feeds expressed both as HTML and RSS. You can think of this is inbound social content from the web.  RSS today is provided by many application as well as social network. RSS 2.0 unpins podcasting and textcasting.  BlueSky and Mastodon produce RSS output you can follow. Mature content systems like WordPress also produce RSS feeds you can consume. By focusing on RSS we take advantage of a wide ecosystem of content ranging from news websites to personal blogs and micro blogs. Even platforms like Substack, Medium and friends produce RSS. Many newsletter systems also produce RSS so you can often started out reading a newsletter in your email, then switch to their RSS feed when you decide you want to aggregated into your Antenna website.

Front Matter provides metadata, data about the document. In the case of your Markdown document defining the feeds you follow the Front matter provides the metadata used in the resulting RSS feed. This include information like title of the aggregation, description of the aggregation. If you are posting to a collection then the front matter in the post will indicate where in the your website's directory structure you want to place the HTML file and use that location when generating the resulting RSS item. Additionally you can include optional item level data like a title and modification date or even enclosed content like the path to an audio or video file if you're curating a podcast with Antenna.

Markdown and Front Matter make it easy to
 
- read, learn and type than HTML or database languages
- generate HTML for posts, pages and aggregations 
- generate RSS feeds for posts and aggregations
- curate a website for your use and if desired to share on the public web

<aside>While there are plenty of simple markup languages out there Markdown has become ubiquitous. John Grubber's Markdown has a nice balance of easy to read and easy to learn while also easy to type.  There are some extensions to Markdown that have also become nearly ubiquitous and those extensions have been aggregated into a markup called [Common Mark](). While Grubber has been be clear about what his Markdown is the community creating the comment often embraces the features of Common Mark while referring to their dialect of markup as Markdown. The Antenna App focuses on Common Mark which is effectively a superset of John Grubber's Markdown. But since most people have heard of Markdown and my not know or make the distinction about the extended features of Common Mark I will also be talking about Markdown. The Antenna app is a feed oriented content manage tool for Markdown documents as well as popular Common Mark features.</aside>

## What are the limitations of Antenna App

1. It's an experimental program, will have bugs
2. It's an experimental program so will evolve over time
3. It doesn't provide a text editor or file system management
4. It's not a GUI tool

## What are it's strengths

1. It makes generating HTML and RSS files easy
2. It is easy to curate a list of feeds as a Markdown document then read them aggregated as an HTML page or RSS feed 
3. You write using your favorite editor pages
4. The resulting website is suitable to view locally using Antenna's preview web server
5. The resulting website can easily be published to the public web via platforms like GitHub pages, Dream host, S3 web buckets
6. Antenna app is suitable for rendering Blogs, Wikis and other websites that use a structured path for content organization

## Why Markdown


Markdown is capable of expressing the core features of HTML documents but is easier to read, learn and type.  There are plenty of editors that you type Markdown. The Markdown document itself is a simple text file. That means it is likely to be readable decades from now unlike old Word documents. There are many tools that can convert Markdown to other formats (e.g. PDF, ePub, Word) such as [Pandoc](https://pandoc.org). That means the Markdown you use with Antenna isn't a dead end.

Markdown expresses all the core features of HTML that make the web a seminal hypertext experience. It supports emphasis, inline source text, links, lists, paragraphs. Since Antenna App also supports Common Mark (Markdown plus common extensions) it can even create tables with it or include images with captions and figures. Markdown does this without getting bogged down in the XML like semantics of HTML. It is a light weight markup intended to remain readable as well as easy to type.

## What is front matter and metadata?

Front matter is term in the Markdown community that refers to metadata (information about the Markdown document) embedded at the start of the Markdown document. The Jekyll static site generator used by GitHub popularized the use of YAML front matter as a means of guiding rendering of the Markdown documents. Similarly systems like RStudio's rMarkdown used YAML front matter to control how RStudio processed the Markdown document. Antenna application follows in that tradition. It uses YAML front matter in the generation of the aggregated RSS feeds from a Markdown document defining a feed collection. It used YAML front matter to know how to process a post and to enhance the metadata provided at the item objects of RSS feeds. 

The front matter at the beginning of a Markdown document starts with a first line formed of three dashes, "---".  It continues until it encounters a matching line of three dashes. Between those Front Matter is typically expressed as YAML. Here's an example of front matter indicating information about the title, author, description, publication date.

~~~YAML
---
title: Building a Blog with Antenna
description: |
  A description of Antenna application and how it uses Markdown and YAML to
  render web pages, posts and websites.

author: R. S. Doiel
datePublished: "2025-09-29"
---
~~~

YAML is a simple way to express structured data. The data can be information you would use in a citation or in a bibliographic record.

## Getting started with Antenna app

The Antenna App is a Markdown friendly content system.  By to make a content system that is feed oriented you need more than the document content typed out in Markdown. A post as an example will need to have a publication date, it probably uses a structured path (e.g. "year/month/day") for organizing content. If you're creating a wiki organized around keywords then that path might be structured differently. In all cases the posts whether blog, wiki or hybrid still need some level of additional metadata to facilitate automatic linkage and feed generation.  In the Markdown community the concept of a documents' [Front Matter](https://en.wikipedia.org/wiki/Book_design#Front_matter) as a means of expressing that. How does that fit with Markdown?


Front Matter is at the very top of the Markdown document. It starts with three dash characters, "---" on a line by themselves and ends with a matching three dashes, "---" on a line by themselves. Anything between the pair of three dash lines is considered structured data expressed in YAML, JSON and sometimes TOML. YAML has come to be the most common form of front matter. What is YAML?

YAML is a simple notation that evolved in response to the challenges of reading and typing JSON much as Markdown evolved in responses to challenges of reading HTML and type it. You can find out more about YAML at [yaml.org](https://yaml.org/). In this post I will illustrate some common metadata as YAML. It is simple enough that you'll probably pickup in the same manner that you learn Markdown. See an example, try it out, then correct things as they come up.

To support feeds the metadata that is most helpful is bibliographic in nature. It's includes information like title, description, author, date create, date modified and date published. Antenna App also includes postPath to know where to put the posted item in your web tree as HTML and how to link to the HTML in the generated RSS file.

The nice thing about using the front matter in you Markdown document for metadata is it is easy to keep up to day when you are writing, can be used by tools like Antenna App and [FlatLake](https://flatlake.app) to generate HTML or enhance your website with a JSON service. And can be rendered into the resulting HTML documents so tools like [PageFind](https://pagefind.app) can provided effective browser side search of your website.

### How is content manage

Antenna app works in a directory where you are managing your web content.  It provides actions used to curate the site. These include the addition of feed collections that can be rendered as aggregated HTML files and RSS feeds. It supports posting and unposting content to a collection (this allows for wikis and blogs). It can also simply be used to translate a Markdown document into an HTML page using a generative process described in YAML.

Feeds aggregated into a collection are defined by a Markdown document containing a list of feeds. The front matter of a collection document is used to populate the channel level data in an RSS feed. The channel metadata provides an description of the collection of the feeds being aggregated.

A collection can be defined without feeds in which it can be curated simply by posting or unposting Markdown documents.

The Antenna app have a generate command that will take the current collections and render out their HTML pages and RSS files.  It includes a harvest command to retrieve any feeds that are being aggregated.

Why does Antenna support aggregated feeds?  The Web is a communications platform and aggregation of feeds allows you to create a social website. One where you expose what you are reading and can comment on via posting to your aggregated feed. The allows for a peer to peer(s) relationship to emerge. If I subscribe to your feeds and you to mine we are "following" one another and can see what we post to our own feeds.

Unlike systems like BlueSky and Mastodon there is no need for centralization. Each person's aggregations and posts works separately and asynchronously. If you need real time coordination there are many means to achieve these from old fashioned phone calls, to video chat, text chat, etc. I believe an asynchronous offline experience can lead to more meaningful and slower discussions. Slower communication can break the rage cycle and doom scrolling so prevalent in the walled gardens of the modern web.

In spite the of much prophesied end of blogging, most news organizations still produced RSS feeds. This is because most also provide podcasts and video casts that depend on RSS to exist. Add to that platforms like WordPress, Mastodon and BlueSky that also produce RSS feeds and you'll find there is a rich ecosystem of feed content out there.  Antenna provides you an convenient way to aggregate that into your own news page. No need to go to Google, Yahoo, Facebook, Instagram to read news. Just curate your list of feeds in a simple Markdown document, and run `antenna harvest` to get the latest.  Run `antenna generate` followed by `antenna preview` and you an read your aggregated feeds in you web browser in the privacy of your own computer.

Curating your Antenna website then is a mater of choosing to post or not and updating your list of feeds. You can see an Antenna generated website at <https://rsdoiel.github.io/antenna>. I've curated lists of feeds around regions and the blogs. The lists are control by mean and reflect my interests. If feed becomes bother some I remove it form my collection. If I discover a new website with a feed I can add it to my collection.  Because the feed is a simple Markdown document it is easy to read and share. I can even choose to include notes about the feeds in the Markdown document because it's just a document and it can be converted easily into HTML.

## Antenna App knits together Markdown documents, front matter to produce a website.

The Antenna application embraces the feed and by that the idea of "post". Posts are written in Markdown with meta data about the post expressed as front matter encoded in YAML.  Similarly defining a collection of feeds to curate is also expressed as a Markdown document.

The tooling provided by the **antenna** command takes care of knitting together the Markdown documents defining collections and Markdown documents hosting posts. The **antenna** command's initialization process takes care of create the YAML configuration files that further define the organization and appearance of the website leaving you to spend your effort on creating content using the simple and familiar Markdown text notation.

### Create and curate your site simply

Antenna takes a different approach to many static site rendering or publishing systems. It relies on simple Markdown documents to define a collection, YAML front matter for metadata curation and a few YAML configuration files. The configuration files are generated by **antenna** automatically but they are easy to customized to your needs.

If you know how to use a text editor, can pickup Markdown for typing your content and can edit and modify YAML that contains HTML elements for laying out the page you're in luck, **antenna** may work for you.

## What is required?

You will need a text editor to create and edit Markdown documents and YAML files. You will need a web browser to review your website. You will need to install the latest version of Antenna App, See <https://github.com/rsdoiel/antennaApp/releases>. Currently Antenna App is experimental, a working proof of concept. It is expected to evolve over time. Getting the latest release is recommended.

The Antenna App is available for Raspberry Pi OS, Linux, macOS and Windows. Find the latest zip file in the releases section of the GitHub repository and download it. You can unzip it and then copy the executable to someplace in your path.  The program runs from a "terminal" and is a command line program. The command line is simple and is based around the idea of an action and parameters needed to complete the action. See the manual page for details, [antenna](antenna.1.md).

Install the latest version of **antenna* executable. Check the installed version

~~~shell
antenna -version
~~~

You should see a version number and release hash.

## How do I set it up?

1. Create a new directory for your project called "blog_demo"
2. Change into it.
3. From the directory run the Antenna init action
4. Create a collection called "index.md"
5. Add the "index.md" collection to Antenna

Here's the steps in a Windows terminal.

~~~shell
mkdir blog_demo
cd cd blog_demo
antenna init
~~~

You need to define at least one collection. In this case the collection I want you to create is called "index.md". Collections are defined by a simple Markdown document. Top level list anchors are treated as feeds. They are formed like this, `- [FEED_LABEL](FEED_URL "FEED_DESCRIPTION")`. If YAML front matter is included in the document then that data is rendered as Channel metadata when the RSS 2.0 file is generated. In our simple blog model we don't have any feeds to include. Our collection definition can look like this.

~~~markdown
---
title: My Blog
---

# My Blog
~~~

Save the above Markdown as "index.md".  You can put any Markdown in this file. Once you've save "index.md" I want to to "add" it to the Antenna collections using this command.

~~~shell
antenna add index.md
~~~

At this point you can check the directory and you should see several files.

antenna.yaml
: This is the main configuration for the antenna project.

index.yaml
: This is a generator configuration for generating index.html form the index.md collection.

index.db
: This is an SQLite3 database with a channels and items table. It forms the core of the content management tool.

Run `antenna generate`, this will generate two additional files.

index.html
: This is a collection aggregation page, it'll be where your posts show up

index.xml
: This is an RSS 2.0 feed file of your blog site.

## How do I add a Post?

Antenna App supports two flavor of posts. Traditionally blog posts have a landing page
containing the contents of the post. On disk they are organized around a simple nested directory structure where the outer directory name is the year, inside is a two digit month directory and inside that is a two digit day directory. The day directory will hold your blog post.

When making a blog post you must provide some YAML front matter. You need to include the following fields at a minimum.

pubDate
: This is a string holding the date in YYYY-MM-DD format.

postPath
: This is the path where the HTML file will be written, example "posts/2025/08/29/mypost.html".

link
: This is the URL the browser will use to read the post.

After this minimal front matter you can write the post in Markdown. Antenna App will translate that to HTML and include the metadata needed for generating RSS 2.0 feeds for the "index.md" collection.

Here's an example of a "helloworld.md" post documents.

~~~markdown
---
title: Hello World!
link: http://localhost/posts/2025/08/29/helloworld.html
postPath: posts/2025/08/29/helloworld.html
---

# Hello World!

This is my first post.

~~~

Save this file as "helloworld.md". Now let's "post" it using the antenna command and add it to the "index.md" collection.

~~~shell
antenna post index.md helloworld.md
~~~

When you do this you should see a directory create for "posts/2025/08/29" and it'll contain "helloworld.html". If you need to update the post you check repeat the command. The link you include int he front matter is treated as the GUID, the unique identifier for the post.  When you run `antenna generate` the RSS feed and aggregation page will be updated to include your new post.

The post front matter includes support for most the item feeds defined in the RSS 2.0 specification.This is includes fields like title, description, author. 

## How do I add a Micro-blog Post?

Micro blogging has become very populate. Micro blog posts are just short blog posts. Usually they do not have a title. Sometimes you want to write something in the flow and not generated a landing page with the commend by itself. Antenna App supports that feature too. The distinction is that you do not include a title, postPath or link in the front matter. You just include the pubDate string of when you want it to show up.

Here's a simple Markdown Micro blog post example.

~~~Markdown
---
pubDate: "2025-08-29"
---

Hi There!
~~~

Save this post as "hithere.md". Then "post" it using the antenna.

~~~shell
antenna post hithere.md
~~~

## Generate the site

Blog posts HTML are create when you post the Markdown document. The rest of the site pages are not regenerated yet. To generate the HTML for the aggregation page (list of blog posts) and the RSS Feed file you use the "generate" action.

~~~shell
antenna generate
~~~

You can then "preview" your website using your web browser at "http://localhost:8000" by run the following command.

~~~shell
antenna preview
~~~

You can press Ctrl-C in the terminal window to end the preview.

## How do I customize the pages?

FIXME: I need to document the YAML and how it works for defining aggregation pages.
