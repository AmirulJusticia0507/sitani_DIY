package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// ErrNotFound dikembalikan saat data yang diminta tidak tersedia.
var ErrNotFound = errors.New("data tidak ditemukan")

// Kesalahan alur draft yang perlu diterjemahkan menjadi kode HTTP berbeda.
var (
	// ErrBukanDraft: permohonan yang dimaksud bukan draft lagi (misalnya
	// sudah diajukan dari tab lain).
	ErrBukanDraft = errors.New("permohonan sudah diajukan sehingga tidak dapat diubah sebagai draft")
	// ErrDraftTidakAda: permohonan diajukan tanpa draft yang tersimpan.
	ErrDraftTidakAda = errors.New("belum ada draft permohonan yang dapat diajukan")
	// ErrSudahMengajukan: pelanggan mencoba menyimpan draft baru padahal masih
	// punya pengajuan resmi yang berjalan.
	ErrSudahMengajukan = errors.New("masih ada permohonan yang berjalan; selesaikan atau batalkan terlebih dahulu")
	// ErrTidakBerhak: peran peminta tidak diizinkan menjalankan operasi ini.
	// Dibedakan dari galat validasi agar handler menjawab 403, bukan 400.
	ErrTidakBerhak = errors.New("peran tidak diizinkan menjalankan operasi ini")
)

// ErrPeran adalah galat peran yang membawa pesan spesifik per operasi; tetap
// memenuhi errors.Is(err, ErrTidakBerhak) agar handler memetakannya ke 403.
func ErrPeran(pesan string) error {
	return fmt.Errorf("%w: %s", ErrTidakBerhak, pesan)
}

// FaseInspeksi adalah jumlah fase inspeksi lapangan yang wajib diselesaikan.
const FaseInspeksi = 4

// Izin per role untuk Fase 1 verifikasi lapangan.
var izinRole = map[string]map[string]bool{
	db.RolePelanggan: {
		"permohonan:buat":  true,
		"permohonan:lihat": true,
		"pembayaran:awal":  true,
	},
	db.RolePengawas: {
		"permohonan:lihat": true,
		"inspeksi:catat":   true,
		"buku_induk:catat": true,
		"registrasi:buat":  true,
		"registrasi:lihat": true,
		"stok:lihat":       true,
		"penjualan:lihat":  true,
		"laporan:lihat":    true,
	},
	db.RoleSuperadmin: {
		"permohonan:buat":        true,
		"permohonan:lihat":       true,
		"pembayaran:awal":        true,
		"inspeksi:catat":         true,
		"buku_induk:catat":       true,
		"registrasi:buat":        true,
		"registrasi:lihat":       true,
		"registrasi:ubah_status": true,
		"stok:lihat":             true,
		"penjualan:buat":         true,
		"penjualan:lihat":        true,
		"penjualan:batal":        true,
		"laporan:lihat":          true,
	},
	db.RoleAdmin: {
		"permohonan:buat":        true,
		"permohonan:lihat":       true,
		"pembayaran:awal":        true,
		"inspeksi:catat":         true,
		"buku_induk:catat":       true,
		"registrasi:buat":        true,
		"registrasi:lihat":       true,
		"registrasi:ubah_status": true,
		"stok:lihat":             true,
		"penjualan:buat":         true,
		"penjualan:lihat":        true,
		"penjualan:batal":        true,
		"laporan:lihat":          true,
	},
	db.RolePenandatangan: {
		"permohonan:lihat": true,
	},
	db.RoleAnalis: {
		"permohonan:lihat": true,
	},
	db.RoleLegalisasi: {
		"permohonan:lihat": true,
	},
	db.RolePercetakan: {
		"permohonan:lihat": true,
	},
}

// boleh memeriksa apakah aktor dengan role memiliki izin aksi.
func boleh(role, aksi string) bool {
	if izin, ok := izinRole[role]; ok {
		return izin[aksi]
	}
	return false
}

// Service mengatur aturan bisnis Fase 1 sertifikasi benih BP3MBTP.
type Service struct {
	st  store.Store
	now func() time.Time
}

func NewService(st store.Store) *Service {
	return &Service{st: st, now: time.Now}
}

