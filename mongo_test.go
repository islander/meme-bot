package main

import "testing"

func TestFindOne(t *testing.T) {
	card, err := findOneCard("чандра факел")
	if err != nil {
		t.Error(err)
	}
	if card.Name != "Чандра, Факел Непокорности" &&
		card.Printing != "KLD" {
		t.Errorf("Expected Чандра, Факел Непокорности [KLD], got %s [%s]",
			card.Name, card.Printing)
	}
}

func TestFindNoDocuments(t *testing.T) {
	_, err := findOneCard("катлетка")
	if err.Error() != "mongo: no documents in result" {
		t.Error(err)
	}
}

func TestListCards(t *testing.T) {
	cards := searchCards("находка")
	if len(cards) < 1 {
		t.Error("Expected some items, got zero.")
	}
}

func TestEmptySearch(t *testing.T) {
	cards := searchCards("катлетка")
	if len(cards) > 1 {
		t.Error("Expected 0 items, got ", len(cards))
	}
}
