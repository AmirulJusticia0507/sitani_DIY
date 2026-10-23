package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

func TestBuatRegistrasiBenihMenyimpanDataSah(t *testing.T) {
	st := newRegistryStore()
	got, err := serviceUji(st).BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid())
	if err != nil {
		t.Fatalf("buat registrasi: %v", err)
	}
	if got.NomorRegistrasi != "RB-LOT-001" || got.NomorLot != "LOT-001" {
		t.Fatalf("nomor registrasi/lot = %q/%q", got.NomorRegistrasi, got.NomorLot)
	}
	if got.Komoditas != "Tomat" || got.Satuan != "kg" {
		t.Fatalf("input tidak dirapikan: %q/%q", got.Komoditas, got.Satuan)
	}
	if got.StatusEdar != db.StatusEdarAktif {
		t.Fatalf("status edar bawaan = %q, ingin aktif", got.StatusEdar)
	}
	if got.SertifikatID == nil || *got.SertifikatID != 10 || got.KodeLabel != "SB-2026-0001" {
		t.Fatalf("dokumen sertifikasi tidak terhubung: %+v", got)
	}
	if st.stok[got.ID].JumlahTersedia != 100 {
		t.Fatalf("stok awal = %v, ingin 100", st.stok[got.ID].JumlahTersedia)
	}
	if len(st.audit) != 1 || st.audit[0].Action != "registrasi_benih:buat" {
		t.Fatalf("audit registrasi tidak tercatat: %+v", st.audit)
	}
}

func TestBuatRegistrasiBenihMenolakInputTidakSah(t *testing.T) {
	kasus := []struct {
		nama  string
		ubah  func(*BuatRegistrasiRequest)
		pesan string
	}{
		{"permohonan kosong", func(r *BuatRegistrasiRequest) { r.PermohonanID = 0 }, "permohonan_id wajib diisi"},
		{"field wajib kosong", func(r *BuatRegistrasiRequest) { r.Varietas = "  " }, "wajib diisi"},
		{"kelas tidak dikenal", func(r *BuatRegistrasiRequest) { r.KelasBenih = "unggul" }, "kelas benih"},
		{"jumlah nol", func(r *BuatRegistrasiRequest) { r.Jumlah = 0 }, "jumlah benih"},
		{"status edar asing", func(r *BuatRegistrasiRequest) { r.StatusEdar = "karantina" }, "status edar"},
		{"format produksi", func(r *BuatRegistrasiRequest) { r.TanggalProduksi = "01-09-2026" }, "tanggal produksi"},
		{"format kedaluwarsa", func(r *BuatRegistrasiRequest) { r.TanggalKedaluwarsa = "" }, "tanggal kedaluwarsa"},
		{"produksi masa depan", func(r *BuatRegistrasiRequest) { r.TanggalProduksi = "2026-10-24" }, "masa depan"},
		{"kedaluwarsa sebelum produksi", func(r *BuatRegistrasiRequest) { r.TanggalKedaluwarsa = "2026-09-01" }, "setelah tanggal produksi"},
		{"lot kedaluwarsa", func(r *BuatRegistrasiRequest) { r.TanggalKedaluwarsa = "2026-10-22" }, "sudah kedaluwarsa"},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			st := newRegistryStore()
			req := registrasiValid()
			k.ubah(&req)
			_, err := serviceUji(st).BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, req)
			if err == nil || !strings.Contains(err.Error(), k.pesan) {
				t.Fatalf("error = %v, ingin memuat %q", err, k.pesan)
			}
			if st.dibuatRegistrasi != 0 || len(st.audit) != 0 {
				t.Fatalf("data tetap tersimpan saat validasi gagal")
			}
		})
	}
}

func TestBuatRegistrasiBenihMenolakDokumenBelumSah(t *testing.T) {
	kasus := []struct {
		nama  string
		ubah  func(*registryStore)
		pesan string
	}{
		{"sertifikat belum ada", func(s *registryStore) { s.sertifikat = nil }, "belum diterbitkan"},
		{"sertifikat belum TTE", func(s *registryStore) { s.sertifikat.TTEStatus = "menunggu" }, "belum ditandatangani"},
		{"label belum legalisasi", func(s *registryStore) { s.label.Status = "dicetak" }, "belum dilegalisasi"},
		{"label belum SIM", func(s *registryStore) { s.label.IntegrasiSIMStatus = "gagal" }, "SIM Perbenihan"},
		{"permohonan ditolak", func(s *registryStore) { s.permohonan.Status = db.StatusPermohonanDitolak }, "ditolak atau dibatalkan"},
		{"permohonan dibatalkan", func(s *registryStore) { s.permohonan.Status = db.StatusPermohonanDibatalkan }, "ditolak atau dibatalkan"},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			st := newRegistryStore()
			k.ubah(st)
			_, err := serviceUji(st).BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid())
			if err == nil || !strings.Contains(err.Error(), k.pesan) {
				t.Fatalf("error = %v, ingin memuat %q", err, k.pesan)
			}
			if st.dibuatRegistrasi != 0 {
				t.Fatalf("registrasi tersimpan walau dokumen belum sah")
			}
		})
	}
}

