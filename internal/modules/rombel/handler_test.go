package rombel

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	items        []ListItem
	total        int64
	detail       Detail
	anggota      AnggotaResponse
	pembelajaran []Pembelajaran
	statistics   Statistics
	err          error
	page, limit  int
	filters      Filters
	id           string
}

func (f *fakeStore) List(_ context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	f.page, f.limit, f.filters = page, limit, filters
	return f.items, f.total, f.err
}
func (f *fakeStore) GetDetail(_ context.Context, id string) (Detail, error) {
	f.id = id
	return f.detail, f.err
}
func (f *fakeStore) ListAnggota(_ context.Context, id string) (AnggotaResponse, error) {
	f.id = id
	return f.anggota, f.err
}
func (f *fakeStore) ListPembelajaran(_ context.Context, id string) ([]Pembelajaran, error) {
	f.id = id
	return f.pembelajaran, f.err
}
func (f *fakeStore) GetStatistics(_ context.Context, id string) (Statistics, error) {
	f.id = id
	return f.statistics, f.err
}
func testServer(store Store) http.Handler {
	mux := http.NewServeMux()
	NewHandler(NewService(store)).RegisterRoutes(mux)
	return mux
}

func TestListPassesFiltersAndPagination(t *testing.T) {
	store := &fakeStore{items: []ListItem{{ID: "a1b2c3d4-e5f6-4711-8223-aabbccddeeff", Nama: "Kelas Contoh"}}, total: 1}
	path := "/api/rombel?sekolah_id=47fe2769-4599-4434-9426-36a7ee7a0cd5&semester_id=20251&tingkat_pendidikan_id=10&jenis_rombel=1&kurikulum_id=2&jurusan_sp_id=05105671-45b1-499b-abdf-bb82dc607bcf&ptk_id=a1b2c3d4-e5f6-4711-8223-aabbccddeeff&search=Contoh&page=2&limit=5"
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if store.page != 2 || store.limit != 5 || store.filters.SekolahID == "" || store.filters.SemesterID != "20251" || store.filters.TingkatPendidikanID != "10" || store.filters.JenisRombel != "1" || store.filters.KurikulumID != "2" || store.filters.JurusanSPID == "" || store.filters.PTKID == "" || store.filters.Search != "Contoh" {
		t.Fatalf("unexpected store arguments: %+v", store)
	}
}

func TestListWithoutLimitReturnsAll(t *testing.T) {
	store := &fakeStore{items: []ListItem{{ID: "a1b2c3d4-e5f6-4711-8223-aabbccddeeff", Nama: "Kelas Contoh"}}, total: 134}
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel?page=2", nil))
	if w.Code != http.StatusOK || store.page != 1 || store.limit != 0 {
		t.Fatalf("status=%d page=%d limit=%d body=%s", w.Code, store.page, store.limit, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"limit":134`) || !strings.Contains(w.Body.String(), `"total_pages":1`) {
		t.Fatalf("unexpected all-results metadata: %s", w.Body.String())
	}
}

func TestListRejectsInvalidFilters(t *testing.T) {
	for _, query := range []string{"?sekolah_id=nope", "?jurusan_sp_id=nope", "?ptk_id=nope", "?tingkat_pendidikan_id=x", "?jenis_rombel=0", "?kurikulum_id=abc"} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			testServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel"+query, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestRombelChildRoutesValidateIDsAndHideErrors(t *testing.T) {
	const id = "a1b2c3d4-e5f6-4711-8223-aabbccddeeff"
	for _, suffix := range []string{"", "/anggota", "/pembelajaran", "/statistik"} {
		t.Run(suffix, func(t *testing.T) {
			store := &fakeStore{}
			w := httptest.NewRecorder()
			testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/"+id+suffix, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	t.Run("not found", func(t *testing.T) {
		store := &fakeStore{err: ErrNotFound}
		w := httptest.NewRecorder()
		testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/"+id, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status=%d", w.Code)
		}
	})
	t.Run("hidden error", func(t *testing.T) {
		store := &fakeStore{err: errors.New("private db detail")}
		w := httptest.NewRecorder()
		testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/"+id+"/anggota", nil))
		if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "private db detail") {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		testServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/not-uuid/statistik", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func TestAnggotaResponseIncludesClassTeacherAndMembers(t *testing.T) {
	const id = "a1b2c3d4-e5f6-4711-8223-aabbccddeeff"
	store := &fakeStore{anggota: AnggotaResponse{RombelID: id, NamaKelas: "X IPA 1", WaliKelas: stringPtr("Guru Contoh"), Anggota: []Anggota{{PesertaDidikID: "b1b2c3d4-e5f6-4711-8223-aabbccddeeff", Nama: "Siswa Contoh"}}}}
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/"+id+"/anggota", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	for _, field := range []string{`"nama_kelas":"X IPA 1"`, `"wali_kelas":"Guru Contoh"`, `"anggota":[{`, `"nama":"Siswa Contoh"`} {
		if !strings.Contains(w.Body.String(), field) {
			t.Errorf("response missing %s: %s", field, w.Body.String())
		}
	}
}

func TestAnggotaResponseUsesEmptyArray(t *testing.T) {
	const id = "a1b2c3d4-e5f6-4711-8223-aabbccddeeff"
	store := &fakeStore{anggota: AnggotaResponse{RombelID: id, NamaKelas: "X IPA 1", Anggota: []Anggota{}}}
	w := httptest.NewRecorder()
	testServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rombel/"+id+"/anggota", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"anggota":[]`) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func stringPtr(value string) *string { return &value }
