package gomorphy

import (
	"fmt"
	"math"
)

// Inflect возвращает объект Parse склонения слова. Если форма слова не найдена, возвращает nil.
//
// Возвращает ошибку в случае некорректного тега.
//
// Есть вспомогательный метод InflectVar() с вариадическими аргументами.
func (p *Parse) Inflect(requiredGrammemes []string) (*Parse, error) {
	if p == nil {
		return nil, fmt.Errorf("parse равен nil")
	}
	if p.morph == nil {
		return nil, fmt.Errorf("parse не связан с MorphAnalyzer")
	}
	if p.Tag.Contains("UNKN") {
		return nil, nil
	}
	reqGrammemesSet := make(map[string]bool, len(requiredGrammemes))
	for _, grammeme := range requiredGrammemes {
		if !p.Tag.grammemeIsKnown(grammeme) {
			return nil, fmt.Errorf("Unknown grammeme: %#v", grammeme)
		}
		reqGrammemesSet[grammeme] = true
	}
	return p.morph.inflect(p, reqGrammemesSet)
}

// тоже самое что и Inflect, только с вариадическими аргументами.
//
// напр., parse[0].InflectVar("tag1", "tag2")
func (p *Parse) InflectVar(requiredGrammemes ...string) (*Parse, error) {
	return p.Inflect(requiredGrammemes)
}

// Return the lexeme this parse belongs to.
func (m *MorphAnalyzer) getLexeme(form *Parse) []*Parse {
	lastMethod := form.MethodsStack[len(form.MethodsStack)-1]
	lexeme := lastMethod.Analyzer.getLexeme(form)
	m.bindParses(lexeme)
	return lexeme
}

func (m *MorphAnalyzer) inflect(form *Parse, requiredGrammemes map[string]bool) (*Parse, error) {
	var possibleResults []*Parse
	for _, f := range m.getLexeme(form) {
		if isSubset(requiredGrammemes, f.Tag.grammemes()) {
			possibleResults = append(possibleResults, f)
		}
	}
	if possibleResults == nil {
		requiredGrammemes = form.Tag.tagClass.fixRareCases(requiredGrammemes)
		for _, f := range m.getLexeme(form) {
			if isSubset(requiredGrammemes, f.Tag.grammemes()) {
				possibleResults = append(possibleResults, f)
			}
		}
	}

	if possibleResults == nil {
		return nil, nil
	}

	grammemes, err := form.Tag.updatedGrammemes(requiredGrammemes)
	if err != nil {
		return nil, err
	}

	return getLargest(possibleResults, grammemes), nil
}

func getLargest(parses []*Parse, grammemes map[string]bool) *Parse {
	largestIdx := 0
	maxKey := math.Inf(-1)

	for i, form := range parses {
		intersectLen := intersectionLen(grammemes, form.Tag.grammemes())
		symDiffLen := symmetricDifferenceLen(grammemes, form.Tag.grammemes())
		if key := float64(intersectLen) - 0.1*float64(symDiffLen); key > maxKey {
			maxKey = key
			largestIdx = i
		}
	}

	return parses[largestIdx]
}