func TestBuatRegistrasiBenihMenolakAktorTidakBerwenang(t *testing.T) {
	st := newRegistryStore()
	_, err := serviceUji(st).BuatRegistrasiBenih(context.Background(), actorID, db.RolePelanggan, registrasiValid())
	if err == nil || !strings.Contains(err.Error(), "tidak diizinkan") {
		t.Fatalf("error = %v, ingin penolakan role", err)
	}
}

func TestBuatRegistrasiBenihMenolakNomorLotGanda(t *testing.T) {
	st := newRegistryStore()
	svc := serviceUji(st)
	if _, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid()); err != nil {
		t.Fatalf("registrasi pertama: %v", err)
	}
	_, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid())
	if err == nil || !strings.Contains(err.Error(), "sudah terdaftar") {
		t.Fatalf("error = %v, ingin penolakan lot ganda", err)
	}
}

func TestBuatRegistrasiBenihPermohonanTidakAda(t *testing.T) {
	st := newRegistryStore()
	st.permohonan = nil
	_, err := serviceUji(st).BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, ingin ErrNotFound", err)
	}
}

func TestJualBenihMengurangiStokDanMencatatAudit(t *testing.T) {
	st := newRegistryStore()
	svc := serviceUji(st)
	if _, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid()); err != nil {
		t.Fatalf("registrasi: %v", err)
	}
	got, err := svc.JualBenih(context.Background(), actorID, db.RoleAdmin, JualBenihRequest{
		NomorDokumen:     "PJ-001",
		RegistrasiID:     1,
		Pembeli:          "Petani Bantul",
		Jumlah:           25,
		HargaSatuan:      25000,
		TanggalTransaksi: tanggalUji,
	})
	if err != nil {
		t.Fatalf("jual benih: %v", err)
	}
	if got.TotalHarga != 625000 {
		t.Fatalf("total = %v, ingin 625000", got.TotalHarga)
	}
	if st.stok[1].JumlahTersedia != 75 {
		t.Fatalf("stok tersisa = %v, ingin 75", st.stok[1].JumlahTersedia)
	}
	last := st.audit[len(st.audit)-1]
	if last.Action != "penjualan_benih:buat" {
		t.Fatalf("audit penjualan = %+v", last)
	}
}

func TestJualBenihMenolakTransaksiTidakSah(t *testing.T) {
	kasus := []struct {
		nama  string
		ubah  func(*registryStore, *JualBenihRequest)
		ingin error
	}{
		{"dokumen kosong", func(_ *registryStore, r *JualBenihRequest) { r.NomorDokumen = " " }, nil},
		{"registrasi kosong", func(_ *registryStore, r *JualBenihRequest) { r.RegistrasiID = 0 }, nil},
		{"jumlah nol", func(_ *registryStore, r *JualBenihRequest) { r.Jumlah = 0 }, nil},
		{"harga negatif", func(_ *registryStore, r *JualBenihRequest) { r.HargaSatuan = -1 }, nil},
		{"tanggal masa depan", func(_ *registryStore, r *JualBenihRequest) { r.TanggalTransaksi = "2026-10-24" }, nil},
		{"stok kurang", func(_ *registryStore, r *JualBenihRequest) { r.Jumlah = 101 }, store.ErrStokTidakCukup},
		{"lot tidak aktif", func(s *registryStore, _ *JualBenihRequest) {
			g := s.registrasi[1]
			g.StatusEdar = db.StatusEdarDitangguhkan
			s.registrasi[1] = g
		}, nil},
		{"lot kedaluwarsa", func(s *registryStore, _ *JualBenihRequest) {
			g := s.registrasi[1]
			g.TanggalKedaluwarsa = "2026-10-22"
			s.registrasi[1] = g
		}, nil},
		{"label tidak cocok", func(s *registryStore, _ *JualBenihRequest) {
			g := s.registrasi[1]
			g.DraftLabelID = nil
			s.registrasi[1] = g
		}, nil},
		{"sertifikat tidak cocok", func(s *registryStore, _ *JualBenihRequest) {
			g := s.registrasi[1]
			g.SertifikatID = nil
			s.registrasi[1] = g
		}, nil},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			st := newRegistryStore()
			svc := serviceUji(st)
			if _, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid()); err != nil {
				t.Fatalf("registrasi: %v", err)
			}
			req := JualBenihRequest{
				NomorDokumen:     "PJ-001",
				RegistrasiID:     1,
				Pembeli:          "Petani Bantul",
				Jumlah:           10,
				HargaSatuan:      25000,
				TanggalTransaksi: tanggalUji,
			}
			k.ubah(st, &req)
			_, err := svc.JualBenih(context.Background(), actorID, db.RoleAdmin, req)
			if err == nil {
				t.Fatalf("penjualan tidak sah diterima")
			}
			if k.ingin != nil && !errors.Is(err, k.ingin) {
				t.Fatalf("error = %v, ingin %v", err, k.ingin)
			}
			if st.dibuatPenjualan != 0 || st.stok[1].JumlahTersedia != 100 {
				t.Fatalf("stok/penjualan berubah walau transaksi ditolak")
			}
		})
	}
}

