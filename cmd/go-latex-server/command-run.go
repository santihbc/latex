/*
  Copyright (c) 2013 José Carlos Nieto, https://menteslibres.net/xiam

  Permission is hereby granted, free of charge, to any person obtaining
  a copy of this software and associated documentation files (the
  "Software"), to deal in the Software without restriction, including
  without limitation the rights to use, copy, modify, merge, publish,
  distribute, sublicense, and/or sell copies of the Software, and to
  permit persons to whom the Software is furnished to do so, subject to
  the following conditions:

  The above copyright notice and this permission notice shall be
  included in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
  MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
  LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
  OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
  WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"menteslibres.net/api/latex"
	"menteslibres.net/gosexy/cli"
	"net"
	"net/http"
	"net/http/fcgi"
	"strings"
)

// This command.
type runCommand struct {
}

// Service prefix
const pathPrefix = "/api/latex/"

// Where to store and look for images.
const imagesEndpoint = pathPrefix + latex.OutputDirectory

// LaTeX service endpoint.
const serviceEndpoint = pathPrefix + "png"

// Cofiguration flags.
var bindIp = flag.String("l", "0.0.0.0", "Bind to this IP.")
var bindPort = flag.Int("p", 9193, "Listen on this port.")
var bindSock = flag.String("s", "", "Socket path.")

var serverType = flag.String("t", "standalone", "Service type ('standalone' or 'fastcgi').")
var prefix = flag.String("r", imagesEndpoint, "Prefix for generated PNG files (i.e: a Content Delivery Network)")

// Blacklisted commands.
var blacklist = []string{
	`\def`,
	`\let`,
	`\futurelet`,
	`\newcommand`,
	`\renewcomment`,
	`\else`,
	`\fi`,
	`\write`,
	`\input`,
	`\include`,
	`\chardef`,
	`\catcode`,
	`\makeatletter`,
	`\noexpand`,
	`\toksdef`,
	`\every`,
	`\errhelp`,
	`\errorstopmode`,
	`\scrollmode`,
	`\nonstopmode`,
	`\batchmode`,
	`\read`,
	`\csname`,
	`\newhelp`,
	`\relax`,
	`\afterground`,
	`\afterassignment`,
	`\expandafter`,
	`\noexpand`,
	`\special`,
	`\command`,
	`\loop`,
	`\repeat`,
	`\toks`,
	`\output`,
	`\line`,
	`\mathcode`,
	`\name`,
	`\section`,
	`\mbox`,
	`\DeclareRobustCommand`,
	`\open`,
	`\aftergroup`,
	`\afterassignment`,
}

// Registers the run command.
func init() {
	cli.Register("run", cli.Entry{
		Name:        "run",
		Description: "Starts the server.",
		Command:     &runCommand{},
	})
}

type handler struct {
	r *latex.Renderer
}

// Expects a ?t=\LaTeX parameter and redirects to a rendered PNG.
func (self *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	r.ParseForm()

	t := r.Form.Get("t")

	if t != "" {
		t, err = chunk(t)

		if err == nil {

			file, err := self.r.Render(t)

			if err == nil {
				redir := *prefix + file
				header := w.Header()
				header.Add("Location", redir)

				w.WriteHeader(301)
				w.Write([]byte(redir))
			} else {
				w.WriteHeader(501)
				// We can't show the exact error as it may give a lot of information.
				w.Write([]byte("Syntax error. Did you remember to add math mode ($...$)?"))
			}
		} else {
			w.WriteHeader(403)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(400)
		w.Write([]byte(http.StatusText(400)))
	}

}

/*
	Inspects a LaTeX string and if it looks clean returns a LaTeX document.
*/
func chunk(latex string) (string, error) {
	latex = strings.Trim(latex, "\r\n\t")

	for _, word := range blacklist {
		if strings.Contains(latex, word) == true {
			return "", fmt.Errorf("Sorry, command %s is not available.", word)
		}
	}

	// Default document properties.
	latex = fmt.Sprintf(`
		\documentclass[draft]{article}
		\usepackage[dvips]{color}
		\usepackage[dvips]{graphicx}
		\usepackage{amsmath}
		\usepackage{amsfonts}
		\usepackage{amssymb}
		\pagestyle{empty}
		\pagecolor{white}
		\begin{document}
		%s
		\end{document}
		`,
		latex,
	)

	return latex, nil
}

// Serves images under the image endpoint.
func serveImages(w http.ResponseWriter, r *http.Request) {
	req := r.URL.Path
	if strings.HasPrefix(req, *prefix) == true {
		if strings.HasSuffix(req, ".png") == true {
			file := req[len(pathPrefix):]
			http.ServeFile(w, r, file)
			return
		}
	}
	w.WriteHeader(404)
	w.Write([]byte(http.StatusText(404)))
}

// The run command executes this function.
func (self *runCommand) Execute() error {

	var err error

	h := &handler{
		r: latex.New(),
	}

	http.Handle(serviceEndpoint, h)

	if strings.HasPrefix(*prefix, "/") {
		*prefix = "/" + strings.TrimLeft(*prefix, "/")
		*prefix = strings.TrimRight(*prefix, "/") + "/"
		http.HandleFunc(*prefix, serveImages)
	}

	addr := *bindSock
	domain := "unix"

	if addr == "" {
		domain = "tcp"
		addr = fmt.Sprintf("%s:%d", *bindIp, *bindPort)
	}

	listener, err := net.Listen(domain, addr)

	if err != nil {
		return err
	}

	defer listener.Close()

	switch *serverType {
	case "standalone":
		err = http.Serve(listener, nil)
	case "fastcgi":
		err = fcgi.Serve(listener, nil)
	default:
		return errors.New("Server type can be either \"fastcgi\" or \"standalone\" only.")
	}

	if err != nil {
		return err
	}

	return nil
}
