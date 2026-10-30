package service

import (
	"context"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// fakeStore menolak semua panggilan yang tidak dioverride oleh fakeStore uji
// spesifik. Dengan begitu, tes hanya mengizinkan alur yang memang relevan
// sehingga kebocoran pemanggilan store tidak lolos diam-diam.
type fakeStore struct{}

func (fakeStore) CreateUser(context.Context, db.User) (int64, error)   { panic("tidak diharapkan") }
func (fakeStore) GetUserByID(context.Context, int64) (*db.User, error) { panic("tidak diharapkan") }
func (fakeStore) GetUserByUsername(context.Context, string) (*db.User, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreatePermohonan(context.Context, db.PermohonanBenih, time.Time) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetPermohonan(context.Context, int64) (*db.PermohonanBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetPermohonanAktifPelanggan(context.Context, int64) (*db.PermohonanBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdatePermohonanDraft(context.Context, int64, db.PermohonanBenih) error {
	panic("tidak diharapkan")
}
func (fakeStore) AjukanPermohonan(context.Context, int64, db.PermohonanBenih, time.Time) error {
	panic("tidak diharapkan")
}
func (fakeStore) SetNomorPermohonan(context.Context, int64, string) error {
	panic("tidak diharapkan")
}
func (fakeStore) UpdatePermohonanStatus(context.Context, int64, string, string) error {
	panic("tidak diharapkan")
}
func (fakeStore) ListPermohonanSiapRegistrasi(context.Context, int64, bool, int) ([]db.SiapRegistrasiPermohonan, error) {
	panic("tidak diharapkan")
}
func (fakeStore) CreatePembayaran(context.Context, db.Pembayaran) (int64, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreateInspeksiFase(context.Context, db.InspeksiFase) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) ListInspeksiFase(context.Context, int64) ([]db.InspeksiFase, error) {
	panic("tidak diharapkan")
}
func (fakeStore) CreateBukuInduk(context.Context, db.BukuInduk) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetBukuInduk(context.Context, int64) (*db.BukuInduk, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreateSampel(context.Context, db.SampelBenih) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetSampel(context.Context, int64) (*db.SampelBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) KodifikasiSampel(context.Context, int64, string, int64) error {
	panic("tidak diharapkan")
}
func (fakeStore) CreateUjiMutu(context.Context, db.UjiMutu) (int64, error) { panic("tidak diharapkan") }
func (fakeStore) ListUjiMutu(context.Context, int64) ([]db.UjiMutu, error) {
	panic("tidak diharapkan")
}
func (fakeStore) CreateLaporanMutu(context.Context, db.LaporanMutu) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetLaporanMutu(context.Context, int64) (*db.LaporanMutu, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreateSertifikat(context.Context, db.SertifikatBenih) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetSertifikat(context.Context, int64) (*db.SertifikatBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetSertifikatByID(context.Context, int64) (*db.SertifikatBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdateSertifikatTTE(context.Context, int64, string, string, *time.Time) error {
	panic("tidak diharapkan")
}
func (fakeStore) CreateDraftLabel(context.Context, db.DraftLabel) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetDraftLabel(context.Context, int64) (*db.DraftLabel, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetDraftLabelByID(context.Context, int64) (*db.DraftLabel, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdateDraftLabel(context.Context, int64, string, string, string) error {
	panic("tidak diharapkan")
}
func (fakeStore) CreateLegalisasi(context.Context, db.LabelLegalisasi) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdateLegalisasi(context.Context, int64, string) error { panic("tidak diharapkan") }
func (fakeStore) CreatePenyerahan(context.Context, db.PenyerahanDokumen) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetPenyerahan(context.Context, int64) (*db.PenyerahanDokumen, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdatePenyerahan(context.Context, int64, bool, bool, string) error {
	panic("tidak diharapkan")
}

func (fakeStore) CreateRegistrasiBenih(context.Context, db.RegistrasiBenih) (int64, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetRegistrasiBenih(context.Context, int64) (*db.RegistrasiBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetRegistrasiBenihByNomorLot(context.Context, string) (*db.RegistrasiBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) ListRegistrasiBenih(context.Context, db.RegistrasiFilter) ([]db.RegistrasiBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) UpdateRegistrasiStatus(context.Context, int64, string) error {
	panic("tidak diharapkan")
}
func (fakeStore) GetStokBenih(context.Context, int64) (*db.StokBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) ListStokBenih(context.Context) ([]db.StokBenih, error) { panic("tidak diharapkan") }
func (fakeStore) ListMutasiStok(context.Context, int64) ([]db.MutasiStokBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) CreatePenjualanBenih(context.Context, db.PenjualanBenih) (*db.PenjualanBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) GetPenjualanBenih(context.Context, int64) (*db.PenjualanBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) BatalkanPenjualanBenih(context.Context, int64, int64, string, time.Time) (*db.PenjualanBenih, error) {
	panic("tidak diharapkan")
}
func (fakeStore) ListPenjualanBenih(context.Context, db.PenjualanFilter) ([]db.PenjualanBenih, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreateAuditLog(context.Context, db.AuditLog) (int64, error) {
	panic("tidak diharapkan")
}

func (fakeStore) CreateDokumen(context.Context, db.Dokumen) (int64, error) { panic("tidak diharapkan") }
func (fakeStore) GetDokumen(context.Context, int64) (*db.Dokumen, error)   { panic("tidak diharapkan") }
func (fakeStore) ListDokumen(context.Context, int64, bool) ([]db.Dokumen, error) {
	panic("tidak diharapkan")
}

func (fakeStore) Ping(context.Context) error { return nil }

var _ store.Store = fakeStore{}
