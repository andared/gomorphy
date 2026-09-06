package gomorphy

import (
	"errors"
	"log"
	"maps"
	"strings"
	"unicode/utf8"
)

// Analyzer units for unknown words with hyphens
// ---------------------------------------------

// class HyphenSeparatedParticleAnalyzer(AnalogyAnalyzerUnit):
//
// Parse the word by analyzing it without
// a particle after a hyphen.
//
// Example: смотри-ка -> смотри + "-ка".
//
// .. note::
//
//	This analyzer doesn't remove particles from the result
//	so for normalization you may need to handle
//	particles at tokenization level.
type hyphenSeparatedParticleAnalyzer struct {
	morph                *MorphAnalyzer
	name                 string
	terminal             bool
	scoreMultiplier      float32
	particlesAfterHyphen []string
	*analogyAnalyzerUnit
}

func newHyphenSeparatedParticleAnalyzer(morph *MorphAnalyzer) *hyphenSeparatedParticleAnalyzer {
	return &hyphenSeparatedParticleAnalyzer{
		morph:                morph,
		name:                 "HyphenSeparatedParticleAnalyzer",
		terminal:             true,
		scoreMultiplier:      0.9,
		particlesAfterHyphen: []string{"-то", "-ка", "-таки", "-де", "-тко", "-тка", "-с", "-ста"},
		// ukr ["-но", "-таки", "-бо", "-от"]
		analogyAnalyzerUnit: newAnalogyAnalyzerUnit(),
	}
}

func (self *hyphenSeparatedParticleAnalyzer) Name() string {
	return self.name
}

func (self *hyphenSeparatedParticleAnalyzer) isTerminal() bool {
	return self.terminal
}

func (self *hyphenSeparatedParticleAnalyzer) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	var result []*Parse
	for _, split := range self.possibleSplits(wordLower) {
		method := Method{Analyzer: self, WordOrStack: split.suffix}
		for _, p := range self.morph.Parse(split.unsuffixedWord) {
			parse := newParse(
				p.Word+split.suffix,
				p.Tag.RawTagsString,
				p.Score*self.scoreMultiplier,
			)
			parse.NormalForm = p.NormalForm + split.suffix
			parse.MethodsStack = append(p.MethodsStack, method)

			if addParseIfNotSeen(parse, seenParses) {
				result = append(result, parse)
			}
		}

		// # If a word ends with with one of the particles,
		// # it can't ends with an another.
		break
	}

	return result
}

func (self *hyphenSeparatedParticleAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	result := make([]*opencorporaTag, 0, 10)
	for _, split := range self.possibleSplits(wordLower) {
		result = append(result, self.morph.Tag(split.unsuffixedWord)...)

		// # If a word ends with with one of the particles,
		// # it can't ends with an another.
		break
	}
	return result
}

func (self *hyphenSeparatedParticleAnalyzer) possibleSplits(word string) []suffixWord {
	if !strings.ContainsRune(word, '-') {
		return nil
	}

	result := make([]suffixWord, 0, len(self.particlesAfterHyphen))
	for _, particle := range self.particlesAfterHyphen {
		if !strings.HasSuffix(word, particle) {
			continue
		}

		unsuffixedWord := stringSliceByRuneIndex(
			word,
			-1,
			utf8.RuneCountInString(word)-utf8.RuneCountInString(particle),
		)

		if unsuffixedWord == "" {
			continue
		}
		result = append(result, suffixWord{unsuffixedWord: unsuffixedWord, suffix: particle})
	}

	return result
}

func (self *hyphenSeparatedParticleAnalyzer) getLexeme(form *Parse) []*Parse {
	baseAnalyzer, thisMethod := self.methodInfo(form)
	particle := thisMethod.WordOrStack.(string)
	f := withoutFixedSuffix(form, len(particle))
	f = withoutLastMethod(f)
	lexeme := baseAnalyzer.getLexeme(f)
	result := make([]*Parse, 0, len(lexeme))
	for _, lex := range lexeme {
		result = append(result, appendMethod(withSuffix(lex, particle), thisMethod))
	}

	return result
}

func (self *hyphenSeparatedParticleAnalyzer) normalized(form *Parse) *Parse {
	baseAnalyzer, thisMethod := self.methodInfo(form)
	particle := thisMethod.WordOrStack.(string)
	f := withoutFixedSuffix(form, len(particle))
	f = withoutLastMethod(f)
	normalForm := baseAnalyzer.normalized(f)
	return appendMethod(withSuffix(normalForm, particle), thisMethod)
}

// class HyphenAdverbAnalyzer(BaseAnalyzerUnit):
//
//	Detect adverbs that starts with "по-".
//
//	Example: по-западному
type hyphenAdverbAnalyzer struct {
	morph           *MorphAnalyzer
	name            string
	selfTag         string
	terminal        bool
	scoreMultiplier float32
}

