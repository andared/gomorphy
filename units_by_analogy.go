package gomorphy

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Analogy analyzer units
// ----------------------
// This module provides analyzer units that analyzes unknown words by looking
// at how similar known words are analyzed.

// class KnownPrefixAnalyzer(_PrefixAnalyzer)
//
// Parse the word by checking if it starts with a known prefix
// and parsing the remainder.
// Example: авиакошка -> (авиа) + кошка.
type knownPrefixAnalyzer struct {
	morph              *MorphAnalyzer
	name               string
	terminal           bool
	scoreMultiplier    float32
	minRemainderLength int
	*prefixAnalyzer
	*prefixMatcher
}

// значения по умолчанию scoreMultiplier=0.75, min_remainder_length=3
func newKnownPrefixAnalyzer(morph *MorphAnalyzer) *knownPrefixAnalyzer {
	anal := &knownPrefixAnalyzer{
		morph:              morph,
		name:               "KnownPrefixAnalyzer",
		scoreMultiplier:    0.75,
		minRemainderLength: 3,
		terminal:           true,
		prefixAnalyzer:     newPrefixAnalyzer(),
		prefixMatcher:      getPrefixMatcherInstance(),
	}

	return anal
}

func (self *knownPrefixAnalyzer) Name() string {
	return self.name
}

func (self *knownPrefixAnalyzer) isTerminal() bool {
	return self.terminal
}

func (self *knownPrefixAnalyzer) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	var result []*Parse
	for _, split := range self.possibleSplits(wordLower) {
		method := Method{Analyzer: self, WordOrStack: split.prefix}

		parses := self.morph.Parse(split.unprefixedWord)
		for _, p := range parses {
			if !p.Tag.isProductive() {
				continue
			}
			parse := newParse(
				split.prefix+p.Word,
				p.Tag.RawTagsString,
				p.Score*self.scoreMultiplier,
			)
			parse.NormalForm = split.prefix + p.NormalForm
			parse.MethodsStack = append(p.MethodsStack, method)

			if addParseIfNotSeen(parse, seenParses) {
				result = append(result, parse)
			}
		}
	}

	return result
}

func (self *knownPrefixAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	result := []*opencorporaTag{}
	for _, split := range self.possibleSplits(wordLower) {
		for _, tag := range self.morph.Tag(split.unprefixedWord) {
			if !tag.isProductive() {
				continue
			}
			if addTagIfNotSeen(tag, seenTags) {
				result = append(result, tag)
			}
		}
	}
	return result
}

func (self *knownPrefixAnalyzer) possibleSplits(word string) []prefixWord {
	wordPrefixes := self.getPrefixes(word)
	slices.SortStableFunc(wordPrefixes, func(a, b string) int {
		return -1 * cmp.Compare(len(a), len(b)) // -1* реверс сразу на месте
	})

	result := make([]prefixWord, 0, len(wordPrefixes))
	for _, prefix := range wordPrefixes {
		unprefixedWord := stringSliceByRuneIndex(word, utf8.RuneCountInString(prefix), -1)
		if utf8.RuneCountInString(unprefixedWord) < self.minRemainderLength {
			continue
		}
		result = append(result, prefixWord{prefix: prefix, unprefixedWord: unprefixedWord})
	}
	return result
}

// class UnknownPrefixAnalyzer(_PrefixAnalyzer):
//
// Parse the word by parsing only the word suffix
// (with restrictions on prefix & suffix lengths).
// Example: байткод -> (байт) + код

type unknownPrefixAnalyzer struct {
	morph           *MorphAnalyzer
	name            string
	terminal        bool
	scoreMultiplier float32
	*prefixAnalyzer
	*prefixMatcher
}

func newUnknownPrefixAnalyzer(morph *MorphAnalyzer) *unknownPrefixAnalyzer {
	u := &unknownPrefixAnalyzer{
		morph:           morph,
		name:            "UnknownPrefixAnalyzer",
		scoreMultiplier: 0.5,
		terminal:        false,
		prefixAnalyzer:  newPrefixAnalyzer(),
		prefixMatcher:   getPrefixMatcherInstance(),
	}
	return u
}

func (self *unknownPrefixAnalyzer) Name() string {
	return self.name
}

func (self *unknownPrefixAnalyzer) isTerminal() bool {
	return self.terminal
}

