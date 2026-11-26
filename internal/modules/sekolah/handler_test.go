package sekolah

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
	schools         []Profile
	total           int64
	profile         Profile
	profileErr      error
	longitudinal    []Longitudinal
	longitudinalErr error
	jurusan         []Jurusan
	jurusanErr      error
	rombel          []Rombel
	rombelErr       error
	rooms           []Room
	roomsErr        error
	gotPage         int
	gotLimit        int
	gotNPSN         string
	gotID           string
	gotSemester     string
}

func (f *fakeStore) ListSchools(_ context.Context, page, limit int, npsn string) ([]Profile, int64, error) {
	f.gotPage, f.gotLimit = page, limit
	f.gotNPSN = npsn
	return f.schools, f.total, nil
}

func (f *fakeStore) GetConfiguredSchoolID(context.Context) (string, error) {
	return "efb28aef-f65a-429e-8bb9-b9da7cba17d0", nil
}

func (f *fakeStore) GetSchool(_ context.Context, id string) (Profile, error) {
	f.gotID = id
	return f.profile, f.profileErr
}

func (f *fakeStore) GetLongitudinal(_ context.Context, id, semester string) ([]Longitudinal, error) {
	f.gotID, f.gotSemester = id, semester
	return f.longitudinal, f.longitudinalErr
}

func (f *fakeStore) ListJurusan(_ context.Context, id string) ([]Jurusan, error) {
	f.gotID = id
	return f.jurusan, f.jurusanErr
}

func (f *fakeStore) ListRombel(_ context.Context, id, semester string) ([]Rombel, error) {
	f.gotID, f.gotSemester = id, semester
	return f.rombel, f.rombelErr
}

func (f *fakeStore) ListRooms(_ context.Context, id string) ([]Room, error) {
	f.gotID = id
	return f.rooms, f.roomsErr
}

func newTestServer(store Store) http.Handler {
	mux := http.NewServeMux()
	NewHandler(NewService(store)).RegisterRoutes(mux)
	return mux
}

func TestListSchools(t *testing.T) {
	store := &fakeStore{
		schools: []Profile{{SekolahID: "efb28aef-f65a-429e-8bb9-b9da7cba17d0", Nama: "Sekolah Contoh"}},
		total:   21,
	}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah?page=2&limit=5&npsn=20227907", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got struct {
		Success bool      `json:"success"`
		Data    []Profile `json:"data"`
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
	if !got.Success || len(got.Data) != 1 || got.Meta.Page != 2 || got.Meta.Limit != 5 || got.Meta.Total != 21 || got.Meta.TotalPages != 5 {
		t.Fatalf("unexpected response: %+v", got)
	}
	if store.gotPage != 2 || store.gotLimit != 5 || store.gotNPSN != "20227907" {
		t.Fatalf("store received page=%d limit=%d npsn=%q, want page=2 limit=5 npsn=20227907", store.gotPage, store.gotLimit, store.gotNPSN)
	}
	for _, blockedField := range []string{`"nik"`, `"no_kk"`, `"npwp"`, `"no_rekening"`} {
		if strings.Contains(w.Body.String(), blockedField) {
			t.Fatalf("response contains prohibited field %s", blockedField)
		}
	}
}

func TestListSchoolsWithoutLimitReturnsAll(t *testing.T) {
	store := &fakeStore{schools: []Profile{{Nama: "Sekolah Contoh"}}, total: 125}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah?page=3", nil))
	if w.Code != http.StatusOK || store.gotPage != 1 || store.gotLimit != 0 {
		t.Fatalf("status=%d page=%d limit=%d body=%s", w.Code, store.gotPage, store.gotLimit, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"limit":125`) || !strings.Contains(w.Body.String(), `"total_pages":1`) {
		t.Fatalf("unexpected all-results metadata: %s", w.Body.String())
	}
}

func TestListSchoolsRejectsInvalidPaginationWhenLimitIsSet(t *testing.T) {
	w := httptest.NewRecorder()
	newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah?limit=10&page=0", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetSchool(t *testing.T) {
	id := "efb28aef-f65a-429e-8bb9-b9da7cba17d0"
	store := &fakeStore{profile: Profile{SekolahID: id, Nama: "Sekolah Contoh"}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah/detail", nil))
	if w.Code != http.StatusOK || store.gotID != id {
		t.Fatalf("status=%d id passed to store=%q body=%s", w.Code, store.gotID, w.Body.String())
	}
}

func TestGetSchoolNotFoundAndInvalidID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		store := &fakeStore{profileErr: ErrNotFound}
		w := httptest.NewRecorder()
		newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah/detail", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
	t.Run("UUID route removed", func(t *testing.T) {
		w := httptest.NewRecorder()
		newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah/not-a-uuid", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestGetLongitudinalPassesSemesterFilter(t *testing.T) {
	id := "efb28aef-f65a-429e-8bb9-b9da7cba17d0"
	store := &fakeStore{longitudinal: []Longitudinal{{SekolahID: id, SemesterID: "20251", Listrik: []Listrik{}, Internet: []Internet{}}}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah/longitudinal?semester_id=20251", nil))
	if w.Code != http.StatusOK || store.gotID != id || store.gotSemester != "20251" {
		t.Fatalf("status=%d id=%q semester=%q body=%s", w.Code, store.gotID, store.gotSemester, w.Body.String())
	}
}

func TestSchoolCollectionEndpoints(t *testing.T) {
	id := "efb28aef-f65a-429e-8bb9-b9da7cba17d0"
	tests := []struct {
		name string
		path string
		call func(*fakeStore) bool
	}{
		{name: "jurusan", path: "/api/sekolah/jurusan", call: func(s *fakeStore) bool { return s.gotID == id }},
		{name: "rombel", path: "/api/sekolah/rombel?semester_id=20251", call: func(s *fakeStore) bool { return s.gotID == id && s.gotSemester == "20251" }},
		{name: "ruang", path: "/api/sekolah/ruang", call: func(s *fakeStore) bool { return s.gotID == id }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{}
			w := httptest.NewRecorder()
			newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, test.path, nil))
			if w.Code != http.StatusOK || !test.call(store) {
				t.Fatalf("status=%d store=%+v body=%s", w.Code, store, w.Body.String())
			}
		})
	}
}

func TestSchoolStoreErrorsAreHidden(t *testing.T) {
	store := &fakeStore{longitudinalErr: errors.New("database connection details")}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sekolah/longitudinal", nil))
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "database connection details") {
		t.Fatalf("unexpected response status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSchoolRoutesRejectNonGETMethods(t *testing.T) {
	id := "efb28aef-f65a-429e-8bb9-b9da7cba17d0"
	paths := []string{
		"/api/sekolah",
		"/api/sekolah/detail",
		"/api/sekolah/longitudinal",
		"/api/sekolah/jurusan",
		"/api/sekolah/rombel",
		"/api/sekolah/ruang",
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
}
