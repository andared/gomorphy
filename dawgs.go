package gomorphy

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

type guide struct {
	units []byte // len = 2 * n_nodes; units[i*2]=child, units[i*2+1]=sibling
}

func (g *guide) child(index uint32) byte   { return g.units[index*2] }
func (g *guide) sibling(index uint32) byte { return g.units[index*2+1] }

func (g *guide) read(r io.Reader) error {
	var size uint32
	if err := binary.Read(r, binary.NativeEndian, &size); err != nil {
		return err
	}
	g.units = make([]byte, size*2)
	_, err := io.ReadFull(r, g.units)
	return err
}

type completer struct {
	dict       *dictionary
	guide      *guide
	key        []byte
	indexStack []uint32
	lastIndex  uint32
}

func newCompleter(d *dictionary, g *guide) *completer {
	return &completer{dict: d, guide: g}
}

func (c *completer) start(index uint32, prefix []byte) {
	c.key = append(c.key[:0], prefix...)
	c.lastIndex = 0
	c.indexStack = c.indexStack[:0]
	c.indexStack = append(c.indexStack, index)
}

func (c *completer) next() bool {
	if len(c.indexStack) == 0 {
		return false
	}
	index := c.indexStack[len(c.indexStack)-1]

	if c.lastIndex != 0 { // not initial call
		childLabel := c.guide.child(index)
		if childLabel != 0 {
			index = c.follow(childLabel, index)
			if index == 0 {
				return false
			}
		} else {
			for {
				siblingLabel := c.guide.sibling(index)
				if len(c.key) > 0 {
					c.key = c.key[:len(c.key)-1]
				}
				c.indexStack = c.indexStack[:len(c.indexStack)-1]
				if len(c.indexStack) == 0 {
					return false
				}
				index = c.indexStack[len(c.indexStack)-1]
				if siblingLabel != 0 {
					index = c.follow(siblingLabel, index)
					if index == 0 {
						return false
					}
					break
				}
			}
		}
	}
	return c.findTerminal(index)
}

func (c *completer) follow(label byte, index uint32) uint32 {
	next, ok := c.dict.followChar(uint32(label), index)
	if !ok {
		return 0
	}
	c.key = append(c.key, label)
	c.indexStack = append(c.indexStack, next)
	return next
}

func (c *completer) findTerminal(index uint32) bool {
	for !c.dict.hasValue(index) {
		label := c.guide.child(index)
		next, ok := c.dict.followChar(uint32(label), index)
		if !ok {
			return false
		}
		c.key = append(c.key, label)
		c.indexStack = append(c.indexStack, next)
		index = next
	}
	c.lastIndex = index
	return true
}

// https://pymorphy2.readthedocs.io/en/0.7/internals/dict.html
// В библиотеке pymorphy3 (как и в её предшественнике pymorphy2) параметры para_id и idx являются внутренними числовыми идентификаторами словаря (OpenCorpora),
// которые однозначно определяют положение конкретной словоформы в языковой матрице.
// para_id (Paradigm ID) — это уникальный идентификатор (номер) парадигмы. Он указывает на конкретную таблицу склонения или спряжения. Все словоформы одного лексического значения (например, все формы слова «стол»: стола, столу, столами и т.д.) имеют один и тот же para_id.
// idx (Index) — это индекс (порядковый номер) формы внутри этой парадигмы. Он определяет конкретную грамматическую форму (например, 0 — именительный падеж единственного числа, 1 — родительный падеж единственного числа и т.д.).
type formParadigm struct {
	cnt    float32
	paraId int
	idx    int
}

type wordsDawg struct {
	dict            *dictionary
	guide           *guide
	isPredSuffix    bool
	dawgPayloadSep  byte // BytesDAWG payload separator byte
	charSubstitutes map[rune]compiledReplaces
}

