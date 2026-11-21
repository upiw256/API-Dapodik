package pesertadidik

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	items       []ListItem
	total       int64
	detail      Detail
	listErr     error
	detailErr   error
	rombel      []Rombel
	rombelErr   error
	prestasi    PrestasiBeasiswa
	prestasiErr error
	page, limit int
	filters     Filters
	id          string
}

func (f *fakeStore) List(_ context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	f.page, f.limit, f.filters = page, limit, filters
	return f.items, f.total, f.listErr
}
func (f *fakeStore) GetDetail(_ context.Context, id string) (Detail, error) {
	f.id = id
	return f.detail, f.detailErr
}
func (f *fakeStore) ListRombel(_ context.Context, id string) ([]Rombel, error) {
	f.id = id
	return f.rombel, f.rombelErr
}
func (f *fakeStore) GetPrestasiBeasiswa(_ context.Context, id string) (PrestasiBeasiswa, error) {
	f.id = id
	return f.prestasi, f.prestasiErr
}

func testServer(store Store) http.Handler {
	mux := http.NewServeMux()
	NewHandler(NewService(store)).RegisterRoutes(mux)
	return mux
}

func TestListParsesFiltersAndProtectsSensitiveFields(t *testing.T) {
	store := &fakeStore{items: []ListItem{{PesertaDidikID: "a1b2c3d4-e5f6-4711-8223-aabbccddeeff", Nama: "Siswa Contoh"}}, total: 1}
	path := "/api/peserta-didik?sekolah_id=47fe2769-4599-4434-9426-36a7ee7a0cd5&rombongan_belajar_id=05105671-45b1-499b-abdf-bb82dc607bcf&semester_id=20251&tingkat_pendidikan_id=10&aktif=true&search=Contoh&page=2&limit=5"
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if store.page != 2 || store.limit != 5 || store.filters.SekolahID == "" || store.filters.SemesterID != "20251" || store.filters.TingkatPendidikanID != "10" || store.filters.Aktif == nil || !*store.filters.Aktif || store.filters.Search != "Contoh" {
		t.Fatalf("unexpected arguments: %+v", store)
	}
	for _, forbidden := range []string{`"nik"`, `"no_kk"`, `"nik_ayah"`, `"nik_ibu"`, `"nik_wali"`, `"rekening_bank"`, `"no_kps"`, `"no_kip"`, `"no_kks"`} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Errorf("response contains forbidden field %s", forbidden)
		}
	}
}

func TestListWithoutPageOrLimitReturnsAllResults(t *testing.T) {
	store := &fakeStore{items: []ListItem{{PesertaDidikID: "a1b2c3d4-e5f6-4711-8223-aabbccddeeff", Nama: "Siswa Contoh"}}, total: 1342}
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if store.page != 1 || store.limit != 0 {
		t.Fatalf("default list should be unpaginated, got page=%d limit=%d", store.page, store.limit)
	}
	if !strings.Contains(w.Body.String(), `"total":1342`) || !strings.Contains(w.Body.String(), `"limit":1342`) || !strings.Contains(w.Body.String(), `"total_pages":1`) {
		t.Fatalf("unexpected unpaginated metadata: %s", w.Body.String())
	}
}

func TestListRejectsInvalidFilters(t *testing.T) {
	for _, query := range []string{"?sekolah_id=invalid", "?rombongan_belajar_id=invalid", "?tingkat_pendidikan_id=x", "?aktif=maybe"} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			testServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik"+query, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestStudentRoutesValidateIDsAndHandleErrors(t *testing.T) {
	const id = "a1b2c3d4-e5f6-4711-8223-aabbccddeeff"
	t.Run("detail not found", func(t *testing.T) {
		store := &fakeStore{detailErr: ErrNotFound}
		w := httptest.NewRecorder()
		testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik/"+id, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status=%d", w.Code)
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		testServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik/nope/rombel", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	})
	for _, suffix := range []string{"/rombel", "/prestasi-beasiswa"} {
		t.Run(suffix, func(t *testing.T) {
			w := httptest.NewRecorder()
			testServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik/"+id+suffix, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	t.Run("store error hidden", func(t *testing.T) {
		store := &fakeStore{listErr: errors.New("private db detail")}
		w := httptest.NewRecorder()
		testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/peserta-didik", nil))
		if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "private db detail") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	})
}
