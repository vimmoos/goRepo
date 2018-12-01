package main

import (
	//"fmt"
	"errors"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// really easy example of read and write on a file
// struct with a Title and a Body of a slice of byte
// NOTE !!!! the fields in every struct must start with a capital letter!!!!!!

type Page struct {
	Title string
	Body  template.HTML
}

func (p *Page) rmScript() error {
	doc := html.NewTokenizer(strings.NewReader(string(p.Body)))
	trim := false
	nBody := ""
	for {
		tag := doc.Next()
		switch tag {
		case html.ErrorToken:
			if err := doc.Err(); err != io.EOF {
				return err
			}
			p.Body = template.HTML(nBody)
			return nil
		case html.SelfClosingTagToken, html.CommentToken:
			fallthrough
		case html.StartTagToken, html.EndTagToken:
			tagName, _ := doc.TagName()

			if string(tagName) == "script" {
				trim = !trim
			} else {
				nBody += string(doc.Raw())
			}
		case html.TextToken:
			if !trim {
				nBody += string(doc.Raw())
			}
		}
	}
}

// a save method of Page struct  which return an error or its nil value if
// the operation was succeffuly execute
func (p *Page) save() error {
	filename := "./paper/" + p.Title + ".txt"

	policy := bluemonday.UGCPolicy()

	nBody := policy.SanitizeBytes([]byte(p.Body))

	p.Body = template.HTML(nBody)

	return ioutil.WriteFile(filename, nBody, 0600)
}

// this function simply read a file given a name; and output the * to this new Page
// the and error if the operation failed
func loadPage(title string) (*Page, error) {
	filename := "./paper/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

// good example of error handeling
func errorHandel(err error, w http.ResponseWriter) {
	if err != nil {
		// send a specific http response code
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// this global is used to execute just one time all the templates
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "home.html"))

// function to Parse and execute a generic template
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	errorHandel(err, w)
}

// look at exampleHttp if u need to remember what an handler do
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		//if  the page does not exit just redirect to the edit version of it
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	// render the template
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	// try to load the relative page
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	// render the template
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	// creating the relative *Page
	// NOTE we have to cast body cause FormValue return a type string
	p := &Page{Title: title, Body: template.HTML(body)}
	// use the method to save the *p
	err := p.save()
	errorHandel(err, w)
	//redirect to the view relative page
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// the regular expression which will must be verified before doing anything
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// extract the title  from an URL but the URL structure of it must respect the
// regular expression
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		// errors.New used to create customize errors
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

// closure which take a particular handler and gives a title (iff it's a valid URL)
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", http.StatusFound)
}

// home handler really tow example
// TODO handle err in the first line of thi func
func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("home.html")
	t.Execute(w, nil)
}

func main() {
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/home", homeHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))

}
