# Go (golang) LaTeX server

This is a server that compiles a LaTeX formula and produces a PNG image using
the standard LaTeX toolchain and ImageMagick's `convert`.

A public version is available at [https://menteslibres.net/api/latex][1]

## Installing

### Requisites

Make sure you have the standard LaTeX toolchain and ImageMagick's `convert`
installed in your system, the server expects these commands to be available:
`latex`, `dvips`, `convert`.

### Getting and compiling the source

Use `go get` to pull the package:

```
go get menteslibres.net/api/latex
```

Install the command line tool (the actual server) to `$GOPATH/bin`:

```
go install menteslibres.net/api/latex/cmd/go-latex-server
```

And start the server

```
go-latex-server run
```

Options for `go-latex-server run`

```
-l="0.0.0.0": Bind to this IP.
-p=9193: Listen on this port.
-r="/api/latex/images": Prefix for generated PNG files (i.e: a Content Delivery Network)
-s="": Socket path.
-t="standalone": Service type ('standalone' or 'fastcgi').
```

### Setup example

This line will start a FastCGI server listening on `127.0.0.1:9096` and will
prepend CDN's URL (`https://ddori5d5np0fr.cloudfront.net/`) to the resulting
image:

```
go-latex-server -t fastcgi -l 127.0.0.1 -p 9096 -r https://ddori5d5np0fr.cloudfront.net/ run &
```

This is a configuration snippet for [nginx](http://nginx.org), this snippet
assumes that we are going to serve the API on `http://example.org/api/latex`
and that we are going to use a CDN to serve images.

```
server {
  # External image server. A CDN could be pointed here.
  server_name direct-images.example.org;
  root /var/www/latex-server/images/;
}

server {
  server_name example.org;

  # FastCGI server
  location ~ ^/api/latex/(.+)$ {
    fastcgi_pass   127.0.0.1:9096;
    include        fastcgi_params;
  }
}
```

## Documentation

Please look at [http://godoc.org/menteslibres.net/api/latex][2] for the
actual renderer docs.

If you want to test it live see [https://menteslibres.net/api/latex][1].

## License

> Copyright (c) 2013 José Carlos Nieto, https://menteslibres.org/xiam
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

[1]: https://menteslibres.net/api/latex
[2]: http://godoc.org/menteslibres.net/api/latex