func TestBatalkanPenjualanMengembalikanStok(t *testing.T) {
	st := newRegistryStore()
	svc := serviceUji(st)
	if _, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid()); err != nil {
		t.Fatalf("registrasi: %v", err)
	}
	jual := JualBenihRequest{
		NomorDokumen:     "PJ-001",
		RegistrasiID:     1,
		Pembeli:          "Petani Bantul",
		Jumlah:           40,
		HargaSatuan:      25000,
		TanggalTransaksi: tanggalUji,
	}
	if _, err := svc.JualBenih(context.Background(), actorID, db.RoleAdmin, jual); err != nil {
		t.Fatalf("jual: %v", err)
	}
	got, err := svc.BatalkanPenjualan(context.Background(), actorID, db.RoleAdmin, 1, " salah input ")
	if err != nil {
		t.Fatalf("batalkan: %v", err)
	}
	if got.Status != "dibatalkan" || got.AlasanPembatalan != "salah input" {
		t.Fatalf("penjualan = %+v", got)
	}
	if st.stok[1].JumlahTersedia != 100 {
		t.Fatalf("stok setelah batal = %v, ingin 100", st.stok[1].JumlahTersedia)
	}
	if len(st.mutasi) != 2 || st.mutasi[1].Jenis != "koreksi" {
		t.Fatalf("mutasi = %+v", st.mutasi)
	}

	if _, err := svc.BatalkanPenjualan(context.Background(), actorID, db.RoleAdmin, 1, "lagi"); !errors.Is(err, store.ErrSudahDibatalkan) {
		t.Fatalf("pembatalan kedua error = %v, ingin ErrSudahDibatalkan", err)
	}
	if _, err := svc.BatalkanPenjualan(context.Background(), actorID, db.RoleAdmin, 1, "  "); err == nil {
		t.Fatalf("alasan kosong seharusnya ditolak")
	}
	if _, err := svc.BatalkanPenjualan(context.Background(), actorID, db.RoleAdmin, 99, "alasan"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("penjualan tidak ada error = %v, ingin ErrNotFound", err)
	}
}

func TestStokBenihMengembalikanMutasi(t *testing.T) {
	st := newRegistryStore()
	svc := serviceUji(st)
	if _, err := svc.BuatRegistrasiBenih(context.Background(), actorID, db.RoleAdmin, registrasiValid()); err != nil {
		t.Fatalf("registrasi: %v", err)
	}
	if _, err := svc.JualBenih(context.Background(), actorID, db.RoleAdmin, JualBenihRequest{
		NomorDokumen:     "PJ-001",
		RegistrasiID:     1,
		Pembeli:          "Petani",
		Jumlah:           5,
		TanggalTransaksi: tanggalUji,
	}); err != nil {
		t.Fatalf("jual: %v", err)
	}
	stok, mutasi, err := svc.StokBenih(context.Background(), db.RolePengawas, 1)
	if err != nil {
		t.Fatalf("stok: %v", err)
	}
	if stok.JumlahTersedia != 95 || len(mutasi) != 1 {
		t.Fatalf("stok = %v, mutasi = %d", stok.JumlahTersedia, len(mutasi))
	}
	if _, _, err := svc.StokBenih(context.Background(), db.RolePelanggan, 1); err == nil {
		t.Fatalf("role pelanggan seharusnya tidak dapat melihat stok")
	}
	if _, _, err := svc.StokBenih(context.Background(), db.RoleAdmin, 99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stok tidak ada error = %v, ingin ErrNotFound", err)
	}
}
