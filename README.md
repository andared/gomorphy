# gomorphy
pymorphy3 port for golang

Частичный порт библиотеки python https://github.com/no-plagiarism/pymorphy3 на язык Golang

Реализована поддержка только русского языка (украинский в планах на будущее).

Частичный форк https://github.com/jus1d/gomorphy . (За основу взят код работы со словарем DAWG.)

Работа со словарем opencorpora в формате DAWG (словарь от 22.05.2026).

Реализованы все эвристические анализаторы.

Добавление пакета:
```
go get github.com/andared/gomorphy
```
Скопируйте каталог словаря opencorpora из пакета gomorphy в папку со своим проектом.

`GetMorphInstance` кэширует анализатор по пути к словарю. Если приложению нужны
независимые экземпляры (например, для нескольких версий словаря), создайте их
через `NewMorphAnalyzer`.

### Пример:
```
package main

import (
	"fmt"
	"log"

	"github.com/andared/gomorphy"
)

func main() {
	morph, err := gomorphy.GetMorphInstance("./opencorpora")
	if err != nil {
		log.Fatal(err)
	}

	word := "кошка"

	parses := morph.Parse(word)

	// if parses[0].Tag.Contains("UNKN")
	// или
	if !morph.IsParsed(parses) {
		fmt.Println("unknown word")
		return
	}

	parse := parses[0]

	fmt.Println("word:", parse.Word)
	fmt.Println("normal form:", parse.NormalForm)
	fmt.Println("raw tags string:", parse.Tag.RawTagsString)
	fmt.Println("tag.POS:", parse.Tag.POS)
	fmt.Println("tag.Case:", parse.Tag.Case)
	fmt.Println("tag.Gender:", parse.Tag.Gender)
	fmt.Println("tag.Number:", parse.Tag.Number)
	fmt.Println("tag.Animate:", parse.Tag.Animate)

	fmt.Println("tags contains `NOUN`:", parse.Tag.Contains("NOUN")) // true
	fmt.Println("tags contains `gent`:", parse.Tag.Contains("gent")) // false

	fmt.Println("prob score:", parse.Score)

	fmt.Println("methods stack:")
	for _, m := range parse.MethodsStack {
		fmt.Println("имя анализатора:", m.Analyzer.Name())
		fmt.Printf("%#v\n\n", m)
	}
	fmt.Println()

	// inf, err := parse.Inflect([]string{"gent", "plur"})
	// или
	inf, err := parse.InflectVar("gent", "plur")
	if err != nil {
		fmt.Println("некорректный тег:", err)
	} else if inf == nil {
		fmt.Println("склонение отсутствует")
	} else {
		fmt.Println("склонение слова род.п., мн.ч.:", inf.Word)
		fmt.Println("inflected normal form:", inf.NormalForm)
		fmt.Println("inflected raw tags string:", inf.Tag.RawTagsString)
		fmt.Println()
	}

	// другие методы

	// parse.Lexeme() // возвращает лексемы, принадлежащие этой форме
	// parse.Normalized() // возвращает объект Parse нормальной формы слова.

	fmt.Println("пять", parse.MakeAgreeWithNumber(5).Word) // пять кошек

	fmt.Println("словарное слово:", parse.IsKnown())
	fmt.Println("словарное слово:", morph.WordIsKnown(word, true))

	if cyr, err := morph.Lat2Cyr("gent"); err != nil {
		fmt.Println("некорректный тег:", err)
	} else {
		fmt.Println("кириллический тег `gent`:", cyr)
	}

	fmt.Printf("нормализованные формы слова: %#v\n", morph.NormalForms(word))

	// fmt.Printf("%#v\n", *morph.Tag(word)[0])
}
```
### Результат:
```
word: кошка
normal form: кошка
raw tags string: NOUN,anim,femn sing,nomn
tag.POS: NOUN
tag.Case: nomn
tag.Gender: femn
tag.Number: sing
tag.Animate: anim
tags contains `NOUN` true
tags contains `gent` false
prob score: 0.9375
methods stack:
имя анализатора: DictionaryAnalyzer
gomorphy.Method{Analyzer:(*gomorphy.dictionaryAnalyzer)(0x154945dbc6c0), WordOrStack:"кошка", ParaIdOrStack:134, Idx:0}


склонение слова род.п., мн.ч.: кошек
inflected normal form: кошка
inflected raw tags string: NOUN,anim,femn plur,gent

пять кошек
словарное слово: true
словарное слово: true
кириллический тег `gent`: рд
нормализованные формы слова: []string{"кошка"}
```

## License

The **Go source code** is licensed under the [MIT License](LICENSE).

The **embedded dictionary data** (`opencorpora/`) is derived from [OpenCorpora](http://opencorpora.org/)
and is licensed under [CC BY-SA 4.0](opencorpora/LICENSE). If you distribute a binary that embeds
this data, you must comply with the CC BY-SA 4.0 terms (attribution + ShareAlike).