func (self *unknownPrefixAnalyzer) parse(word, wordLower string, seenParses map[seenParse]bool) []*Parse {
	var result []*Parse
	for _, split := range wordSplits(wordLower) {
		method := Method{Analyzer: self, WordOrStack: split.prefix}

		// analyzers[0] — DictionaryAnalyzer
		parses := self.morph.analyzers[0].parse(split.unprefixedWord, split.unprefixedWord, seenParses)
		for _, p := range parses {
			if !p.Tag.isProductive() {
				continue
			}

			parse := newParse(
				split.prefix+p.Word,
				p.Tag.RawTagsString,
				p.Score*self.scoreMultiplier,
			)
			parse.NormalForm = split.prefix + p.NormalForm
			parse.MethodsStack = append(p.MethodsStack, method)

			if addParseIfNotSeen(parse, seenParses) {
				result = append(result, parse)
			}
		}
	}

	return result
}

func (self *unknownPrefixAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	result := []*opencorporaTag{}
	for _, split := range wordSplits(wordLower) {
		tags := self.morph.words.tag(split.unprefixedWord, split.unprefixedWord, seenTags)
		for _, tag := range tags {
			if !tag.isProductive() {
				continue
			}
			if addTagIfNotSeen(tag, seenTags) {
				result = append(result, tag)
			}
		}
	}
	return result
}

type paradigmPrefix struct {
	prefixId int
	prefix   string
}

// class KnownSuffixAnalyzer(AnalogyAnalyzerUnit):
//
// Parse the word by checking how the words with similar suffixes
// are parsed.
// Example: бутявкать -> ...вкать
type knownSuffixAnalyzer struct {
	morph            *MorphAnalyzer
	name             string
	terminal         bool
	minWordLength    int
	scoreMultiplier  float32
	paradigmPrefixes [3]paradigmPrefix
	predictionSplits []int
	fakeDict         *fakeDictionary
	*analogyAnalyzerUnit
}

// значения по умолчанию scoreMultiplier=0.5,  minWordLength=4
func newKnownSuffixAnalyzer(morph *MorphAnalyzer, dawg *wordsDawg) *knownSuffixAnalyzer {
	a := &knownSuffixAnalyzer{
		morph:               morph,
		name:                "KnownSuffixAnalyzer",
		terminal:            true,
		minWordLength:       4,
		scoreMultiplier:     0.5,
		paradigmPrefixes:    [3]paradigmPrefix{},
		predictionSplits:    make([]int, dawg.dict.maxSuffixLength),
		fakeDict:            newFakeDictionary(dawg),
		analogyAnalyzerUnit: newAnalogyAnalyzerUnit(),
	}

	// self.paradigmPrefixes = [(2, 'наи'), (1, 'по'), (0, '')]
	for i, j := len(dawg.dict.paradigmPrefixes)-1, 0; i >= 0; i, j = i-1, j+1 {
		a.paradigmPrefixes[j] = paradigmPrefix{
			prefixId: i,
			prefix:   dawg.dict.paradigmPrefixes[i],
		}
	}

	// self.predictionSplits = [5, 4, 3, 2, 1]
	for i, j := dawg.dict.maxSuffixLength, 0; i >= 1; i, j = i-1, j+1 {
		a.predictionSplits[j] = i
	}

	return a
}

func (self *knownSuffixAnalyzer) Name() string {
	return self.name
}

func (self *knownSuffixAnalyzer) isTerminal() bool {
	return self.terminal
}

type knownSuffixParse struct {
	cnt          float32
	word         string
	tag          string
	normalForm   string
	prefixId     int
	methodsStack []Method
}

func (self *knownSuffixAnalyzer) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {

	if utf8.RuneCountInString(word) < self.minWordLength {
		return nil
	}

	totalCount := make([]float32, len(self.paradigmPrefixes))
	for i := range len(self.paradigmPrefixes) {
		totalCount[i] = 1
	}

	wordLowerLen := utf8.RuneCountInString(wordLower)
	res := []knownSuffixParse{}
	for _, posbPfx := range self.possiblePrefixes(wordLower) {

		for _, splitIdx := range self.predictionSplits {
			// # XXX: this should be counted once, not for each prefix
			wordStart := stringSliceByRuneIndex(wordLower, -1, max(wordLowerLen-splitIdx, 0))
			wordEnd := stringSliceByRuneIndex(wordLower, max(wordLowerLen-splitIdx, 0), -1)
			paraData := posbPfx.suffixesDawg.similarItems(wordEnd)

			for _, para := range paraData {
				fixedSuffix := para.fixedWord
				parses := para.paradigms
				fixedWord := wordStart + fixedSuffix

				for _, p := range parses {
					tag := newOpencorporaTag(self.morph.words.dict.buildTagInfo(p.paraId, p.idx))

					// # skip non-productive tags
					if !tag.isProductive() {
						continue
					}

					totalCount[posbPfx.prefixId] += p.cnt

					// # avoid duplicate parses
					reducedParse := seenParse{word: fixedWord, tag: tag.RawTagsString, paraId: strconv.Itoa(p.paraId)}
					if _, ok := seenParses[reducedParse]; ok {
						continue
					}
					seenParses[reducedParse] = true

					// # ok, build the result
					normalForm := self.morph.words.dict.buildNormalForm(p, fixedWord)
					methods := []Method{
						{Analyzer: self.fakeDict, WordOrStack: fixedWord, ParaIdOrStack: p.paraId, Idx: p.idx},
						{Analyzer: self, WordOrStack: fixedSuffix},
					}
					parse := knownSuffixParse{
						cnt:          p.cnt,
						word:         fixedWord,
						tag:          tag.RawTagsString,
						normalForm:   normalForm,
						prefixId:     posbPfx.prefixId,
						methodsStack: methods,
					}

					res = append(res, parse)
				}
			}

			if totalCount[posbPfx.prefixId] > 1 {
				break
			}
		}
	}

	result := make([]*Parse, len(res))
	for i, p := range res {
		parse := newParse(
			p.word,
			p.tag,
			p.cnt/totalCount[p.prefixId]*self.scoreMultiplier,
		)
		parse.NormalForm = p.normalForm
		parse.MethodsStack = p.methodsStack
		result[i] = parse
	}

	slices.SortStableFunc(result, func(a, b *Parse) int {
		return -1 * cmp.Compare(a.Score, b.Score)
	})

	return result
}