// BuatPermohonan mencatat permohonan sertifikasi benih baru oleh Pelanggan.
// Nomor permohonan dibuat server (SPB-YYYYMMDD-XXXX) dan tidak diterima dari
// client; nomor itu baru ada setelah permohonan tersimpan.
func (s *Service) BuatPermohonan(ctx context.Context, pelangganID int64, pelangganRole string, req BuatPermohonanRequest) (*db.PermohonanBenih, error) {
	if !boleh(pelangganRole, "permohonan:buat") {
		return nil, ErrPeran("role tidak diizinkan membuat permohonan")
	}
	if err := validasiIsianPermohonan(&req); err != nil {
		return nil, err
	}

	now := s.now()
	p := db.PermohonanBenih{
		PelangganID:          pelangganID,
		JenisBenih:           req.JenisBenih,
		Varietas:             req.Varietas,
		LuasLahan:            req.LuasLahan,
		LokasiKebun:          req.LokasiKebun,
		Status:               db.StatusPermohonanDiajukan,
		PembayaranAwalStatus: "belum",
	}
	id, err := s.st.CreatePermohonan(ctx, p, now)
	if err != nil {
		return nil, fmt.Errorf("buat permohonan: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pelangganID,
		ActorRole: pelangganRole,
		Action:    "permohonan:buat",
		Entity:    "permohonan_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("jenis_benih=%s", req.JenisBenih),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return s.permohonanTersimpan(ctx, id)
}

// SimpanDraft menyimpan permohonan sebagai draft. Draft boleh disimpan walau
// isian belum lengkap; validasi lengkap dilakukan saat permohonan diajukan.
// Bila pelanggan belum punya draft atau pengajuan berjalan, draft baru dibuat;
// bila sudah ada draft, isiannya diperbarui sehingga tombol "Simpan Draft"
// dapat ditekan berkali-kali tanpa membuat baris ganda.
func (s *Service) SimpanDraft(ctx context.Context, pelangganID int64, pelangganRole string, draftID int64, req BuatPermohonanRequest) (*db.PermohonanBenih, error) {
	if !boleh(pelangganRole, "permohonan:buat") {
		return nil, ErrPeran("role tidak diizinkan menyimpan draft permohonan")
	}
	req.JenisBenih = strings.TrimSpace(req.JenisBenih)
	req.Varietas = strings.TrimSpace(req.Varietas)
	req.LokasiKebun = strings.TrimSpace(req.LokasiKebun)
	if req.LuasLahan < 0 {
		return nil, errors.New("luas lahan tidak boleh negatif")
	}

	if draftID > 0 {
		err := s.st.UpdatePermohonanDraft(ctx, draftID, db.PermohonanBenih{
			PelangganID: pelangganID,
			JenisBenih:  req.JenisBenih,
			Varietas:    req.Varietas,
			LuasLahan:   req.LuasLahan,
			LokasiKebun: req.LokasiKebun,
		})
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, ErrNotFound
		case errors.Is(err, store.ErrBukanDraft):
			return nil, ErrBukanDraft
		case err != nil:
			return nil, fmt.Errorf("simpan draft permohonan: %w", err)
		}
		return s.catatDraftTersimpan(ctx, draftID, pelangganID, pelangganRole, req)
	}

	aktif, err := s.permohonanAktif(ctx, pelangganID)
	if err != nil {
		return nil, err
	}
	if aktif != nil && aktif.Status != db.StatusPermohonanDraft {
		return nil, ErrSudahMengajukan
	}
	if aktif != nil {
		return s.SimpanDraft(ctx, pelangganID, pelangganRole, aktif.ID, req)
	}

	id, err := s.st.CreatePermohonan(ctx, db.PermohonanBenih{
		PelangganID:          pelangganID,
		JenisBenih:           req.JenisBenih,
		Varietas:             req.Varietas,
		LuasLahan:            req.LuasLahan,
		LokasiKebun:          req.LokasiKebun,
		Status:               db.StatusPermohonanDraft,
		PembayaranAwalStatus: "belum",
	}, s.now())
	if err != nil {
		return nil, fmt.Errorf("buat draft permohonan: %w", err)
	}
	return s.catatDraftTersimpan(ctx, id, pelangganID, pelangganRole, req)
}

// AjukanPermohonan mengajukan permohonan resmi. Semua isian wajib lengkap;
// setelah diajukan, isian tidak dapat diubah pelanggan dan nomor permohonan
// sudah final.
//
// draftID menentukan sumber draft:
//   - > 0: draft yang sudah tersimpan (misal hasil "Simpan Draft") diajukan;
//   - 0: draft aktif milik pelanggan diajukan bila ada, atau draft baru dibuat
//     lalu langsung diajukan dalam satu langkah.
func (s *Service) AjukanPermohonan(ctx context.Context, pelangganID int64, pelangganRole string, draftID int64, req BuatPermohonanRequest) (*db.PermohonanBenih, error) {
	if !boleh(pelangganRole, "permohonan:buat") {
		return nil, ErrPeran("role tidak diizinkan mengajukan permohonan")
	}
	if err := validasiIsianPermohonan(&req); err != nil {
		return nil, err
	}

	if draftID == 0 {
		aktif, err := s.permohonanAktif(ctx, pelangganID)
		if err != nil {
			return nil, err
		}
		switch {
		case aktif != nil && aktif.Status == db.StatusPermohonanDraft:
			draftID = aktif.ID
		case aktif != nil:
			// Pengajuan resmi yang masih berjalan tidak boleh digandakan.
			return nil, ErrSudahMengajukan
		default:
			draftID, err = s.st.CreatePermohonan(ctx, db.PermohonanBenih{
				PelangganID:          pelangganID,
				Status:               db.StatusPermohonanDraft,
				PembayaranAwalStatus: "belum",
			}, s.now())
			if err != nil {
				return nil, fmt.Errorf("buat draft permohonan: %w", err)
			}
		}
	}

	if err := s.st.AjukanPermohonan(ctx, draftID, db.PermohonanBenih{
		PelangganID: pelangganID,
		JenisBenih:  req.JenisBenih,
		Varietas:    req.Varietas,
		LuasLahan:   req.LuasLahan,
		LokasiKebun: req.LokasiKebun,
	}, s.now()); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, ErrNotFound
		case errors.Is(err, store.ErrBukanDraft):
			return nil, ErrBukanDraft
		}
		return nil, fmt.Errorf("ajukan permohonan: %w", err)
	}

	tersimpan, err := s.permohonanTersimpan(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pelangganID,
		ActorRole: pelangganRole,
		Action:    "permohonan:ajukan",
		Entity:    "permohonan_benih",
		EntityID:  &draftID,
		Detail:    fmt.Sprintf("nomor_permohonan=%s jenis_benih=%s", tersimpan.NomorPermohonan, tersimpan.JenisBenih),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return tersimpan, nil
}

