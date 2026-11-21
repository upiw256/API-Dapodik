package sekolah

import "context"

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListSchools(ctx context.Context, page, limit int, npsn string) ([]Profile, int64, error) {
	return s.store.ListSchools(ctx, page, limit, npsn)
}

func (s *Service) GetSchool(ctx context.Context, id string) (Profile, error) {
	return s.store.GetSchool(ctx, id)
}

func (s *Service) GetLongitudinal(ctx context.Context, id, semesterID string) ([]Longitudinal, error) {
	return s.store.GetLongitudinal(ctx, id, semesterID)
}

func (s *Service) ListJurusan(ctx context.Context, id string) ([]Jurusan, error) {
	return s.store.ListJurusan(ctx, id)
}

func (s *Service) ListRombel(ctx context.Context, id, semesterID string) ([]Rombel, error) {
	return s.store.ListRombel(ctx, id, semesterID)
}

func (s *Service) ListRooms(ctx context.Context, id string) ([]Room, error) {
	return s.store.ListRooms(ctx, id)
}
