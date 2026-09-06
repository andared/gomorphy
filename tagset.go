package gomorphy

import (
	"fmt"
	"maps"
	"strings"
	"sync"
)

type opencorporaTag struct {
	prob           float32 // для applyToTags() и knownSuffixAnalyzer.tag()
	tags           map[string]bool
	grammemesCache map[string]bool // рабочая копия tags
	cacheMu        sync.RWMutex
	grammemesTuple []string
	POS            string
	Case           string
	Number         string
	Gender         string
	Animate        string
	RawTagsString  string
	*tagClass
}

func newOpencorporaTag(rawTagString string) *opencorporaTag {

	tag := &opencorporaTag{
		RawTagsString:  rawTagString,
		tags:           map[string]bool{},
		grammemesTuple: make([]string, 0, 10),
		tagClass:       getTagClassInstance(),
	}

	for part := range strings.SplitSeq(rawTagString, " ") {
		for g := range strings.SplitSeq(part, ",") {

			tag.grammemesTuple = append(tag.grammemesTuple, g)

			// заполяет tag.pos, tag.case... и мапу tags
			if tag.tagClass.partsOfSpeech[g] {
				tag.POS = g
			} else if tag.tagClass.cases[g] {
				tag.Case = g
			} else if tag.tagClass.numbers[g] {
				tag.Number = g
			} else if tag.tagClass.genders[g] {
				tag.Gender = g
			} else if tag.tagClass.animacy[g] {
				tag.Animate = g
			}
			tag.tags[g] = true
		}
	}

	return tag
}

// проверяет наличие тега в объекте Parse
func (o *opencorporaTag) Contains(tag string) bool {
	return o.tags[tag]
}

func (o *opencorporaTag) numeralAgreementGrammemes(num int) []string {

	var index int
	if (num%10 == 1) && (num%100 != 11) {
		index = 0
	} else if (num%10 >= 2) && (num%10 <= 4) && (num%100 < 10 || num%100 >= 20) {
		index = 1
	} else {
		index = 2
	}
	if o.POS != "NOUN" && o.POS != "ADJF" && o.POS != "PRTF" {
		return nil // []string{}
	}

	var grammemes []string
	if o.POS == "NOUN" && (o.Case != "nomn" && o.Case != "accs") {
		if index == 0 {
			grammemes = []string{"sing", o.Case}
		} else {
			grammemes = []string{"plur", o.Case}
		}
	} else if index == 0 {
		if o.Case == "nomn" {
			grammemes = o.tagClass.numeralAgreementGrammemes[0]
		} else {
			grammemes = o.tagClass.numeralAgreementGrammemes[1]
		}
	} else if o.POS == "NOUN" && index == 1 {
		grammemes = o.tagClass.numeralAgreementGrammemes[2]
	} else if (o.POS == "ADJF" || o.POS == "PRTF") && o.Gender == "femn" && index == 1 {
		grammemes = o.tagClass.numeralAgreementGrammemes[3]
	} else {
		grammemes = o.tagClass.numeralAgreementGrammemes[4]
	}

	return grammemes
}

func (o *opencorporaTag) grammemes() map[string]bool {
	o.cacheMu.RLock()
	cache := o.grammemesCache
	if cache != nil {
		grammemes := make(map[string]bool, len(cache))
		maps.Copy(grammemes, cache)
		o.cacheMu.RUnlock()
		return grammemes
	}
	o.cacheMu.RUnlock()

	o.cacheMu.Lock()
	if o.grammemesCache == nil {
		// Кэш — неизменяемая копия tags. Каждый вызов возвращает свою копию,
		// чтобы эвристические анализаторы не меняли состояние Parse и не
		// сталкивались друг с другом при работе из разных горутин.
		o.grammemesCache = make(map[string]bool, len(o.tags))
		maps.Copy(o.grammemesCache, o.tags)
	}
	grammemes := make(map[string]bool, len(o.grammemesCache))
	maps.Copy(grammemes, o.grammemesCache)
	o.cacheMu.Unlock()
	return grammemes
}

// Return a new set of grammemes with “required“ grammemes added and incompatible grammemes removed.
func (o *opencorporaTag) updatedGrammemes(required map[string]bool) (map[string]bool, error) {
	newGrammemes := make(map[string]bool, len(o.grammemes()))
	maps.Copy(newGrammemes, o.grammemes())

	for grammeme := range required {
		if !o.grammemeIsKnown(grammeme) {
			return nil, fmt.Errorf("Unknown grammeme: %#v", grammeme)
		}

		if !newGrammemes[grammeme] {
			newGrammemes[grammeme] = true

			// несовместимые грамемы, напр.
			// "sing": {"plur"},
			// "plur": {"ms-f", "femn", "sing", "GNdr", "masc", "neut"},
			for _, g := range o.grammemeIncompatible[grammeme] {
				if newGrammemes[g] {
					delete(newGrammemes, g) // newGrammemes[g] = false
				}
			}
		}
	}

	return newGrammemes, nil
}

func (o *opencorporaTag) grammemeIsKnown(grammeme string) bool {
	_, ok := o.tagClass.knownGrammemes[grammeme]
	return ok
}

func (o *opencorporaTag) isUnknown() bool {
	return !o.tagClass.partsOfSpeech[o.POS]
}

func (o *opencorporaTag) isProductive() bool {
	return !(intersectionLen(o.grammemes(), o.tagClass.nonProductiveGrammemes) != 0)
}