// load reads a words.dawg file from r.
// File layout: Dictionary data | Guide data (concatenated).
func (w *wordsDawg) load(r io.Reader) error {
	w.dict = new(dictionary)
	if err := w.dict.read(r); err != nil {
		return err
	}

	w.isPredSuffix = false
	w.dawgPayloadSep = 0x01

	w.guide = new(guide)
	return w.guide.read(r)
}

// Parse a word using this dictionary.
func (w *wordsDawg) parse(anal analyzer, word string, wordLower string, seenParses map[seenParse]bool) []*Parse {
	res := make([]*Parse, 0, 25) // 25 примерное среднее число форм слова с запасиком
	paraData := w.similarItems(wordLower)

	for _, para := range paraData {
		fixedWord := para.fixedWord
		parses := para.paradigms
		for _, f := range parses {
			normalForm := w.dict.buildNormalForm(f, fixedWord)

			tag := w.dict.buildTagInfo(f.paraId, f.idx)
			methods := []Method{{anal, fixedWord, f.paraId, f.idx}}
			res = append(res, &Parse{
				Word:         fixedWord,
				Tag:          newOpencorporaTag(tag),
				NormalForm:   normalForm,
				Score:        1.0,
				MethodsStack: methods,
			})
		}
	}

	return res
}

// Tag a word using this dictionary.
func (w *wordsDawg) tag(word string, wordLower string, seenTags map[string]bool) []*opencorporaTag {

	paraData := w.similarItemValues(wordLower)

	result := make([]*opencorporaTag, 0, 10) // 10 примерное сред. кол-во тегов у слова
	for _, parse := range paraData {
		for _, p := range parse.paradigms {
			paragigm := w.dict.paradigms[p.paraId]
			tagId := paragigm[(len(paragigm)/3)+p.idx]
			result = append(result, newOpencorporaTag(w.dict.gramtab[tagId]))
		}
	}

	return result
}

// Returns a list of (key, value) tuples for all variants of “key“
// in this DAWG according to “replaces“.
// “replaces“ is an object obtained from
// “DAWG.compile_replaces(mapping)“ where mapping is a dict
// that maps single-char unicode strings to (one or more) single-char
// unicode strings.
func (w *wordsDawg) similarItems(key string) []similarItem {
	return w.getSimilarItems("", key, w.dict.root)
}

type similarItem struct {
	fixedWord string
	paradigms []formParadigm
}

func (w *wordsDawg) getSimilarItems(
	currentPrefix string,
	key string,
	index uint32,
) []similarItem {
	res := []similarItem{}
	startPos := utf8.RuneCountInString(currentPrefix)
	endPos := utf8.RuneCountInString(key)
	wordPos := startPos

	loopIsBreaked := false
	for wordPos < endPos {
		bStep, _ := runeAt(key, wordPos)
		replaceChars, ok := w.charSubstitutes[bStep]
		if ok {
			nextIndex := index
			if nextIndex, ok = w.dict.followBytes(replaceChars.bReplaceChar, nextIndex); ok {
				prefix := currentPrefix + stringSliceByRuneIndex(key, startPos, wordPos) + replaceChars.uReplaceChar
				extraItems := w.getSimilarItems(prefix, key, nextIndex)
				res = append(res, extraItems...)
			}
		}
		if index, ok = w.dict.followBytes([]byte(string(bStep)), index); !ok {
			loopIsBreaked = true
			break
		}
		wordPos++
	}

	var ok bool
	if !loopIsBreaked {
		if index, ok = w.dict.followBytes([]byte{w.dawgPayloadSep}, index); ok {
			foundKey := currentPrefix + stringSliceByRuneIndex(key, startPos, -1)
			value := w.valueForIndex(index)
			res = append([]similarItem{{foundKey, value}}, res...)
		}
	}

	return res
}

