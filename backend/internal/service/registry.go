package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

const (
	batasHariKedaluwarsa = 3650
	tanggalFormat        = "2006-01-02"
)

// BuatRegistrasiRequest adalah input pendaftaran lot benih yang sudah tersertifikasi.
type BuatRegistrasiRequest struct {
	PermohonanID       int64
	Komoditas          string
	Varietas           string
	Produsen           string
	SumberBenih        string
	NomorLot           string
	KelasBenih         string
	Jumlah             float64
	Satuan             string
	TanggalProduksi    string
	TanggalKedaluwarsa string
	StatusEdar         string
}

// JualBenihRequest adalah input transaksi penjualan lot benih.
type JualBenihRequest struct {
	NomorDokumen     string
	RegistrasiID     int64
	Pembeli          string
	Jumlah           float64
	HargaSatuan      float64
	TanggalTransaksi string
}

// BuatRegistrasiBenih mendaftarkan lot benih yang telah memiliki sertifikat
// bertanda tangan TTE dan label yang sudah dilegalisasi serta terkirim ke SIM Perbenihan.
func (s *Service) BuatRegistrasiBenih(ctx context.Context, actorID int64, actorRole string, req BuatRegistrasiRequest) (*db.RegistrasiBenih, error) {
	if !boleh(actorRole, "registrasi:buat") {
		return nil, errors.New("role tidak diizinkan mendaftarkan benih")
	}
	req.Komoditas = strings.TrimSpace(req.Komoditas)
	req.Varietas = strings.TrimSpace(req.Varietas)
	req.Produsen = strings.TrimSpace(req.Produsen)
	req.SumberBenih = strings.TrimSpace(req.SumberBenih)
	req.NomorLot = strings.TrimSpace(req.NomorLot)
	req.Satuan = strings.TrimSpace(req.Satuan)
	if req.PermohonanID <= 0 {
		return nil, errors.New("permohonan_id wajib diisi")
	}
	if req.Komoditas == "" || req.Varietas == "" || req.Produsen == "" || req.SumberBenih == "" || req.NomorLot == "" || req.Satuan == "" {
		return nil, errors.New("komoditas, varietas, produsen, sumber benih, nomor lot, dan satuan wajib diisi")
	}
	if !kelasBenihValid(req.KelasBenih) {
		return nil, errors.New("kelas benih harus penjenis, dasar, pokok, atau sebar")
	}
	if req.Jumlah <= 0 {
		return nil, errors.New("jumlah benih harus lebih dari nol")
	}
	if req.StatusEdar == "" {
		req.StatusEdar = db.StatusEdarAktif
	}
	if !statusEdarValid(req.StatusEdar) {
		return nil, errors.New("status edar tidak valid")
	}

	produksi, err := parseTanggal(req.TanggalProduksi)
	if err != nil {
		return nil, errors.New("tanggal produksi harus berformat YYYY-MM-DD")
	}
	kedaluwarsa, err := parseTanggal(req.TanggalKedaluwarsa)
	if err != nil {
		return nil, errors.New("tanggal kedaluwarsa harus berformat YYYY-MM-DD")
	}
	hariIni := awalHari(s.now())
	if produksi.After(hariIni) {
		return nil, errors.New("tanggal produksi tidak boleh di masa depan")
	}
	if !kedaluwarsa.After(produksi) {
		return nil, errors.New("tanggal kedaluwarsa harus setelah tanggal produksi")
	}
	if kedaluwarsa.Before(hariIni) {
		return nil, errors.New("lot benih sudah kedaluwarsa")
	}

	permohonan, err := s.st.GetPermohonan(ctx, req.PermohonanID)
	if err != nil {
		return nil, fmt.Errorf("ambil permohonan: %w", err)
	}
	if permohonan == nil {
		return nil, ErrNotFound
	}
	if permohonan.Status == db.StatusPermohonanDitolak || permohonan.Status == db.StatusPermohonanDibatalkan {
		return nil, errors.New("permohonan sudah ditolak atau dibatalkan dan tidak dapat diregistrasi")
	}
	sertifikat, label, err := s.validasiDokumenSertifikasi(ctx, req.PermohonanID)
	if err != nil {
		return nil, err
	}
	existing, err := s.st.GetRegistrasiBenihByNomorLot(ctx, req.NomorLot)
	if err != nil {
		return nil, fmt.Errorf("periksa nomor lot: %w", err)
	}
	if existing != nil {
		return nil, errors.New("nomor lot sudah terdaftar")
	}

	registrasi := db.RegistrasiBenih{
		NomorRegistrasi:    "RB-" + req.NomorLot,
		PermohonanID:       req.PermohonanID,
		SertifikatID:       &sertifikat.ID,
		DraftLabelID:       &label.ID,
		NomorSertifikat:    sertifikat.NomorSertifikat,
		KodeLabel:          label.KodeBarcode,
		Komoditas:          req.Komoditas,
		Varietas:           req.Varietas,
		Produsen:           req.Produsen,
		SumberBenih:        req.SumberBenih,
		NomorLot:           req.NomorLot,
		KelasBenih:         req.KelasBenih,
		Jumlah:             req.Jumlah,
		Satuan:             req.Satuan,
		TanggalProduksi:    produksi.Format(tanggalFormat),
		TanggalKedaluwarsa: kedaluwarsa.Format(tanggalFormat),
		StatusEdar:         req.StatusEdar,
		CreatedBy:          actorID,
	}
	id, err := s.st.CreateRegistrasiBenih(ctx, registrasi)
	if err != nil {
		return nil, fmt.Errorf("simpan registrasi benih: %w", err)
	}
	registrasi.ID = id
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &actorID,
		ActorRole: actorRole,
		Action:    "registrasi_benih:buat",
		Entity:    "registrasi_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("nomor_lot=%s jumlah=%.3f %s", registrasi.NomorLot, registrasi.Jumlah, registrasi.Satuan),
	}); err != nil {
		return nil, fmt.Errorf("catat audit registrasi: %w", err)
	}
	return &registrasi, nil
}

