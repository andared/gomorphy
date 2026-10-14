package gomorphy

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestDictionaryAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "DictionaryAnalyzer"

	dictAnalyzer := morph.analyzers[0]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// `положившим`
			// Parse(word='положившим', tag=OpencorporaTag('PRTF,perf,tran,past,actv masc,sing,ablt'), normal_form='положить', score=0.3333333333333333, methods_stack=((DictionaryAnalyzer(), 'положившим', 668, 20),))
			// Parse(word='положившим', tag=OpencorporaTag('PRTF,perf,tran,past,actv neut,sing,ablt'), normal_form='положить', score=0.3333333333333333, methods_stack=((DictionaryAnalyzer(), 'положившим', 668, 33),))
			// Parse(word='положившим', tag=OpencorporaTag('PRTF,perf,tran,past,actv plur,datv'), normal_form='положить', score=0.3333333333333333, methods_stack=((DictionaryAnalyzer(), 'положившим', 668, 37),))
			// normalized
			// Parse(word='положить', tag=OpencorporaTag('INFN,perf,tran'), normal_form='положить', score=1.0, methods_stack=((DictionaryAnalyzer(), 'положить', 668, 0),))
			// inflect accs,femn,sing
			// Parse(word='положившую', tag=OpencorporaTag('PRTF,perf,tran,past,actv femn,sing,accs'), normal_form='положить', score=1.0, methods_stack=((DictionaryAnalyzer(), 'положившую', 668, 25),))

			word: "положившим",
			expParses: []Parse{
				{Word: "положившим", Tag: newOpencorporaTag("PRTF,perf,tran,past,actv masc,sing,ablt"), NormalForm: "положить", Score: 0.3333333333333333, MethodsStack: []Method{{dictAnalyzer, "положившим", 668, 20}}},
				{Word: "положившим", Tag: newOpencorporaTag("PRTF,perf,tran,past,actv neut,sing,ablt"), NormalForm: "положить", Score: 0.3333333333333333, MethodsStack: []Method{{dictAnalyzer, "положившим", 668, 33}}},
				{Word: "положившим", Tag: newOpencorporaTag("PRTF,perf,tran,past,actv plur,datv"), NormalForm: "положить", Score: 0.3333333333333333, MethodsStack: []Method{{dictAnalyzer, "положившим", 668, 37}}},
			},
			expNormalized: Parse{Word: "положить", Tag: newOpencorporaTag("INFN,perf,tran"), NormalForm: "положить", Score: 1.0, MethodsStack: []Method{{dictAnalyzer, "положить", 668, 0}}},
			inflectTag:    []string{"accs", "femn", "sing"},
			expInflect:    Parse{Word: "положившую", Tag: newOpencorporaTag("PRTF,perf,tran,past,actv femn,sing,accs"), NormalForm: "положить", Score: 1.0, MethodsStack: []Method{{dictAnalyzer, "положившую", 668, 25}}},
		},
		{
			// `слезу`
			// Parse(word='слезу', tag=OpencorporaTag('NOUN,inan,femn sing,accs'), normal_form='слеза', score=0.75, methods_stack=((DictionaryAnalyzer(), 'слезу', 2896, 3),))
			// Parse(word='слезу', tag=OpencorporaTag('VERB,perf,intr sing,1per,futr,indc'), normal_form='слезть', score=0.25, methods_stack=((DictionaryAnalyzer(), 'слезу', 790, 5),))
			// normalized
			// Parse(word='слеза', tag=OpencorporaTag('NOUN,inan,femn sing,nomn'), normal_form='слеза', score=1.0, methods_stack=((DictionaryAnalyzer(), 'слеза', 2896, 0),))
			// inflect gen1,femn,plur
			// Parse(word='слёз', tag=OpencorporaTag('NOUN,inan,femn plur,gent'), normal_form='слеза', score=1.0, methods_stack=((DictionaryAnalyzer(), 'слёз', 2896, 8),))

			word: "слезу",
			expParses: []Parse{
				{Word: "слезу", Tag: newOpencorporaTag("NOUN,inan,femn sing,accs"), NormalForm: "слеза", Score: 0.75, MethodsStack: []Method{{dictAnalyzer, "слезу", 2896, 3}}},
				{Word: "слезу", Tag: newOpencorporaTag("VERB,perf,intr sing,1per,futr,indc"), NormalForm: "слезть", Score: 0.25, MethodsStack: []Method{{dictAnalyzer, "слезу", 790, 5}}},
			},
			expNormalized: Parse{Word: "слеза", Tag: newOpencorporaTag("NOUN,inan,femn sing,nomn"), NormalForm: "слеза", Score: 1.0, MethodsStack: []Method{{dictAnalyzer, "слеза", 2896, 0}}},
			inflectTag:    []string{"gen1", "femn", "plur"},
			expInflect:    Parse{Word: "слёз", Tag: newOpencorporaTag("NOUN,inan,femn plur,gent"), NormalForm: "слеза", Score: 1.0, MethodsStack: []Method{{dictAnalyzer, "слёз", 2896, 8}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestNumberAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "NumberAnalyzer"

	numbAnalyzer := morph.analyzers[3]
	unkAnalyzer := morph.analyzers[13]

	testCases := []struct {
		word      string
		expParses []Parse
	}{
		{
			// 1234
			// Parse(word='1234', tag=OpencorporaTag('NUMB,intg'), normal_form='1234', score=1.0, methods_stack=((NumberAnalyzer(score=0.9), '1234'),))

			word: "1234",
			expParses: []Parse{
				{Word: "1234", Tag: newOpencorporaTag("NUMB,intg"), NormalForm: "1234", Score: 1.0, MethodsStack: []Method{{Analyzer: numbAnalyzer, WordOrStack: "1234"}}},
			},
		},
		{
			// 12_34
			// Parse(word='12_34', tag=OpencorporaTag('NUMB,intg'), normal_form='12_34', score=1.0, methods_stack=((NumberAnalyzer(score=0.9), '12_34'),))

			word: "12_34",
			expParses: []Parse{
				{Word: "12_34", Tag: newOpencorporaTag("NUMB,intg"), NormalForm: "12_34", Score: 1.0, MethodsStack: []Method{{Analyzer: numbAnalyzer, WordOrStack: "12_34"}}},
			},
		},
		{
			// -1234
			// Parse(word='-1234', tag=OpencorporaTag('NUMB,intg'), normal_form='-1234', score=1.0, methods_stack=((NumberAnalyzer(score=0.9), '-1234'),))

			word: "-1234",
			expParses: []Parse{
				{Word: "-1234", Tag: newOpencorporaTag("NUMB,intg"), NormalForm: "-1234", Score: 1.0, MethodsStack: []Method{{Analyzer: numbAnalyzer, WordOrStack: "-1234"}}},
			},
		},
		{
			// 12.34
			// Parse(word='12.34', tag=OpencorporaTag('NUMB,real'), normal_form='12.34', score=1.0, methods_stack=((NumberAnalyzer(score=0.9), '12.34'),))

			word: "12.34",
			expParses: []Parse{
				{Word: "12.34", Tag: newOpencorporaTag("NUMB,real"), NormalForm: "12.34", Score: 1.0, MethodsStack: []Method{{Analyzer: numbAnalyzer, WordOrStack: "12.34"}}},
			},
		},
		{
			// -123e4
			// Parse(word='-123e4', tag=OpencorporaTag('NUMB,real'), normal_form='-123e4', score=1.0, methods_stack=((NumberAnalyzer(score=0.9), '-123e4'),))

			word: "-123e4",
			expParses: []Parse{
				{Word: "-123e4", Tag: newOpencorporaTag("NUMB,real"), NormalForm: "-123e4", Score: 1.0, MethodsStack: []Method{{Analyzer: numbAnalyzer, WordOrStack: "-123e4"}}},
			},
		},
		{
			// _123_4
			// Parse(word='_123_4', tag=OpencorporaTag('UNKN'), normal_form='_123_4', score=1.0, methods_stack=((UnknAnalyzer(), '_123_4'),))

			word: "_123_4",
			expParses: []Parse{
				{Word: "_123_4", Tag: newOpencorporaTag("UNKN"), NormalForm: "_123_4", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: "_123_4"}}},
			},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}
	}
}

func TestPunctuationAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "PunctuationAnalyzer"

	punctAnalyzer := morph.analyzers[4]

	testCases := []struct {
		word      string
		expParses []Parse
	}{
		{
			// .
			// Parse(word='.', tag=OpencorporaTag('PNCT'), normal_form='.', score=1.0, methods_stack=((PunctuationAnalyzer(score=0.9), '.'),))

			word: ".",
			expParses: []Parse{
				{Word: ".", Tag: newOpencorporaTag("PNCT"), NormalForm: ".", Score: 1.0, MethodsStack: []Method{{Analyzer: punctAnalyzer, WordOrStack: "."}}},
			},
		},
		{
			// ,
			// Parse(word=',', tag=OpencorporaTag('PNCT'), normal_form=',', score=1.0, methods_stack=((PunctuationAnalyzer(score=0.9), ','),))

			word: ",",
			expParses: []Parse{
				{Word: ",", Tag: newOpencorporaTag("PNCT"), NormalForm: ",", Score: 1.0, MethodsStack: []Method{{Analyzer: punctAnalyzer, WordOrStack: ","}}},
			},
		},
		{
			// !?
			// Parse(word='!?', tag=OpencorporaTag('PNCT'), normal_form='!?', score=1.0, methods_stack=((PunctuationAnalyzer(score=0.9), '!?'),))

			word: "!?",
			expParses: []Parse{
				{Word: "!?", Tag: newOpencorporaTag("PNCT"), NormalForm: "!?", Score: 1.0, MethodsStack: []Method{{Analyzer: punctAnalyzer, WordOrStack: "!?"}}},
			},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}
	}
}

func TestRomanAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "RomanNumberAnalyzer"

	romanNumbAnalyzer := morph.analyzers[5]
	latnAnalyzer := morph.analyzers[6]

	testCases := []struct {
		word      string
		expParses []Parse
	}{
		{
			// VII
			// Parse(word='vii', tag=OpencorporaTag('ROMN'), normal_form='vii', score=0.966666, methods_stack=((RomanNumberAnalyzer(score=0.9), 'VII'),))
			// Parse(word='vii', tag=OpencorporaTag('LATN'), normal_form='vii', score=0.033333, methods_stack=((LatinAnalyzer(score=0.9), 'VII'),))
			word: "VII",
			expParses: []Parse{
				{Word: "vii", Tag: newOpencorporaTag("ROMN"), NormalForm: "vii", Score: 0.966666, MethodsStack: []Method{{Analyzer: romanNumbAnalyzer, WordOrStack: "VII"}}},
				{Word: "vii", Tag: newOpencorporaTag("LATN"), NormalForm: "vii", Score: 0.033333, MethodsStack: []Method{{Analyzer: latnAnalyzer, WordOrStack: "VII"}}},
			},
		},
		{
			// MCXXI
			// Parse(word='mcxxi', tag=OpencorporaTag('ROMN'), normal_form='mcxxi', score=0.5, methods_stack=((RomanNumberAnalyzer(score=0.9), 'MCXXI'),))
			// Parse(word='mcxxi', tag=OpencorporaTag('LATN'), normal_form='mcxxi', score=0.5, methods_stack=((LatinAnalyzer(score=0.9), 'MCXXI'),))

			word: "MCXXI",
			expParses: []Parse{
				{Word: "mcxxi", Tag: newOpencorporaTag("ROMN"), NormalForm: "mcxxi", Score: 0.5, MethodsStack: []Method{{Analyzer: romanNumbAnalyzer, WordOrStack: "MCXXI"}}},
				{Word: "mcxxi", Tag: newOpencorporaTag("LATN"), NormalForm: "mcxxi", Score: 0.5, MethodsStack: []Method{{Analyzer: latnAnalyzer, WordOrStack: "MCXXI"}}},
			},
		},
		{
			// IIX
			// Parse(word='iix', tag=OpencorporaTag('LATN'), normal_form='iix', score=1.0, methods_stack=((LatinAnalyzer(score=0.9), 'IIX'),))
			word: "IIX",
			expParses: []Parse{
				{Word: "iix", Tag: newOpencorporaTag("LATN"), NormalForm: "iix", Score: 1.0, MethodsStack: []Method{{Analyzer: latnAnalyzer, WordOrStack: "IIX"}}},
			},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}
	}
}

func TestUnknAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "UnknAnalyzer"

	unkAnalyzer := morph.analyzers[13]

	testCases := []struct {
		word      string
		expParses []Parse
	}{
		{
			word: "",
			expParses: []Parse{
				{Word: "", Tag: newOpencorporaTag("UNKN"), NormalForm: "", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: ""}}},
			},
		},
		{
			// ' '
			// Parse(word=' ', tag=OpencorporaTag('UNKN'), normal_form=' ', score=1.0, methods_stack=((UnknAnalyzer(), ' '),))

			word: " ",
			expParses: []Parse{
				{Word: " ", Tag: newOpencorporaTag("UNKN"), NormalForm: " ", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: " "}}},
			},
		},
		{
			// абвabc
			// Parse(word='абвabc', tag=OpencorporaTag('UNKN'), normal_form='абвabc', score=1.0, methods_stack=((UnknAnalyzer(), 'абвabc'),))

			word: "абвabc",
			expParses: []Parse{
				{Word: "абвabc", Tag: newOpencorporaTag("UNKN"), NormalForm: "абвabc", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: "абвabc"}}},
			},
		},
		{
			// абв123
			// Parse(word='абв123', tag=OpencorporaTag('UNKN'), normal_form='абв123', score=1.0, methods_stack=((UnknAnalyzer(), 'абв123'),))

			word: "абв123",
			expParses: []Parse{
				{Word: "абв123", Tag: newOpencorporaTag("UNKN"), NormalForm: "абв123", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: "абв123"}}},
			},
		},
		{
			// 123..
			// Parse(word='123..', tag=OpencorporaTag('UNKN'), normal_form='123..', score=1.0, methods_stack=((UnknAnalyzer(), '123..'),))

			word: "123..",
			expParses: []Parse{
				{Word: "123..", Tag: newOpencorporaTag("UNKN"), NormalForm: "123..", Score: 1.0, MethodsStack: []Method{{Analyzer: unkAnalyzer, WordOrStack: "123.."}}},
			},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}
	}
}

func TestHyphenSeparatedParticleAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "HyphenSeparatedParticleAnalyzer"

	dictAnalyzer := morph.analyzers[0]
	hyphenSepParticleAnalyzer := morph.analyzers[7]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// бюджет-то
			// Parse(word='бюджет-то', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='бюджет-то', score=0.5, methods_stack=((DictionaryAnalyzer(), 'бюджет', 34, 0), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-то')))
			// Parse(word='бюджет-то', tag=OpencorporaTag('NOUN,inan,masc sing,accs'), normal_form='бюджет-то', score=0.5, methods_stack=((DictionaryAnalyzer(), 'бюджет', 34, 3), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-то')))
			// normalized
			// Parse(word='бюджет-то', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='бюджет-то', score=0.5, methods_stack=((DictionaryAnalyzer(), 'бюджет', 34, 0), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-то')))
			// inflect 'datv','plur'
			// Parse(word='бюджетам-то', tag=OpencorporaTag('NOUN,inan,masc plur,datv'), normal_form='бюджет-то', score=1.0, methods_stack=((DictionaryAnalyzer(), 'бюджетам', 34, 8), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-то')))

			word: "бюджет-то",
			expParses: []Parse{
				{Word: "бюджет-то", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "бюджет-то", Score: 0.5, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "бюджет", ParaIdOrStack: 34, Idx: 0}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-то"}}},
				{Word: "бюджет-то", Tag: newOpencorporaTag("NOUN,inan,masc sing,accs"), NormalForm: "бюджет-то", Score: 0.5, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "бюджет", ParaIdOrStack: 34, Idx: 3}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-то"}}},
			},
			expNormalized: Parse{Word: "бюджет-то", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "бюджет-то", Score: 0.5, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "бюджет", ParaIdOrStack: 34, Idx: 0}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-то"}}},
			inflectTag:    []string{"datv", "plur"},
			expInflect:    Parse{Word: "бюджетам-то", Tag: newOpencorporaTag("NOUN,inan,masc plur,datv"), NormalForm: "бюджет-то", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "бюджетам", ParaIdOrStack: 34, Idx: 8}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-то"}}},
		},
		{
			// пойдем-ка
			// Parse(word='пойдём-ка', tag=OpencorporaTag('VERB,perf,intr plur,1per,futr,indc'), normal_form='пойти-ка', score=0.5, methods_stack=((DictionaryAnalyzer(), 'пойдём', 2533, 6), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-ка')))
			// Parse(word='пойдём-ка', tag=OpencorporaTag('VERB,perf,intr sing,impr,incl'), normal_form='пойти-ка', score=0.5, methods_stack=((DictionaryAnalyzer(), 'пойдём', 2533, 11), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-ка')))
			// normalized
			// Parse(word='пойти-ка', tag=OpencorporaTag('INFN,perf,intr'), normal_form='пойти-ка', score=1.0, methods_stack=((DictionaryAnalyzer(), 'пойти', 2533, 0), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-ка')))
			// inflect 'datv','neut'
			// Parse(word='пошедшему-ка', tag=OpencorporaTag('PRTF,perf,intr,past,actv neut,sing,datv'), normal_form='пойти-ка', score=1.0, methods_stack=((DictionaryAnalyzer(), 'пошедшему', 2533, 35), (HyphenSeparatedParticleAnalyzer(particles_after_hyphen=['-то', '-ка', '-таки', '-де', '-тко', '-тка', '-с', '-ста'], score_multiplier=0.9), '-ка')))

			word: "пойдем-ка",
			expParses: []Parse{
				{Word: "пойдём-ка", Tag: newOpencorporaTag("VERB,perf,intr plur,1per,futr,indc"), NormalForm: "пойти-ка", Score: 0.5, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "пойдём", ParaIdOrStack: 2533, Idx: 6}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-ка"}}},
				{Word: "пойдём-ка", Tag: newOpencorporaTag("VERB,perf,intr sing,impr,incl"), NormalForm: "пойти-ка", Score: 0.5, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "пойдём", ParaIdOrStack: 2533, Idx: 11}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-ка"}}},
			},
			expNormalized: Parse{Word: "пойти-ка", Tag: newOpencorporaTag("INFN,perf,intr"), NormalForm: "пойти-ка", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "пойти", ParaIdOrStack: 2533, Idx: 0}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-ка"}}},
			inflectTag:    []string{"datv", "neut"},
			expInflect:    Parse{Word: "пошедшему-ка", Tag: newOpencorporaTag("PRTF,perf,intr,past,actv neut,sing,datv"), NormalForm: "пойти-ка", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "пошедшему", ParaIdOrStack: 2533, Idx: 35}, {Analyzer: hyphenSepParticleAnalyzer, WordOrStack: "-ка"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestHyphenAdverbAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "HyphenAdverbAnalyzer"

	hyphAdverbAnalyzer := morph.analyzers[8]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
	}{
		{
			// `по-серьезному`
			// Parse(word='по-серьезному', tag=OpencorporaTag('ADVB'), normal_form='по-серьезному', score=1.0, methods_stack=((HyphenAdverbAnalyzer(score_multiplier=0.7), 'по-серьезному'),))
			// normalized
			// Parse(word='по-серьезному', tag=OpencorporaTag('ADVB'), normal_form='по-серьезному', score=1.0, methods_stack=((HyphenAdverbAnalyzer(score_multiplier=0.7), 'по-серьезному'),))

			word: "по-серьезному",
			expParses: []Parse{
				{Word: "по-серьезному", Tag: newOpencorporaTag("ADVB"), NormalForm: "по-серьезному", Score: 1.0, MethodsStack: []Method{{Analyzer: hyphAdverbAnalyzer, WordOrStack: "по-серьезному"}}},
			},
			expNormalized: Parse{Word: "по-серьезному", Tag: newOpencorporaTag("ADVB"), NormalForm: "по-серьезному", Score: 1.0, MethodsStack: []Method{{Analyzer: hyphAdverbAnalyzer, WordOrStack: "по-серьезному"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}
	}
}

func TestHyphenatedWordsAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "HyphenatedWordsAnalyzer"

	dictAnalyzer := morph.analyzers[0]
	hyphenatedWordsAnal := morph.analyzers[9]
	knownPfxAnalyzer := morph.analyzers[10]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// `полуприцеп-контейнеровоз`
			// Parse(word='полуприцеп-контейнеровоз', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='полуприцеп-контейнеровоз', score=0.5, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), ((DictionaryAnalyzer(), 'прицеп', 34, 0),), ((DictionaryAnalyzer(), 'контейнеровоз', 34, 0),)), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'полу')))
			// Parse(word='полуприцеп-контейнеровоз', tag=OpencorporaTag('NOUN,inan,masc sing,accs'), normal_form='полуприцеп-контейнеровоз', score=0.5, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), ((DictionaryAnalyzer(), 'прицеп', 34, 3),), ((DictionaryAnalyzer(), 'контейнеровоз', 34, 3),)), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'полу')))
			// normalized
			// Parse(word='полуприцеп-контейнеровоз', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='полуприцеп-контейнеровоз', score=1.0, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), ((DictionaryAnalyzer(), 'прицеп', 34, 0),), ((DictionaryAnalyzer(), 'контейнеровоз', 34, 0),)), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'полу')))
			// inflect loc1,masc,sing
			// Parse(word='полуприцепе-контейнеровозе', tag=OpencorporaTag('NOUN,inan,masc sing,loct'), normal_form='полуприцеп-контейнеровоз', score=1.0, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), ((DictionaryAnalyzer(), 'прицепе', 34, 5),), ((DictionaryAnalyzer(), 'контейнеровозе', 34, 5),)), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'полу')))

			word: "полуприцеп-контейнеровоз",
			expParses: []Parse{
				{Word: "полуприцеп-контейнеровоз", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "полуприцеп-контейнеровоз", Score: 0.5, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "прицеп", ParaIdOrStack: 34, Idx: 0}}, ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "контейнеровоз", ParaIdOrStack: 34, Idx: 0}}}, {Analyzer: knownPfxAnalyzer, WordOrStack: "полу"}}},
				{Word: "полуприцеп-контейнеровоз", Tag: newOpencorporaTag("NOUN,inan,masc sing,accs"), NormalForm: "полуприцеп-контейнеровоз", Score: 0.5, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "прицеп", ParaIdOrStack: 34, Idx: 3}}, ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "контейнеровоз", ParaIdOrStack: 34, Idx: 3}}}, {Analyzer: knownPfxAnalyzer, WordOrStack: "полу"}}},
			},
			expNormalized: Parse{Word: "полуприцеп-контейнеровоз", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "полуприцеп-контейнеровоз", Score: 0.5, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "прицеп", ParaIdOrStack: 34, Idx: 0}}, ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "контейнеровоз", ParaIdOrStack: 34, Idx: 0}}}, {Analyzer: knownPfxAnalyzer, WordOrStack: "полу"}}},
			inflectTag:    []string{"loc1", "masc", "sing"},
			expInflect:    Parse{Word: "полуприцепе-контейнеровозе", Tag: newOpencorporaTag("NOUN,inan,masc sing,loct"), NormalForm: "полуприцеп-контейнеровоз", Score: 1.0, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "прицепе", ParaIdOrStack: 34, Idx: 5}}, ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "контейнеровозе", ParaIdOrStack: 34, Idx: 5}}}, {Analyzer: knownPfxAnalyzer, WordOrStack: "полу"}}},
		},
		{
			// `насосно-компрессорных`
			// Parse(word='насосно-компрессорных', tag=OpencorporaTag('ADJF plur,gent'), normal_form='насосно-компрессорный', score=0.3333333333333333, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), 'насосно', ((DictionaryAnalyzer(), 'компрессорных', 35, 21),)),))
			// Parse(word='насосно-компрессорных', tag=OpencorporaTag('ADJF anim,plur,accs'), normal_form='насосно-компрессорный', score=0.3333333333333333, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), 'насосно', ((DictionaryAnalyzer(), 'компрессорных', 35, 23),)),))
			// Parse(word='насосно-компрессорных', tag=OpencorporaTag('ADJF plur,loct'), normal_form='насосно-компрессорный', score=0.3333333333333333, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), 'насосно', ((DictionaryAnalyzer(), 'компрессорных', 35, 26),)),))
			// normalized
			// Parse(word='насосно-компрессорный', tag=OpencorporaTag('ADJF masc,sing,nomn'), normal_form='насосно-компрессорный', score=1.0, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), 'насосно', ((DictionaryAnalyzer(), 'компрессорный', 35, 0),)),))
			// inflect gent,femn,sing
			// Parse(word='насосно-компрессорной', tag=OpencorporaTag('ADJF femn,sing,gent'), normal_form='насосно-компрессорный', score=1.0, methods_stack=((HyphenatedWordsAnalyzer(score_multiplier=0.75, skip_prefixes=<...>), 'насосно', ((DictionaryAnalyzer(), 'компрессорной', 35, 8),)),))

			word: "насосно-компрессорных",
			expParses: []Parse{
				{Word: "насосно-компрессорных", Tag: newOpencorporaTag("ADJF plur,gent"), NormalForm: "насосно-компрессорный", Score: 0.3333333333333333, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: "насосно", ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "компрессорных", ParaIdOrStack: 35, Idx: 21}}}}},
				{Word: "насосно-компрессорных", Tag: newOpencorporaTag("ADJF anim,plur,accs"), NormalForm: "насосно-компрессорный", Score: 0.3333333333333333, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: "насосно", ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "компрессорных", ParaIdOrStack: 35, Idx: 23}}}}},
				{Word: "насосно-компрессорных", Tag: newOpencorporaTag("ADJF plur,loct"), NormalForm: "насосно-компрессорный", Score: 0.3333333333333333, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: "насосно", ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "компрессорных", ParaIdOrStack: 35, Idx: 26}}}}},
			},
			expNormalized: Parse{Word: "насосно-компрессорный", Tag: newOpencorporaTag("ADJF masc,sing,nomn"), NormalForm: "насосно-компрессорный", Score: 1.0, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: "насосно", ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "компрессорный", ParaIdOrStack: 35, Idx: 0}}}}},
			inflectTag:    []string{"gent", "femn", "sing"},
			expInflect:    Parse{Word: "насосно-компрессорной", Tag: newOpencorporaTag("ADJF femn,sing,gent"), NormalForm: "насосно-компрессорный", Score: 1.0, MethodsStack: []Method{{Analyzer: hyphenatedWordsAnal, WordOrStack: "насосно", ParaIdOrStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "компрессорной", ParaIdOrStack: 35, Idx: 8}}}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestKnownPrefixAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "KnownPrefixAnalyzer"

	dictAnalyzer := morph.analyzers[0]
	knownPfxAnalyzer := morph.analyzers[10]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// взаимопосещение
			// Parse(word='взаимопосещение', tag=OpencorporaTag('NOUN,inan,neut sing,nomn'), normal_form='взаимопосещение', score=0.7, methods_stack=((DictionaryAnalyzer(), 'посещение', 77, 0), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'взаимо')))
			// Parse(word='взаимопосещение', tag=OpencorporaTag('NOUN,inan,neut sing,accs'), normal_form='взаимопосещение', score=0.3, methods_stack=((DictionaryAnalyzer(), 'посещение', 77, 6), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'взаимо')))
			// normalized
			// Parse(word='взаимопосещение', tag=OpencorporaTag('NOUN,inan,neut sing,nomn'), normal_form='взаимопосещение', score=0.7, methods_stack=((DictionaryAnalyzer(), 'посещение', 77, 0), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'взаимо')))
			// inflect 'loc2','neut','plur'
			// Parse(word='взаимопосещениях', tag=OpencorporaTag('NOUN,inan,neut plur,loct'), normal_form='взаимопосещение', score=1.0, methods_stack=((DictionaryAnalyzer(), 'посещениях', 77, 22), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'взаимо')))

			word: "взаимопосещение",
			expParses: []Parse{
				{Word: "взаимопосещение", Tag: newOpencorporaTag("NOUN,inan,neut sing,nomn"), NormalForm: "взаимопосещение", Score: 0.7, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "посещение", ParaIdOrStack: 77, Idx: 0}, {Analyzer: knownPfxAnalyzer, WordOrStack: "взаимо"}}},
				{Word: "взаимопосещение", Tag: newOpencorporaTag("NOUN,inan,neut sing,accs"), NormalForm: "взаимопосещение", Score: 0.3, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "посещение", ParaIdOrStack: 77, Idx: 6}, {Analyzer: knownPfxAnalyzer, WordOrStack: "взаимо"}}},
			},
			expNormalized: Parse{Word: "взаимопосещение", Tag: newOpencorporaTag("NOUN,inan,neut sing,nomn"), NormalForm: "взаимопосещение", Score: 0.7, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "посещение", ParaIdOrStack: 77, Idx: 0}, {Analyzer: knownPfxAnalyzer, WordOrStack: "взаимо"}}},
			inflectTag:    []string{"loc2", "neut", "plur"},
			expInflect:    Parse{Word: "взаимопосещениях", Tag: newOpencorporaTag("NOUN,inan,neut plur,loct"), NormalForm: "взаимопосещение", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "посещениях", ParaIdOrStack: 77, Idx: 22}, {Analyzer: knownPfxAnalyzer, WordOrStack: "взаимо"}}},
		},
		{
			// постмодерна
			// Parse(word='постмодерна', tag=OpencorporaTag('NOUN,inan,masc sing,gent'), normal_form='постмодерн', score=1.0, methods_stack=((DictionaryAnalyzer(), 'модерна', 34, 1), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'пост')))
			// normalized
			// Parse(word='постмодерн', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='постмодерн', score=1.0, methods_stack=((DictionaryAnalyzer(), 'модерн', 34, 0), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'пост')))
			// inflect 'voct','masc','plur'
			// Parse(word='постмодерны', tag=OpencorporaTag('NOUN,inan,masc plur,nomn'), normal_form='постмодерн', score=1.0, methods_stack=((DictionaryAnalyzer(), 'модерны', 34, 6), (KnownPrefixAnalyzer(known_prefixes=<...>, min_remainder_length=3, score_multiplier=0.75), 'пост')))

			word: "постмодерна",
			expParses: []Parse{
				{Word: "постмодерна", Tag: newOpencorporaTag("NOUN,inan,masc sing,gent"), NormalForm: "постмодерн", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "модерна", ParaIdOrStack: 34, Idx: 1}, {Analyzer: knownPfxAnalyzer, WordOrStack: "пост"}}},
			},
			expNormalized: Parse{Word: "постмодерн", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "постмодерн", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "модерн", ParaIdOrStack: 34, Idx: 0}, {Analyzer: knownPfxAnalyzer, WordOrStack: "пост"}}},
			inflectTag:    []string{"voct", "masc", "plur"},
			expInflect:    Parse{Word: "постмодерны", Tag: newOpencorporaTag("NOUN,inan,masc plur,nomn"), NormalForm: "постмодерн", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "модерны", ParaIdOrStack: 34, Idx: 6}, {Analyzer: knownPfxAnalyzer, WordOrStack: "пост"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestUnknownPrefixAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "UnknownPrefixAnalyzer"

	dictAnalyzer := morph.analyzers[0]
	unknownPfxAnalyzer := morph.analyzers[11]
	knownSfxAnalyzer := morph.analyzers[12]
	fakeDict := knownSfxAnalyzer.(*knownSuffixAnalyzer).fakeDict

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// медколледжа
			// Parse(word='медколледжа', tag=OpencorporaTag('NOUN,inan,masc sing,gent'), normal_form='медколледж', score=1.0, methods_stack=((DictionaryAnalyzer(), 'колледжа', 80, 1), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'мед')))
			// normalized
			// Parse(word='медколледж', tag=OpencorporaTag('NOUN,inan,masc sing,nomn'), normal_form='медколледж', score=1.0, methods_stack=((DictionaryAnalyzer(), 'колледж', 80, 0), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'мед')))
			// inflect 'gent','masc','plur'
			// Parse(word='медколледжей', tag=OpencorporaTag('NOUN,inan,masc plur,gent'), normal_form='медколледж', score=1.0, methods_stack=((DictionaryAnalyzer(), 'колледжей', 80, 7), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'мед')))

			word: "медколледжа",
			expParses: []Parse{
				{Word: "медколледжа", Tag: newOpencorporaTag("NOUN,inan,masc sing,gent"), NormalForm: "медколледж", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "колледжа", ParaIdOrStack: 80, Idx: 1}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "мед"}}},
			},
			expNormalized: Parse{Word: "медколледж", Tag: newOpencorporaTag("NOUN,inan,masc sing,nomn"), NormalForm: "медколледж", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "колледж", ParaIdOrStack: 80, Idx: 0}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "мед"}}},
			inflectTag:    []string{"gent", "masc", "plur"},
			expInflect:    Parse{Word: "медколледжей", Tag: newOpencorporaTag("NOUN,inan,masc plur,gent"), NormalForm: "медколледж", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "колледжей", ParaIdOrStack: 80, Idx: 7}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "мед"}}},
		},
		{
			// шлицешлифовальные
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('ADJF plur,nomn'), normal_form='шлицешлифовальный', score=0.33337780149413027, methods_stack=((DictionaryAnalyzer(), 'шлифовальные', 60, 20), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'шлице')))
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('ADJF inan,plur,accs'), normal_form='шлицешлифовальный', score=0.33337780149413027, methods_stack=((DictionaryAnalyzer(), 'шлифовальные', 60, 24), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'шлице')))
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('ADJF plur,nomn'), normal_form='шлицешлифовальный', score=0.16542155816435436, methods_stack=((FakeDictionary(), 'шлицешлифовальные', 10, 20), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'льные')))
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('ADJF inan,plur,accs'), normal_form='шлицешлифовальный', score=0.16542155816435436, methods_stack=((FakeDictionary(), 'шлицешлифовальные', 10, 24), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'льные')))
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('NOUN,inan,femn plur,nomn'), normal_form='шлицешлифовальная', score=0.0012006403415154752, methods_stack=((FakeDictionary(), 'шлицешлифовальные', 117, 7), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'льные')))
			// Parse(word='шлицешлифовальные', tag=OpencorporaTag('NOUN,inan,femn plur,accs'), normal_form='шлицешлифовальная', score=0.0012006403415154752, methods_stack=((FakeDictionary(), 'шлицешлифовальные', 117, 10), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'льные')))
			// normalized
			// Parse(word='шлицешлифовальный', tag=OpencorporaTag('ADJF masc,sing,nomn'), normal_form='шлицешлифовальный', score=1.0, methods_stack=((DictionaryAnalyzer(), 'шлифовальный', 60, 0), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'шлице')))
			// inflect 'gent','femn','sing'
			// Parse(word='шлицешлифовальной', tag=OpencorporaTag('ADJF femn,sing,gent'), normal_form='шлицешлифовальный', score=1.0, methods_stack=((DictionaryAnalyzer(), 'шлифовальной', 60, 8), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'шлице')))

			word: "шлицешлифовальные",
			expParses: []Parse{
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("ADJF plur,nomn"), NormalForm: "шлицешлифовальный", Score: 0.33337780149413027, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "шлифовальные", ParaIdOrStack: 60, Idx: 20}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "шлице"}}},
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("ADJF inan,plur,accs"), NormalForm: "шлицешлифовальный", Score: 0.33337780149413027, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "шлифовальные", ParaIdOrStack: 60, Idx: 24}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "шлице"}}},
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("ADJF plur,nomn"), NormalForm: "шлицешлифовальный", Score: 0.16542155816435436, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "шлицешлифовальные", ParaIdOrStack: 10, Idx: 20}, {Analyzer: knownSfxAnalyzer, WordOrStack: "льные"}}},
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("ADJF inan,plur,accs"), NormalForm: "шлицешлифовальный", Score: 0.16542155816435436, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "шлицешлифовальные", ParaIdOrStack: 10, Idx: 24}, {Analyzer: knownSfxAnalyzer, WordOrStack: "льные"}}},
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("NOUN,inan,femn plur,nomn"), NormalForm: "шлицешлифовальная", Score: 0.0012006403415154752, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "шлицешлифовальные", ParaIdOrStack: 117, Idx: 7}, {Analyzer: knownSfxAnalyzer, WordOrStack: "льные"}}},
				{Word: "шлицешлифовальные", Tag: newOpencorporaTag("NOUN,inan,femn plur,accs"), NormalForm: "шлицешлифовальная", Score: 0.0012006403415154752, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "шлицешлифовальные", ParaIdOrStack: 117, Idx: 10}, {Analyzer: knownSfxAnalyzer, WordOrStack: "льные"}}},
			},
			expNormalized: Parse{Word: "шлицешлифовальный", Tag: newOpencorporaTag("ADJF masc,sing,nomn"), NormalForm: "шлицешлифовальный", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "шлифовальный", ParaIdOrStack: 60, Idx: 0}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "шлице"}}},
			inflectTag:    []string{"gent", "femn", "sing"},
			expInflect:    Parse{Word: "шлицешлифовальной", Tag: newOpencorporaTag("ADJF femn,sing,gent"), NormalForm: "шлицешлифовальный", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "шлифовальной", ParaIdOrStack: 60, Idx: 8}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "шлице"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestKnownSuffixAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "KnownSuffixAnalyzer"

	dictAnalyzer := morph.analyzers[0]
	unknownPfxAnalyzer := morph.analyzers[11]
	knownSfxAnalyzer := morph.analyzers[12]
	fakeDict := knownSfxAnalyzer.(*knownSuffixAnalyzer).fakeDict

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
		inflectTag    []string
		expInflect    Parse
	}{
		{
			// бутявкать
			// Parse(word='бутявкать', tag=OpencorporaTag('INFN,impf,intr'), normal_form='бутявкать', score=0.20454545454545453, methods_stack=((DictionaryAnalyzer(), 'тявкать', 15, 0), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бу')))
			// Parse(word='бутявкать', tag=OpencorporaTag('NOUN,anim,femn,Name sing,voct,Infr'), normal_form='бутявкатя', score=0.20454545454545453, methods_stack=((DictionaryAnalyzer(), 'кать', 208, 6), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бутяв')))
			// Parse(word='бутявкать', tag=OpencorporaTag('NOUN,anim,femn,Name plur,gent'), normal_form='бутявкатя', score=0.20454545454545453, methods_stack=((DictionaryAnalyzer(), 'кать', 208, 8), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бутяв')))
			// Parse(word='бутявкать', tag=OpencorporaTag('NOUN,anim,femn,Name plur,accs'), normal_form='бутявкатя', score=0.20454545454545453, methods_stack=((DictionaryAnalyzer(), 'кать', 208, 10), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бутяв')))
			// Parse(word='бутявкать', tag=OpencorporaTag('INFN,perf,intr'), normal_form='бутявкать', score=0.1818181818181818, methods_stack=((FakeDictionary(), 'бутявкать', 749, 0), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'вкать')))
			// normalized
			// Parse(word='бутявкать', tag=OpencorporaTag('INFN,impf,intr'), normal_form='бутявкать', score=0.20454545454545453, methods_stack=((DictionaryAnalyzer(), 'тявкать', 15, 0), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бу')))
			// inflect 'gent','neut','sing'
			// Parse(word='бутявкающего', tag=OpencorporaTag('PRTF,impf,intr,pres,actv neut,sing,gent'), normal_form='бутявкать', score=1.0, methods_stack=((DictionaryAnalyzer(), 'тявкающего', 15, 28), (UnknownPrefixAnalyzer(score_multiplier=0.5), 'бу')))

			word: "бутявкать",
			expParses: []Parse{
				{Word: "бутявкать", Tag: newOpencorporaTag("INFN,impf,intr"), NormalForm: "бутявкать", Score: 0.20454545454545453, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "тявкать", ParaIdOrStack: 15, Idx: 0}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бу"}}},
				{Word: "бутявкать", Tag: newOpencorporaTag("NOUN,anim,femn,Name sing,voct,Infr"), NormalForm: "бутявкатя", Score: 0.20454545454545453, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "кать", ParaIdOrStack: 208, Idx: 6}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бутяв"}}},
				{Word: "бутявкать", Tag: newOpencorporaTag("NOUN,anim,femn,Name plur,gent"), NormalForm: "бутявкатя", Score: 0.20454545454545453, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "кать", ParaIdOrStack: 208, Idx: 8}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бутяв"}}},
				{Word: "бутявкать", Tag: newOpencorporaTag("NOUN,anim,femn,Name plur,accs"), NormalForm: "бутявкатя", Score: 0.20454545454545453, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "кать", ParaIdOrStack: 208, Idx: 10}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бутяв"}}},
				{Word: "бутявкать", Tag: newOpencorporaTag("INFN,perf,intr"), NormalForm: "бутявкать", Score: 0.1818181818181818, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "бутявкать", ParaIdOrStack: 749, Idx: 0}, {Analyzer: knownSfxAnalyzer, WordOrStack: "вкать"}}},
			},
			expNormalized: Parse{Word: "бутявкать", Tag: newOpencorporaTag("INFN,impf,intr"), NormalForm: "бутявкать", Score: 0.20454545454545453, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "тявкать", ParaIdOrStack: 15, Idx: 0}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бу"}}},
			inflectTag:    []string{"gent", "neut", "sing"},
			expInflect:    Parse{Word: "бутявкающего", Tag: newOpencorporaTag("PRTF,impf,intr,pres,actv neut,sing,gent"), NormalForm: "бутявкать", Score: 1.0, MethodsStack: []Method{{Analyzer: dictAnalyzer, WordOrStack: "тявкающего", ParaIdOrStack: 15, Idx: 28}, {Analyzer: unknownPfxAnalyzer, WordOrStack: "бу"}}},
		},
		{
			// наружка
			// Parse(word='наружка', tag=OpencorporaTag('NOUN,inan,femn sing,nomn'), normal_form='наружка', score=0.5, methods_stack=((FakeDictionary(), 'наружка', 9, 0), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'ружка')))
			// Parse(word='наружка', tag=OpencorporaTag('NOUN,inan,masc sing,gent'), normal_form='наружок', score=0.5, methods_stack=((FakeDictionary(), 'наружка', 141, 1), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'ружка')))
			// normalized
			// Parse(word='наружка', tag=OpencorporaTag('NOUN,inan,femn sing,nomn'), normal_form='наружка', score=0.5, methods_stack=((FakeDictionary(), 'наружка', 9, 0), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'ружка')))
			// inflect 'datv','femn','plur'
			// Parse(word='наружкам', tag=OpencorporaTag('NOUN,inan,femn plur,datv'), normal_form='наружка', score=1.0, methods_stack=((FakeDictionary(), 'наружкам', 9, 9), (KnownSuffixAnalyzer(min_word_length=4, score_multiplier=0.5), 'ружка')))

			word: "наружка",
			expParses: []Parse{
				{Word: "наружка", Tag: newOpencorporaTag("NOUN,inan,femn sing,nomn"), NormalForm: "наружка", Score: 0.5, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "наружка", ParaIdOrStack: 9, Idx: 0}, {Analyzer: knownSfxAnalyzer, WordOrStack: "ружка"}}},
				{Word: "наружка", Tag: newOpencorporaTag("NOUN,inan,masc sing,gent"), NormalForm: "наружок", Score: 0.5, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "наружка", ParaIdOrStack: 141, Idx: 1}, {Analyzer: knownSfxAnalyzer, WordOrStack: "ружка"}}},
			},
			expNormalized: Parse{Word: "наружка", Tag: newOpencorporaTag("NOUN,inan,femn sing,nomn"), NormalForm: "наружка", Score: 0.5, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "наружка", ParaIdOrStack: 9, Idx: 0}, {Analyzer: knownSfxAnalyzer, WordOrStack: "ружка"}}},
			inflectTag:    []string{"datv", "femn", "plur"},
			expInflect:    Parse{Word: "наружкам", Tag: newOpencorporaTag("NOUN,inan,femn plur,datv"), NormalForm: "наружка", Score: 1.0, MethodsStack: []Method{{Analyzer: fakeDict, WordOrStack: "наружкам", ParaIdOrStack: 9, Idx: 9}, {Analyzer: knownSfxAnalyzer, WordOrStack: "ружка"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}

		// inflect
		inf, _ := parses[0].Inflect(tc.inflectTag)
		if inf == nil {
			t.Errorf("[%s] word: %#v - inflect is nil", aName, tc.word)
			continue
		}
		if !parsesIsEqual(inf, &tc.expInflect) {
			t.Errorf("[%s] word: %#v - inflect error", aName, tc.word)
		}
	}
}

