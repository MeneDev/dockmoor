package dockref

import (
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

func FromOriginal(original string) (ref Reference, e error) {
	r, e := reference.ParseAnyReference(original);
	if e != nil {
		return
	}

	var name string
	if named, ok := r.(reference.Named); ok {
		name = named.Name()
	}

	var tag string
	if tagged, ok := r.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	var dig string
	if digested, ok := r.(reference.Digested); ok {
		dig = string(digested.Digest())
	}

	ref = dockref{
		original: original,
		name:     name,
		tag:      tag,
		digest:   dig,
	}
	return
}

type Reference interface {
	Name() string
	Tag() string
	DigestString() string
	Digest() digest.Digest
	Original() string
}

var _ Reference = (*dockref)(nil)

type dockref struct {
	name     string
	original string
	tag      string
	digest   string
}

func (r dockref) Name() string {
	return r.name
}

func (r dockref) Tag() string {
	return r.tag
}

func (r dockref) DigestString() string {
	return r.digest
}

func (r dockref) Digest() digest.Digest {
	return digest.Digest(r.digest)
}

func (r dockref) Original() string {
	return r.original
}
