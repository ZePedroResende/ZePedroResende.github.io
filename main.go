package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Post struct {
	Path  string
	Title string
}

func generate_posts_html(name string) (*Post, error) {
	log.Printf("Generating post %v", name)
	input := "posts/" + name + ".md"
	path := "posts/" + name + ".html"
	output := "generated/" + path

	tmp := "/tmp/post"

	cmd := exec.Command("pandoc", "-f", "markdown", input, "-o", tmp)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running pandoc: %v", err)
	}

	f, err := os.Create(output)
	if err != nil {
		// handle error
		log.Printf("Error creating file: %v", err)
		return nil, err
	}

	tmpl := template.Must(template.ParseFiles("template/post.html"))

	content := template.Must(template.ParseFiles(tmp))
	fmt.Println(content.Name())
	_, err = tmpl.AddParseTree(content.Name(), content.Tree)
	if err != nil {
		log.Printf("Error adding template: %v", err)
		return nil, err
	}

	error := tmpl.Execute(f, nil)
	if error != nil {
		log.Printf("Error executing template: %v", error)
		return nil, error
	}

	return &Post{Path: path, Title: name}, nil
}

func get_posts() []string {
	posts, err := os.ReadDir("posts")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	post_names := []string{}
	for _, post := range posts {

		stripped := strings.Split(post.Name(), ".")[0]
		post_names = append(post_names, stripped)
	}

	return post_names
}

func generate_index_html(posts []Post) {
	log.Printf("Generating index.html")
	output := "generated/index.html"
	// Create the file
	f, err := os.Create(output)
	if err != nil {
		// handle error
		log.Printf("Error creating file: %v", err)
		return
	}

	tmpl := template.Must(template.ParseFiles("template/index.html"))
	error := tmpl.Execute(f, posts)
	if error != nil {
		log.Printf("Error executing template: %v", error)
	}
}

func main() {
	posts_list := get_posts()

	posts := []Post{}
	for _, post_path := range posts_list {
		post, err := generate_posts_html(post_path)
		if err != nil {
			log.Printf("Error generating post: %v", err)
			return
		}
		posts = append(posts, *post)
	}

	generate_index_html(posts)
}
