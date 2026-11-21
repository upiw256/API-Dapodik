package pesertadidik

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"dapodik-readonly-api/internal/response"
	"dapodik-readonly-api/internal/utils"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/peserta-didik", h.list)
	mux.HandleFunc("GET /api/peserta-didik/{peserta_didik_id}", h.getDetail)
	mux.HandleFunc("GET /api/peserta-didik/{peserta_didik_id}/rombel", h.listRombel)
	mux.HandleFunc("GET /api/peserta-didik/{peserta_didik_id}/prestasi-beasiswa", h.getPrestasiBeasiswa)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pagination, err := utils.ParseListPagination(q)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	filters := Filters{SekolahID: strings.TrimSpace(q.Get("sekolah_id")), RombonganBelajarID: strings.TrimSpace(q.Get("rombongan_belajar_id")), SemesterID: strings.TrimSpace(q.Get("semester_id")), TingkatPendidikanID: strings.TrimSpace(q.Get("tingkat_pendidikan_id")), Search: strings.TrimSpace(q.Get("search"))}
	if filters.SekolahID != "" && !isUUID(filters.SekolahID) {
		response.WriteError(w, http.StatusBadRequest, "sekolah_id harus berupa UUID yang valid")
		return
	}
	if filters.RombonganBelajarID != "" && !isUUID(filters.RombonganBelajarID) {
		response.WriteError(w, http.StatusBadRequest, "rombongan_belajar_id harus berupa UUID yang valid")
		return
	}
	if filters.TingkatPendidikanID != "" {
		n, e := strconv.ParseInt(filters.TingkatPendidikanID, 10, 64)
		if e != nil || n < 1 {
			response.WriteError(w, http.StatusBadRequest, "tingkat_pendidikan_id harus berupa bilangan bulat positif")
			return
		}
	}
	if aktif := strings.TrimSpace(q.Get("aktif")); aktif != "" {
		value, e := strconv.ParseBool(aktif)
		if e != nil {
			response.WriteError(w, http.StatusBadRequest, "aktif harus bernilai true atau false")
			return
		}
		filters.Aktif = &value
	}
	items, total, err := h.service.List(r.Context(), pagination.Page, pagination.Limit, filters)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil daftar peserta didik")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "daftar peserta didik berhasil diambil", items, utils.PaginationMeta(total, pagination))
}

func (h *Handler) getDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := validStudentID(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "peserta didik tidak ditemukan")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil detail peserta didik")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "detail peserta didik berhasil diambil", item, nil)
}

func (h *Handler) listRombel(w http.ResponseWriter, r *http.Request) {
	id, ok := validStudentID(w, r)
	if !ok {
		return
	}
	items, err := h.service.ListRombel(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil riwayat rombel peserta didik")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "riwayat rombel peserta didik berhasil diambil", items, nil)
}

func (h *Handler) getPrestasiBeasiswa(w http.ResponseWriter, r *http.Request) {
	id, ok := validStudentID(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetPrestasiBeasiswa(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "gagal mengambil prestasi dan beasiswa peserta didik")
		return
	}
	response.WriteSuccess(w, http.StatusOK, "prestasi dan beasiswa peserta didik berhasil diambil", item, nil)
}

func validStudentID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("peserta_didik_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "peserta_didik_id harus berupa UUID yang valid")
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
