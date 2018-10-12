package main

import (
	"bytes"
	"fmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestEmptyPredicates(t *testing.T) {
	fo := &MatchingOptions{}
	err := verifyMatchOptions(fo)
	assert.Nil(t, err)
}

func TestIsSetPredicateByName(t *testing.T) {
	for _, name := range predicateNames {

		t.Run(name, func(t *testing.T) {
			mo := MatchingOptions{}

			before := mo.isSetPredicateByName(name)
			assert.False(t, before)

			applyPredicatesByName(&mo, name)

			after := mo.isSetPredicateByName(name)

			assert.True(t, after)
		})
	}
}

func applyPredicatesByName(fo *MatchingOptions, names ...string) {

	for _, name := range names {
		switch {
		case equalsAnyString(outdatedPred, name):
			fo.TagPredicates.Outdated = true
		case equalsAnyString(unpinnedPred, name):
			fo.DigestPredicates.Unpinned = true
		case equalsAnyString(latestPred, name):
			fo.TagPredicates.Latest = true
		case equalsAnyString(domainPred, name):
			fo.DomainPredicates.Domains = []string{"a", "b"}
		case equalsAnyString(namePred, name):
			fo.NamePredicates.Names = []string{"a", "b"}
		case equalsAnyString(untaggedPred, name):
			fo.TagPredicates.Untagged = true
		case equalsAnyString(tagPred, name):
			fo.TagPredicates.Tags = []string{"a", "b"}
		case equalsAnyString(unpinnedPred, name):
			fo.DigestPredicates.Unpinned = true
		case equalsAnyString(digestsPred, name):
			fo.DigestPredicates.Digests = []string{"a", "b"}
		case equalsAnyString(familiarNamePred, name):
			fo.NamePredicates.FamiliarNames = []string{"a", "b"}
		case equalsAnyString(pathPred, name):
			fo.NamePredicates.Paths = []string{"a", "b"}
		default:
			panic(fmt.Sprintf("Unknown predicate name '%s'", name))
		}
	}
}

func TestAllPredicateNamesAreRecognizedByParser(t *testing.T) {
	options := MatchingOptions{}
	parser := flags.NewParser(&options, flags.PassDoubleDash)

	optsToCheck := append([]string(nil), predicateNames...)

	groups := parser.Groups()
	for _, root := range groups {
		for _, group := range root.Groups() {
			opts := group.Options()
			for _, opt := range opts {
				longName := opt.LongName

				var err error
				optsToCheck, err = without(longName, optsToCheck)
				assert.Nil(t, err)
			}
		}
	}

	assert.Empty(t, optsToCheck)
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

func TestMultipleFromSameGroupFail(t *testing.T) {

	for i1, p1 := range predicateNames {

		for i2, p2 := range predicateNames {
			if i1 <= i2 {
				continue
			}

			idx := indexOf(p2, exclusives[p1])

			if idx >= 0 {
				t.Run(p1+" and "+p2, func(t *testing.T) {
					fo := &MatchingOptions{}
					applyPredicatesByName(fo, p1, p2)

					err := verifyMatchOptions(fo)
					assert.Error(t, err)
					//assert.Equal(t, ErrAtMostOnePredicate[groupName], err)
				})
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
					names := make([]string, 0)

					if untagged {
						set++
						names = append(names, untaggedPred)
					}
					if latest {
						set++
						names = append(names, latestPred)
					}
					if outdated {
						set++
						names = append(names, outdatedPred)
					}
					if tags != nil {
						set++
						names = append(names, tagPred)
					}

					fo.TagPredicates.Untagged = untagged
					fo.TagPredicates.Latest = latest
					fo.TagPredicates.Outdated = outdated
					fo.TagPredicates.Tags = tags

					testCase := fmt.Sprintf("tag combination: %s", strings.Join(names, ", "))
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

			fo.DigestPredicates.Unpinned = unpinned
			fo.DigestPredicates.Digests = digests

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

func TestAnyPredicateWhenNoFlagWithContains(t *testing.T) {
	fo := &MatchingOptions{}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.AnyPredicateNew(), predicate)
}

func TestDomainsPredicateWhenDomainsSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.DomainPredicates.Domains = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.DomainsPredicateNew([]string{"a", "b"}), predicate)
}

func TestNamesPredicateWhenNamesSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.NamePredicates.Names = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.NamesPredicateNew([]string{"a", "b"}), predicate)
}

func TestFamiliarNamePredicateWhenFamiliarNameSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.NamePredicates.FamiliarNames = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.FamiliarNamesPredicateNew(nil), predicate)
}

func TestPathsPredicateWhenPathsSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.NamePredicates.Paths = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.PathsPredicateNew(nil), predicate)
}

func TestTagsPredicateWhenTagsSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.TagPredicates.Tags = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.TagsPredicateNew([]string{"a", "b"}), predicate)
}

func TestUntaggedPredicateWhenUntaggedSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.TagPredicates.Untagged = true

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.UntaggedPredicateNew(), predicate)
}

//func TestOutdatedPredicateWhenOutdatedSet(t *testing.T) {
//	fo := &MatchingOptions{}
//	fo.TagPredicates.Outdated = true
//
//	predicate := fo.getPredicate()
//
//	assert.IsType(t, dockproc.OutdatedPredicateNew(), predicate)
//}

func TestLatestPredicateWhenLatestSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.TagPredicates.Latest = true

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.LatestPredicateNew(), predicate)
}

func TestDigestsPredicateWhenDomainSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.DigestPredicates.Digests = []string{"a", "b"}

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.DigestsPredicateNew([]string{"a", "b"}), predicate)
}

func TestUnpinnedPredicateWhenUnpinnedSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.DigestPredicates.Unpinned = true

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.UnpinnedPredicateNew(), predicate)
}

func TestAndPredicateWhenUnpinnedAndLatestSet(t *testing.T) {
	fo := &MatchingOptions{}
	fo.DigestPredicates.Unpinned = true
	fo.TagPredicates.Latest = true

	predicate := fo.getPredicate()

	assert.IsType(t, dockproc.AndPredicateNew(nil), predicate)

	andPredicate, _ := predicate.(dockproc.AndPredicate)
	andPredicate.Predicates()

	assert.IsType(t, dockproc.AndPredicateNew(nil), predicate)

	matches := 0
	for _, p := range andPredicate.Predicates() {
		t := reflect.TypeOf(p)
		switch t {
		case reflect.TypeOf(dockproc.UnpinnedPredicateNew()):
			fallthrough
		case reflect.TypeOf(dockproc.LatestPredicateNew()):
			matches++
		}
	}

	assert.Equal(t, 2, matches)
}

var unimplemented = []string{outdatedPred}

func TestHelpContainsImplementedPredicates(t *testing.T) {

	mo := MatchingOptions{}
	parser := flags.NewParser(&mo, flags.HelpFlag)
	buffer := bytes.NewBuffer(nil)
	parser.WriteHelp(buffer)

	for _, name := range predicateNames {
		if indexOf(name, unimplemented) < 0 {
			t.Run(fmt.Sprintf("Contains %s", name), func(t *testing.T) {
				assert.Contains(t, buffer.String(), "--"+name)
			})
		} else {
			t.Run(fmt.Sprintf("Not Contains %s", name), func(t *testing.T) {
				assert.NotContains(t, buffer.String(), "--"+name)
			})
		}
	}
}