func newHyphenAdverbAnalyzer(morph *MorphAnalyzer) *hyphenAdverbAnalyzer {
	return &hyphenAdverbAnalyzer{
		morph:           morph,
		name:            "HyphenAdverbAnalyzer",
		selfTag:         "ADVB",
		terminal:        true,
		scoreMultiplier: 0.7,
	}
}

func (self *hyphenAdverbAnalyzer) Name() string {
	return self.name
}

func (self *hyphenAdverbAnalyzer) isTerminal() bool {
	return self.terminal
}

func (self *hyphenAdverbAnalyzer) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	if !self.shouldParse(wordLower) {
		return nil
	}

	parse := newParse(
		wordLower,
		self.selfTag,
		self.scoreMultiplier,
	)
	parse.NormalForm = wordLower
	parse.MethodsStack = []Method{{Analyzer: self, WordOrStack: word}}

	seenParses[seenParse{word: parse.Word, tag: self.selfTag}] = true

	return []*Parse{parse}
}

func (self *hyphenAdverbAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	if !self.shouldParse(wordLower) || seenTags[self.selfTag] {
		return nil
	}
	seenTags[self.selfTag] = true
	return []*opencorporaTag{newOpencorporaTag(self.selfTag)}
}

func (self *hyphenAdverbAnalyzer) shouldParse(word string) bool {
	if utf8.RuneCountInString(word) < 5 {
		return false
	}

	if w, ok := strings.CutPrefix(word, "по-"); ok {
		for _, tag := range self.morph.Tag(w) {
			if tag.Contains("ADJF") && tag.Contains("sing") && tag.Contains("datv") {
				return true
			}
		}
	}
	return false
}

func (self *hyphenAdverbAnalyzer) getLexeme(form *Parse) []*Parse {
	return []*Parse{form}
}

func (self *hyphenAdverbAnalyzer) normalized(form *Parse) *Parse {
	return form
}

// class HyphenatedWordsAnalyzer(BaseAnalyzerUnit):
//
//	Parse the word by parsing its hyphen-separated parts.
//
//	Examples:
//
//	    * интернет-магазин -> "интернет-" + магазин
//	    * человек-гора -> человек + гора
type hyphenatedWordsAnalyzer struct {
	morph            *MorphAnalyzer
	name             string
	terminal         bool
	scoreMultiplier  float32
	considerTheSame  map[string]string
	similarFeatures  map[string]string
	featureGrammemes map[string]bool
	*prefixMatcher
}

func newHyphenatedWordsAnalyzer(morph *MorphAnalyzer) *hyphenatedWordsAnalyzer {
	a := &hyphenatedWordsAnalyzer{
		morph:           morph,
		name:            "HyphenatedWordsAnalyzer",
		terminal:        true,
		scoreMultiplier: 0.75,
		prefixMatcher:   getPrefixMatcherInstance(),
	}

	a.considerTheSame = map[string]string{
		"V-oy": "V-ey",
		"gen1": "gent",
		"loc1": "loct",
		// # 'acc1': 'accs',
	}

	a.similarFeatures = map[string]string{"gen1": "gent", "loc1": "loct"}

	a.featureGrammemes = map[string]bool{
		"PRTF": true, "NPRO": true, "CONJ": true, "COMP": true, "INFN": true, "NUMR": true,
		"ADJF": true, "voct": true, "ablt": true, "ADVB": true, "past": true, "loct": true,
		"nomn": true, "loc2": true, "pres": true, "PRTS": true, "2per": true, "loc1": true,
		"gen1": true, "INTJ": true, "PREP": true, "datv": true, "plur": true, "futr": true,
		"NOUN": true, "acc2": true, "gen2": true, "GRND": true, "ADJS": true, "1per": true,
		"sing": true, "VERB": true, "PRED": true, "3per": true, "accs": true, "gent": true,
		"PRCL": true,
	}

	return a
}

func (self *hyphenatedWordsAnalyzer) Name() string {
	return self.name
}

func (self *hyphenatedWordsAnalyzer) isTerminal() bool {
	return self.terminal
}

func (self *hyphenatedWordsAnalyzer) parse(word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	if !self.shouldParse(wordLower) {
		return nil
	}

	wSplit := strings.SplitN(wordLower, "-", 2) // 2 - max кол-во подстрок
	if len(wSplit) < 2 {
		return nil
	}
	left, right := wSplit[0], wSplit[1]
	leftParses := self.morph.Parse(left)
	rightParses := self.morph.Parse(right)

	result := self.parseAsVariableBoth(leftParses, rightParses)

	// # We copy `seen_parses` to preserve parses even if similar parses
	// # were observed at previous step (they may have different lexemes).
	seen := map[seenParse]bool{}
	maps.Copy(seen, seenParses)
	result = append(result, self.parseAsFixedLeft(rightParses, left)...)
	maps.Copy(seenParses, seen)

	return result
}