func TestAbbrAnalyzer(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	aName := "AbbrAnalyzer"

	abbrFirstNameAnalyzer := morph.analyzers[1]
	abbrPatronymicAnalyzer := morph.analyzers[2]

	testCases := []struct {
		word          string
		expParses     []Parse
		expNormalized Parse
	}{
		{

			// 'Х'
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,nomn'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,gent'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,datv'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,accs'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,ablt'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,loct'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,nomn'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,gent'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,datv'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,accs'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,ablt'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,loct'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))

			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,nomn'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,gent'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,datv'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,accs'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,ablt'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,loct'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,nomn'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,gent'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,datv'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,accs'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,ablt'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,loct'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedPatronymicAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Patr,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))

			// normalized
			// Parse(word='х', tag=OpencorporaTag('NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,nomn'), normal_form='х', score=0.04166666666666666, methods_stack=((AbbreviatedFirstNameAnalyzer(letters='АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЭЮЯ', score=0.1, tag_pattern='NOUN,anim,%(gender)s,Sgtm,Name,Fixd,Abbr,Init sing,%(case)s'), 'Х'),))

			word: "Х",
			expParses: []Parse{
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,nomn"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,gent"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,datv"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,accs"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,ablt"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,loct"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,nomn"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,gent"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,datv"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,accs"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,ablt"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Name,Fixd,Abbr,Init sing,loct"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},

				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,nomn"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,gent"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,datv"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,accs"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,ablt"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Patr,Fixd,Abbr,Init sing,loct"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,nomn"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,gent"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,datv"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,accs"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,ablt"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
				{Word: "х", Tag: newOpencorporaTag("NOUN,anim,femn,Sgtm,Patr,Fixd,Abbr,Init sing,loct"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrPatronymicAnalyzer, WordOrStack: "Х"}}},
			},
			expNormalized: Parse{Word: "х", Tag: newOpencorporaTag("NOUN,anim,masc,Sgtm,Name,Fixd,Abbr,Init sing,nomn"), NormalForm: "х", Score: 0.04166666666666666, MethodsStack: []Method{{Analyzer: abbrFirstNameAnalyzer, WordOrStack: "Х"}}},
		},
	}

	for _, tc := range testCases {
		parses := morph.Parse(tc.word)
		if len(parses) != len(tc.expParses) {
			t.Errorf("[%s] word: %#v - parses length not equal", aName, tc.word)
			continue
		}
		for i, p := range parses {
			if !parsesIsEqual(p, &tc.expParses[i]) {
				t.Errorf("[%s] word: %#v - wrong parse", aName, tc.word)
			}
		}

		// normalized
		gotNormalized := parses[0].Normalized()
		if !parsesIsEqual(gotNormalized, &tc.expNormalized) {
			t.Errorf("[%s] word: %#v - wrong normalized", aName, tc.word)
		}
	}
}

