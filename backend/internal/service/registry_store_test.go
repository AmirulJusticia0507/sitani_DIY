package service

import (
	"context"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

const (
	actorID    = int64(7)
	tanggalUji = "2026-10-23"
)

// registryStore menyediakan data minimum untuk menguji aturan bisnis
// registrasi, persediaan, dan penjualan benih. Semua pemanggilan dicatat agar
// tes dapat memastikan tidak ada perubahan data yang terjadi saat validasi gagal.
type registryStore struct {
	fakeStore

	permohonan *db.PermohonanBenih
	registrasi map[int64]db.RegistrasiBenih
	stok       map[int64]db.StokBenih
	mutasi     []db.MutasiStokBenih
	penjualan  map[int64]db.PenjualanBenih
	sertifikat *db.SertifikatBenih
	label      *db.DraftLabel

	dibuatRegistrasi int
	dibuatPenjualan  int
	audit            []db.AuditLog
}

func newRegistryStore() *registryStore {
	return &registryStore{
		permohonan: &db.PermohonanBenih{ID: 1, Status: "lulus"},
		registrasi: map[int64]db.RegistrasiBenih{},
		stok:       map[int64]db.StokBenih{},
		penjualan:  map[int64]db.PenjualanBenih{},
		sertifikat: &db.SertifikatBenih{
			ID:              10,
			PermohonanID:    1,
			NomorSertifikat: "SRT-001",
			TTEStatus:       "ditandatangani",
		},
		label: &db.DraftLabel{
			ID:                 20,
			Status:             "legalisasi",
			IntegrasiSIMStatus: "terkirim",
			KodeBarcode:        "SB-2026-0001",
		},
	}
}

func (s *registryStore) GetPermohonan(context.Context, int64) (*db.PermohonanBenih, error) {
	return s.permohonan, nil
}

func (s *registryStore) GetSertifikat(context.Context, int64) (*db.SertifikatBenih, error) {
	return s.sertifikat, nil
}

func (s *registryStore) GetDraftLabel(context.Context, int64) (*db.DraftLabel, error) {
	return s.label, nil
}

func (s *registryStore) GetRegistrasiBenihByNomorLot(_ context.Context, nomorLot string) (*db.RegistrasiBenih, error) {
	for _, r := range s.registrasi {
		if r.NomorLot == nomorLot {
			salin := r
			return &salin, nil
		}
	}
	return nil, nil
}

func (s *registryStore) CreateRegistrasiBenih(_ context.Context, r db.RegistrasiBenih) (int64, error) {
	s.dibuatRegistrasi++
	id := int64(len(s.registrasi) + 1)
	r.ID = id
	s.registrasi[id] = r
	s.stok[id] = db.StokBenih{
		ID:                 id,
		RegistrasiID:       id,
		NomorLot:           r.NomorLot,
		JumlahTersedia:     r.Jumlah,
		Satuan:             r.Satuan,
		TanggalKedaluwarsa: r.TanggalKedaluwarsa,
		StatusEdar:         r.StatusEdar,
	}
	return id, nil
}

func (s *registryStore) GetRegistrasiBenih(_ context.Context, id int64) (*db.RegistrasiBenih, error) {
	r, ok := s.registrasi[id]
	if !ok {
		return nil, nil
	}
	return &r, nil
}

func (s *registryStore) GetStokBenih(_ context.Context, registrasiID int64) (*db.StokBenih, error) {
	st, ok := s.stok[registrasiID]
	if !ok {
		return nil, nil
	}
	return &st, nil
}

func (s *registryStore) ListMutasiStok(_ context.Context, registrasiID int64) ([]db.MutasiStokBenih, error) {
	var hasil []db.MutasiStokBenih
	for _, m := range s.mutasi {
		if m.RegistrasiID == registrasiID {
			hasil = append(hasil, m)
		}
	}
	return hasil, nil
}

func (s *registryStore) CreatePenjualanBenih(_ context.Context, p db.PenjualanBenih) (*db.PenjualanBenih, error) {
	stok := s.stok[p.RegistrasiID]
	if stok.JumlahTersedia < p.Jumlah {
		return nil, store.ErrStokTidakCukup
	}
	s.dibuatPenjualan++
	p.ID = int64(len(s.penjualan) + 1)
	stok.JumlahTersedia -= p.Jumlah
	s.stok[p.RegistrasiID] = stok
	s.penjualan[p.ID] = p
	s.mutasi = append(s.mutasi, db.MutasiStokBenih{
		RegistrasiID: p.RegistrasiID,
		Jenis:        "keluar",
		Jumlah:       p.Jumlah,
		SaldoSetelah: stok.JumlahTersedia,
	})
	salin := p
	return &salin, nil
}

func (s *registryStore) BatalkanPenjualanBenih(_ context.Context, id int64, actorID int64, alasan string, waktu time.Time) (*db.PenjualanBenih, error) {
	p, ok := s.penjualan[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	if p.Status == "dibatalkan" {
		return nil, store.ErrSudahDibatalkan
	}
	p.Status = "dibatalkan"
	p.AlasanPembatalan = alasan
	p.DibatalkanOleh = &actorID
	p.DibatalkanPada = &waktu
	stok := s.stok[p.RegistrasiID]
	stok.JumlahTersedia += p.Jumlah
	s.stok[p.RegistrasiID] = stok
	s.penjualan[id] = p
	s.mutasi = append(s.mutasi, db.MutasiStokBenih{
		RegistrasiID: p.RegistrasiID,
		Jenis:        "koreksi",
		Jumlah:       p.Jumlah,
		SaldoSetelah: stok.JumlahTersedia,
	})
	salin := p
	return &salin, nil
}

func (s *registryStore) CreateAuditLog(_ context.Context, a db.AuditLog) (int64, error) {
	s.audit = append(s.audit, a)
	return int64(len(s.audit)), nil
}

// serviceUji membuat service dengan jam tetap 23 Oktober 2026 agar validasi
// tanggal deterministik dan tidak bergantung pada waktu mesin.
func serviceUji(st *registryStore) *Service {
	svc := NewService(st)
	svc.now = func() time.Time {
		return time.Date(2026, 10, 23, 9, 0, 0, 0, time.Local)
	}
	return svc
}

func registrasiValid() BuatRegistrasiRequest {
	return BuatRegistrasiRequest{
		PermohonanID:       1,
		Komoditas:          "  Tomat ",
		Varietas:           "Rama",
		Produsen:           "Kelompok Tani Sleman",
		SumberBenih:        "BP3MBTP",
		NomorLot:           " LOT-001 ",
		KelasBenih:         db.KelasBenihSebar,
		Jumlah:             100,
		Satuan:             " kg ",
		TanggalProduksi:    "2026-09-01",
		TanggalKedaluwarsa: "2027-03-01",
	}
}
