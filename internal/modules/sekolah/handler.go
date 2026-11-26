package sekolah

import (
	"errors"
	"net/http"
	"strings"

	"dapodik-readonly-api/internal/response"
	"dapodik-readonly-api/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/sekolah", h.listSchools)
	mux.HandleFunc("GET /api/sekolah/", h.listSchoolsPath)
	mux.HandleFunc("GET /api/sekolah/detail", h.getSchool)
	mux.HandleFunc("GET /api/sekolah/longitudinal", h.getLongitudinal)
	mux.HandleFunc("GET /api/sekolah/jurusan", h.listJurusan)
	mux.HandleFunc("GET /api/sekolah/rombel", h.listRombel)
	mux.HandleFunc("GET /api/sekolah/ruang", h.listRooms)
}

func (h *Handler) listSchoolsPath(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/sekolah/" {
		http.NotFound(w, r)
		return
	}
	h.listSchools(w, r)
}

func (h *Handler) listSchools(w http.ResponseWriter, r *http.Request) {
	pagination, err := utils.ParseListPagination(r.URL.Query())
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	npsn := strings.TrimSpace(r.URL.Query().Get("npsn"))
	items, total, err := h.service.ListSchools(r.Context(), pagination.Page, pagination.Limit, npsn)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil daftar sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "daftar sekolah berhasil diambil", items, utils.PaginationMeta(total, pagination))
}

func (h *Handler) getSchool(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetConfiguredSchool(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "sekolah tidak ditemukan")
			return
		}
		response.WriteInternalError(w, r, "gagal mengambil detail sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "detail sekolah berhasil diambil", item, nil)
}

func (h *Handler) getLongitudinal(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetConfiguredLongitudinal(r.Context(), r.URL.Query().Get("semester_id"))
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil data longitudinal sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "data longitudinal sekolah berhasil diambil", items, nil)
}

func (h *Handler) listJurusan(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListConfiguredJurusan(r.Context())
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil jurusan sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "jurusan sekolah berhasil diambil", items, nil)
}

func (h *Handler) listRombel(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListConfiguredRombel(r.Context(), r.URL.Query().Get("semester_id"))
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil rombongan belajar sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "rombongan belajar sekolah berhasil diambil", items, nil)
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListConfiguredRooms(r.Context())
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil ruang sekolah", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "ruang sekolah berhasil diambil", items, nil)
}
