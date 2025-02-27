package shell

import (
	"context"
	"io"
	"os"
	"path/filepath"

	files "github.com/ipfs/go-ipfs-files"
)

type object struct {
	Hash string
}

type AddOpts = func(*RequestBuilder) error

func OnlyHash(enabled bool) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("only-hash", enabled)
		return nil
	}
}

func Pin(enabled bool) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("pin", enabled)
		return nil
	}
}

func Progress(enabled bool) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("progress", enabled)
		return nil
	}
}

func RawLeaves(enabled bool) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("raw-leaves", enabled)
		return nil
	}
}

// Hash allows for selecting the multihash type
func Hash(hash string) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("hash", hash)
		return nil
	}
}

// CidVersion allows for selecting the CID version that ipfs should use.
func CidVersion(version int) AddOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("cid-version", version)
		return nil
	}
}

// Add adds a file to ipfs pinning it with the given options
func (s *Shell) Add(r io.Reader, options ...AddOpts) (string, error) {
	fr := files.NewReaderFile(r)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", fr)})
	fileReader := files.NewMultiFileReader(slf, true)

	var out object
	rb := s.Request("add")
	for _, option := range options {
		option(rb)
	}
	return out.Hash, rb.Body(fileReader).Exec(context.Background(), &out)
}

// AddNoPin adds a file to ipfs without pinning it
// Deprecated: Use Add() with option functions instead
func (s *Shell) AddNoPin(r io.Reader) (string, error) {
	return s.Add(r, Pin(false))
}

// AddWithOpts adds a file to ipfs with some additional options
// Deprecated: Use Add() with option functions instead
func (s *Shell) AddWithOpts(r io.Reader, pin bool, rawLeaves bool) (string, error) {
	return s.Add(r, Pin(pin), RawLeaves(rawLeaves))
}

func (s *Shell) AddLink(target string) (string, error) {
	link := files.NewLinkFile(target, nil)
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("", link)})
	reader := files.NewMultiFileReader(slf, true)

	var out object
	return out.Hash, s.Request("add").Body(reader).Exec(context.Background(), &out)
}

// AddDir adds a directory recursively with all of the files under it
func (s *Shell) AddDir(dir string, options ...AddOpts) (string, error) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	sf, err := files.NewSerialFile(dir, false, stat)
	if err != nil {
		return "", err
	}
	slf := files.NewSliceDirectory([]files.DirEntry{files.FileEntry(filepath.Base(dir), sf)})
	reader := files.NewMultiFileReader(slf, true)

	var out object
	rb := s.Request("add").Option("recursive", true)
	for _, option := range options {
		option(rb)
	}
	return out.Hash, rb.Body(reader).Exec(context.Background(), &out)
}
