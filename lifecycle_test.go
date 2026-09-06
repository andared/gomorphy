package gomorphy

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestGetMorphInstanceIsIndependentForDifferentPaths(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "opencorpora")
	if _, err := GetMorphInstance(missingPath); err == nil {
		t.Fatal("ожидалась ошибка для отсутствующего словаря")
	}

	morph, err := GetMorphInstance("./opencorpora")
	if err != nil {
		t.Fatalf("словарь после ошибочной загрузки не открылся: %v", err)
	}

	cached, err := GetMorphInstance("./opencorpora")
	if err != nil {
		t.Fatalf("повторная загрузка словаря завершилась ошибкой: %v", err)
	}
	if morph != cached {
		t.Fatal("один и тот же путь должен возвращать кэшированный экземпляр")
	}
}

func TestParseKeepsAnalyzerOwner(t *testing.T) {
	first, err := NewMorphAnalyzer("./opencorpora")
	if err != nil {
		t.Fatalf("не удалось создать первый анализатор: %v", err)
	}
	second, err := NewMorphAnalyzer("./opencorpora")
	if err != nil {
		t.Fatalf("не удалось создать второй анализатор: %v", err)
	}

	firstParse := first.Parse("дом")[0]
	secondParse := second.Parse("дом")[0]
	if firstParse.morph != first || secondParse.morph != second {
		t.Fatal("разбор должен ссылаться на создавший его анализатор")
	}

	firstInflected, err := firstParse.InflectVar("plur", "gent")
	if err != nil || firstInflected == nil {
		t.Fatalf("не удалось склонить слово первым анализатором: %v", err)
	}
	secondInflected, err := secondParse.InflectVar("plur", "gent")
	if err != nil || secondInflected == nil {
		t.Fatalf("не удалось склонить слово вторым анализатором: %v", err)
	}
	if firstInflected.Word != secondInflected.Word {
		t.Fatalf("разные анализаторы дали разные формы: %q и %q", firstInflected.Word, secondInflected.Word)
	}
}

func TestConcurrentInflectIsRaceFree(t *testing.T) {
	morph, err := NewMorphAnalyzer("./opencorpora")
	if err != nil {
		t.Fatalf("не удалось создать анализатор: %v", err)
	}
	parse := morph.Parse("дом")[0]

	const workers = 16
	const iterations = 100
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for range iterations {
				inflected := parse.MakeAgreeWithNumber(5)
				if inflected == nil || inflected.Word != "домов" {
					t.Errorf("неожиданная форма: %#v", inflected)
					return
				}
			}
		})
	}
	wg.Wait()
}
