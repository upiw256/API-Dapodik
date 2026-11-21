package referensi

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("referensi tidak ditemukan")

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListSemesters(ctx context.Context) ([]Semester, error) {
	return s.store.ListSemesters(ctx)
}

func (s *Service) ListTahunAjaran(ctx context.Context) ([]TahunAjaran, error) {
	return s.store.ListTahunAjaran(ctx)
}

func (s *Service) GetWilayahHierarchy(ctx context.Context, kode string) (WilayahHierarchy, error) {
	return s.store.GetWilayahHierarchy(ctx, kode)
}
