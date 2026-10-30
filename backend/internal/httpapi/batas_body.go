package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// MaksBodyJSON adalah batas ukuran body permintaan non-unggah. Isi JSON portal
// jauh di bawah nilai ini; tanpa batas, satu permintaan raksasa — bahkan ke
// endpoint yang belum login seperti /auth/dev/login dan /auth/callback/... —
// ditampung utuh di memori sebelum ditolak.
const MaksBodyJSON int64 = 1 << 20

// Methods that may carry a request body.
var metodeBerBody = map[string]bool{
	http.MethodPost:  true,
	http.MethodPut:   true,
	http.MethodPatch: true,
}

// pesanBatasBody menerjemahkan *http.MaxBytesError menjadi balasan 413 yang
// menyebut batasnya. Galat lain (JSON cacat, isian salah) mengembalikan false
// agar pemanggil memakai balasan 400 seperti biasa.
func pesanBatasBody(w http.ResponseWriter, err error) bool {
	var terlaluBesar *http.MaxBytesError
	if !errors.As(err, &terlaluBesar) {
		return false
	}
	pesan := fmt.Sprintf("body permintaan melebihi batas %d KB", terlaluBesar.Limit>>10)
	if terlaluBesar.Limit > MaksBodyJSON {
		// Batas dari rute unggah = ukuran berkas maks + margin parse.
		pesan = fmt.Sprintf("berkas melebihi batas unggah %d MB", terlaluBesar.Limit>>20)
	}
	respondError(w, http.StatusRequestEntityTooLarge, pesan)
	return true
}

// batasBodyJSON membatasi body permintaan JSON ke MaksBodyJSON. Dipanggil di
// setiap handler yang membaca JSON agar rute unggah (multipart) tidak ikut
// terkena: handler unggah memasang batasnya sendiri sesuai ukuran berkas.
func batasBodyJSON(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaksBodyJSON)
}

// bacaJSON membaca body JSON ke v setelah memasang batas ukuran. Mengembalikan
// false bila balasan (413 atau 400) sudah ditulis ke w.
func bacaJSON(w http.ResponseWriter, r *http.Request, v any) (bool, error) {
	batasBodyJSON(w, r)
	err := json.NewDecoder(r.Body).Decode(v)
	if err == nil {
		return true, nil
	}
	if pesanBatasBody(w, err) {
		return false, io.ErrUnexpectedEOF
	}
	return false, err
}
