package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// permohonanStore meniru penyimpanan permohonan: nomor dibuat oleh store
// (bukan oleh service) dan kembali saat dibaca ulang.
type permohonanStore struct {
	fakeStore

	data       map[int64]db.PermohonanBenih
	urutHarian map[string]int64
	siap       []db.SiapRegistrasiPermohonan
	audit      []db.AuditLog
	nextID     int64
	simpan     int
}

func newPermohonanStore() *permohonanStore {
	return &permohonanStore{
		data:       map[int64]db.PermohonanBenih{},
		urutHarian: map[string]int64{},
		nextID:     1,
	}
}

func (s *permohonanStore) CreatePermohonan(_ context.Context, p db.PermohonanBenih, tanggal time.Time) (int64, error) {
	id := s.nextID
	s.nextID++
	p.ID = id
	p.CreatedAt = tanggal
	p.NomorPermohonan = fmt.Sprintf("DRAFT-%d", id)
	if p.Status != db.StatusPermohonanDraft {
		// Pengajuan langsung mendapat nomor resmi seperti di store sungguhan.
		hari := tanggal.Format("20060102")
		s.urutHarian[hari]++
		p.NomorPermohonan = store.FormatNomorPermohonan(tanggal, s.urutHarian[hari])
		waktu := tanggal
		p.DiajukanPada = &waktu
	}
	s.data[id] = p
	s.simpan++
	return id, nil
}

func (s *permohonanStore) GetPermohonan(_ context.Context, id int64) (*db.PermohonanBenih, error) {
	p, ok := s.data[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return &p, nil
}

func (s *permohonanStore) GetPermohonanAktifPelanggan(_ context.Context, pelangganID int64) (*db.PermohonanBenih, error) {
	var terpilih *db.PermohonanBenih
	for _, p := range s.data {
		if p.PelangganID != pelangganID {
			continue
		}
		switch p.Status {
		case db.StatusPermohonanDitolak, db.StatusPermohonanDibatalkan, "divalidasi":
			continue
		}
		if terpilih == nil {
			salin := p
			terpilih = &salin
			continue
		}
		// Draft selalu dipilih lebih dulu, lalu pengajuan ber-ID terbesar.
		if p.Status == db.StatusPermohonanDraft && terpilih.Status != db.StatusPermohonanDraft {
			salin := p
			terpilih = &salin
		}
	}
	return terpilih, nil
}

func (s *permohonanStore) UpdatePermohonanDraft(_ context.Context, id int64, p db.PermohonanBenih) error {
	lama, ok := s.data[id]
	if !ok || lama.PelangganID != p.PelangganID {
		return store.ErrNotFound
	}
	if lama.Status != db.StatusPermohonanDraft {
		return store.ErrBukanDraft
	}
	lama.JenisBenih = p.JenisBenih
	lama.Varietas = p.Varietas
	lama.LuasLahan = p.LuasLahan
	lama.LokasiKebun = p.LokasiKebun
	s.data[id] = lama
	return nil
}

func (s *permohonanStore) AjukanPermohonan(_ context.Context, id int64, p db.PermohonanBenih, tanggal time.Time) error {
	lama, ok := s.data[id]
	if !ok || lama.PelangganID != p.PelangganID {
		return store.ErrNotFound
	}
	if lama.Status != db.StatusPermohonanDraft {
		return store.ErrBukanDraft
	}
	hari := tanggal.Format("20060102")
	s.urutHarian[hari]++
	lama.JenisBenih = p.JenisBenih
	lama.Varietas = p.Varietas
	lama.LuasLahan = p.LuasLahan
	lama.LokasiKebun = p.LokasiKebun
	lama.NomorPermohonan = store.FormatNomorPermohonan(tanggal, s.urutHarian[hari])
	lama.Status = db.StatusPermohonanDiajukan
	waktu := tanggal
	lama.DiajukanPada = &waktu
	s.data[id] = lama
	return nil
}

func (s *permohonanStore) UpdatePermohonanStatus(_ context.Context, id int64, status, pembayaran string) error {
	p, ok := s.data[id]
	if !ok {
		return store.ErrNotFound
	}
	p.Status = status
	p.PembayaranAwalStatus = pembayaran
	s.data[id] = p
	return nil
}

func (s *permohonanStore) ListPermohonanSiapRegistrasi(_ context.Context, pemilikID int64, semua bool, _ int) ([]db.SiapRegistrasiPermohonan, error) {
	var hasil []db.SiapRegistrasiPermohonan
	for _, item := range s.siap {
		if semua || s.data[item.PermohonanID].PelangganID == pemilikID {
			hasil = append(hasil, item)
		}
	}
	return hasil, nil
}

func (s *permohonanStore) CreateAuditLog(_ context.Context, a db.AuditLog) (int64, error) {
	s.audit = append(s.audit, a)
	return int64(len(s.audit)), nil
}

func servicePermohonan(st *permohonanStore) *Service {
	svc := NewService(st)
	svc.now = func() time.Time {
		return time.Date(2026, 10, 23, 9, 15, 0, 0, time.Local)
	}
	return svc
}

func TestBuatPermohonanMenerapkanFormatNomorDanMenyimpanNomornya(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	pertama, err := svc.BuatPermohonan(context.Background(), 10, db.RolePelanggan, BuatPermohonanRequest{
		JenisBenih:  "  Tomat ",
		Varietas:    " Rama ",
		LuasLahan:   0.5,
		LokasiKebun: "  Sleman ",
	})
	if err != nil {
		t.Fatalf("permohonan pertama: %v", err)
	}
	if pertama.NomorPermohonan != "SPB-20261023-0001" {
		t.Fatalf("nomor = %q, ingin SPB-20261023-0001", pertama.NomorPermohonan)
	}
	if pertama.JenisBenih != "Tomat" || pertama.Varietas != "Rama" {
		t.Fatalf("input tidak dirapikan: %q/%q", pertama.JenisBenih, pertama.Varietas)
	}
	if st.data[pertama.ID].NomorPermohonan != "SPB-20261023-0001" {
		t.Fatalf("nomor tidak tersimpan di store: %+v", st.data[pertama.ID])
	}
	if len(st.audit) != 1 || st.audit[0].Detail != "jenis_benih=Tomat" {
		t.Fatalf("audit permohonan = %+v", st.audit)
	}

	kedua, err := svc.BuatPermohonan(context.Background(), 11, db.RolePelanggan, BuatPermohonanRequest{
		JenisBenih:  "Padi",
		Varietas:    "IR64",
		LuasLahan:   1,
		LokasiKebun: "Bantul",
	})
	if err != nil {
		t.Fatalf("permohonan kedua: %v", err)
	}
	if kedua.NomorPermohonan != "SPB-20261023-0002" {
		t.Fatalf("nomor kedua = %q, ingin SPB-20261023-0002", kedua.NomorPermohonan)
	}
}

func TestBuatPermohonanMenolakInputTidakSah(t *testing.T) {
	kasus := []struct {
		nama  string
		req   BuatPermohonanRequest
		role  string
		pesan string
	}{
		{"jenis kosong", BuatPermohonanRequest{JenisBenih: "   "}, db.RolePelanggan, "jenis benih wajib"},
		{"luas negatif", BuatPermohonanRequest{JenisBenih: "Padi", Varietas: "IR64", LuasLahan: -1, LokasiKebun: "Bantul"}, db.RolePelanggan, "luas lahan"},
		{"luas nol", BuatPermohonanRequest{JenisBenih: "Padi", Varietas: "IR64", LuasLahan: 0, LokasiKebun: "Bantul"}, db.RolePelanggan, "luas lahan"},
		{"varietas kosong", BuatPermohonanRequest{JenisBenih: "Padi", LokasiKebun: "Bantul", LuasLahan: 1}, db.RolePelanggan, "varietas wajib"},
		{"lokasi kosong", BuatPermohonanRequest{JenisBenih: "Padi", Varietas: "IR64", LuasLahan: 1}, db.RolePelanggan, "lokasi kebun wajib"},
		{"role terlarang", BuatPermohonanRequest{JenisBenih: "Padi"}, db.RoleAnalis, "tidak diizinkan"},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			st := newPermohonanStore()
			_, err := servicePermohonan(st).BuatPermohonan(context.Background(), 10, k.role, k.req)
			if err == nil || !contains(err.Error(), k.pesan) {
				t.Fatalf("error = %v, ingin memuat %q", err, k.pesan)
			}
			if st.simpan != 0 {
				t.Fatalf("permohonan tersimpan walau input ditolak")
			}
		})
	}
}