// ListPermohonanSiapRegistrasi mengembalikan pilihan permohonan untuk form
// registrasi lot. Hanya permohonan yang sertifikatnya sudah ditandatangani TTE,
// labelnya sudah dilegalisasi dan terkirim ke SIM Perbenihan, serta belum
// memiliki lot terdaftar. Pelanggan hanya melihat permohonannya sendiri.
func (s *Service) ListPermohonanSiapRegistrasi(ctx context.Context, actorID int64, actorRole string) ([]db.SiapRegistrasiPermohonan, error) {
	if !boleh(actorRole, "registrasi:lihat") && !boleh(actorRole, "permohonan:lihat") {
		return nil, errors.New("role tidak diizinkan melihat permohonan")
	}
	semua := actorRole == db.RoleAdmin
	items, err := s.st.ListPermohonanSiapRegistrasi(ctx, actorID, semua, batasLimit(0))
	if err != nil {
		return nil, fmt.Errorf("daftar permohonan siap registrasi: %w", err)
	}
	return items, nil
}

// GetRegistrasiBenih mengambil detail satu lot benih.
func (s *Service) GetRegistrasiBenih(ctx context.Context, actorRole string, id int64) (*db.RegistrasiBenih, error) {
	if !boleh(actorRole, "registrasi:lihat") {
		return nil, errors.New("role tidak diizinkan melihat registrasi benih")
	}
	registrasi, err := s.st.GetRegistrasiBenih(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ambil registrasi benih: %w", err)
	}
	if registrasi == nil {
		return nil, ErrNotFound
	}
	return registrasi, nil
}

// ListRegistrasiBenih mengembalikan daftar lot benih sesuai filter.
func (s *Service) ListRegistrasiBenih(ctx context.Context, actorRole string, filter db.RegistrasiFilter) ([]db.RegistrasiBenih, error) {
	if !boleh(actorRole, "registrasi:lihat") {
		return nil, errors.New("role tidak diizinkan melihat registrasi benih")
	}
	if filter.StatusEdar != "" && !statusEdarValid(filter.StatusEdar) {
		return nil, errors.New("status edar tidak valid")
	}
	filter.Limit = batasLimit(filter.Limit)
	items, err := s.st.ListRegistrasiBenih(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("daftar registrasi benih: %w", err)
	}
	return items, nil
}

