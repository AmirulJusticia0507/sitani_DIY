package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// verifikasiNikResponse adalah respons dari endpoint verifikasi NIK.
type verifikasiNikResponse struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
}

// verifikasiNikRequest adalah request untuk verifikasi NIK manual.
type verifikasiNikRequest struct {
	NIK string `json:"nik" validate:"required,nik16"`
}

// nik16 validasi NIK 16 digit.
// Pola ini sesuai agriculture-be: `regexp.MustCompile(`^[0-9]{16}$`)`.
var nik16Pattern = regexp.MustCompile(`^[0-9]{16}$`)

func nik16(s string) bool {
	return nik16Pattern.MatchString(s)
}

// registerVerifikasiNikRoutes mendaftarkan endpoint verifikasi NIK.
// Fungsi ini dideklarasikan dalam handler.go, file ini hanya berisi handler.
var _ = registerVerifikasiNikRoutes // hindari re-deklarasi (dilindungi di handler.go)

// handleVerifikasiNikManual menangani input NIK manual.
func handleVerifikasiNikManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req verifikasiNikRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body request tidak valid")
		return
	}

	// Validasi NIK 16 digit
	if !nik16(req.NIK) {
		respondError(w, http.StatusBadRequest, "NIK harus terdiri dari 16 digit")
		return
	}

	// Generate verification ID
	verificationID := "ver-" + req.NIK[:8] + "-" + req.NIK[8:12]

	resp := verifikasiNikResponse{
		VerificationID: verificationID,
		Status:         "pending",
		Message:        "Antrean verifikasi NIK dibuat, silakan tunggu verifikator menyetujui",
	}

	respondJSON(w, http.StatusCreated, resp)
}

// handleVerifikasiNikFromKTP menangani unggah foto e-KTP untuk ekstraksi NIK.
func handleVerifikasiNikFromKTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form max 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "gagal parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "gagal ambil file: "+err.Error())
		return
	}
	defer file.Close()

	// Ekstrak NIK dari file e-KTP (simulasi OCR)
	extractedNIK := extractNIKFromFilename(header.Filename)
	if extractedNIK == "" {
		respondError(w, http.StatusBadRequest, "tidak dapat mengekstrak NIK dari file e-KTP")
		return
	}

	// Validasi NIK yang diekstrak
	if !nik16(extractedNIK) {
		respondError(w, http.StatusBadRequest, "NIK yang diekstrak tidak valid (harus 16 digit)")
		return
	}

	// Generate verification ID
	verificationID := "ver-" + extractedNIK[:8] + "-" + extractedNIK[8:12]

	resp := verifikasiNikResponse{
		VerificationID: verificationID,
		Status:         "pending",
		Message:        "NIK berhasil diekstrak dari e-KTP, antrean verifikasi dibuat",
	}

	respondJSON(w, http.StatusCreated, resp)
}

// extractNIKFromFilename mengambil NIK dari nama file e-KTP.
// Di implementasi nyata, library OCR akan digunakan.
func extractNIKFromFilename(filename string) string {
	// Cari pola 16 digit di filename
	for _, part := range strings.Split(filename, ".") {
		if len(part) >= 16 {
			// Ambil 16 digit pertama yang ditemukan
			var digits strings.Builder
			for _, c := range part {
				if c >= '0' && c <= '9' {
					digits.WriteRune(c)
					if digits.Len() == 16 {
						break
					}
				}
			}
			if digits.Len() == 16 {
				return digits.String()
			}
		}
	}

	// Fallback: cek apakah ada kata "ktp" di filename (case-insensitive)
	if strings.Contains(strings.ToLower(filename), "ktp") {
		// Return NIK contoh untuk development
		return "1234567890123456"
	}

	return ""
}