# graffio

Graffio is a command-line tool for posting content to a shared blog site. It takes the provided text (plaintext or Markdown) and converts it into HTML, which is then pushed to the server for publication. Each post has a 1729 character limit.

Website: https://graffio.xyz

## Installation
`brew tap shubhxms/tools`
`brew install graffio`
in your preferred terminal after installing [Homebrew](https:/homebrew.sh)

Or checkout the [releases](https://github.com/shubhxms/graffio/releases)

## Usage
Usage: graffio text [options]

Example:

1. Make a post with default settings:
   graffio "This is my blog content. This is a [link](https://graffio.xyz) to the main _site_."
2. Make a post with custom settings:
   graffio "This is my blog content. This is some **bold text**." -fontFamily "monospace" -fontSize 2

Options:
-fontSize int
range: 1-10

-fontFamily string
serif, sans, or monospace

-alignment string
left, right or center

-color string
hex code, name or rgba

-width int
range: 1-10

Note:

- If no formatting options are provided, the program will use default values for the font size (5), font family (serif), alignment (left), color (black), and width (5).
- If a username is not set, the program will default to "anon" as the author name.
- Use the -u or --username commands to change your username.
- You can use the `graffio -d "postID"` to delete a post within 15 minutes of creating it. The postID is given to you after you make a post.
- You can add images using Markdown syntax: [alt text](https://link-to-image.com/image.jpg)
- Use --help to display this guide to Graffio.
