package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

func isianLengkap() BuatPermohonanRequest {
	return BuatPermohonanRequest{
		JenisBenih:  "Tomat",
		Varietas:    "Rama",
		LuasLahan:   0.5,
		LokasiKebun: "Sleman",
	}
}

func TestSimpanDraftBolehBelumLengkapDanTidakMembuatBarisGanda(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	pertama, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"})
	if err != nil {
		t.Fatalf("simpan draft pertama: %v", err)
	}
	if pertama.Status != db.StatusPermohonanDraft {
		t.Fatalf("status = %q, ingin draft", pertama.Status)
	}
	if pertama.DiajukanPada != nil {
		t.Fatalf("draft tidak boleh punya waktu pengajuan: %v", pertama.DiajukanPada)
	}

	// Tanpa draft_id, draft aktif yang sama diperbarui.
	kedua, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, isianLengkap())
	if err != nil {
		t.Fatalf("simpan draft kedua: %v", err)
	}
	if kedua.ID != pertama.ID {
		t.Fatalf("draft ganda: id %d lalu %d", pertama.ID, kedua.ID)
	}
	if len(st.data) != 1 {
		t.Fatalf("jumlah baris = %d, ingin 1", len(st.data))
	}
	if kedua.Varietas != "Rama" || kedua.LokasiKebun != "Sleman" {
		t.Fatalf("isian draft tidak tersimpan: %+v", kedua)
	}

	// Penyimpanan lewat draft_id yang sama juga memperbarui, bukan menduplikasi.
	ketiga, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, kedua.ID, BuatPermohonanRequest{JenisBenih: "Cabai", LuasLahan: 1, LokasiKebun: "Bantul"})
	if err != nil {
		t.Fatalf("simpan draft ketiga: %v", err)
	}
	if ketiga.ID != pertama.ID || ketiga.JenisBenih != "Cabai" {
		t.Fatalf("draft ketiga = %+v", ketiga)
	}

	// Setiap penyimpanan tercatat pada audit dengan aksi yang dapat dibedakan.
	if len(st.audit) != 3 {
		t.Fatalf("audit = %d catatan, ingin 3", len(st.audit))
	}
	for _, a := range st.audit {
		if a.Action != "permohonan:draft_simpan" {
			t.Fatalf("aksi audit = %q", a.Action)
		}
	}
}

func TestSimpanDraftMenolakDraftMilikPelangganLain(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)
	draft, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"})
	if err != nil {
		t.Fatalf("simpan draft: %v", err)
	}

	if _, err := svc.SimpanDraft(context.Background(), 11, db.RolePelanggan, draft.ID, isianLengkap()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("pelanggan lain: error = %v, ingin ErrNotFound", err)
	}
	if st.data[draft.ID].JenisBenih != "Tomat" {
		t.Fatalf("draft pelanggan lain ikut berubah: %+v", st.data[draft.ID])
	}

	if _, err := svc.SimpanDraft(context.Background(), 10, db.RoleAnalis, 0, isianLengkap()); err == nil {
		t.Fatalf("role analis tidak boleh menyimpan draft")
	}
}

func TestAjukanPermohonanMengubahDraftMenjadiNomorResmi(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	draft, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"})
	if err != nil {
		t.Fatalf("simpan draft: %v", err)
	}
	resmi, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, draft.ID, isianLengkap())
	if err != nil {
		t.Fatalf("ajukan permohonan: %v", err)
	}
	if resmi.Status != db.StatusPermohonanDiajukan {
		t.Fatalf("status = %q, ingin diajukan", resmi.Status)
	}
	if resmi.NomorPermohonan != "SPB-20261023-0001" {
		t.Fatalf("nomor = %q, ingin SPB-20261023-0001", resmi.NomorPermohonan)
	}
	if resmi.DiajukanPada == nil {
		t.Fatalf("waktu pengajuan tidak tercatat")
	}
	if resmi.LuasLahan != 0.5 || resmi.LokasiKebun != "Sleman" {
		t.Fatalf("isian final tidak tersimpan: %+v", resmi)
	}
	if len(st.audit) != 2 || st.audit[1].Action != "permohonan:ajukan" {
		t.Fatalf("audit pengajuan = %+v", st.audit)
	}

	// Draft yang sudah diajukan tidak dapat dibuat draft lagi.
	if _, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, draft.ID, isianLengkap()); !errors.Is(err, ErrBukanDraft) {
		t.Fatalf("simpan ulang setelah diajukan: error = %v, ingin ErrBukanDraft", err)
	}
	if _, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, draft.ID, isianLengkap()); !errors.Is(err, ErrBukanDraft) {
		t.Fatalf("ajukan dua kali: error = %v, ingin ErrBukanDraft", err)
	}
	if st.data[draft.ID].NomorPermohonan != "SPB-20261023-0001" {
		t.Fatalf("nomor berubah setelah pengajuan ulang: %+v", st.data[draft.ID])
	}
}

