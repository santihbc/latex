# Go (golang) LaTeX server

This is a server that compiles a LaTeX formula and produces a PNG image using
the standard LaTeX toolchain and ImageMagick's `convert`.

A public version is available at [https://menteslibres.net/api/latex][1]

## Installing

### Requisites

Make sure you have the standard LaTeX toolchain and ImageMagick's `convert`
installed in your system, the server expects to use these commands: `latex`,
`dvips`, `convert`.

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

[1]: https://menteslibres.net/api/latex