// UbahStatusEdar mengubah status edar lot (misalnya ditangguhkan atau dicabut).
func (s *Service) UbahStatusEdar(ctx context.Context, actorID int64, actorRole string, id int64, status string) error {
	if !boleh(actorRole, "registrasi:ubah_status") {
		return errors.New("role tidak diizinkan mengubah status edar")
	}
	if !statusEdarValid(status) {
		return errors.New("status edar tidak valid")
	}
	registrasi, err := s.st.GetRegistrasiBenih(ctx, id)
	if err != nil {
		return fmt.Errorf("ambil registrasi benih: %w", err)
	}
	if registrasi == nil {
		return ErrNotFound
	}
	if err := s.st.UpdateRegistrasiStatus(ctx, id, status); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("ubah status edar: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &actorID,
		ActorRole: actorRole,
		Action:    "registrasi_benih:ubah_status",
		Entity:    "registrasi_benih",
		EntityID:  &id,
		Detail:    fmt.Sprintf("status_edar=%s", status),
	}); err != nil {
		return fmt.Errorf("catat audit status edar: %w", err)
	}
	return nil
}

// StokBenih mengembalikan saldo stok dan riwayat mutasinya.
func (s *Service) StokBenih(ctx context.Context, actorRole string, registrasiID int64) (*db.StokBenih, []db.MutasiStokBenih, error) {
	if !boleh(actorRole, "stok:lihat") {
		return nil, nil, errors.New("role tidak diizinkan melihat stok benih")
	}
	stok, err := s.st.GetStokBenih(ctx, registrasiID)
	if err != nil {
		return nil, nil, fmt.Errorf("ambil stok benih: %w", err)
	}
	if stok == nil {
		return nil, nil, ErrNotFound
	}
	mutasi, err := s.st.ListMutasiStok(ctx, registrasiID)
	if err != nil {
		return nil, nil, fmt.Errorf("daftar mutasi stok: %w", err)
	}
	return stok, mutasi, nil
}