func (self *hyphenatedWordsAnalyzer) tag(word, wordLower string, seenTags map[string]bool) []*opencorporaTag {
	// # By default .tag() uses .parse().
	// # Usually it is possible to write a more efficient implementation;
	// # analyzers should do it when possible.
	parses := self.parse(word, wordLower, map[seenParse]bool{})
	result := make([]*opencorporaTag, 0, len(parses))
	for _, p := range parses {
		if addTagIfNotSeen(p.Tag, seenTags) {
			result = append(result, p.Tag)
		}
	}
	return result
}

func (self *hyphenatedWordsAnalyzer) parseAsFixedLeft(rightParses []*Parse, left string) []*Parse {
	// Step 1: Assume that the left part is an immutable prefix.
	// Examples: интернет-магазин, воздушно-капельный
	result := make([]*Parse, 0, len(rightParses))

	for _, p := range rightParses {
		if p.Tag.isUnknown() {
			continue
		}

		newMethodsStack := []Method{{Analyzer: self, WordOrStack: left, ParaIdOrStack: p.MethodsStack}}

		parse := newParse(
			left+"-"+p.Word,
			p.Tag.RawTagsString,
			p.Score*self.scoreMultiplier,
		)
		parse.NormalForm = left + "-" + p.NormalForm
		parse.MethodsStack = newMethodsStack

		result = append(result, parse)
	}

	return result
}

func (self *hyphenatedWordsAnalyzer) parseAsVariableBoth(leftParses, rightParses []*Parse) []*Parse {
	// Step 2: if left and right can be parsed the same way,
	// then it may be the case that both parts should be inflected.
	// Examples: человек-гора, команд-участниц, компания-производитель

	result := make([]*Parse, 0, len(rightParses))
	rightFeatures := make([]map[string]bool, len(rightParses))
	for i, p := range rightParses {
		rightFeatures[i] = self.similarityFeatures(p.Tag)
	}

	for _, leftParse := range leftParses {
		leftTag := leftParse.Tag

		if leftTag.isUnknown() {
			continue
		}

		leftFeat := self.similarityFeatures(leftTag)

		for parseIndex, rightParse := range rightParses {
			rightFeat := rightFeatures[parseIndex]
			if !maps.Equal(leftFeat, rightFeat) {
				continue
			}

			leftMethods := leftParse.MethodsStack
			rightMethods := rightParse.MethodsStack
			newMethodsStack := []Method{{Analyzer: self, WordOrStack: leftMethods, ParaIdOrStack: rightMethods}}

			// # tag
			parse := newParse(
				leftParse.Word+"-"+rightParse.Word,
				leftTag.RawTagsString,
				leftParse.Score*self.scoreMultiplier,
			)
			parse.NormalForm = leftParse.NormalForm + "-" + rightParse.NormalForm
			parse.MethodsStack = newMethodsStack

			result = append(result, parse)
		}
	}
	return result
}

func (self *hyphenatedWordsAnalyzer) similarityFeatures(tag *opencorporaTag) map[string]bool {
	return replaceGrammemes(
		intersection(tag.tags, self.featureGrammemes),
		self.similarFeatures,
	)
}

func (self *hyphenatedWordsAnalyzer) shouldParse(word string) bool {
	if !strings.ContainsRune(word, '-') {
		return false
	}

	wordStripped := strings.Trim(word, "-")
	if wordStripped != word {
		// # don't handle words that start of end with a hyphen
		return false
	}
	if strings.Count(wordStripped, "-") != 1 {
		// # require exactly 1 hyphen, in the middle of the word
		return false
	}
	if self.prefixMatcher.isPrefixed(word) {
		// # such words should really be parsed by KnownPrefixAnalyzer
		return false
	}

	return true
}

func (self *hyphenatedWordsAnalyzer) getLexeme(form *Parse) []*Parse {
	return self.iterLexeme(form)
}

func (self *hyphenatedWordsAnalyzer) normalized(form *Parse) *Parse {
	return self.iterLexeme(form)[0]
}

