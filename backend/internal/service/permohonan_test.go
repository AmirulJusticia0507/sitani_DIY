package service

import (
	"context"
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
	hari := tanggal.Format("20060102")
	s.urutHarian[hari]++
	id := s.nextID
	s.nextID++
	p.ID = id
	p.NomorPermohonan = store.FormatNomorPermohonan(tanggal, s.urutHarian[hari])
	p.CreatedAt = tanggal
	s.data[id] = p
	s.simpan++
	return id, nil
}

func (s *permohonanStore) GetPermohonan(_ context.Context, id int64) (*db.PermohonanBenih, error) {
	p, ok := s.data[id]
	if !ok {
		return nil, nil
	}
	return &p, nil
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
		JenisBenih: "  Tomat ",
		Varietas:   " Rama ",
		LuasLahan:  0.5,
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

	kedua, err := svc.BuatPermohonan(context.Background(), 10, db.RolePelanggan, BuatPermohonanRequest{JenisBenih: "Padi"})
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
		{"luas negatif", BuatPermohonanRequest{JenisBenih: "Padi", LuasLahan: -1}, db.RolePelanggan, "luas lahan"},
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
	milikSaya, err := svc.BuatPermohonan(context.Background(), 10, db.RolePelanggan, BuatPermohonanRequest{JenisBenih: "Tomat"})
	if err != nil {
		t.Fatalf("permohonan pemilik: %v", err)
	}
	milikOrang, err := svc.BuatPermohonan(context.Background(), 11, db.RolePelanggan, BuatPermohonanRequest{JenisBenih: "Padi"})
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
