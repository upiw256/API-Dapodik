package sekolah

import (
	"encoding/hex"
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
	mux.HandleFunc("GET /api/sekolah/{sekolah_id}", h.getSchool)
	mux.HandleFunc("GET /api/sekolah/{sekolah_id}/longitudinal", h.getLongitudinal)
	mux.HandleFunc("GET /api/sekolah/{sekolah_id}/jurusan", h.listJurusan)
	mux.HandleFunc("GET /api/sekolah/{sekolah_id}/rombel", h.listRombel)
	mux.HandleFunc("GET /api/sekolah/{sekolah_id}/ruang", h.listRooms)
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
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil daftar sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "daftar sekolah berhasil diambil", items, utils.PaginationMeta(total, pagination))
}

func (h *Handler) getSchool(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sekolah_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	item, err := h.service.GetSchool(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "sekolah tidak ditemukan")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil detail sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "detail sekolah berhasil diambil", item, nil)
}

func (h *Handler) getLongitudinal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sekolah_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	items, err := h.service.GetLongitudinal(r.Context(), id, r.URL.Query().Get("semester_id"))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil data longitudinal sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "data longitudinal sekolah berhasil diambil", items, nil)
}

func (h *Handler) listJurusan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sekolah_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	items, err := h.service.ListJurusan(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil jurusan sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "jurusan sekolah berhasil diambil", items, nil)
}

func (h *Handler) listRombel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sekolah_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	items, err := h.service.ListRombel(r.Context(), id, r.URL.Query().Get("semester_id"))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil rombongan belajar sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "rombongan belajar sekolah berhasil diambil", items, nil)
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sekolah_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	items, err := h.service.ListRooms(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil ruang sekolah")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "ruang sekolah berhasil diambil", items, nil)
}

func isUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded, err := hex.DecodeString(compact)
	return err == nil && len(decoded) == 16
}