// тест целостности исходного Parse после Inflect
func TestImmutableParse(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	parses := morph.Parse("кошка")

	type tParse struct {
		word     string
		tag      *opencorporaTag
		normForm string
		score    float32
		stack    string
	}

	exceptedParse := tParse{
		word:     parses[0].Word,
		tag:      newOpencorporaTag(parses[0].Tag.RawTagsString),
		normForm: parses[0].NormalForm,
		score:    parses[0].Score,
		stack:    fmt.Sprintf("%v", parses[0].MethodsStack),
	}

	parses[0].InflectVar("datv", "plur")

	testedParse := tParse{
		word:     parses[0].Word,
		tag:      parses[0].Tag,
		normForm: parses[0].NormalForm,
		score:    parses[0].Score,
		stack:    fmt.Sprintf("%v", parses[0].MethodsStack),
	}

	if !reflect.DeepEqual(testedParse, exceptedParse) {
		t.Errorf("[ImmutableParse] source parse was changed after inflect")
	}
}

func TestTags(t *testing.T) {
	testCases := []struct {
		word         string
		expectedTags []string
	}{
		{
			word: "беговой",
			expectedTags: []string{
				"NOUN,anim,femn,Sgtm,Surn sing,gent",
				"NOUN,anim,femn,Sgtm,Surn sing,datv",
				"NOUN,anim,femn,Sgtm,Surn sing,ablt",
				"NOUN,anim,femn,Sgtm,Surn sing,loct",
				"ADJF masc,sing,nomn",
				"ADJF inan,masc,sing,accs",
				"ADJF femn,sing,gent",
				"ADJF femn,sing,datv",
				"ADJF femn,sing,ablt",
				"ADJF femn,sing,loct",
			},
		},
		{
			word: "метро",
			expectedTags: []string{
				"NOUN,inan,neut,Fixd sing,loct",
				"NOUN,inan,neut,Fixd sing,gent",
				"NOUN,inan,neut,Fixd sing,nomn",
				"NOUN,inan,neut,Fixd sing,datv",
				"NOUN,inan,neut,Fixd sing,accs",
				"NOUN,inan,neut,Fixd sing,ablt",
				"NOUN,inan,neut,Fixd plur,nomn",
				"NOUN,inan,neut,Fixd plur,gent",
				"NOUN,inan,neut,Fixd plur,datv",
				"NOUN,inan,neut,Fixd plur,accs",
				"NOUN,inan,neut,Fixd plur,ablt",
				"NOUN,inan,neut,Fixd plur,loct",
			},
		},
		{
			word: "вооруженной",
			expectedTags: []string{
				"ADJF femn,sing,gent",
				"ADJF femn,sing,datv",
				"ADJF femn,sing,ablt",
				"ADJF femn,sing,loct",
				"PRTF,perf,tran,past,pssv femn,sing,gent",
				"PRTF,perf,tran,past,pssv femn,sing,datv",
				"PRTF,perf,tran,past,pssv femn,sing,ablt",
				"PRTF,perf,tran,past,pssv femn,sing,loct",
			},
		},
	}

	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	for _, tc := range testCases {
		tags := morph.Tag(tc.word)
		if len(tags) != len(tc.expectedTags) {
			t.Errorf("[tags] word: tags length not equal")
			continue
		}
		for i, tag := range tags {
			if tag.RawTagsString != tc.expectedTags[i] {
				t.Errorf("[tags] word: wrong tags")
			}
		}
	}
}

