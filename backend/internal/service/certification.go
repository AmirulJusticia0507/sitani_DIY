package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// ErrNotFound dikembalikan saat data yang diminta tidak tersedia.
var ErrNotFound = errors.New("data tidak ditemukan")

// FaseInspeksi adalah jumlah fase inspeksi lapangan yang wajib diselesaikan.
const FaseInspeksi = 4

// Izin per role untuk Fase 1 verifikasi lapangan.
var izinRole = map[string]map[string]bool{
	db.RolePelanggan: {
		"permohonan:buat": true,
		"permohonan:lihat": true,
		"pembayaran:awal": true,
	},
	db.RolePengawas: {
		"permohonan:lihat": true,
		"inspeksi:catat": true,
		"buku_induk:catat": true,
	},
	db.RoleAdmin: {
		"permohonan:buat": true,
		"permohonan:lihat": true,
		"pembayaran:awal": true,
		"inspeksi:catat": true,
		"buku_induk:catat": true,
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
func (s *Service) BuatPermohonan(ctx context.Context, pelangganID int64, pelangganRole string, req BuatPermohonanRequest) (*db.PermohonanBenih, error) {
	if !boleh(pelangganRole, "permohonan:buat") {
		return nil, errors.New("role tidak diizinkan membuat permohonan")
	}
	if req.JenisBenih == "" {
		return nil, errors.New("jenis benih wajib diisi")
	}

	now := s.now()
	p := db.PermohonanBenih{
		PelangganID:          pelangganID,
		JenisBenih:           req.JenisBenih,
		Varietas:             req.Varietas,
		LuasLahan:            req.LuasLahan,
		LokasiKebun:          req.LokasiKebun,
		Status:               "diajukan",
		PembayaranAwalStatus: "belum",
	}
	id, err := s.st.CreatePermohonan(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("buat permohonan: %w", err)
	}
	p.ID = id
	p.NomorPermohonan = fmt.Sprintf("SPB-%d-%d", id, now.Unix())
	if err := s.st.UpdatePermohonanStatus(ctx, id, "diajukan", "belum"); err != nil {
		return nil, fmt.Errorf("set nomor permohonan: %w", err)
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
	p.NomorPermohonan = fmt.Sprintf("SPB-%d-%d", id, now.Unix())
	return &p, nil
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
	if p.Status != "dibayar" {
		return errors.New("inspeksi hanya dapat dilakukan setelah pembayaran awal lunas")
	}

	i := db.InspeksiFase{
		PermohonanID:     permohonanID,
		Fase:             fase,
		PengawasID:       pengawasID,
		Hasil:            hasil,
		Catatan:          catatan,
		TanggalInspeksi:  tanggalInspeksi,
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
	JenisBenih   string
	Varietas     string
	LuasLahan    float64
	LokasiKebun  string
}
