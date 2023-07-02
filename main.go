package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/k3a/html2text"
	"net/http"
	"os"
	"strings"
	"time"
)

type metadata struct {
	Name string `json:"name"`
}

func main() {

	//USAGE
	usage := `
   Graffio is a command-line tool for posting content to a shared blog site. It takes the provided text (plaintext or Markdown) and converts it into HTML, which is then pushed to the server for publication. Each post has a 1729 character limit. 

   Website: https://graffio.xyz
 
Usage: graffio text [options] 

Example:
  1. Make a post with default settings:
    graffio "This is my blog content. This is a [link](https://graffio.xyz) to the main *site*."
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
   - You can use the graffio -d "postID" to delete a post within 15 minutes of creating it. The postID is given to you after you make a post. 
   - You can add images using Markdown syntax: [alt text](https://link-to-image.com/image.jpg)
   - Use --help to display this guide to Graffio.`
	// ARGS
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	msg := os.Args[1]
	if msg == "-h" || msg == "--help" {
		fmt.Println(usage)
		os.Exit(1)
	} else if msg == "-u" || msg == "--username" {
		setUsername()
		os.Exit(1)
	} else if msg == "-d" || msg == "--delete" {
		key := os.Args[2]
		req, err := http.NewRequest("DELETE", "https://shared-blog-server.sav1tr.repl.co/delete", strings.NewReader(key))
		if err != nil {
			fmt.Println("Error sending request:", err)
			os.Exit(1)
		}
		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			os.Exit(1)
		}

		if resp.StatusCode == http.StatusForbidden {
			fmt.Println("can only delete within 15 minutes of posting.")
		}else if resp.StatusCode == http.StatusOK {
			fmt.Printf("deleted msg id %s", key)
		}
		
		os.Exit(1)
	}
	os.Args = os.Args[1:]

	// FLAGS
	fontSizePtr := flag.Int("fontSize", 5, "the font size between 1 and 10")
	fontPtr := flag.String("fontFamily", "serif", "the font family - serif or sans or mono")
	alignPtr := flag.String("alignment", "left", "the alignment of the text - left orright or center")
	colorPtr := flag.String("color", "black", "the color of the text in hex")
	widthPtr := flag.Int("width", 5, "the width of the text")

	flag.Parse()

	// USERNAME
	file, err := os.OpenFile("blog-meta.json", os.O_RDONLY, 0)

	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create("blog-meta.json")
		if err != nil {
			fmt.Println("could not create file for storing username")
		}
		file.Close()
		setUsername()
	}

	defer file.Close()

	dat, err := os.ReadFile("blog-meta.json")
	if err != nil {
		fmt.Println("did not find a username. defaulting to anon")
	}

	var meta metadata
	json.Unmarshal(dat, &meta)

	// DATA ASSEMBLING
	md := []byte(msg)
	html := string(mdToHtml(md))
	plain := html2text.HTML2Text(html)
	timestamp := time.Now().Unix()
	date := time.Unix(time.Now().Unix(), 0)
	author := meta.Name

	// VALIDATING DATA
	if len(plain) > 1729 {
		fmt.Println("The message is too long. Max length allowed is 1729 characters.")
		os.Exit(1)
	}

	if strings.TrimSpace(plain) == "test" || strings.TrimSpace(plain) == "hello" || strings.TrimSpace(plain) == "trial" {
		fmt.Printf("'%s'? Really? Do better.\n", strings.TrimSpace(plain))
		os.Exit(1)
	}

	if author == "" {
		author = "anon"
	}

	font := [3]string{"sans", "serif", "monospace"}
	fontInFont := false
	for _, v := range font {
		if v == *fontPtr {
			fontInFont = true
			break
		}
	}
	if !fontInFont {
		*fontPtr = "serif"
	}

	align := [3]string{"left", "right", "center"}
	alignInAlign := false
	for _, v := range align {
		if v == *fontPtr {
			alignInAlign = true
			break
		}
	}
	if !alignInAlign {
		*alignPtr = "left"
	}

	// PUSHING TO DATABASE
	item := map[string]interface{}{
		"md":         string(md),
		"html":       html,
		"plain":      plain,
		"timestamp":  timestamp,
		"date":       date,
		"author":     author,
		"fontSize":   *fontSizePtr,
		"color":      *colorPtr,
		"alignment":  *alignPtr,
		"fontFamily": *fontPtr,
		"width":      *widthPtr,
	}
	// fmt.Println(item)
	postBody, err := json.Marshal(item)
	if err != nil {
		fmt.Println(err)
	}

	requestBody := bytes.NewBuffer(postBody)

	responseBody, err := http.Post("https://shared-blog-server.sav1tr.repl.co/post", "application/json", requestBody)
	if err != nil {
		fmt.Println("Something went wrong! :(")
	}
	key := responseBody.Header.Get("key")

	fmt.Printf("posted to the wall :).\nmessage id: %s\n", key)
}

func mdToHtml(md []byte) []byte {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)

}

func setUsername() {
	fmt.Print("enter new username: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("could not read the username")
	}

	name := strings.TrimSuffix(input, "\n")
	meta := metadata{
		Name: name,
	}

	newData, err := json.Marshal(&meta)
	if err != nil {
		fmt.Println("could not convert data to json to update username")
	}

	file, err := os.OpenFile("blog-meta.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("could not open file to update username")
	}

	_, err = file.Write(newData)
	if err != nil {
		fmt.Println("could not update username")
	}

	fmt.Println(name)
}