func TestAgreeWithNumber(t *testing.T) {

	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	testCases := []struct {
		num          int
		expectedWord string
	}{
		{-10, "полуприцепов-контейнеровозов"}, // оригинальные результаты pymorhy3
		{-5, "полуприцепов-контейнеровозов"},
		{-1, "полуприцепов-контейнеровозов"},
		{0, "полуприцепов-контейнеровозов"},
		{1, "полуприцеп-контейнеровоз"},
		{2, "полуприцепа-контейнеровоза"},
		{3, "полуприцепа-контейнеровоза"},
		{5, "полуприцепов-контейнеровозов"},
		{10, "полуприцепов-контейнеровозов"},
		{11, "полуприцепов-контейнеровозов"},
		{20, "полуприцепов-контейнеровозов"},
		{21, "полуприцеп-контейнеровоз"},
		{24, "полуприцепа-контейнеровоза"},
		{26, "полуприцепов-контейнеровозов"},
		{99, "полуприцепов-контейнеровозов"},
	}

	parses := morph.Parse("полуприцеп-контейнеровоз")
	for _, tc := range testCases {
		gotWord := parses[0].MakeAgreeWithNumber(tc.num).Word
		if gotWord != tc.expectedWord {
			t.Errorf("[agreeWithNumber] num: %d, got: %s, expected: %s", tc.num, gotWord, tc.expectedWord)
		}
	}

	testCases = []struct {
		num          int
		expectedWord string
	}{
		{-5, "полуприцепам-контейнеровозам"}, // оригинальные результаты pymorhy3
		{-1, "полуприцепам-контейнеровозам"},
		{0, "полуприцепам-контейнеровозам"},
		{1, "полуприцепу-контейнеровозу"},
		{2, "полуприцепам-контейнеровозам"},
		{10, "полуприцепам-контейнеровозам"},
		{21, "полуприцепу-контейнеровозу"},
		{99, "полуприцепам-контейнеровозам"},
	}

	parses = morph.Parse("полуприцепам-контейнеровозам")
	for _, tc := range testCases {
		gotWord := parses[0].MakeAgreeWithNumber(tc.num).Word
		if gotWord != tc.expectedWord {
			t.Errorf("[agreeWithNumber] num: %d, got: %s, expected: %s", tc.num, gotWord, tc.expectedWord)
		}
	}

	testCases = []struct {
		num          int
		expectedWord string
	}{
		{-10, "сосисок"}, // оригинальные результаты pymorhy3
		{-5, "сосисок"},
		{-1, "сосисок"},
		{0, "сосисок"},
		{1, "сосиска"},
		{2, "сосиски"},
		{3, "сосиски"},
		{6, "сосисок"},
		{10, "сосисок"},
		{11, "сосисок"},
		{20, "сосисок"},
		{21, "сосиска"},
		{22, "сосиски"},
		{24, "сосиски"},
		{25, "сосисок"},
		{99, "сосисок"},
		{241, "сосиска"},
	}

	parses = morph.Parse("сосиска")
	for _, tc := range testCases {
		gotWord := parses[0].MakeAgreeWithNumber(tc.num).Word
		if gotWord != tc.expectedWord {
			t.Errorf("[agreeWithNumber] num: %d, got: %s, expected: %s", tc.num, gotWord, tc.expectedWord)
		}
	}
}

