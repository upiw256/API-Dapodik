package referensi

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
	semesters   []Semester
	semesterErr error
	tahunAjaran []TahunAjaran
	tahunErr    error
	wilayah     WilayahHierarchy
	wilayahErr  error
	gotKode     string
}

func (f *fakeStore) ListSemesters(context.Context) ([]Semester, error) {
	return f.semesters, f.semesterErr
}

func (f *fakeStore) ListTahunAjaran(context.Context) ([]TahunAjaran, error) {
	return f.tahunAjaran, f.tahunErr
}

func (f *fakeStore) GetWilayahHierarchy(_ context.Context, kode string) (WilayahHierarchy, error) {
	f.gotKode = kode
	return f.wilayah, f.wilayahErr
}

func newTestServer(store Store) http.Handler {
	mux := http.NewServeMux()
	NewHandler(NewService(store)).RegisterRoutes(mux)
	return mux
}

func TestListSemesters(t *testing.T) {
	store := &fakeStore{semesters: []Semester{{
		SemesterID:     "20251",
		TahunAjaranID:  2025,
		Nama:           "2025/2026 Ganjil",
		Semester:       1,
		PeriodeAktif:   1,
		TanggalMulai:   "2025-07-01",
		TanggalSelesai: "2025-12-31",
	}}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ref/semester", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got struct {
		Success bool       `json:"success"`
		Data    []Semester `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Success || len(got.Data) != 1 || got.Data[0].SemesterID != "20251" || got.Data[0].PeriodeAktif != 1 {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestListTahunAjaran(t *testing.T) {
	store := &fakeStore{tahunAjaran: []TahunAjaran{{
		TahunAjaranID:  2025,
		Nama:           "2025/2026",
		PeriodeAktif:   1,
		TanggalMulai:   "2025-07-01",
		TanggalSelesai: "2026-06-30",
	}}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ref/tahun-ajaran", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got struct {
		Success bool          `json:"success"`
		Data    []TahunAjaran `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Success || len(got.Data) != 1 || got.Data[0].TahunAjaranID != 2025 || got.Data[0].PeriodeAktif != 1 {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestGetWilayahHierarchy(t *testing.T) {
	store := &fakeStore{wilayah: WilayahHierarchy{
		DesaKelurahan: &Region{KodeWilayah: "0101011001", Nama: "Contoh Kelurahan"},
		Kecamatan:     &Region{KodeWilayah: "010101", Nama: "Contoh Kecamatan"},
		KabupatenKota: &Region{KodeWilayah: "010100", Nama: "Contoh Kota"},
		Provinsi:      &Region{KodeWilayah: "010000", Nama: "Contoh Provinsi"},
	}}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ref/wilayah/0101011001", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got struct {
		Success bool             `json:"success"`
		Data    WilayahHierarchy `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Success || store.gotKode != "0101011001" {
		t.Fatalf("success=%v, kode passed to store=%q", got.Success, store.gotKode)
	}
	if got.Data.DesaKelurahan == nil || got.Data.Kecamatan == nil || got.Data.KabupatenKota == nil || got.Data.Provinsi == nil {
		t.Fatalf("expected all hierarchy levels, got %+v", got.Data)
	}
}

func TestGetWilayahHierarchyNotFound(t *testing.T) {
	store := &fakeStore{wilayahErr: ErrNotFound}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ref/wilayah/unknown", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestReferenceStoreErrorsAreHidden(t *testing.T) {
	store := &fakeStore{semesterErr: errors.New("database connection details")}
	w := httptest.NewRecorder()
	newTestServer(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/ref/semester", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if body := w.Body.String(); body == "" || strings.Contains(body, "database connection details") {
		t.Fatalf("unexpected error response: %s", body)
	}
}

func TestReferenceRoutesRejectNonGETMethods(t *testing.T) {
	for _, path := range []string{"/api/ref/semester", "/api/ref/tahun-ajaran", "/api/ref/wilayah/010000"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			newTestServer(&fakeStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, nil))
			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
