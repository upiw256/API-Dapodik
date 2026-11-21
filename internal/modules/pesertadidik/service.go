package pesertadidik

import "context"

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) List(ctx context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	return s.store.List(ctx, page, limit, filters)
}
func (s *Service) GetDetail(ctx context.Context, id string) (Detail, error) {
	return s.store.GetDetail(ctx, id)
}
func (s *Service) ListRombel(ctx context.Context, id string) ([]Rombel, error) {
	return s.store.ListRombel(ctx, id)
}
func (s *Service) GetPrestasiBeasiswa(ctx context.Context, id string) (PrestasiBeasiswa, error) {
	return s.store.GetPrestasiBeasiswa(ctx, id)
}
