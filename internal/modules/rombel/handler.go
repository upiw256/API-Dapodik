package rombel

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
	mux.HandleFunc("GET /api/pembelajaran", h.listAllPembelajaran)
	mux.HandleFunc("GET /api/rombel", h.list)
	mux.HandleFunc("GET /api/rombel/{rombongan_belajar_id}", h.getDetail)
	mux.HandleFunc("GET /api/rombel/{rombongan_belajar_id}/anggota", h.listAnggota)
	mux.HandleFunc("GET /api/rombel/{rombongan_belajar_id}/pembelajaran", h.listPembelajaran)
	mux.HandleFunc("GET /api/rombel/{rombongan_belajar_id}/statistik", h.getStatistics)
}

func (h *Handler) listAllPembelajaran(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAllPembelajaran(r.Context())
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil seluruh pembelajaran", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "seluruh pembelajaran berhasil diambil", items, nil)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	p, err := utils.ParseListPagination(r.URL.Query())
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	q := r.URL.Query()
	f := Filters{SekolahID: strings.TrimSpace(q.Get("sekolah_id")), SemesterID: strings.TrimSpace(q.Get("semester_id")), TingkatPendidikanID: strings.TrimSpace(q.Get("tingkat_pendidikan_id")), JenisRombel: strings.TrimSpace(q.Get("jenis_rombel")), KurikulumID: strings.TrimSpace(q.Get("kurikulum_id")), JurusanSPID: strings.TrimSpace(q.Get("jurusan_sp_id")), PTKID: strings.TrimSpace(q.Get("ptk_id")), Search: strings.TrimSpace(q.Get("search"))}
	for name, id := range map[string]string{"sekolah_id": f.SekolahID, "jurusan_sp_id": f.JurusanSPID, "ptk_id": f.PTKID} {
		if id != "" && !isUUID(id) {
			response.WriteError(w, http.StatusBadRequest, name+" harus berupa UUID yang valid")
			return
		}
	}
	for name, value := range map[string]string{"tingkat_pendidikan_id": f.TingkatPendidikanID, "jenis_rombel": f.JenisRombel, "kurikulum_id": f.KurikulumID} {
		if value != "" {
			n, e := strconv.ParseInt(value, 10, 64)
			if e != nil || n < 1 {
				response.WriteError(w, http.StatusBadRequest, name+" harus berupa bilangan bulat positif")
				return
			}
		}
	}
	items, total, err := h.service.List(r.Context(), p.Page, p.Limit, f)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil daftar rombel", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "daftar rombel berhasil diambil", items, utils.PaginationMeta(total, p))
}

func (h *Handler) getDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := validRombelID(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "rombel tidak ditemukan")
			return
		}
		response.WriteInternalError(w, r, "gagal mengambil detail rombel", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "detail rombel berhasil diambil", item, nil)
}
func (h *Handler) listAnggota(w http.ResponseWriter, r *http.Request) {
	id, ok := validRombelID(w, r)
	if !ok {
		return
	}
	item, err := h.service.ListAnggota(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "rombel tidak ditemukan")
			return
		}
		response.WriteInternalError(w, r, "gagal mengambil anggota rombel", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "anggota rombel berhasil diambil", item, nil)
}
func (h *Handler) listPembelajaran(w http.ResponseWriter, r *http.Request) {
	id, ok := validRombelID(w, r)
	if !ok {
		return
	}
	detail, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "rombel tidak ditemukan")
			return
		}
		response.WriteInternalError(w, r, "gagal mengambil detail rombel", err)
		return
	}
	items, err := h.service.ListPembelajaran(r.Context(), id)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil pembelajaran rombel", err)
		return
	}
	response.WriteJSON(w, http.StatusOK, struct {
		Success   bool           `json:"success"`
		Message   string         `json:"message"`
		NamaKelas string         `json:"nama_kelas"`
		WaliKelas *string        `json:"wali_kelas"`
		Data      []Pembelajaran `json:"data"`
	}{
		Success:   true,
		Message:   "pembelajaran rombel berhasil diambil",
		NamaKelas: detail.Nama,
		WaliKelas: detail.WaliKelas,
		Data:      items,
	})
}
func (h *Handler) getStatistics(w http.ResponseWriter, r *http.Request) {
	id, ok := validRombelID(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetStatistics(r.Context(), id)
	if err != nil {
		response.WriteInternalError(w, r, "gagal mengambil statistik rombel", err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "statistik rombel berhasil diambil", item, nil)
}

func validRombelID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("rombongan_belajar_id")
	if !isUUID(id) {
		response.WriteError(w, http.StatusBadRequest, "rombongan_belajar_id harus berupa UUID yang valid")
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
