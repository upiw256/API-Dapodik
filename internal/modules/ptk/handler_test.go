package ptk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	items           []ListItem
	total           int64
	listErr         error
	detail          Detail
	detailErr       error
	pembelajaran    []Pembelajaran
	pembelajaranErr error
	tugasTambahan   []TugasTambahan
	tugasErr        error
	riwayat         Riwayat
	riwayatErr      error
	gotPage         int
	gotLimit        int
	gotFilters      Filters
	gotID           string
	gotSemester     string
}

func (f *fakeStore) List(_ context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	f.gotPage, f.gotLimit, f.gotFilters = page, limit, filters
	return f.items, f.total, f.listErr
}

func (f *fakeStore) GetDetail(_ context.Context, id string) (Detail, error) {
	f.gotID = id
	return f.detail, f.detailErr
}

func (f *fakeStore) ListPembelajaran(_ context.Context, id, semester string) ([]Pembelajaran, error) {
	f.gotID, f.gotSemester = id, semester
	return f.pembelajaran, f.pembelajaranErr
}

func (f *fakeStore) ListTugasTambahan(_ context.Context, id string) ([]TugasTambahan, error) {
	f.gotID = id
	return f.tugasTambahan, f.tugasErr
}

func (f *fakeStore) GetRiwayat(_ context.Context, id string) (Riwayat, error) {
	f.gotID = id
	return f.riwayat, f.riwayatErr
}

func newTestServer(store Store) http.Handler {
	mux := http.NewServeMux()
	NewHandler(NewService(store)).RegisterRoutes(mux)
	return mux
}

func stringPointer(value string) *string {
	return &value
}

func TestListPassesFiltersAndPagination(t *testing.T) {
	schoolID := "47fe2769-4599-4434-9426-36a7ee7a0cd5"
	store := &fakeStore{
		items: []ListItem{{PTKID: "05105671-45b1-499b-abdf-bb82dc607bcf", Nama: "Guru Contoh", NIK: "test-nik", NoKK: stringPointer("test-kk"), NPWP: stringPointer("test-npwp"), RekeningBank: stringPointer("test-rekening"), RekeningAtasNama: stringPointer("Pemilik Test"), Karpeg: stringPointer("test-karpeg"), Karpas: stringPointer("test-karpas")}},
		total: 21,
	}
	w := httptest.NewRecorder()
	path := "/api/ptk?sekolah_id=" + schoolID + "&tahun_ajaran_id=2025&jenis_ptk_id=1&search=Guru&page=2&limit=10"
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got struct {
		Success bool       `json:"success"`
		Data    []ListItem `json:"data"`
		Meta    struct {
			Page       int64 `json:"page"`
			Limit      int64 `json:"limit"`
			Total      int64 `json:"total"`
			TotalPages int64 `json:"total_pages"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Success || len(got.Data) != 1 || got.Meta.Page != 2 || got.Meta.Limit != 10 || got.Meta.Total != 21 || got.Meta.TotalPages != 3 {
		t.Fatalf("unexpected response: %+v", got)
	}
	if store.gotPage != 2 || store.gotLimit != 10 || store.gotFilters.SekolahID != schoolID || store.gotFilters.TahunAjaranID != "2025" || store.gotFilters.JenisPTKID != "1" || store.gotFilters.Search != "Guru" {
		t.Fatalf("unexpected store arguments: %+v", store)
	}
	for _, requiredField := range []string{`"nik":"test-nik"`, `"no_kk":"test-kk"`, `"npwp":"test-npwp"`, `"rekening_bank":"test-rekening"`, `"rekening_atas_nama":"Pemilik Test"`, `"karpeg":"test-karpeg"`, `"karpas":"test-karpas"`} {
		if !strings.Contains(w.Body.String(), requiredField) {
			t.Fatalf("response is missing required field %s", requiredField)
		}
	}
}

func TestListWithoutLimitReturnsAll(t *testing.T) {
	store := &fakeStore{items: []ListItem{{PTKID: "05105671-45b1-499b-abdf-bb82dc607bcf", Nama: "Guru Contoh"}}, total: 1200}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk?page=4", nil))
	if w.Code != http.StatusOK || store.gotPage != 1 || store.gotLimit != 0 {
		t.Fatalf("status=%d page=%d limit=%d body=%s", w.Code, store.gotPage, store.gotLimit, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"limit":1200`) || !strings.Contains(w.Body.String(), `"total_pages":1`) {
		t.Fatalf("unexpected all-results metadata: %s", w.Body.String())
	}
}

func TestListRejectsInvalidFilters(t *testing.T) {
	for _, query := range []string{
		"?sekolah_id=not-a-uuid",
		"?tahun_ajaran_id=abc",
		"?jenis_ptk_id=0",
	} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk"+query, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetDetailAndNotFound(t *testing.T) {
	id := "05105671-45b1-499b-abdf-bb82dc607bcf"
	t.Run("success", func(t *testing.T) {
		store := &fakeStore{detail: Detail{PTKID: id, Nama: "Guru Contoh"}}
		w := httptest.NewRecorder()
		newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk/"+id, nil))
		if w.Code != http.StatusOK || store.gotID != id {
			t.Fatalf("status=%d id=%q body=%s", w.Code, store.gotID, w.Body.String())
		}
	})
	t.Run("not found", func(t *testing.T) {
		store := &fakeStore{detailErr: ErrNotFound}
		w := httptest.NewRecorder()
		newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk/"+id, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
	t.Run("invalid UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk/not-a-uuid", nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestPembelajaranPassesSemesterFilter(t *testing.T) {
	id := "05105671-45b1-499b-abdf-bb82dc607bcf"
	store := &fakeStore{pembelajaran: []Pembelajaran{{ID: "example", SemesterID: "20251"}}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk/"+id+"/pembelajaran?semester_id=20251", nil))
	if w.Code != http.StatusOK || store.gotID != id || store.gotSemester != "20251" {
		t.Fatalf("status=%d id=%q semester=%q body=%s", w.Code, store.gotID, store.gotSemester, w.Body.String())
	}
}

func TestPTKCollectionRoutes(t *testing.T) {
	id := "05105671-45b1-499b-abdf-bb82dc607bcf"
	for _, path := range []string{"/api/ptk/" + id + "/tugas-tambahan", "/api/ptk/" + id + "/riwayat"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
			}
		})
	}
}

func TestStoreErrorsAreHiddenAndRoutesRejectNonGET(t *testing.T) {
	t.Run("hidden store error", func(t *testing.T) {
		store := &fakeStore{pembelajaranErr: errors.New("database connection details")}
		w := httptest.NewRecorder()
		newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ptk/05105671-45b1-499b-abdf-bb82dc607bcf/pembelajaran", nil))
		if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "database connection details") {
			t.Fatalf("unexpected response status=%d body=%s", w.Code, w.Body.String())
		}
	})
	t.Run("non-GET methods", func(t *testing.T) {
		paths := []string{
			"/api/ptk",
			"/api/ptk/05105671-45b1-499b-abdf-bb82dc607bcf",
			"/api/ptk/05105671-45b1-499b-abdf-bb82dc607bcf/pembelajaran",
			"/api/ptk/05105671-45b1-499b-abdf-bb82dc607bcf/tugas-tambahan",
			"/api/ptk/05105671-45b1-499b-abdf-bb82dc607bcf/riwayat",
		}
		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				w := httptest.NewRecorder()
				newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, nil))
				if w.Code != http.StatusMethodNotAllowed {
					t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
				}
			})
		}
	})
}
