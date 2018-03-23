# Duplo - Detect Similar or Duplicate Images

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/rivo/duplo)
[![Go Report](https://img.shields.io/badge/go%20report-A%2B-brightgreen.svg)](https://goreportcard.com/report/github.com/rivo/duplo)

This Go library allows you to perform a visual query on a set of images, returning the results in the order of similarity. This allows you to effectively detect duplicates with minor modifications (e.g. some colour correction or watermarks).

It is an implementation of [Fast Multiresolution Image Querying](http://grail.cs.washington.edu/projects/query/mrquery.pdf) by Jacobs et al. which uses truncated Haar wavelet transforms to create visual hashes of the images. The same method has previously been used in the [imgSeek](http://www.imgseek.net) software and the [retrievr](http://labs.systemone.at/retrievr) website.

## Installation

```
go get github.com/rivo/duplo
```

## Usage

```go
import "github.com/rivo/duplo"

// Create an empty store.
store := duplo.New()

// Add image "img" to the store.
hash, _ := duplo.CreateHash(img)
store.Add("myimage", hash)

// Query the store based on image "query".
hash, _ = duplo.CreateHash(query)
matches := store.Query(hash)
sort.Sort(matches)
// matches[0] is the best match.
```

## Documentation

http://godoc.org/github.com/rivo/duplo

## Possible Applications

* Identify copyright violations
* Save disk space by detecting and removing duplicate images
* Search for images by similarity

## Projects Using This Package

* [imgdup2go](https://github.com/rif/imgdup2go): A visual image duplicate finder.

## More Information

For more information, please go to http://rentafounder.com/find-similar-images-with-duplo/ or get in touch.
