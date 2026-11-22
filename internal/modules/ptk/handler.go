package ptk

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
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
	mux.HandleFunc("GET /api/ptk", h.list)
	mux.HandleFunc("GET /api/ptk/{ptk_id}", h.getDetail)
	mux.HandleFunc("GET /api/ptk/{ptk_id}/pembelajaran", h.listPembelajaran)
	mux.HandleFunc("GET /api/ptk/{ptk_id}/tugas-tambahan", h.listTugasTambahan)
	mux.HandleFunc("GET /api/ptk/{ptk_id}/riwayat", h.getRiwayat)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	pagination, err := utils.ParseListPagination(r.URL.Query())
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	filters := Filters{
		SekolahID:     strings.TrimSpace(r.URL.Query().Get("sekolah_id")),
		TahunAjaranID: strings.TrimSpace(r.URL.Query().Get("tahun_ajaran_id")),
		JenisPTKID:    strings.TrimSpace(r.URL.Query().Get("jenis_ptk_id")),
		Search:        strings.TrimSpace(r.URL.Query().Get("search")),
	}
	if filters.SekolahID != "" && !isUUID(filters.SekolahID) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	if !isPositiveInteger(filters.TahunAjaranID) || !isPositiveInteger(filters.JenisPTKID) {
		response.WriteError(w, http.StatusBadRequest, "tahun_ajaran_id dan jenis_ptk_id harus berupa bilangan bulat positif")
		return
	}
	items, total, err := h.service.List(r.Context(), pagination.Page, pagination.Limit, filters)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil daftar PTK", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "daftar PTK berhasil diambil", items, utils.PaginationMeta(total, pagination))
}

func (h *Handler) getDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ptk_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "ptk_id harus berupa UUID yang valid")
		return
	}
	item, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "PTK tidak ditemukan")
			return
		}
		response.WriteInternalError(w, r, "gagal mengambil detail PTK", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "detail PTK berhasil diambil", item, nil)
}

func (h *Handler) listPembelajaran(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathPTKID(w, r)
	if !ok {
		return
	}
	items, err := h.service.ListPembelajaran(r.Context(), id, strings.TrimSpace(r.URL.Query().Get("semester_id")))
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil data pembelajaran PTK", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "data pembelajaran PTK berhasil diambil", items, nil)
}

func (h *Handler) listTugasTambahan(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathPTKID(w, r)
	if !ok {
		return
	}
	items, err := h.service.ListTugasTambahan(r.Context(), id)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil tugas tambahan PTK", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "tugas tambahan PTK berhasil diambil", items, nil)
}

func (h *Handler) getRiwayat(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathPTKID(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetRiwayat(r.Context(), id)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil riwayat PTK", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "riwayat PTK berhasil diambil", item, nil)
}

func validPathPTKID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("ptk_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "ptk_id harus berupa UUID yang valid")
		return "", false
	}
	return id, true
}

func isUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func isPositiveInteger(value string) bool {
	if value == "" {
		return true
	}
	number, err := strconv.ParseInt(value, 10, 64)
	return err == nil && number > 0
}
