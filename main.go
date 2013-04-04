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

package latex

import (
	"bytes"
	"crypto"
	"errors"
	"fmt"
	"io"
	"menteslibres.net/gosexy/checksum"
	"menteslibres.net/gosexy/to"
	"os"
	"os/exec"
	"path"
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
	Rendered type stores settings for the rendering toolchain.
*/
type Renderer struct {
	Density  int
	UseCache bool
}

/*
	Returns a new *Renderer with default options.
*/
func New() *Renderer {
	self := &Renderer{}
	self.Density = 144
	self.UseCache = true
	return self
}

/*
	Renders a LaTeX string, returns a PNG image path.
*/
func (self *Renderer) Render(latex string) (string, error) {
	var err error

	name := checksum.String(latex, crypto.SHA1)

	// Relative output directory.
	relPath := name[0:4] + PS + name[4:8] + PS + name[8:12] + PS + name[12:16] + PS + name[16:] + ".png"

	// Setting output directory.
	pngPath := OutputDirectory + PS + relPath

	err = os.MkdirAll(path.Dir(pngPath), 0755)

	if err != nil {
		return "", err
	}

	// Does the output file already exists?
	if self.UseCache == true {
		_, err = os.Stat(pngPath)

		if err == nil {
			return relPath, nil
		}
	}

	// Setting working directory.
	workdir := WorkingDirectory + PS + name

	err = os.MkdirAll(workdir, 0755)

	if err != nil {
		return "", err
	}

	// Will clean the directory at the end.
	defer os.RemoveAll(workdir)

	// Writing LaTeX to a file.
	texFile, err := os.Create(workdir + PS + "output.tex")

	if err != nil {
		return "", err
	}

	defer texFile.Close()

	_, err = io.WriteString(texFile, latex)

	if err != nil {
		return "", err
	}

	// Temp files.
	dviPath := workdir + PS + "output.dvi"
	epsPath := workdir + PS + "output.eps"

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
			to.String(self.Density),
			"-antialias",
			epsPath,
			pngPath,
		),
	}

	// Executing toolchain.
	for _, cmd := range batch {
		err = cmd.Run()

		if err != nil {
			// Trying to catch error
			logPath := workdir + PS + "output.log"

			logFile, err := os.Open(logPath)

			if err == nil {
				buf := bytes.NewBuffer(nil)
				buf.ReadFrom(logFile)

				logFile.Close()

				if buf.Len() > 0 {
					return "", errors.New(string(buf.Bytes()))
				}
			}

			return "", err
		}
	}

	// Had success?
	stat, err := os.Stat(pngPath)

	if err != nil || stat == nil {
		return "", errors.New("Failed to create PNG file.")

	}

	return relPath, err
}
