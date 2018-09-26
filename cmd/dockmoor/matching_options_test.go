package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyPredicates(t *testing.T) {
	fo := &MatchingOptions{}
	err := verifyMatchOptions(fo)
	assert.Nil(t, err)
}

var domainPredicateNames = []string{"domains"}
var namePredicateNames = []string{"names"}
var tagPredicateNames = []string{"latest", "outdated", "untagged", "tags"}
var digestPredicateNames = []string{"digests", "unpinned"}

var predicateGroups = [][]string{
	domainPredicateNames,
	namePredicateNames,
	tagPredicateNames,
	digestPredicateNames,
}

var predicateNames = append(
	append(
		append(
			domainPredicateNames,
			namePredicateNames...),
		tagPredicateNames...),
	digestPredicateNames...)

func applyPredicatesByName(fo *MatchingOptions, names ...string) {

	for _, name := range names {
		switch {
		case equalsAnyString("outdated", name):
			fo.Predicates.Outdated = true
		case equalsAnyString("unpinned", name):
			fo.Predicates.Unpinned = true
		case equalsAnyString("latest", name):
			fo.Predicates.Latest = true
		case equalsAnyString("domains", name):
			fo.Predicates.Domains = []string{"a", "b"}
		case equalsAnyString("names", name):
			fo.Predicates.Names = []string{"a", "b"}
		case equalsAnyString("untagged", name):
			fo.Predicates.Untagged = true
		case equalsAnyString("tags", name):
			fo.Predicates.Tags = []string{"a", "b"}
		case equalsAnyString("unpinned", name):
			fo.Predicates.Unpinned = true
		case equalsAnyString("digests", name):
			fo.Predicates.Digests = []string{"a", "b"}
		default:
			panic(fmt.Sprintf("Unknown predicate name '%s'", names))
		}
	}
}

func TestSinglePredicatesIsValid(t *testing.T) {
	for _, a := range predicateNames {
		t.Run(a, func(t *testing.T) {

			fo := &MatchingOptions{}
			applyPredicatesByName(fo, a)

			err := verifyMatchOptions(fo)
			assert.Nil(t, err)
		})
	}
}

func TestNonGlobalPredicatesCanBeCombinedWithOther(t *testing.T) {

	for iGroupA, groupA := range predicateGroups {
		if iGroupA == 0 { // global
			continue
		}

		for iGroupB, groupB := range predicateGroups {
			if iGroupB == 0 { // global
				continue
			}
			if iGroupA == iGroupB { // global
				continue
			}
			for _, a := range groupA {
				for _, b := range groupB {

					t.Run(fmt.Sprintf("%s and %s", a, b), func(t *testing.T) {
						fo := &MatchingOptions{}
						applyPredicatesByName(fo, a, b)

						err := verifyMatchOptions(fo)
						assert.Nil(t, err)
					})
				}
			}
		}

	}
}

func TestMultipleFromSameGroupFail(t *testing.T) {

	for _, group := range predicateGroups {
		if len(group) <= 1 {
			continue
		}

		for _, a := range group {
			for _, b := range group {
				if a == b {
					continue
				}

				t.Run(a+" and "+b, func(t *testing.T) {
					fo := &MatchingOptions{}
					applyPredicatesByName(fo, a, b)

					err := verifyMatchOptions(fo)
					assert.Error(t, err)
					assert.Equal(t, ErrAtMostOnePredicate, err)
				})
			}
		}
	}
}

func TestAllExclusivePredicatesAtOnceFail(t *testing.T) {
	fo := &MatchingOptions{}
	fo.Predicates.Outdated = true
	fo.Predicates.Unpinned = true
	fo.Predicates.Latest = true
	err := verifyMatchOptions(fo)
	assert.Equal(t, ErrAtMostOnePredicate, err)
}

func TestNonExclusivePredicatesCanBeCombined(t *testing.T) {
	fo := &MatchingOptions{}

	for _, domain := range [][]string{nil, {"a", "b"}} {
		fo.Predicates.Domains = domain
		for _, name := range [][]string{nil, {"a", "b"}} {
			fo.Predicates.Names = name
			for _, tag := range [][]string{nil, {"a", "b"}} {
				fo.Predicates.Tags = tag
				for _, digest := range []string{"unpinned"} {
					unpinned := digest == "unpinned"
					fo.Predicates.Unpinned = unpinned

					testCase := fmt.Sprintf("predicates can be combined: %s/%s:%s@%t", domain, name, tag, unpinned)
					t.Run(testCase, func(t *testing.T) {
						err := verifyMatchOptions(fo)
						assert.Nil(t, err)
					})
				}
			}
		}
	}
}

func TestTagPredicateCombinations(t *testing.T) {
	fo := &MatchingOptions{}

	for _, untagged := range []bool{true, false} {
		for _, latest := range []bool{true, false} {
			for _, outdated := range []bool{true, false} {
				for _, tags := range [][]string{nil, {"a", "b"}} {
					set := 0
					if untagged {
						set++
					}
					if latest {
						set++
					}
					if outdated {
						set++
					}
					if tags != nil {
						set++
					}

					fo.Predicates.Untagged = untagged
					fo.Predicates.Latest = latest
					fo.Predicates.Outdated = outdated
					fo.Predicates.Tags = tags

					testCase := fmt.Sprintf("domain tag combination: untagged:%t-latest:%t-outdated%t-tags:%s", untagged, latest, outdated, tags)
					t.Run(testCase, func(t *testing.T) {
						err := verifyMatchOptions(fo)
						switch {
						case set > 1:
							assert.Error(t, err)
						case set <= 1:
							assert.Nil(t, err)
						}
					})
				}
			}
		}
	}
}

func TestDigestPredicateCombinations(t *testing.T) {
	fo := &MatchingOptions{}

	for _, unpinned := range []bool{true, false} {
		for _, digests := range [][]string{nil, {"a", "b"}} {
			set := 0
			if unpinned {
				set++
			}
			if digests != nil {
				set++
			}

			fo.Predicates.Unpinned = unpinned
			fo.Predicates.Digests = digests

			testCase := fmt.Sprintf("domain tag combination: unpinned:%t-digests:%s", unpinned, digests)
			t.Run(testCase, func(t *testing.T) {
				err := verifyMatchOptions(fo)
				switch {
				case set > 1:
					assert.Error(t, err)
				case set == 1:
					assert.Nil(t, err)
				}
			})
		}
	}
}