// JualBenih mencatat penjualan setelah memvalidasi registrasi, sertifikat,
// label, masa berlaku, status edar, dan kecukupan stok.
func (s *Service) JualBenih(ctx context.Context, actorID int64, actorRole string, req JualBenihRequest) (*db.PenjualanBenih, error) {
	if !boleh(actorRole, "penjualan:buat") {
		return nil, errors.New("role tidak diizinkan mencatat penjualan benih")
	}
	req.NomorDokumen = strings.TrimSpace(req.NomorDokumen)
	req.Pembeli = strings.TrimSpace(req.Pembeli)
	if req.NomorDokumen == "" || req.Pembeli == "" {
		return nil, errors.New("nomor dokumen dan pembeli wajib diisi")
	}
	if req.RegistrasiID <= 0 {
		return nil, errors.New("registrasi_id wajib diisi")
	}
	if req.Jumlah <= 0 {
		return nil, errors.New("jumlah penjualan harus lebih dari nol")
	}
	if req.HargaSatuan < 0 {
		return nil, errors.New("harga satuan tidak boleh negatif")
	}
	tanggal, err := parseTanggal(req.TanggalTransaksi)
	if err != nil {
		return nil, errors.New("tanggal transaksi harus berformat YYYY-MM-DD")
	}
	hariIni := awalHari(s.now())
	if tanggal.After(hariIni) {
		return nil, errors.New("tanggal transaksi tidak boleh di masa depan")
	}

	registrasi, err := s.st.GetRegistrasiBenih(ctx, req.RegistrasiID)
	if err != nil {
		return nil, fmt.Errorf("ambil registrasi benih: %w", err)
	}
	if registrasi == nil {
		return nil, ErrNotFound
	}
	if registrasi.StatusEdar != db.StatusEdarAktif {
		return nil, fmt.Errorf("lot benih tidak aktif (status edar: %s)", registrasi.StatusEdar)
	}
	kedaluwarsa, err := parseTanggal(registrasi.TanggalKedaluwarsa)
	if err != nil {
		return nil, errors.New("tanggal kedaluwarsa registrasi tidak valid")
	}
	if kedaluwarsa.Before(hariIni) {
		return nil, errors.New("masa berlaku lot benih sudah berakhir")
	}

	sertifikat, label, err := s.validasiDokumenSertifikasi(ctx, registrasi.PermohonanID)
	if err != nil {
		return nil, err
	}
	if registrasi.SertifikatID == nil || *registrasi.SertifikatID != sertifikat.ID {
		return nil, errors.New("sertifikat lot tidak sesuai dengan registrasi")
	}
	if registrasi.DraftLabelID == nil || *registrasi.DraftLabelID != label.ID {
		return nil, errors.New("label lot tidak sesuai dengan registrasi")
	}

	stok, err := s.st.GetStokBenih(ctx, req.RegistrasiID)
	if err != nil {
		return nil, fmt.Errorf("ambil stok benih: %w", err)
	}
	if stok == nil {
		return nil, ErrNotFound
	}
	if stok.JumlahTersedia < req.Jumlah {
		return nil, store.ErrStokTidakCukup
	}

	penjualan := db.PenjualanBenih{
		NomorDokumen:     req.NomorDokumen,
		RegistrasiID:     req.RegistrasiID,
		NomorLot:         registrasi.NomorLot,
		Pembeli:          req.Pembeli,
		Jumlah:           req.Jumlah,
		HargaSatuan:      req.HargaSatuan,
		TotalHarga:       math.Round(req.Jumlah*req.HargaSatuan*100) / 100,
		TanggalTransaksi: tanggal.Format(tanggalFormat),
		Status:           "selesai",
		CreatedBy:        actorID,
	}
	saved, err := s.st.CreatePenjualanBenih(ctx, penjualan)
	if err != nil {
		if errors.Is(err, store.ErrStokTidakCukup) {
			return nil, store.ErrStokTidakCukup
		}
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("simpan penjualan: %w", err)
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &actorID,
		ActorRole: actorRole,
		Action:    "penjualan_benih:buat",
		Entity:    "penjualan_benih",
		EntityID:  &saved.ID,
		Detail:    fmt.Sprintf("dokumen=%s lot=%s jumlah=%.3f total=%.2f", saved.NomorDokumen, saved.NomorLot, saved.Jumlah, saved.TotalHarga),
	}); err != nil {
		return nil, fmt.Errorf("catat audit penjualan: %w", err)
	}
	return saved, nil
}

// BatalkanPenjualan membatalkan transaksi tanpa menghapus histori dan
// mengembalikan jumlah ke stok melalui mutasi koreksi.
func (s *Service) BatalkanPenjualan(ctx context.Context, actorID int64, actorRole string, penjualanID int64, alasan string) (*db.PenjualanBenih, error) {
	if !boleh(actorRole, "penjualan:batal") {
		return nil, errors.New("role tidak diizinkan membatalkan penjualan")
	}
	alasan = strings.TrimSpace(alasan)
	if alasan == "" {
		return nil, errors.New("alasan pembatalan wajib diisi")
	}
	saved, err := s.st.BatalkanPenjualanBenih(ctx, penjualanID, actorID, alasan, s.now())
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, ErrNotFound
		case errors.Is(err, store.ErrSudahDibatalkan):
			return nil, store.ErrSudahDibatalkan
		default:
			return nil, fmt.Errorf("batalkan penjualan: %w", err)
		}
	}
	if _, err := s.st.CreateAuditLog(ctx, db.AuditLog{
		ActorID:   &actorID,
		ActorRole: actorRole,
		Action:    "penjualan_benih:batal",
		Entity:    "penjualan_benih",
		EntityID:  &penjualanID,
		Detail:    fmt.Sprintf("alasan=%s", alasan),
	}); err != nil {
		return nil, fmt.Errorf("catat audit pembatalan: %w", err)
	}
	return saved, nil
}

// LaporanRegistrasi mengembalikan laporan registrasi lot benih.
func (s *Service) LaporanRegistrasi(ctx context.Context, actorRole string, filter db.RegistrasiFilter) ([]db.RegistrasiBenih, error) {
	if filter.Limit <= 0 {
		filter.Limit = 1000
	}
	return s.ListRegistrasiBenih(ctx, actorRole, filter)
}

