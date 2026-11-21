package rombel

import "context"

type Store interface {
	List(context.Context, int, int, Filters) ([]ListItem, int64, error)
	GetDetail(context.Context, string) (Detail, error)
	ListAnggota(context.Context, string) ([]Anggota, error)
	ListPembelajaran(context.Context, string) ([]Pembelajaran, error)
	GetStatistics(context.Context, string) (Statistics, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) List(ctx context.Context, page, limit int, f Filters) ([]ListItem, int64, error) {
	return s.store.List(ctx, page, limit, f)
}
func (s *Service) GetDetail(ctx context.Context, id string) (Detail, error) {
	return s.store.GetDetail(ctx, id)
}
func (s *Service) ListAnggota(ctx context.Context, id string) ([]Anggota, error) {
	return s.store.ListAnggota(ctx, id)
}
func (s *Service) ListPembelajaran(ctx context.Context, id string) ([]Pembelajaran, error) {
	return s.store.ListPembelajaran(ctx, id)
}
func (s *Service) GetStatistics(ctx context.Context, id string) (Statistics, error) {
	return s.store.GetStatistics(ctx, id)
}