// PermohonanAktif mengembalikan draft atau pengajuan berjalan milik pelanggan
// agar form dapat diisi ulang. Hasil nil berarti belum ada yang berjalan.
func (s *Service) PermohonanAktif(ctx context.Context, pelangganID int64, pelangganRole string) (*db.PermohonanBenih, error) {
	if !boleh(pelangganRole, "permohonan:buat") {
		return nil, ErrPeran("role tidak diizinkan melihat permohonan sendiri")
	}
	return s.permohonanAktif(ctx, pelangganID)
}

func (s *Service) permohonanAktif(ctx context.Context, pelangganID int64) (*db.PermohonanBenih, error) {
	p, err := s.st.GetPermohonanAktifPelanggan(ctx, pelangganID)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan aktif: %w", err)
	}
	return p, nil
}

// catatDraftTersimpan mencatat audit simpan draft dan mengembalikan isi draft
// yang benar-benar tersimpan.
func (s *Service) catatDraftTersimpan(ctx context.Context, id, pelangganID int64, pelangganRole string, req BuatPermohonanRequest) (*db.PermohonanBenih, error) {
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pelangganID,
		ActorRole: pelangganRole,
		Action:    "permohonan:draft_simpan",
		Entity:    "permohonan_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("jenis_benih=%s", req.JenisBenih),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return s.permohonanTersimpan(ctx, id)
}

// permohonanTersimpan membaca ulang agar nilai yang dikembalikan (termasuk
// nomor) sama dengan yang benar-benar tersimpan di database.
func (s *Service) permohonanTersimpan(ctx context.Context, id int64) (*db.PermohonanBenih, error) {
	tersimpan, err := s.st.GetPermohonan(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan tersimpan: %w", err)
	}
	if tersimpan == nil {
		return nil, ErrNotFound
	}
	return tersimpan, nil
}

// validasiIsianPermohonan merapikan spasi dan memastikan isian wajib terisi.
// Dipakai saat mengajukan permohonan; draft boleh belum lengkap.
func validasiIsianPermohonan(req *BuatPermohonanRequest) error {
	if req == nil {
		return errors.New("isian permohonan tidak tersedia")
	}
	req.JenisBenih = strings.TrimSpace(req.JenisBenih)
	req.Varietas = strings.TrimSpace(req.Varietas)
	req.LokasiKebun = strings.TrimSpace(req.LokasiKebun)
	if req.JenisBenih == "" {
		return errors.New("jenis benih wajib diisi")
	}
	if req.Varietas == "" {
		return errors.New("varietas wajib diisi")
	}
	if req.LuasLahan <= 0 {
		return errors.New("luas lahan harus lebih dari nol")
	}
	if req.LokasiKebun == "" {
		return errors.New("lokasi kebun wajib diisi")
	}
	if len(req.JenisBenih) > 100 || len(req.Varietas) > 100 {
		return errors.New("jenis benih atau varietas terlalu panjang")
	}
	return nil
}