func (self *knownSuffixAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	// # XXX: the result order may be different from ``self.parse(...)``.

	if utf8.RuneCountInString(word) < self.minWordLength {
		return nil
	}

	result := []*opencorporaTag{}
	for _, posbPfx := range self.possiblePrefixes(wordLower) {
		for _, splitIdx := range self.predictionSplits {
			// # XXX: end should be counted once, not for each prefix
			end := stringSliceByRuneIndex(wordLower, max(utf8.RuneCountInString(wordLower)-splitIdx, 0), -1)
			paraData := posbPfx.suffixesDawg.similarItems(end)
			found := false
			for _, para := range paraData {
				for _, parse := range para.paradigms {
					tag := newOpencorporaTag(self.morph.words.dict.buildTagInfo(parse.paraId, parse.idx))
					if !tag.isProductive() {
						continue
					}
					tag.prob = parse.cnt
					found = true
					if seenTags[tag.RawTagsString] {
						continue
					}

					seenTags[tag.RawTagsString] = true

					result = append(result, tag)
				}
			}
			if found {
				break
			}
		}
	}

	// сортировка идентичная питону [ [1, OpencorporaTag], [1, OpencorporaTag] ]
	// Если первые элементы одинаковые, сравнение по вторым.
	// В pymorphy OpencorporaTag сравнение (__lt__  __gt__) по _grammemes_tuple
	slices.SortStableFunc(result, func(a, b *opencorporaTag) int {
		if a.prob != b.prob {
			return -1 * cmp.Compare(a.prob, b.prob) // -1* реверс сразу на месте
		}
		return -1 * slices.Compare(a.grammemesTuple, b.grammemesTuple)
	})

	return result
}

type prefixSuffix struct {
	prefixId     int
	prefix       string
	suffixesDawg *wordsDawg
}

func (self *knownSuffixAnalyzer) possiblePrefixes(word string) []prefixSuffix {
	result := make([]prefixSuffix, 0, len(self.paradigmPrefixes))
	for _, p := range self.paradigmPrefixes {
		if !strings.HasPrefix(word, p.prefix) {
			continue
		}
		result = append(result, prefixSuffix{
			prefixId:     p.prefixId,
			prefix:       p.prefix,
			suffixesDawg: self.morph.predSuffixes.predictSfxDawgs[p.prefixId],
		})

	}
	return result
}

// class FakeDictionary(DictionaryAnalyzer):
//
//	""" This is just a DictionaryAnalyzer with different __repr__ """
type fakeDictionary struct {
	name     string
	terminal bool
	*wordsDawg
}

func newFakeDictionary(dawg *wordsDawg) *fakeDictionary {
	return &fakeDictionary{
		name:      "FakeDictionary",
		terminal:  false,
		wordsDawg: dawg,
	}
}

func (self *fakeDictionary) Name() string {
	return self.name
}

func (self *fakeDictionary) isTerminal() bool {
	return self.terminal
}

func (self *fakeDictionary) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	return self.wordsDawg.parse(self, word, wordLower, seenParses)
}

func (self *fakeDictionary) getLexeme(form *Parse) []*Parse {
	return self.dict.getLexeme(self, form)
}

func (self *fakeDictionary) normalized(form *Parse) *Parse {
	return self.dict.normalized(self, form)
}

// метод tag "наследуется" от wordsDawg
