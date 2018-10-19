package dockref

import (
	_ "crypto/sha256" // side effect: register sha256
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

func deliberatelyUnsued(err error) {
	// noop
}

// FromOriginalNoError same functionallity as FromOriginal, but hides errors.
// Use this function only when you know that the input cannot have an error
func FromOriginalNoError(original string) Reference {
	ref, e := FromOriginal(original)
	deliberatelyUnsued(e)

	return ref
}

func FromOriginal(original string) (ref Reference, e error) {
	r, e := reference.ParseAnyReference(original)
	if e != nil {
		return
	}

	var name string
	var domain string
	var path string
	var named reference.Named
	var ok bool
	if named, ok = r.(reference.Named); ok {
		name = named.Name()
		domain = reference.Domain(named)
		path = reference.Path(named)
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
		domain:   domain,
		name:     name,
		tag:      tag,
		digest:   dig,
		path:     path,
		named:    named,
	}
	return
}

type Reference interface {
	Name() string
	Tag() string
	DigestString() string
	Digest() digest.Digest
	Original() string
	Domain() string
	Path() string
	Named() reference.Named
	Format() Format
	Formatted(format Format) string
}

type Format int

const (
	FormatHasName   Format = 1 << iota
	FormatHasTag    Format = 1 << iota
	FormatHasDomain Format = 1 << iota
	FormatHasDigest Format = 1 << iota
)

func (format Format) hasName() bool {
	return format&FormatHasName != 0
}
func (format Format) hasTag() bool {
	return format&FormatHasTag != 0
}
func (format Format) hasDomain() bool {
	return format&FormatHasDomain != 0
}
func (format Format) hasDigest() bool {
	return format&FormatHasDigest != 0
}

var _ Reference = (*dockref)(nil)

type dockref struct {
	name     string
	original string
	tag      string
	digest   string
	domain   string
	path     string
	named    reference.Named
}

func (r dockref) Format() Format {
	var format Format

	if r.named != nil {
		fn := reference.FamiliarString(r.named)
		if fn != r.original {
			format |= FormatHasDomain
		}
	}

	if r.name != "" {
		format |= FormatHasName
	}
	if r.tag != "" {
		format |= FormatHasTag
	}
	if r.DigestString() != "" {
		format |= FormatHasDigest
	}

	return format
}

func (r dockref) Formatted(format Format) string {

	s := ""

	var name string

	if format.hasName() {
		if format.hasDomain() {
			name = r.name
		} else {
			name = reference.FamiliarName(r.named)
		}
		s += name
	}

	if format.hasTag() {
		s += ":" + r.tag
	}

	if format.hasDigest() {
		if format.hasName() {
			s += "@" + r.DigestString()
		} else {
			s += r.Digest().Hex()
		}
	}

	return s
}

func (r dockref) Named() reference.Named {
	return r.named
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

func (r dockref) Domain() string {
	return r.domain
}

func (r dockref) Path() string {
	return r.path
}
