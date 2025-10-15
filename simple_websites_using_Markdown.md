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

The Web is a __networked hypertext system__. A human friendly way of expressing hypertext is [Markdown](https://en.wikipedia.org/wiki/Markdown). Markdown has allowed many people to participate in creating content for the web. I think it deserves more focus. Between Markdown and a sprinkle of CSS and you can create entire websites using your own computer or hosted on a static site service[^1]. In this post I show how Markdown can be used for basic websites[^2]. All it requires is knowing a little Markdown, a text editor, web browser and the [Antenna App](https://rsdoiel.github.io/antennaApp).

[^1]: Examples: [GitHub Pages](https://pages.github.com), S3 buckets setup as static websites at AWS, Dreamhost, my favorite [Mythic Beasts](https://www.mythic-beasts.com/)

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

Chances are you have the first four if you're reading this. What I describe in this ports should work on computers running Windows, macOS and Linux based systems. If you need to acquire a computer and already have a HDMI television I recommend getting a Raspberry Pi Computer. The Raspberry Pi 500 Kit is a nice middle range device that runs about $120.00 US. It includes a book about using Raspberry Pi computers. If you can afford a bit more the Raspberry Pi 500+ Kit should be out by Christmas 2025 and that'll probably run in the ballpark of $240.00 US. That's a nice computer with lots of storage and the bells and whistles to do image, audio and light video work. If you are on a budget then you can squeak by with a Raspberry Pi Zero 2W running the software. That setup will set you back about $50.00 US with the minimum accessories needed like case, keyboard, mouse, power supply and cables to connect things up. The Official Raspberry Pi Handbook print edition runs between $25.00 US but the electronic edition that you can read on your phone or tablet is free. My point is that for a little bit of investment or using your existing resources you have the means of building websites for yourself and others. You don't need to be a programmer or even a web designer if you are comfortable editing text and doing a bit of writing. 

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
antenna init
~~~

If you look at the contents of that directory.

~~~shell

ls
~~~

You should see the following files in the simple directory.

~~~shell

antenna.yaml
page.yaml
~~~

These files are used by Antenna App to generate your website. The "antenna.yaml" is the main configuration file. The "page.yaml" file is used by Antenna App to know how to construct an HTML page.

For step three we need to use a text editor. On Raspberry Pi mousepad is the the default editor. On Windows it is notepad and macOS its textedit. Many people use [VS Code](https://code.visualstudio.com/download) editor which is available on macOS, Windows and Raspberry Pi OS. That is what I'll be using in my examples. You can change the "code" references to the editor you prefer.

On my Raspberry Pi I can run the editor creating "index.md" when I save it. Here is the command I type into my terminal to do that.

~~~shell

code index.md
~~~

In the editor window you should see an empty file. Type in the following and save the file as "index.md" in the simple directory.

~~~markdown

Hello World! 

~~~

Now we want to use Antenna App "page" command to turn that "index.md" into our home web page. This is the command you type into your Terminal window.

~~~shell

antenna page index.md
~~~

If you check the contents of the simple directory.

~~~shell

ls
~~~

You should see three four files now.

~~~

antenna.yaml
index.md
index.html
page.yaml
~~~

We can new preview the website by run the Antenna App's "preview' command. This command runs a little web server on your computer so you can view the results of your handy work in your web browser.

~~~shell

antenna preview
~~~

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

