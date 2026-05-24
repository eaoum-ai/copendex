package search

import "github.com/eaoum-ai/copendex/internal/index"

type Service struct {
	store *index.Store
}

func New(store *index.Store) Service {
	return Service{store: store}
}

func (s Service) All(query string) ([]index.SearchResult, error) {
	return s.store.SearchAll(query)
}

func (s Service) Symbols(query string) ([]index.Symbol, error) {
	return s.store.SearchSymbols(query)
}