func TestLat2Cyr(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	word := "кошка"
	expectedCyrTag := "СУЩ,од,жр ед,им"
	expectedCyrPOS := "СУЩ"

	parse := morph.Parse(word)[0]
	if got, _ := morph.Lat2Cyr(parse.Tag.RawTagsString); got != expectedCyrTag {
		t.Errorf("[Lat2Cyr] got: %#v, expected: %#v", got, expectedCyrTag)
	}

	if got, _ := morph.Lat2Cyr(parse.Tag.POS); got != expectedCyrPOS {
		t.Errorf("[Lat2Cyr] got: %#v, expected: %#v", got, expectedCyrPOS)
	}

	word = "зеленый"
	expectedCyrTag = "ПРИЛ,кач мр,ед,дт"
	expectedCyrCase := "дт"

	parse = morph.Parse(word)[0]
	inf, _ := parse.InflectVar("datv")

	if got, _ := morph.Lat2Cyr(inf.Tag.RawTagsString); got != expectedCyrTag {
		t.Errorf("[Lat2Cyr] got: %#v, expected: %#v", got, expectedCyrTag)
	}

	if got, _ := morph.Lat2Cyr(inf.Tag.Case); got != expectedCyrCase {
		t.Errorf("[Lat2Cyr] got: %#v, expected: %#v", got, expectedCyrCase)
	}
}