func TestListPermohonanSiapRegistrasiMembatasiPemilik(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)
	milikSaya, err := svc.BuatPermohonan(context.Background(), 10, db.RolePelanggan, BuatPermohonanRequest{JenisBenih: "Tomat", Varietas: "Rama", LuasLahan: 0.5, LokasiKebun: "Sleman"})
	if err != nil {
		t.Fatalf("permohonan pemilik: %v", err)
	}
	milikOrang, err := svc.BuatPermohonan(context.Background(), 11, db.RolePelanggan, BuatPermohonanRequest{JenisBenih: "Padi", Varietas: "IR64", LuasLahan: 1, LokasiKebun: "Bantul"})
	if err != nil {
		t.Fatalf("permohonan pelanggan lain: %v", err)
	}
	st.siap = []db.SiapRegistrasiPermohonan{
		{PermohonanID: milikSaya.ID, NomorPermohonan: milikSaya.NomorPermohonan},
		{PermohonanID: milikOrang.ID, NomorPermohonan: milikOrang.NomorPermohonan},
	}

	pelanggan, err := svc.ListPermohonanSiapRegistrasi(context.Background(), 10, db.RolePelanggan)
	if err != nil {
		t.Fatalf("daftar pelanggan: %v", err)
	}
	if len(pelanggan) != 1 || pelanggan[0].PermohonanID != milikSaya.ID {
		t.Fatalf("pelanggan hanya boleh melihat permohonannya sendiri: %+v", pelanggan)
	}

	admin, err := svc.ListPermohonanSiapRegistrasi(context.Background(), 1, db.RoleAdmin)
	if err != nil {
		t.Fatalf("daftar admin: %v", err)
	}
	if len(admin) != 2 {
		t.Fatalf("admin melihat %d permohonan, ingin 2", len(admin))
	}

	if _, err := svc.ListPermohonanSiapRegistrasi(context.Background(), 10, "tidak_dikenal"); err == nil {
		t.Fatalf("role analis tidak boleh melihat permohonan untuk registrasi")
	}
}

func contains(text, bagian string) bool {
	return strings.Contains(text, bagian)
}