// Returns a list of values for all variants of the “key“
// in this DAWG according to “replaces“.
//
// “replaces“ is an object obtained from
// “DAWG.compile_replaces(mapping)“ where mapping is a dict
// that maps single-char unicode strings to (one or more) single-char
// unicode strings.
func (w *wordsDawg) similarItemValues(key string) []similarItem {
	return w.getSimilarItemValues(0, key, w.dict.root)
}

func (w *wordsDawg) getSimilarItemValues(
	startPos int,
	key string,
	index uint32,
) []similarItem {

	res := []similarItem{}
	endPos := utf8.RuneCountInString(key)
	wordPos := startPos

	loopIsBreaked := false
	for wordPos < endPos {
		bStep, _ := runeAt(key, wordPos)
		replaceChars, ok := w.charSubstitutes[bStep]
		if ok {
			nextIndex := index
			if nextIndex, ok = w.dict.followBytes(replaceChars.bReplaceChar, nextIndex); ok {
				extraItems := w.getSimilarItemValues(wordPos+1, key, nextIndex)
				res = append(res, extraItems...)
			}
		}
		if index, ok = w.dict.followBytes([]byte(string(bStep)), index); !ok {
			loopIsBreaked = true
			break
		}
		wordPos++
	}

	var ok bool
	if !loopIsBreaked {
		if index, ok = w.dict.followBytes([]byte{w.dawgPayloadSep}, index); ok {
			value := w.valueForIndex(index)
			res = append([]similarItem{{paradigms: value}}, res...)
		}
	}

	return res
}

func (w *wordsDawg) valueForIndex(index uint32) []formParadigm {
	res := []formParadigm{}
	completer := newCompleter(w.dict, w.guide)
	completer.start(index, nil)
	for completer.next() {
		key := completer.key
		if len(key) > 0 && key[len(key)-1] == '\n' {
			key = key[:len(key)-1]
		}

		decoded, err := base64.StdEncoding.DecodeString(string(key))
		if err != nil {
			continue
		}

		if w.isPredSuffix && len(decoded) < 8 {
			continue
		} else if !w.isPredSuffix && len(decoded) < 4 {
			continue
		}

		if w.isPredSuffix {
			// python struct.unpack >IHH
			res = append(res, formParadigm{
				cnt:    float32(binary.BigEndian.Uint32(decoded[0:4])),
				paraId: int(binary.BigEndian.Uint16(decoded[4:6])),
				idx:    int(binary.BigEndian.Uint16(decoded[6:8])),
			})

		} else {
			// python struct.unpack >HH
			res = append(res, formParadigm{
				paraId: int(binary.BigEndian.Uint16(decoded[0:2])),
				idx:    int(binary.BigEndian.Uint16(decoded[2:4])),
			})
		}

	}

	return res
}

func (w *wordsDawg) contains(key string) bool {
	_, ok := w.dict.followKey(key, []byte{w.dawgPayloadSep})
	return ok
}

// Check if a “word“ is in the dictionary.
//
// To allow some fuzzyness pass “substitutes_compiled“ argument;
// it should be a result of :meth:`DAWG.compile_replaces()`.
// This way you can e.g. handle ё letters replaced with е in the
// input words.
//
// .. note::
//
//	Dictionary words are not always correct words;
//	the dictionary also contains incorrect forms which
//	are commonly used. So for spellchecking tasks this
//	method should be used with extra care.
func (w *wordsDawg) wordIsKnown(wordLower string, substitutesCompiled map[rune]compiledReplaces) bool {
	if substitutesCompiled != nil {
		return len(w.similarKeys(wordLower, substitutesCompiled)) != 0
	}
	return w.contains(wordLower)
}

// Returns all variants of “key“ in this DAWG according to “replaces“.
//
// “replaces“ is an object obtained from
// “DAWG.compile_replaces(mapping)“ where mapping is a dict
// that maps single-char unicode strings to (one or more) single-char
// unicode strings.
//
// This may be useful e.g. for handling single-character umlauts.
func (w *wordsDawg) similarKeys(key string, replaces map[rune]compiledReplaces) []string {
	return w.getSimilarKeys("", key, 0, replaces)
}