func TestWordIsKnown(t *testing.T) {
	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	word := "кошка"
	expWordIsKnown := true
	expWordIsKnownStrict := true
	expIsKnown := true
	if got := morph.WordIsKnown(word, false); got != expWordIsKnown {
		t.Errorf("[wordIsKnown] (strict=false) word: %#v, got: %t, expected: %t", word, got, expWordIsKnown)
	}
	if got := morph.WordIsKnown(word, true); got != expWordIsKnownStrict {
		t.Errorf("[wordIsKnown] (strict=true) word: %#v, got: %t, expected: %t", word, got, expWordIsKnownStrict)
	}
	if got := morph.Parse(word)[0].IsKnown(); got != expIsKnown {
		t.Errorf("[wordIsKnown] (isKnown) word: %#v, got: %t, expected: %t", word, got, expIsKnown)
	}

	word = "ежик"
	expWordIsKnown = true
	expWordIsKnownStrict = false
	expIsKnown = true
	if got := morph.WordIsKnown(word, false); got != expWordIsKnown {
		t.Errorf("[wordIsKnown] (strict=false) word: %#v, got: %t, expected: %t", word, got, expWordIsKnown)
	}
	if got := morph.WordIsKnown(word, true); got != expWordIsKnownStrict {
		t.Errorf("[wordIsKnown] (strict=true) word: %#v, got: %t, expected: %t", word, got, expWordIsKnownStrict)
	}
	if got := morph.Parse(word)[0].IsKnown(); got != expIsKnown {
		t.Errorf("[wordIsKnown] (isKnown) word: %#v, got: %t, expected: %t", word, got, expIsKnown)
	}

	word = "кудяплик"
	expWordIsKnown = false
	expWordIsKnownStrict = false
	expIsKnown = false
	if got := morph.WordIsKnown(word, false); got != expWordIsKnown {
		t.Errorf("[wordIsKnown] (strict=false) word: %#v, got: %t, expected: %t", word, got, expWordIsKnown)
	}
	if got := morph.WordIsKnown(word, true); got != expWordIsKnownStrict {
		t.Errorf("[wordIsKnown] (strict=true) word: %#v, got: %t, expected: %t", word, got, expWordIsKnownStrict)
	}
	if got := morph.Parse(word)[0].IsKnown(); got != expIsKnown {
		t.Errorf("[wordIsKnown] (isKnown) word: %#v, got: %t, expected: %t", word, got, expIsKnown)
	}
}

func parsesIsEqual(gotP, expP *Parse) bool {
	if gotP.Word != expP.Word {
		return false
	}

	if !reflect.DeepEqual(gotP.Tag, expP.Tag) {
		return false
	}

	if gotP.NormalForm != expP.NormalForm {
		return false
	}

	// score могут не совпадать из-за разного округления, приводим к общему виду
	gotPscore := float32(int(gotP.Score*1e6)) / 1e6
	expPscore := float32(int(gotP.Score*1e6)) / 1e6
	if gotPscore != expPscore {
		return false
	}

	if !reflect.DeepEqual(gotP.MethodsStack, expP.MethodsStack) {
		return false
	}

	return true
}
