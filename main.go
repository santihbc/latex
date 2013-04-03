/*
  Copyright (c) 2013 José Carlos Nieto, https://menteslibres.org/xiam

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

package glatex

import (
	"crypto"
	"errors"
	"fmt"
	"io"
	"menteslibres.net/gosexy/checksum"
	"menteslibres.net/gosexy/to"
	"os"
	"os/exec"
	"strings"
)

const (
	// Must be write friendly.
	WorkingDirectory = "tmp"
	// Must have lots of space.
	OutputDirectory = "images"
	// Path separator.
	PS = string(os.PathSeparator)
)

/*
	Renders a LaTeX string, returns a PNG image path.
*/
func Render(tex string, density int) (string, error) {
	var err error

	name := checksum.String(tex, crypto.SHA1)

	//filename := OutputDirectory + PS + name
	workdir := WorkingDirectory + PS + name

	err = os.MkdirAll(workdir, 0755)

	if err != nil {
		return "", err
	}

	snippet := strings.Trim(tex, "\r\n\t")

	// Default formatting
	if strings.HasPrefix(snippet, `\begin{document}`) == false {
		snippet = fmt.Sprintf(`
			\begin{document}
			%s
			\end{document}`,
			snippet,
		)
	}

	if strings.Contains(snippet, `\documentclass`) == false {
		snippet = fmt.Sprintf(`
			\documentclass[draft]{article}
			\usepackage[dvips]{color}
			\usepackage[dvips]{graphicx}
			\usepackage{amsmath}
			\usepackage{amsfonts}
			\usepackage{amssymb}
			\pagestyle{empty}
			\pagecolor{white}
			%s`,
			snippet,
		)
	}

	// Writing LaTeX to a file.
	texFile, err := os.Create(workdir + PS + "output.tex")

	if err != nil {
		return "", err
	}

	_, err = io.WriteString(texFile, snippet)

	if err != nil {
		return "", err
	}

	texFile.Close()

	dviPath := workdir + PS + "output.dvi"
	epsPath := workdir + PS + "output.eps"
	pngPath := workdir + PS + "output.png"

	// LaTeX toolchain
	batch := []*exec.Cmd{
		exec.Command("latex",
			"-no-shell-escape",
			"-interaction=batchmode",
			fmt.Sprintf("-output-directory=%s", workdir),
			texFile.Name(),
		),
		exec.Command(
			"dvips",
			"-E",
			dviPath,
			"-o",
			epsPath,
		),
		exec.Command(
			"convert",
			"+adjoin",
			"-density",
			to.String(density),
			"-antialias",
			epsPath,
			pngPath,
		),
	}

	for _, cmd := range batch {
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	// Had success?
	stat, err := os.Stat(pngPath)

	if err != nil || stat == nil {
		return "", errors.New("Failed to create PNG file.")
	}

	return pngPath, err
}