// GetPermohonan mengembalikan detail permohonan berdasarkan ID.
func (s *Service) GetPermohonan(ctx context.Context, id int64) (*db.PermohonanBenih, error) {
	p, err := s.st.GetPermohonan(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

// BayarAwal mencatat pembayaran awal oleh Pelanggan melalui Gateway Pembayaran.
func (s *Service) BayarAwal(ctx context.Context, permohonanID int64, pelangganID int64, pelangganRole string, jumlah float64) error {
	if !boleh(pelangganRole, "pembayaran:awal") {
		return errors.New("role tidak diizinkan membayar awal")
	}
	if jumlah <= 0 {
		return errors.New("jumlah pembayaran harus lebih dari nol")
	}

	p, err := s.st.GetPermohonan(ctx, permohonanID)
	if err != nil {
		return fmt.Errorf("ambil permohonan: %w", err)
	}
	if p.PelangganID != pelangganID {
		return errors.New("bukan pemilik permohonan")
	}
	if p.Status == db.StatusPermohonanDraft {
		return errors.New("draft belum diajukan; ajukan permohonan terlebih dahulu sebelum membayar")
	}
	if p.Status != db.StatusPermohonanDiajukan {
		return fmt.Errorf("pembayaran awal hanya dapat dilakukan pada permohonan berstatus diajukan (status saat ini: %s)", p.Status)
	}
	if p.PembayaranAwalStatus == "lunas" {
		return errors.New("pembayaran awal sudah lunas")
	}

	now := s.now()
	pb := db.Pembayaran{
		PermohonanID: permohonanID,
		Jenis:        "awal",
		Jumlah:       jumlah,
		Status:       "lunas",
	}
	if _, err := s.st.CreatePembayaran(ctx, pb); err != nil {
		return fmt.Errorf("buat pembayaran: %w", err)
	}
	if err := s.st.UpdatePermohonanStatus(ctx, permohonanID, "dibayar", "lunas"); err != nil {
		return fmt.Errorf("update status permohonan: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pelangganID,
		ActorRole: pelangganRole,
		Action:    "pembayaran:awal",
		Entity:    "permohonan_benih",
		EntityID:  &permohonanID,
		Detail:    fmt.Sprintf("jumlah=%.2f", jumlah),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	_ = now
	return nil
}

// CatatInspeksi mencatat hasil satu fase inspeksi lapangan oleh Pengawas.
func (s *Service) CatatInspeksi(ctx context.Context, permohonanID int64, fase int, pengawasID int64, pengawasRole string, hasil string, catatan string, tanggalInspeksi string) error {
	if !boleh(pengawasRole, "inspeksi:catat") {
		return errors.New("role tidak diizinkan catat inspeksi")
	}
	if fase < 1 || fase > FaseInspeksi {
		return errors.New("fase inspeksi harus antara 1 dan 4")
	}
	if hasil != "lulus" && hasil != "perlu_perbaikan" && hasil != "tidak_lulus" {
		return errors.New("hasil inspeksi tidak valid")
	}

	p, err := s.st.GetPermohonan(ctx, permohonanID)
	if err != nil {
		return fmt.Errorf("ambil permohonan: %w", err)
	}
	if p.Status == db.StatusPermohonanDraft {
		return errors.New("draft belum diajukan sehingga belum dapat diinspeksi")
	}
	if p.Status != db.StatusPermohonanDibayar {
		return errors.New("inspeksi hanya dapat dilakukan setelah pembayaran awal lunas")
	}

	i := db.InspeksiFase{
		PermohonanID:    permohonanID,
		Fase:            fase,
		PengawasID:      pengawasID,
		Hasil:           hasil,
		Catatan:         catatan,
		TanggalInspeksi: tanggalInspeksi,
	}
	if _, err := s.st.CreateInspeksiFase(ctx, i); err != nil {
		return fmt.Errorf("catat inspeksi: %w", err)
	}

	fases, err := s.st.ListInspeksiFase(ctx, permohonanID)
	if err != nil {
		return fmt.Errorf("daftar inspeksi: %w", err)
	}
	lulus := 0
	for _, f := range fases {
		if f.Hasil == "lulus" {
			lulus++
		}
	}
	if lulus >= FaseInspeksi {
		if err := s.st.UpdatePermohonanStatus(ctx, permohonanID, "diverifikasi", p.PembayaranAwalStatus); err != nil {
			return fmt.Errorf("update status terverifikasi: %w", err)
		}
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pengawasID,
		ActorRole: pengawasRole,
		Action:    "inspeksi:catat",
		Entity:    "permohonan_benih",
		EntityID:  &permohonanID,
		Detail:    fmt.Sprintf("fase=%d hasil=%s", fase, hasil),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// CatatBukuInduk mencatat pencatatan buku induk setelah verifikasi lapangan.
func (s *Service) CatatBukuInduk(ctx context.Context, permohonanID int64, nomorBukuInduk string, catatan string, userID int64, userRole string) error {
	if !boleh(userRole, "buku_induk:catat") {
		return errors.New("role tidak diizinkan catat buku induk")
	}
	if nomorBukuInduk == "" {
		return errors.New("nomor buku induk wajib diisi")
	}

	p, err := s.st.GetPermohonan(ctx, permohonanID)
	if err != nil {
		return fmt.Errorf("ambil permohonan: %w", err)
	}
	if p.Status != "diverifikasi" {
		return errors.New("buku induk hanya dapat dicatat setelah verifikasi selesai")
	}

	b := db.BukuInduk{
		PermohonanID:   permohonanID,
		NomorBukuInduk: nomorBukuInduk,
		Catatan:        catatan,
		DibuatOleh:     userID,
	}
	if _, err := s.st.CreateBukuInduk(ctx, b); err != nil {
		return fmt.Errorf("buat buku induk: %w", err)
	}
	if err := s.st.UpdatePermohonanStatus(ctx, permohonanID, "diinspeksi", p.PembayaranAwalStatus); err != nil {
		return fmt.Errorf("update status buku induk: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &userID,
		ActorRole: userRole,
		Action:    "buku_induk:catat",
		Entity:    "permohonan_benih",
		EntityID:  &permohonanID,
		Detail:    fmt.Sprintf("nomor_buku_induk=%s", nomorBukuInduk),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// BuatPermohonanRequest adalah input pembuatan permohonan.
type BuatPermohonanRequest struct {
	JenisBenih  string
	Varietas    string
	LuasLahan   float64
	LokasiKebun string
}

// BuatSertifikat membuat sertifikat benih untuk permohonan yang lulus (Fase 3, jalur lulus).
func (s *Service) BuatSertifikat(ctx context.Context, permohonanID int64, penandatanganID int64, penandatanganRole string, nomorSertifikat string) (*db.SertifikatBenih, error) {
	if !boleh(penandatanganRole, "permohonan:lihat") {
		return nil, errors.New("role tidak diizinkan membuat sertifikat")
	}
	p, err := s.st.GetPermohonan(ctx, permohonanID)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	srt := db.SertifikatBenih{
		PermohonanID:    permohonanID,
		NomorSertifikat: nomorSertifikat,
		PenandatanganID: penandatanganID,
		TTEStatus:       "belum",
	}
	id, err := s.st.CreateSertifikat(ctx, srt)
	if err != nil {
		return nil, fmt.Errorf("buat sertifikat: %w", err)
	}
	srt.ID = id
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &penandatanganID,
		ActorRole: penandatanganRole,
		Action:    "sertifikat:buat",
		Entity:    "sertifikat_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("nomor_sertifikat=%s", nomorSertifikat),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return &srt, nil
}

// TandaTanganiTTE menandatangani sertifikat oleh Pejabat melalui TTE Sadewa.
func (s *Service) TandaTanganiTTE(ctx context.Context, permohonanID int64, penandatanganID int64, penandatanganRole string, ttd time.Time, dokumenURL string) error {
	if !boleh(penandatanganRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan menandatangani TTE")
	}
	srt, err := s.st.GetSertifikat(ctx, permohonanID)
	if err != nil {
		return fmt.Errorf("ambil sertifikat: %w", err)
	}
	if srt == nil {
		return ErrNotFound
	}
	if err := s.st.UpdateSertifikatTTE(ctx, permohonanID, "ditandatangani", dokumenURL, &ttd); err != nil {
		return fmt.Errorf("update TTE: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &penandatanganID,
		ActorRole: penandatanganRole,
		Action:    "sertifikat:ttd_tte",
		Entity:    "sertifikat_benih",
		EntityID:  &permohonanID,
		Detail:    fmt.Sprintf("tte_status=ditandatangani dokumen=%s", dokumenURL),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// BuatDraftLabel membuat draft label dengan kode unik/barcode terintegrasi SIM Perbenihan.
func (s *Service) BuatDraftLabel(ctx context.Context, sertifikatID int64, kodeBarcode string, pembuatID int64, pembuatRole string) (*db.DraftLabel, error) {
	if !boleh(pembuatRole, "permohonan:lihat") {
		return nil, errors.New("role tidak diizinkan membuat draft label")
	}
	if kodeBarcode == "" {
		return nil, errors.New("kode barcode wajib diisi")
	}
	srt, err := s.st.GetSertifikatByID(ctx, sertifikatID)
	if err != nil {
		return nil, fmt.Errorf("ambil sertifikat: %w", err)
	}
	if srt == nil {
		return nil, ErrNotFound
	}
	lbl := db.DraftLabel{
		SertifikatID:       sertifikatID,
		KodeBarcode:        kodeBarcode,
		Status:             "draft",
		IntegrasiSIMStatus: "belum",
	}
	id, err := s.st.CreateDraftLabel(ctx, lbl)
	if err != nil {
		return nil, fmt.Errorf("buat draft label: %w", err)
	}
	lbl.ID = id
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pembuatID,
		ActorRole: pembuatRole,
		Action:    "label:draft",
		Entity:    "draft_label",
		EntityID:  &id,
		Detail:    fmt.Sprintf("kode_barcode=%s", kodeBarcode),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return &lbl, nil
}

// IntegrasiSIM mengintegrasikan kode barcode label ke SIM Perbenihan.
func (s *Service) IntegrasiSIM(ctx context.Context, draftLabelID int64, pembuatID int64, pembuatRole string) error {
	if !boleh(pembuatRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan integrasi SIM Perbenihan")
	}
	lbl, err := s.st.GetDraftLabelByID(ctx, draftLabelID)
	if err != nil {
		return fmt.Errorf("ambil draft label: %w", err)
	}
	if lbl == nil {
		return ErrNotFound
	}
	if err := s.st.UpdateDraftLabel(ctx, draftLabelID, lbl.Status, "terkirim", lbl.KodeBarcode); err != nil {
		return fmt.Errorf("update integrasi SIM: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pembuatID,
		ActorRole: pembuatRole,
		Action:    "label:integrasi_sim",
		Entity:    "draft_label",
		EntityID:  &draftLabelID,
		Detail:    fmt.Sprintf("kode_barcode=%s", lbl.KodeBarcode),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// CetakLabel menandai label dicetak oleh Percetakan.
func (s *Service) CetakLabel(ctx context.Context, draftLabelID int64, cetakID int64, cetakRole string) error {
	if !boleh(cetakRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan mencetak label")
	}
	lbl, err := s.st.GetDraftLabelByID(ctx, draftLabelID)
	if err != nil {
		return fmt.Errorf("ambil draft label: %w", err)
	}
	if lbl == nil {
		return ErrNotFound
	}
	if err := s.st.UpdateDraftLabel(ctx, draftLabelID, "dicetak", lbl.IntegrasiSIMStatus, lbl.KodeBarcode); err != nil {
		return fmt.Errorf("cetak label: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &cetakID,
		ActorRole: cetakRole,
		Action:    "label:cetak",
		Entity:    "draft_label",
		EntityID:  &draftLabelID,
		Detail:    "label dicetak",
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// LegalisasiLabel memverifikasi dan me-legalisasi label oleh petugas Legal.
func (s *Service) LegalisasiLabel(ctx context.Context, draftLabelID int64, legalisasiID int64, legalisasiRole string, status string, catatan string) error {
	if !boleh(legalisasiRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan legalisasi label")
	}
	if status != "diverifikasi" && status != "ditolak" {
		return errors.New("status legalisasi harus diverifikasi atau ditolak")
	}
	lbl, err := s.st.GetDraftLabelByID(ctx, draftLabelID)
	if err != nil {
		return fmt.Errorf("ambil draft label: %w", err)
	}
	if lbl == nil {
		return ErrNotFound
	}
	id, err := s.st.CreateLegalisasi(ctx, db.LabelLegalisasi{
		DraftLabelID: draftLabelID,
		LegalisasiID: legalisasiID,
		Status:       status,
		Catatan:      catatan,
	})
	if err != nil {
		return fmt.Errorf("buat legalisasi: %w", err)
	}
	if err := s.st.UpdateLegalisasi(ctx, id, status); err != nil {
		return fmt.Errorf("update legalisasi: %w", err)
	}
	if status == "diverifikasi" {
		if err := s.st.UpdateDraftLabel(ctx, draftLabelID, "legalisasi", lbl.IntegrasiSIMStatus, lbl.KodeBarcode); err != nil {
			return fmt.Errorf("update status label legalisasi: %w", err)
		}
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &legalisasiID,
		ActorRole: legalisasiRole,
		Action:    "label:legalisasi",
		Entity:    "draft_label",
		EntityID:  &draftLabelID,
		Detail:    fmt.Sprintf("status=%s", status),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// TerimaDokumenPembayaranAkhir mencatat penyerahan sertifikat & label ke pelanggan dan pembayaran akhir.
func (s *Service) TerimaDokumenPembayaranAkhir(ctx context.Context, permohonanID int64, diterimaOleh string, diterimaOlehRole string, sertifikatDikirim, labelDikirim bool, pembayaranAkhirStatus string) error {
	if !boleh(diterimaOlehRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan terima dokumen")
	}
	if pembayaranAkhirStatus != "lunas" && pembayaranAkhirStatus != "gagal" {
		return errors.New("status pembayaran akhir harus lunas atau gagal")
	}
	now := time.Now()
	if _, err := s.st.CreatePenyerahan(ctx, db.PenyerahanDokumen{
		PermohonanID:          permohonanID,
		SertifikatDikirim:     sertifikatDikirim,
		LabelDikirim:          labelDikirim,
		PembayaranAkhirStatus: pembayaranAkhirStatus,
		TanggalPenyerahan:     &now,
		DiterimaOleh:          diterimaOleh,
	}); err != nil {
		return fmt.Errorf("buat penyerahan: %w", err)
	}
	if err := s.st.UpdatePenyerahan(ctx, permohonanID, sertifikatDikirim, labelDikirim, pembayaranAkhirStatus); err != nil {
		return fmt.Errorf("update penyerahan: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   nil,
		ActorRole: diterimaOlehRole,
		Action:    "dokumen:penyerahan",
		Entity:    "permohonan_benih",
		EntityID:  &permohonanID,
		Detail:    fmt.Sprintf("sertifikat=%v label=%v pembayaran=%s", sertifikatDikirim, labelDikirim, pembayaranAkhirStatus),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// JalurTidakLulus mengirimkan laporan lengkap pengujian ke pelanggan dan menyelesaikan proses (Fase 3, jalur tidak lulus).
func (s *Service) JalurTidakLulus(ctx context.Context, sampelID int64, adminID int64, adminRole string) error {
	if !boleh(adminRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan jalur tidak lulus")
	}
	lapor, err := s.st.GetLaporanMutu(ctx, sampelID)
	if err != nil {
		return fmt.Errorf("ambil laporan mutu: %w", err)
	}
	if lapor == nil {
		return ErrNotFound
	}
	if lapor.Kesimpulan != "tidak_lulus" {
		return errors.New("jalur tidak lulus hanya untuk hasil uji tidak lulus")
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &adminID,
		ActorRole: adminRole,
		Action:    "laporan:kirim_tidak_lulus",
		Entity:    "laporan_mutu",
		EntityID:  &lapor.ID,
		Detail:    fmt.Sprintf("sampel_id=%d kesimpulan=%s", sampelID, lapor.Kesimpulan),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// ProsesHasilMutu adalah auto-trigger percabangan setelah laporan mutu disusun:
// lulus → alur sertifikat & label; tidak_lulus → notifikasi gagal ke pelanggan (proses selesai).
func (s *Service) ProsesHasilMutu(ctx context.Context, sampelID int64, adminID int64, adminRole string) (string, error) {
	if !boleh(adminRole, "permohonan:lihat") {
		return "", errors.New("role tidak diizinkan memproses hasil mutu")
	}
	lapor, err := s.st.GetLaporanMutu(ctx, sampelID)
	if err != nil {
		return "", fmt.Errorf("ambil laporan mutu: %w", err)
	}
	if lapor == nil {
		return "", ErrNotFound
	}
	if lapor.Kesimpulan == "lulus" {
		return "lulus", nil
	}
	if lapor.Kesimpulan == "tidak_lulus" {
		if err := s.JalurTidakLulus(ctx, sampelID, adminID, adminRole); err != nil {
			return "", fmt.Errorf("jalur tidak lulus: %w", err)
		}
		return "tidak_lulus", nil
	}
	return "", errors.New("kesimpulan laporan mutu tidak valid")
}

// AmbilSampel mencatat pengambilan sampel benih oleh Pengawas (Fase 2).
func (s *Service) AmbilSampel(ctx context.Context, permohonanID int64, pengawasID int64, pengawasRole string, jumlahSampel int, tanggalPengambilan string) (*db.SampelBenih, error) {
	if !boleh(pengawasRole, "inspeksi:catat") {
		return nil, errors.New("role tidak diizinkan mengambil sampel")
	}
	p, err := s.st.GetPermohonan(ctx, permohonanID)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	if p.Status != "diinspeksi" {
		return nil, errors.New("sampel hanya dapat diambil setelah buku induk dicatat")
	}
	samp := db.SampelBenih{
		PermohonanID:       permohonanID,
		JumlahSampel:       jumlahSampel,
		TanggalPengambilan: tanggalPengambilan,
		PengawasID:         pengawasID,
		Status:             "dikumpulkan",
	}
	id, err := s.st.CreateSampel(ctx, samp)
	if err != nil {
		return nil, fmt.Errorf("buat sampel: %w", err)
	}
	samp.ID = id
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &pengawasID,
		ActorRole: pengawasRole,
		Action:    "sampel:ambil",
		Entity:    "sampel_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("permohonan_id=%d jumlah=%d", permohonanID, jumlahSampel),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return &samp, nil
}

// KodifikasiSampel memberi kode unik pengujian pada sampel oleh Analis.
func (s *Service) KodifikasiSampel(ctx context.Context, sampelID int64, kodeSampel string, analisID int64, analisRole string) error {
	if !boleh(analisRole, "permohonan:lihat") {
		return errors.New("role tidak diizinkan mengkodifikasi sampel")
	}
	if kodeSampel == "" {
		return errors.New("kode sampel wajib diisi")
	}
	samp, err := s.st.GetSampel(ctx, sampelID)
	if err != nil {
		return fmt.Errorf("ambil sampel: %w", err)
	}
	if samp == nil {
		return ErrNotFound
	}
	if samp.Status != "dikumpulkan" {
		return errors.New("sampel belum dalam status dikumpulkan")
	}
	if err := s.st.KodifikasiSampel(ctx, sampelID, kodeSampel, analisID); err != nil {
		return fmt.Errorf("kodifikasi: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &analisID,
		ActorRole: analisRole,
		Action:    "sampel:kodifikasi",
		Entity:    "sampel_benih",
		EntityID:  &sampelID,
		Detail:    fmt.Sprintf("kode_sampel=%s", kodeSampel),
	}); err != nil {
		return fmt.Errorf("catat audit: %w", err)
	}
	return nil
}

// InputHasilUji mencatat hasil pengujian lab oleh Analis.
func (s *Service) InputHasilUji(ctx context.Context, sampelID int64, parameter string, hasil string, statusLulus bool, catatan string, diujiOleh int64, diujiOlehRole string, tanggalUji string) (*db.UjiMutu, error) {
	if !boleh(diujiOlehRole, "permohonan:lihat") {
		return nil, errors.New("role tidak diizinkan input hasil uji")
	}
	samp, err := s.st.GetSampel(ctx, sampelID)
	if err != nil {
		return nil, fmt.Errorf("ambil sampel: %w", err)
	}
	if samp == nil {
		return nil, ErrNotFound
	}
	if samp.Status != "dikodekan" {
		return nil, errors.New("sampel harus dikodekan sebelum diuji")
	}
	u := db.UjiMutu{
		SampelID:    sampelID,
		Parameter:   parameter,
		Hasil:       hasil,
		StatusLulus: statusLulus,
		Catatan:     catatan,
		DiujiOleh:   diujiOleh,
		TanggalUji:  tanggalUji,
	}
	id, err := s.st.CreateUjiMutu(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("buat uji mutu: %w", err)
	}
	u.ID = id
	if _, err := s.st.ListUjiMutu(ctx, sampelID); err != nil {
		return nil, fmt.Errorf("daftar uji mutu: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &diujiOleh,
		ActorRole: diujiOlehRole,
		Action:    "uji_mutu:input",
		Entity:    "uji_mutu",
		EntityID:  &id,
		Detail:    fmt.Sprintf("parameter=%s lulus=%v", parameter, statusLulus),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return &u, nil
}

// SusunLaporanMutu menyusun kesimpulan pengujian mutu oleh Analis.
func (s *Service) SusunLaporanMutu(ctx context.Context, sampelID int64, kesimpulan string, isiLaporan string, disusunOleh int64, disusunOlehRole string) (*db.LaporanMutu, error) {
	if !boleh(disusunOlehRole, "permohonan:lihat") {
		return nil, errors.New("role tidak diizinkan menyusun laporan mutu")
	}
	if kesimpulan != "lulus" && kesimpulan != "tidak_lulus" {
		return nil, errors.New("kesimpulan harus lulus atau tidak_lulus")
	}
	samp, err := s.st.GetSampel(ctx, sampelID)
	if err != nil {
		return nil, fmt.Errorf("ambil sampel: %w", err)
	}
	if samp == nil {
		return nil, ErrNotFound
	}
	uji, err := s.st.ListUjiMutu(ctx, sampelID)
	if err != nil {
		return nil, fmt.Errorf("daftar uji mutu: %w", err)
	}
	if len(uji) == 0 {
		return nil, errors.New("belum ada hasil uji mutu untuk sampel ini")
	}
	l := db.LaporanMutu{
		SampelID:    sampelID,
		Kesimpulan:  kesimpulan,
		IsiLaporan:  isiLaporan,
		DisusunOleh: disusunOleh,
	}
	id, err := s.st.CreateLaporanMutu(ctx, l)
	if err != nil {
		return nil, fmt.Errorf("buat laporan mutu: %w", err)
	}
	l.ID = id
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &disusunOleh,
		ActorRole: disusunOlehRole,
		Action:    "laporan_mutu:susun",
		Entity:    "laporan_mutu",
		EntityID:  &id,
		Detail:    fmt.Sprintf("kesimpulan=%s", kesimpulan),
	}); err != nil {
		return nil, fmt.Errorf("catat audit: %w", err)
	}
	return &l, nil
}