// AjukanPermohonanTanpaDraftIDMemakaiDraftAktif memastikan pengajuan tanpa
// draft_id membuat draft lalu mengajukannya dalam satu langkah untuk pelanggan
// yang belum punya permohonan, dan menolak duplikasi bila sudah ada pengajuan.
func TestAjukanPermohonanTanpaDraftIDMemakaiDraftAktif(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	resmi, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, 0, isianLengkap())
	if err != nil {
		t.Fatalf("ajukan tanpa draft tersimpan: %v", err)
	}
	if resmi.Status != db.StatusPermohonanDiajukan {
		t.Fatalf("status = %q", resmi.Status)
	}
	if resmi.NomorPermohonan != "SPB-20261023-0001" {
		t.Fatalf("nomor = %q", resmi.NomorPermohonan)
	}
	if resmi.Varietas != "Rama" || resmi.LokasiKebun != "Sleman" {
		t.Fatalf("isian tidak tersimpan: %+v", resmi)
	}
	if len(st.data) != 1 {
		t.Fatalf("jumlah baris = %d, ingin 1", len(st.data))
	}

	// Pengajuan kedua saat masih ada pengajuan berjalan ditolak.
	if _, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, 0, isianLengkap()); !errors.Is(err, ErrSudahMengajukan) {
		t.Fatalf("ajukan ulang: error = %v, ingin ErrSudahMengajukan", err)
	}
	if len(st.data) != 1 {
		t.Fatalf("pengajuan ganda membuat baris baru: %d", len(st.data))
	}
}

func TestAjukanPermohonanMemakaiDraftAktifBilaAda(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	draft, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"})
	if err != nil {
		t.Fatalf("simpan draft: %v", err)
	}
	resmi, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, 0, isianLengkap())
	if err != nil {
		t.Fatalf("ajukan draft aktif: %v", err)
	}
	if resmi.ID != draft.ID {
		t.Fatalf("draft baru dibuat padahal draft aktif ada: %d vs %d", draft.ID, resmi.ID)
	}
	if len(st.data) != 1 {
		t.Fatalf("jumlah baris = %d, ingin 1", len(st.data))
	}
}

func TestAjukanPermohonanMenolakIsianTidakLengkap(t *testing.T) {
	kasus := []BuatPermohonanRequest{
		{JenisBenih: "", Varietas: "Rama", LuasLahan: 1, LokasiKebun: "Sleman"},
		{JenisBenih: "Tomat", Varietas: "", LuasLahan: 1, LokasiKebun: "Sleman"},
		{JenisBenih: "Tomat", Varietas: "Rama", LuasLahan: 0, LokasiKebun: "Sleman"},
		{JenisBenih: "Tomat", Varietas: "Rama", LuasLahan: 1, LokasiKebun: "  "},
	}
	for _, req := range kasus {
		st := newPermohonanStore()
		svc := servicePermohonan(st)
		draft, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"})
		if err != nil {
			t.Fatalf("simpan draft: %v", err)
		}
		if _, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, draft.ID, req); err == nil {
			t.Fatalf("isian tidak lengkap diterima: %+v", req)
		}
		if st.data[draft.ID].Status != db.StatusPermohonanDraft {
			t.Fatalf("status berubah walau ditolak: %+v", st.data[draft.ID])
		}
		// Tanpa draft_id pun tidak ada baris baru yang dibuat saat isian tidak sah.
		if _, err := svc.AjukanPermohonan(context.Background(), 11, db.RolePelanggan, 0, req); err == nil {
			t.Fatalf("isian tidak lengkap diterima tanpa draft: %+v", req)
		}
		if len(st.data) != 1 {
			t.Fatalf("isian tidak sah membuat baris baru: %d", len(st.data))
		}
	}
}

func TestPermohonanAktifMengembalikanDraftLaluPengajuan(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)

	kosong, err := svc.PermohonanAktif(context.Background(), 10, db.RolePelanggan)
	if err != nil || kosong != nil {
		t.Fatalf("tanpa permohonan: %+v, %v", kosong, err)
	}

	if _, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Tomat"}); err != nil {
		t.Fatalf("simpan draft: %v", err)
	}
	aktif, err := svc.PermohonanAktif(context.Background(), 10, db.RolePelanggan)
	if err != nil || aktif == nil || aktif.Status != db.StatusPermohonanDraft {
		t.Fatalf("permohonan aktif = %+v, %v", aktif, err)
	}

	// Draft milik pelanggan lain tidak ikut terbaca.
	lain, err := svc.PermohonanAktif(context.Background(), 11, db.RolePelanggan)
	if err != nil || lain != nil {
		t.Fatalf("permohonan pelanggan lain = %+v, %v", lain, err)
	}

	// Setelah diajukan, draft baru ditolak; pengajuan berjalan yang dikembalikan.
	if _, err := svc.AjukanPermohonan(context.Background(), 10, db.RolePelanggan, 0, isianLengkap()); err != nil {
		t.Fatalf("ajukan: %v", err)
	}
	if _, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, BuatPermohonanRequest{JenisBenih: "Padi"}); !errors.Is(err, ErrSudahMengajukan) {
		t.Fatalf("draft baru saat ada pengajuan: error = %v, ingin ErrSudahMengajukan", err)
	}
	aktif, err = svc.PermohonanAktif(context.Background(), 10, db.RolePelanggan)
	if err != nil || aktif == nil || aktif.Status != db.StatusPermohonanDiajukan {
		t.Fatalf("pengajuan aktif = %+v, %v", aktif, err)
	}
}

func TestBayarAwalDitolakUntukDraft(t *testing.T) {
	st := newPermohonanStore()
	svc := servicePermohonan(st)
	draft, err := svc.SimpanDraft(context.Background(), 10, db.RolePelanggan, 0, isianLengkap())
	if err != nil {
		t.Fatalf("simpan draft: %v", err)
	}

	err = svc.BayarAwal(context.Background(), draft.ID, 10, db.RolePelanggan, 100000)
	if err == nil || !contains(err.Error(), "draft belum diajukan") {
		t.Fatalf("bayar draft: error = %v", err)
	}
	if st.data[draft.ID].PembayaranAwalStatus != "belum" {
		t.Fatalf("status pembayaran berubah untuk draft: %+v", st.data[draft.ID])
	}
}