func (w *wordsDawg) getSimilarKeys(
	currentPrefix string,
	key string,
	index uint32,
	replaces map[rune]compiledReplaces,
) []string {

	res := []string{}
	startPos := utf8.RuneCountInString(currentPrefix)
	endPos := utf8.RuneCountInString(key)
	wordPos := startPos

	loopIsBreaked := false
	for wordPos < endPos {
		bStep, _ := runeAt(key, wordPos)
		replaceChars, ok := replaces[bStep]
		if ok {
			nextIndex := index
			if nextIndex, ok = w.dict.followBytes(replaceChars.bReplaceChar, nextIndex); ok {
				prefix := currentPrefix + stringSliceByRuneIndex(key, startPos, wordPos) + replaceChars.uReplaceChar
				extraKeys := w.getSimilarKeys(prefix, key, nextIndex, replaces)
				res = append(res, extraKeys...)
			}
		}
		if index, ok = w.dict.followBytes([]byte(string(bStep)), index); !ok {
			loopIsBreaked = true
			break
		}
		wordPos++
	}

	if !loopIsBreaked {
		if _, ok := w.dict.followBytes([]byte{w.dawgPayloadSep}, index); ok {
			foundKey := currentPrefix + stringSliceByRuneIndex(key, startPos, -1)
			res = append([]string{foundKey}, res...)
		}
	}

	return res
}

type intDawg struct {
	dict  *dictionary
	guide *guide
}

// load reads a ints dawg file from r.
// File layout: Dictionary data | Guide data (concatenated).
func (w *intDawg) load(r io.Reader) error {
	w.dict = new(dictionary)
	if err := w.dict.read(r); err != nil {
		return err
	}
	w.guide = new(guide)
	return w.guide.read(r)
}

func (w *intDawg) prob(word string, tag string) float32 {
	dawgKey := word + ":" + tag
	return float32(w.get(dawgKey, 0)) / 1000000 // MULTIPLIER = 1000000
}

// аргумент word в формате "wordLower:tag" (ex. "живой:NOUN,anim,masc sing,nomn")
// аргумент def - значение по умолчанию, если запись не найдена
func (w *intDawg) get(word string, def int64) int64 {
	res := w.bGetValue(word)
	if res == -1 { // LOOKUP_ERROR
		return def
	}
	return res
}

func (w *intDawg) bGetValue(key string) int64 {
	return w.dict.find([]byte(key))
}

type predictionSuffixesDawgs struct {
	predictSfxDawgs [3]*wordsDawg
}

func newPredictionSuffixesDawgs(opencorporaPath string) (predictionSuffixesDawgs, error) {
	pds := predictionSuffixesDawgs{
		predictSfxDawgs: [3]*wordsDawg{
			{isPredSuffix: true, dawgPayloadSep: 0x01},
			{isPredSuffix: true, dawgPayloadSep: 0x01},
			{isPredSuffix: true, dawgPayloadSep: 0x01},
		},
	}
	err := pds.load(opencorporaPath)
	return pds, err
}

// load reads a words.dawg file from r.
// File layout: Dictionary data | Guide data (concatenated).
func (w *predictionSuffixesDawgs) load(opencorporaPath string) error {
	for i, wd := range w.predictSfxDawgs {
		fName := fmt.Sprintf("prediction-suffixes-%d.dawg", i)
		fPath := filepath.Join(opencorporaPath, fName)
		raw, err := os.ReadFile(fPath)
		if err != nil {
			return err
		}

		reader := bytes.NewReader(raw)

		wd.dict = new(dictionary)
		if err := wd.dict.read(reader); err != nil {
			return err
		}
		wd.guide = new(guide)
		if err = wd.guide.read(reader); err != nil {
			return nil
		}
	}

	return nil
}