// LaporanStok mengembalikan laporan persediaan seluruh lot.
func (s *Service) LaporanStok(ctx context.Context, actorRole string) ([]db.StokBenih, error) {
	if !boleh(actorRole, "laporan:lihat") {
		return nil, errors.New("role tidak diizinkan melihat laporan")
	}
	items, err := s.st.ListStokBenih(ctx)
	if err != nil {
		return nil, fmt.Errorf("laporan stok: %w", err)
	}
	return items, nil
}

// LaporanKedaluwarsa mengembalikan lot yang kedaluwarsa dalam jumlah hari tertentu.
func (s *Service) LaporanKedaluwarsa(ctx context.Context, actorRole string, hari int) ([]db.RegistrasiBenih, error) {
	if !boleh(actorRole, "laporan:lihat") {
		return nil, errors.New("role tidak diizinkan melihat laporan")
	}
	if hari <= 0 {
		hari = 90
	}
	if hari > batasHariKedaluwarsa {
		return nil, fmt.Errorf("rentang hari maksimal %d", batasHariKedaluwarsa)
	}
	batas := awalHari(s.now()).AddDate(0, 0, hari)
	return s.st.ListRegistrasiBenih(ctx, db.RegistrasiFilter{KedaluwarsaSebelum: &batas, Limit: 1000})
}

// LaporanPenjualan mengembalikan transaksi penjualan sesuai filter.
func (s *Service) LaporanPenjualan(ctx context.Context, actorRole string, filter db.PenjualanFilter) ([]db.PenjualanBenih, error) {
	if !boleh(actorRole, "laporan:lihat") {
		return nil, errors.New("role tidak diizinkan melihat laporan")
	}
	if filter.Status != "" && filter.Status != "selesai" && filter.Status != "dibatalkan" {
		return nil, errors.New("status penjualan harus selesai atau dibatalkan")
	}
	filter.Limit = batasLimit(filter.Limit)
	items, err := s.st.ListPenjualanBenih(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("laporan penjualan: %w", err)
	}
	return items, nil
}

// validasiDokumenSertifikasi memastikan sertifikat sudah ditandatangani TTE dan
// label sudah dilegalisasi serta terkirim ke SIM Perbenihan.
func (s *Service) validasiDokumenSertifikasi(ctx context.Context, permohonanID int64) (*db.SertifikatBenih, *db.DraftLabel, error) {
	sertifikat, err := s.st.GetSertifikat(ctx, permohonanID)
	if err != nil {
		return nil, nil, fmt.Errorf("ambil sertifikat: %w", err)
	}
	if sertifikat == nil {
		return nil, nil, errors.New("sertifikat benih belum diterbitkan")
	}
	if sertifikat.TTEStatus != "ditandatangani" {
		return nil, nil, errors.New("sertifikat belum ditandatangani TTE")
	}
	label, err := s.st.GetDraftLabel(ctx, sertifikat.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("ambil label: %w", err)
	}
	if label == nil {
		return nil, nil, errors.New("label benih belum dibuat")
	}
	if label.Status != "legalisasi" {
		return nil, nil, errors.New("label belum dilegalisasi")
	}
	if label.IntegrasiSIMStatus != "terkirim" {
		return nil, nil, errors.New("label belum terintegrasi ke SIM Perbenihan")
	}
	return sertifikat, label, nil
}

func parseTanggal(value string) (time.Time, error) {
	return time.ParseInLocation(tanggalFormat, strings.TrimSpace(value), time.Local)
}

func awalHari(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func kelasBenihValid(kelas string) bool {
	switch kelas {
	case db.KelasBenihPenjenis, db.KelasBenihDasar, db.KelasBenihPokok, db.KelasBenihSebar:
		return true
	default:
		return false
	}
}

func statusEdarValid(status string) bool {
	switch status {
	case db.StatusEdarAktif, db.StatusEdarDitangguhkan, db.StatusEdarDicabut, db.StatusEdarKedaluwarsa:
		return true
	default:
		return false
	}
}

func batasLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}
