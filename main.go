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
	Usage: sharedblog text [options] 
	
	Options:
	  -fontSize int
			the font size between 1 and 10 (default 5)
	  -fontFamily string
			the font family - serif or sans or monospace (default "serif")
	  -alignment string
			the alignment of the text - left or right or center (default "left")
	  -color string
			the color of the text in hex (default "black")
	  -width int
			the width of the text (default 5)
	
	Description:
	  sharedblog is a command-line tool for posting blog content to a shared blog server. It takes the provided text and converts it into markdown, HTML, and plain text formats. The converted content is then pushed to the server for publication.
	
	Arguments:
	  text
			The blog content to be posted. It should be provided as a single string enclosed in quotes.
	
	Flags:
	  -h, --help
			Display this help message and exit. Use this flag to learn more about the available options and how to use the program.
	
	Examples:
	  1. Post a blog with default settings:
		  sharedblog "This is my blog content."
		
	  2. Post a blog with custom options:
		  sharedblog -fontSize 8 -fontFamily sans -alignment center -color "#333333" -width 7 "This is my blog content."
	
	Note:
	  - If no options are provided, the program will use default values for the font size, font family, alignment, color, and width.
	  - The maximum allowed length for the blog content is 1729 characters.
	  - If a username is not set, the program will default to "anon" as the author name.
	
	For more information and updates, please visit: <URL>`
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
		fmt.Println("Response Status:", resp.Status)

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
