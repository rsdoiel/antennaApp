---
title: Simple websites using Markdown
author: R. S. Doiel
dateCreated: "2025-10-10"
dateModified: "2025-10-14"
datePublished: "2025-10-14"
description: |
  Markdown allows many more people to participate in creating content for the web. The [Antenna App](https://rsdoiel.github.io/antennaApp) is a tool for creating Markdown focused websites. This post explores creating a simple website using Antenna App,
  your favorite text editor and your web browser.

keywords:
  - Markdown
  - CSS
  - Web
  - aggregation
  - site generator
  - microblog
  - blog
  - link blog
---

# Simple Websites using Markdown

> The [Antenna App](/antennaApp) is shaping the way I write and build websites

By R. S. Doiel, 2025-10-14

The Web is a __networked hypertext system__. A human friendly way of expressing hypertext is [Markdown](https://en.wikipedia.org/wiki/Markdown). Markdown has allowed many people to participate in creating content for the web. I think it deserves more focus. Between Markdown and a sprinkle of CSS[^2] and you can create entire websites using your own computer or hosted on a static site service[^1]. In this post I show how Markdown can be used for basic websites[^2]. All it requires is knowing a little Markdown, a text editor, web browser and the [Antenna App](https://rsdoiel.github.io/antennaApp).

[^1]: Examples: [GitHub Pages](https://pages.github.com), S3 buckets setup as static websites at AWS, Dreamhost, my favorite [Mythic Beasts](https://www.mythic-beasts.com/)

[^2]: Oatcake CSS is a nice example of Markdown oriented CSS styling, see <https://www.seanh.cc/oatcake/>

Demonstrated in this post:

- A means of creating a web page (HTML page) using Markdown and the Antenna App
- A means of creating a simple multi page website with navigation using an Antenna App theme
- A means of styling the web site using CSS expressed as part of an Antenna App theme 

[^2]: Antenna supports creating many types of websites. Examples include [microblogs](https://en.wikipedia.org/wiki/Microblogging), [blogs](https://en.wikipedia.org/wiki/Blog), [linkblogs](https://en.wikipedia.org/wiki/Linklog), [wikis](https://en.wikipedia.org/wiki/Wiki) and hybrids of these. Pretty much what you need to be social on the web without resorting to the silos of Big Corp.

## What you need to get started

- A computer
- A terminal app
- A text editor
- A web browser
- Antenna App

Chances are you have the first four if you're reading this. What I describe in this ports should work on computers running Windows, macOS and Linux based systems. If you need to acquire a computer and already have a HDMI television I recommend getting a Raspberry Pi Computer. The [Raspberry Pi 500 Desktop Kit](https://www.pishop.us/product/raspberry-pi-500-us-complete-kit/?searchid=0&search_query=Raspberry+Pi+500+Desktop+) is a nice middle range device that runs about $120.00 US. It includes a book about using Raspberry Pi computers. If you can afford a bit more the [Raspberry Pi 500+ Desktop Kit](https://www.pishop.us/product/raspberry-pi-500-desktop-kit/) runs $230.00 US. That's a __nice computer__ with solid state storage and 16 Gig of RAM.  That is enough bells and whistles to do light duty image, audio and video work. If you are on a budget then you can squeak by with a Raspberry Pi Zero 2W running the Antenna App software and a plain old text editor in the console. That setup will set you back about $50.00 US with the minimum accessories needed like case, keyboard, mouse, power supply and cables to connect things up. The Official Raspberry Pi Handbook print edition runs between $25.00 US but the electronic edition that you can read on your phone or tablet is free. My point is that for a little bit of investment or using your existing resources you have the means of building websites for yourself and others. You don't need to be a programmer or even a web designer if you are comfortable editing text and doing a bit of writing. 

The [Antenna App](https://rsdoiel.github.io/antennaApp) is software I wrote to prove to myself that building websites could be made simpler. I currently use it to curate my own websites include my blog and personal news site, <https://rsdoiel.github.io/antenna> and <https://rsdoiel.github.io/antennaApp/INSTALL.html>.

## Installing Antenna App

If you need to install the Antenna App you need to use your terminal to run the following command on macOS and Linux like systems.

~~~shell
curl https://rsdoiel.github.io/antennaApp/installer.sh | sh
~~~

or on Windows in Powershell you can use this command.

~~~shell
irm https://rsdoiel.github.io/antennaApp/installer.ps1 | iex
~~~

You can find detailed instructions in getting the latest version of the Antenna App at <https://rsdoiel.github.io/antennaApp/INSTALL.html>.

## Build a simple website

On your computer open the Terminal application. The basic steps you'll be taking are the following.

1. Create a project directory, "simple", and change into it
2. Use the Antenna App to setup your site using the "init" command
3. Create a home page in Markdown called "index.md"
4. Generate the HTML page, "index.html" using Antenna
5. Preview your website in your web browser

Here is what you type in the terminal for steps one through two.

~~~shell

mkdir simple
cd simple
antenna
~~~

The terminal will clear and you'll be prompted by a question

~~~
	antenna.yaml does not exist. Create it? yes/no 
~~~

Type "yes" and press the enter key. This will create the project files used to manage your website. The project files include antenna.yaml, pages.md and pages.db. I'll talk more about these as we go.  Assuming you answer "yes" and pressed the enter key you should see a screen that looks like this.

~~~
Curate Collections

	 1: pages.md, An Antenna Website

(1/1, [h]elp or [q]uit): 
~~~

A simple website or blog may only have one collection. You can press "1" and the enter key so we can manage our collection. This presents the definition for the collection defined by the Markdown document pages.md.  From here you can curate pages, posts and items. These are the three basic types Antenna App supports.

pages
: These are Markdown documents that will be turned into web pages. When your website is generate they'll be included the sitemap but will NOT be included in any feeds associated with the site. 

posts
: Posts are Markdown documents that are aggregated into the website feed(s). These include things like blog posts. They are organized temporially and tend to accumulate over time. Unlike pages, these are included in both the sitemap and feeds. Normally you'll put these in a directory structured around the date, example "posts/2026/01/09", could be used as the path for posts made on January 9th, 2026.

items
: Are similar to posts except they come from other websites. These are used to present aggregated posts used in personal news sites. The individual items will not appear in the sitemap but will appear in the aggregated feed for the collection. You can aggreate many items from many feeds. Feeds are defined using Markdown in the collection's Markdown definition file. You define them by included a list of links.

In our first effort we're going to work with pages. If you've entered the pages.md collection you should a menu like this now.

~~~
Curate pages.md:

	[pa]ges
	[po]osts
	[i]tems


([h]elp or [q]uit): 
~~~

This is the list of things to curate. The "[" and "]" indicate the menu shortcut to type in before pressing the enter key. Since we want to curate some pages you should type "pa" and then press the enter key. That should leave you with a menu like this.

~~~
Pages in pages.md


(1/0, [h]elp or [q]uit): 
~~~

At this point we don't have any pages to curate so the list is empty. Type "h" and press enter. This shows you help for curating pages. You can add, delete and list pages.

add
: This adds an existing Markdown document to the collection

del
: This removes the Markdown content from a collection but does not remove the Markdown document from the disk.

When we generate our website the "add" operation generates and updated HTML file. Also when we later generate our whole website each page in the collection as well as each post and the aggregation pages for items will get automatically generated in their HTML versions (along with other useful files).

To back out of this menu type "q" and press enter. This will take you back to our curate pages.md menu, repeat typing "q" and press enter. This backs us out to the first menu listing the collections associated with this website. Type "q" and press enter one my time and we should be back at our terminal.

Now it's time to create two Markdown documents, "index.md" and "about.md".

index.md
: This will be our welcome page. This is the default page shown for the website.

about.md
: This is a short descriptive page about our simple website.

We are goint to create these two files using a text editor. There don't have to be long. Here's example content for each Markdown document.

index.md:

~~~markdown

# Welcome

This is a simple webpage for a simple website. See [about](about.html) for more information.

~~~

about.md

~~~markdown

# About a simple website

This is just a demo of a two page website generated with [Antenna App](https://rsdoiel.github.io/antennaApp)

~~~

Now we have two Markdown documents. We need to add them to our pages.md collection. Fire up antenna and enter the following text.

~~~
antenna
1
pa
a index.md

a about.md

~~~

NOTE: the each newline is input by pressing the enter key. When you add index.md it'll give you a chance to also set the
the output name to associated with index.md. By default that would be "index.html". You can just press enter without typing anything else to accept the default.

You should now see a menu like this.

~~~
Pages in pages.md

   1 about.md
	"about.html" 2026-01-09T17:12:56-08:00
   2 index.md
	"index.html" 2026-01-09T17:12:36-08:00

(1/2, [h]elp or [q]uit): 
~~~

The timestamp will of course be different but you should see the Markdown document added to the collection as a page along with the name used when rendering to HTML.

We can now quit our way back to the terminal prompt.

~~~
q
q
q
~~~

You can preview the website using `antenna preview`, type that into the terminal.

~~~shell
antenna preview
~~~

You should see output like this.

~~~shell
$ antenna preview
2026/01/09 17:16:30 Document root 
2026/01/09 17:16:30 Listening for http://localhost:8000
~~~

You can now start your web browser and input the URL "http://localhost:8000". Do you see our Welcome page?  Is there a link to the about page? Try clicking the links to confirm all is working. If you look back in the terminal window you should see something like this.

~~~shell
2026/01/09 17:19:40 Document root 
2026/01/09 17:19:40 Listening for http://localhost:8000
2026/01/09 17:19:42 request Method: GET Path: / RemoteAddr: 127.0.0.1:57295 UserAgent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:146.0) Gecko/20100101 Firefox/146.0
2026/01/09 17:19:44 request Method: GET Path: /about.html RemoteAddr: 127.0.0.1:57295 UserAgent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:146.0) Gecko/20100101 Firefox/146.0
~~~

Each line is what we call a log line. It shows for each page you click on. This helps us identify when we have problems and that the preview is working.  To exit the preview hold the control key down and type the letter "c". This is referred to as "ctrl+c" in computer jargon. At this point you should be back at the terminal prompt. Take a look at the files in the directory using the "ls" command.  You should see something like this.

~~~shell
$ ls
about.html   about.md     antenna.yaml index.html   index.md     page.yaml    pages.db     pages.md
~~~

You've create your first website! Let me explain each of the files.

index.md and about.md
: These are Markdown files that Antenna App will turn into HTML pages.

about.html and index.html
: These are the files you web browser understands and uses as hypertext.

antenna.yaml
: A configuration file used by and maintained by Antenna App.

page.yaml
: A configuration file used Antenna App describing how to layout a web page (more on that later)

pages.md
: A Markdown document used to configure the collection. The any front matter (more on that later) provided will be used when generate feeds. Also if you are aggregated feeds from other websites they needed to be included in a Markdown list of links in the body of the Markdown document.

pages.db
: This is an SQLite3 database used by Antenna App for managing your collection of items, pages and posts.

## Improving our simple site

The site right now is very plain. There is only the text we typed into the web pages. Usually you want to include a simple means of navigating between web pages without requiring explicit linking in our Markdown documents. You can do that be configuring site navigation in a "theme" for your website.  Let's create that in a momement. A theme is a folder of Markdown documents named for each part of the webpage. A theme can be associated with an Antenna App collection.  Antenna App is unusual in that you can define the elements of the page using Markdown. The page.yaml file describes how to use each Markdown document to render a part of the web page. Supported elements in a webpage are head, header, navigation, top content, bottom content and footer. These are defined using the following file names in the theme folder.

head.yaml
: This is a YAML document that defines the head element in the HTML page. Don't worry about it yet. Antenna App will make that easier to edit than having to learn YAML.

header.md
: If present this will cause a header element to be created at the start of the body element in the HTML page. Usually you would have things like a logo, sitename or other branding content in the header.md file.

nav.md
; If present this will create a nav element towards the top of the body element in the HTML page. Usually you use a Markdown list of links to define navigation around you website.

top_content.md
: This is Markdown to be included before your Markdown document.

bottom_content.md
: This is Markdown to be included after your Markdown document.

footer.md
: If presnet this will cause the footer element to be created just before the closing of the body lement in the HTML page. Usually sites include things like copyright information, or less used navigation details, links to feeds, link to an about page, contact information here.

Let's create Markdown files of the theme above.

>>>>>> FIXME: show a basic step through of theming and how to apply them here

When you are previewing your website you see how the theme is applied to the webiste.

Started your web browser and open the URL <http://localhost:8000>.  You should see your Hello World homepage. If you open <view-source:http://localhost:8000> you'll see the HTML source generated by the "page" command. When you are finished go back to your Terminal window and hold down the control key and press the "C" key (Ctrl-C). This exits the preview command. If you reload the pages in your web browser you'll get your browser's default page indicating that the pages are not available.

### Expanding our site

Our site isn't useful website yet. There are no links. Modify "index.md" file to look like this.

~~~markdown

# Hello World!

Learn about [Markdown](https://en.wikipedia.org/wiki/Markdown) on
John Grubber's website <https://daringfireball.net/projects/markdown/>.

~~~

Once you've saved this revision use the page command to update the HTML. After that run the preview command like before. Switch back to your terminal and type the following.

~~~shell

antenna page index.md
antenna preview
~~~

Refresh (reload) your web browser page for the URL <http://localhost:8000>.  Our "Hello World" homepage, "index.md", now should be transformed with "Hello World" displaying as a heading and our added paragraph with links to Wikipedia and Daring Fireball websites. Refresh the "view-source" version of the pages and see the additional HTML markup generated by the page command.

### Elaboration

A single page website is very limited.  You can create more pages going through the process we used for "index.md". Here's the basic sequence.

1. create or edit the markdown document in your editor, save it
2. use the page command to render the HTML version
3. use the preview command and your web browser to see your handy work

Let's add a page called "fruit.md" with a list of heading and list fruits.

~~~shell

code fruit.md
~~~

Here's the text of the fruits.md page.

~~~Markdown

# Fruit

- Dragon Fruit
- Lime
- Mango

~~~

You turn the page into HTML using the page command and then preview the website again

~~~shell

antenna page fruit.md
antenna preview
~~~

You have two pages in your website now. The URL are

- http://localhost:8000/index.html
- http://localhost:8000/fruit.html

Look at each in your web browser. Notice the pages aren't linked in anyway. That's not ideal. What we need is an easy way to navigation the main pages of our site. Antenna App supports a concept called themes. This let's you define how a page is generated through assigning content to specific parts of the page. One of the parts of an Antenna generated page is called "nav" and it is expressed as Markdown. Other parts include header, footer, top_content and bottom_content (the last two are before and after the section holding the contents our   original Markdown documents). 

### theming

To create a theme we need to first create a directory. Then we need to include the desired parts in the theme. Let's create a theme that has the nav element for our website that will make it easy to switch from the homepage to the fruit page. The steps we'll take to create our theme follows.

1. create a directory called "theme"
2. create a file in the "theme" directory called "nav.md"
3. Add two links using Markdown to "nav.md" and save the file
4. "apply" our theme
5. regenerate our HTML pages and preview the site

Here's what I'd type in the terminal for steps one and two.

~~~shell

mkdir theme
code theme/nav.md
~~~

The content to type into "nav.md" is as follows.

~~~Markdown

[home](index.html) [fruit](fruit.html)
~~~

Save the result then in the terminal window type the following to finish steps four and five.

~~~shell

antenna apply theme
antenna page index.md
antenna page fruit.md
antenna preview
~~~

You can now refresh the browser pages and exploring our site by clicking on the links provided by the navigation.

## Antenna Themes

An Antenna App theme is defined as a directory with Markdown files to express the visible common HTML elements. Other non-visible elements in the head element of an HTML page are also able to be set with in a theme. Those aren't expressed in Markdown. Here's the names of files which are recognized by the Antenna App as valid parts of a theme.

- header.md
- nav.md
- top_content.md
- bottom_content.md
- footer.md
- head.yaml
- style.css

The files ending with ".md" are Markdown files. Markdown describes the content we want to include in those sections of our webpage. Any Markdown content can be used. If one of the files does not exist in the theme then that element is not populated in the resulting page. Example if you include nav.md then navigation will be included otherwise it will not be included. The page order of the Markdown elements are 

1. header.md
2. nav.md
3. top_content.md
4. The content of the Markdown document used with the page command
5. bottom_content.md
6. footer.md

When we create a file like "index.md" and "fruits.md" the resulting HTML is inserted between the "top_content" and "bottom_content" sections of the webpage. 

To keep things simple I started by showing you how to create the "nav" element using the "nav.md" site your theme folder.
The theme name can be anything, it's just a directory. Here's the basic steps in creating a new theme

1. create a directory to hold the theme parts
2. Inside your directory create Markdown documents for the parts of the them you want defined
3. Apply the theme by using the directory name and the apply command 
4. Generate or regenerate HTML for each page in your site
5. Preview the site

Editing the theme is just a matter of adding, editing or removing elements from the theme folder. The two elements that are not Markdown documents are "style.css" and "head.yaml" if you choose to include them. The "style.css" file should contain valid CSS.
The "head.yaml" gives you find grained control of the meta, link and script elements that get rendered into the head element of your web page.

### Spicing pages up with CSS

So far the website is functional but relies on the defaults provided by your web browser. The visual appearance can be enhanced through using CSS. CSS is a language that describes to the web browser how it should layout the HTML page contents. The Antenna App's themes supports embedding a style element in the HTML page's head element by including a "style.css" file inside the theme directory. Let's do that now.

Use your terminal to create "theme/style.css"

~~~shell

code theme/style.css
~~~

Paste in the follow CSS. It'll turn your top level page titles vertical using CSS.

~~~css

/* Turn the H1 elements vertical */
h1 {
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  text-orientation: mixed;
}
~~~

After creating and save the "style.css" you theme directory should have two files.

~~~

theme/nav.md
theme/style.css
~~~

You can now apply the theme, regenerate the HTML pages and preview them.

~~~shell

antenna apply theme
antenna page index.md
antenna page fruits.md
antenna preview
~~~

When you preview the site you should see the effect the CSS had on how the page displayed. The changes to layout using CSS can be very elaborate and CSS is a topic I encourage you to explore on your own. A good reference and tutorial website for CSS is <https://developer.mozilla.org/en-US/docs/Web/CSS>. A historic example of what people have done with CSS can be seen at [CSS Zen Garden](https://csszengarden.com/). A web search on [DuckDuckGo](https://duckduckgo.com?q=learning CSS) or other search engine will turn up lots of additional resources. There are also book available through [Open Library](https://openlibrary.org/search?q=CSS&mode=everything).

## Next steps

- Create a few more pages for your website using Markdown for the text content
- Explore adding additional theme elements like "header" and "footer"

Happy site building!