func (self *hyphenatedWordsAnalyzer) iterLexeme(form *Parse) []*Parse {
	if len(form.MethodsStack) != 1 {
		log.Fatal(errors.New("len(methods_stack) != 1"))
	} else if form.MethodsStack[0].Analyzer.Name() != self.name {
		log.Fatal(errors.New("form.MethodsStack[0].Analyzer.Name() != self.name"))
	}

	thisMethod := form.MethodsStack[0].Analyzer
	leftMethods := form.MethodsStack[0].WordOrStack
	rightMethods := form.MethodsStack[0].ParaIdOrStack.([]Method)

	// если left_methods строка (не tuple)
	if self.fixedLeftMethodWasUsed(leftMethods) {

		// # Form is obtained by parsing right part,
		// # assuming that left part is an uninflected prefix.
		// # Lexeme can be calculated from the right part in this case:

		prefix := leftMethods.(string) + "-"
		rightForm := withoutFixedPrefix(
			replaceMethodsStack(form, rightMethods),
			len(prefix),
		)

		baseAnalyzer := rightMethods[len(rightMethods)-1].Analyzer
		lexeme := baseAnalyzer.getLexeme(rightForm)
		result := make([]*Parse, len(lexeme))
		for i, f := range lexeme {
			result[i] = replaceMethodsStack(
				withPrefix(f, prefix),
				[]Method{{Analyzer: thisMethod, WordOrStack: leftMethods, ParaIdOrStack: f.MethodsStack}},
			)
		}

		return result
	}

	// # Form is obtained by parsing both parts.
	// # Compute lexemes for left and right parts,
	// # then merge them.

	leftMethodsAsSlice := leftMethods.([]Method)

	leftForm := self.withoutRightPart(
		replaceMethodsStack(form, leftMethodsAsSlice),
	)

	rightForm := self.withoutLeftPart(
		replaceMethodsStack(form, rightMethods),
	)

	leftAnalyzer := leftMethodsAsSlice[len(leftMethodsAsSlice)-1].Analyzer
	leftLexeme := leftAnalyzer.getLexeme(leftForm)

	rightAnalyzer := rightMethods[len(rightMethods)-1].Analyzer
	rightLexeme := rightAnalyzer.getLexeme(rightForm)

	return self.mergeLexemes(leftLexeme, rightLexeme)
}

func (self *hyphenatedWordsAnalyzer) mergeLexemes(leftLexeme, rightLexeme []*Parse) []*Parse {
	pairParses := self.alignLexemeForms(leftLexeme, rightLexeme)
	result := make([]*Parse, len(pairParses))
	for i, pair := range pairParses {
		parse := newParse(
			pair.left.Word+"-"+pair.right.Word,
			pair.left.Tag.RawTagsString,
			(pair.left.Score+pair.right.Score)/2,
		)

		methodStack := []Method{{Analyzer: self, WordOrStack: pair.left.MethodsStack, ParaIdOrStack: pair.right.MethodsStack}}

		parse.NormalForm = pair.left.NormalForm + "-" + pair.right.NormalForm
		parse.MethodsStack = methodStack

		result[i] = parse
	}

	return result
}

type pairParse struct {
	left  *Parse
	right *Parse
}

func (self *hyphenatedWordsAnalyzer) alignLexemeForms(leftLexeme, rightLexeme []*Parse) []pairParse {
	result := make([]pairParse, len(rightLexeme))

	for i, right := range rightLexeme {
		min_dist := 1e6
		var closest *Parse
		gr_right := replaceGrammemes(right.Tag.grammemes(), self.considerTheSame)

		for _, left := range leftLexeme {
			gr_left := replaceGrammemes(left.Tag.grammemes(), self.considerTheSame)
			dist := float64(len(symmetricDifference(gr_left, gr_right)))
			if dist < min_dist {
				min_dist = dist
				closest = left
			}
		}
		result[i] = pairParse{closest, right}
	}

	return result
}

func (self *hyphenatedWordsAnalyzer) withoutRightPart(form *Parse) *Parse {
	p := newParse(
		form.Word[:strings.IndexRune(form.Word, '-')],
		form.Tag.RawTagsString,
		form.Score,
	)
	p.NormalForm = form.NormalForm[:strings.IndexRune(form.NormalForm, '-')]
	p.MethodsStack = form.MethodsStack
	return p
}

func (self *hyphenatedWordsAnalyzer) withoutLeftPart(form *Parse) *Parse {
	p := newParse(
		form.Word[strings.IndexRune(form.Word, '-')+1:], // +utf8.RuneLen('-')
		form.Tag.RawTagsString,
		form.Score,
	)
	p.NormalForm = form.NormalForm[strings.IndexRune(form.NormalForm, '-')+1:]
	p.MethodsStack = form.MethodsStack
	return p
}

func (self *hyphenatedWordsAnalyzer) fixedLeftMethodWasUsed(leftMethods any) bool {
	_, ok := leftMethods.([]Method)
	return !ok
}

func replaceGrammemes(grammemes map[string]bool, replaces map[string]string) map[string]bool {
	for gr, replace := range replaces {
		if grammemes[gr] {
			delete(grammemes, gr) // grammemes[gr] = false
			grammemes[replace] = true
		}
	}
	return grammemes
}
