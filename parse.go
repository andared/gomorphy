package gomorphy

import (
	"cmp"
	"slices"
	"strings"
)

type Method struct {
	Analyzer      analyzer // анализатор, интерфейс analyzer
	WordOrStack   any      // string или []Method (nil?)
	ParaIdOrStack any      // int, nil или []Method
	Idx           any      // int или nil, (Idx может отсутствовать, поэтому interface{} для сравнения с nil)
}

type Parse struct {
	Word         string
	Tag          *opencorporaTag
	NormalForm   string
	Score        float32
	MethodsStack []Method
	morph        *MorphAnalyzer
}

func newParse(word string, tag string, score float32) *Parse {
	wordLower := strings.ToLower(word)
	return &Parse{
		Word:         wordLower,
		Tag:          newOpencorporaTag(tag),
		NormalForm:   wordLower,
		Score:        score,
		MethodsStack: []Method{},
	}
}

// Parse анализирует слово и возвращает список вариантов разбора.
//
// В соответствии с pymorphy3, метод Parse всегда возвращает хотя бы один вариант разбора:
// если разбор не удался, то вместо пустого списка возвращается список
// с одним элементом с тегом UNKN (анализатор UnknAnalyzer).
//
// Для проверки успешности разбора, используйте метод morph.IsParsed(parses)
func (m *MorphAnalyzer) Parse(word string) []*Parse {
	res := make([]*Parse, 0, 25)

	wordLower := strings.ToLower(word)
	seen := make(map[seenParse]bool, 13)

	for _, analyzer := range m.analyzers {
		// parses := analyzer.parse(word, wordLower, seen)
		res = append(res, analyzer.parse(word, wordLower, seen)...)
		if analyzer.isTerminal() && len(res) != 0 {
			break
		}
	}

	// расчет score
	m.applyToParses(word, wordLower, res)
	m.bindParses(res)
	return res
}

// IsParsed возвращает true, если слово найдено в словаре или распознано с помощью эвристик.
// Иначе - false
func (*MorphAnalyzer) IsParsed(parses []*Parse) bool {
	return !parses[0].Tag.Contains("UNKN")
}

// MakeAgreeWithNumber склоненяет слово в соответствии с числом num ('дом'(5) -> 'домов')
func (p *Parse) MakeAgreeWithNumber(num int) *Parse {
	if p == nil {
		return nil
	}
	res, _ := p.Inflect(p.Tag.numeralAgreementGrammemes(num))
	return res
}

// Lexeme возвращает лексемы, принадлежащие этой форме
func (p *Parse) Lexeme() []*Parse {
	if p == nil || p.morph == nil {
		return nil
	}
	return p.morph.getLexeme(p)
}

// Normalized возвращает объект Parse нормальной формы слова.
func (p *Parse) Normalized() *Parse {
	if p == nil || len(p.MethodsStack) == 0 {
		return nil
	}
	normalized := p.MethodsStack[len(p.MethodsStack)-1].Analyzer.normalized(p)
	if normalized != nil {
		normalized.morph = p.morph
	}
	return normalized
}

// True if this form is a known dictionary form.
func (p *Parse) IsKnown() bool {
	if p == nil || p.morph == nil {
		return false
	}
	return p.morph.words.wordIsKnown(p.Word, p.morph.charSubstitutes)
}

func (m *MorphAnalyzer) applyToParses(word string, wordLower string, parses []*Parse) {
	probs := make([]float32, len(parses))
	for i, parse := range parses {
		probs[i] = m.scores.prob(wordLower, parse.Tag.RawTagsString)
	}

	var sumProbs float32
	for _, p := range probs {
		sumProbs += p
	}

	if sumProbs == 0 {
		// no P(t|w) information is available; return normalized estimate
		var sumScores float32
		for _, p := range parses {
			sumScores += p.Score
		}
		k := 1.0 / sumScores
		for _, p := range parses {
			p.Score = p.Score * k
		}
		return
	}

	// replace score with P(t|w) probability
	for i := range len(parses) {
		parses[i].Score = probs[i]
	}

	slices.SortStableFunc(parses, func(a, b *Parse) int {
		return -1 * cmp.Compare(a.Score, b.Score)
	})
}

func (m *MorphAnalyzer) Tag(word string) []*opencorporaTag {
	tags := make([]*opencorporaTag, 0, 10)
	seen := make(map[string]bool, 10)
	wordLower := strings.ToLower(word)

	for _, analyzer := range m.analyzers {
		tags = append(tags, analyzer.tag(word, wordLower, seen)...)
		if analyzer.isTerminal() && len(tags) != 0 {
			break
		}
	}
	// расчет score
	m.applyToTags(wordLower, tags)
	return tags
}

func (m *MorphAnalyzer) applyToTags(wordLower string, tags []*opencorporaTag) {
	if len(tags) == 0 {
		return
	}

	for i, tag := range tags {
		tags[i].prob = m.scores.prob(wordLower, tag.RawTagsString)
	}

	slices.SortStableFunc(tags, func(a, b *opencorporaTag) int {
		return -1 * cmp.Compare(a.prob, b.prob)
	})
}
