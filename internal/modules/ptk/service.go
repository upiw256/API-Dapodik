package ptk

import "context"

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	return s.store.List(ctx, page, limit, filters)
}

func (s *Service) GetDetail(ctx context.Context, id string) (Detail, error) {
	return s.store.GetDetail(ctx, id)
}

func (s *Service) ListPembelajaran(ctx context.Context, id, semesterID string) ([]Pembelajaran, error) {
	return s.store.ListPembelajaran(ctx, id, semesterID)
}

func (s *Service) ListTugasTambahan(ctx context.Context, id string) ([]TugasTambahan, error) {
	return s.store.ListTugasTambahan(ctx, id)
}

func (s *Service) GetRiwayat(ctx context.Context, id string) (Riwayat, error) {
	return s.store.GetRiwayat(ctx, id)
}
