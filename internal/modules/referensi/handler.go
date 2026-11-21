package referensi

import (
	"errors"
	"net/http"

	"dapodik-readonly-api/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ref/semester", h.listSemesters)
	mux.HandleFunc("GET /api/ref/tahun-ajaran", h.listTahunAjaran)
	mux.HandleFunc("GET /api/ref/wilayah/{kode_wilayah}", h.getWilayahHierarchy)
}

func (h *Handler) listSemesters(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListSemesters(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil data semester")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "data semester berhasil diambil", items, nil)
}

func (h *Handler) listTahunAjaran(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListTahunAjaran(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil data tahun ajaran")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "data tahun ajaran berhasil diambil", items, nil)
}

func (h *Handler) getWilayahHierarchy(w http.ResponseWriter, r *http.Request) {
	kode := r.PathValue("kode_wilayah")
	item, err := h.service.GetWilayahHierarchy(r.Context(), kode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "kode wilayah tidak ditemukan")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil hierarki wilayah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "hierarki wilayah berhasil diambil", item, nil)
}
