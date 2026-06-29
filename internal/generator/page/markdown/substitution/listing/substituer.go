package listing

import "strings"

type (
	Substituer[T Printer] struct {
		placeHolder   string
		lister        Lister[T]
		itemSeparator string
	}

	Printer interface {
		Print() string
	}

	Lister[T Printer] interface {
		ListPrinters() ([]T, error)
	}
)

func NewSubstituer[T Printer](placeHolder string, lister Lister[T], itemSeparator string) *Substituer[T] {
	return &Substituer[T]{
		placeHolder:   placeHolder,
		itemSeparator: itemSeparator,
		lister:        lister,
	}
}

func (s *Substituer[T]) Placeholder() string {
	return s.placeHolder
}

func (s *Substituer[T]) Resolve() (string, error) {
	printers, err := s.lister.ListPrinters()
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, printer := range printers {
		result.WriteString(printer.Print())
		result.WriteString(s.itemSeparator)
	}

	return result.String(), nil
}
