# Asisten AI (Proxy Backend)

## Status

Fondasi proxy asisten AI tersedia di backend: adaptor `internal/ai` memanggil
penyedia yang kompatibel dengan API OpenAI (mis. GripHub Router), dan endpoint
`POST /v1/ai/chat` meneruskan percakapan setelah memverifikasi identitas SSO.
Pemanggilan langsung ke penyedia **belum diverifikasi** karena `AI_API_KEY`
belum tersedia; perilaku transport diuji dengan server HTTP palsu.

Ini bukan chatbot siap produksi: belum ada basis pengetahuan resmi DPKP,
sitasi sumber, evaluasi kualitas jawaban, atau persetujuan penggunaan model.

## Pola yang diadopsi

Pola mengikuti adaptor klien eksternal pada `diygov-starter-be`:

- konfigurasi hanya dari environment, tanpa URL/kredensial di kode;
- backend tidak gagal start saat konfigurasi kosong; endpoint menjawab `503`
  sampai diisi (gagal saat dipakai, bukan saat start);
- API key hanya berada di backend dan tidak pernah dicatat ke log atau
  dikirim ke browser;
- timeout, batas ukuran respons, dan pemetaan status error dibatasi di adaptor;
- body balasan penyedia tidak diteruskan apa adanya ke pengguna;
- model dikunci dari environment sehingga klien tidak dapat mengganti model
  atau tujuan permintaan.

## Konfigurasi

Isi environment backend (jangan commit nilai rahasia):

```text
AI_BASE_URL=https://griphubrouter.web.id/v1
AI_API_KEY=
AI_MODEL=deepseek-v4-flash
AI_TIMEOUT_SECONDS=90
AI_MAX_REQUESTS_PER_MINUTE=10
```

`AI_MODEL` memakai slug persis dari katalog penyedia, mis. `deepseek-v4-flash`
atau `deepseek-v4-pro`.

## Endpoint

`POST /v1/ai/chat`, dilindungi middleware SSO. Peran yang diizinkan:
`superadmin`, `admin`, `pengawas`, `analis`, `penandatangan`, `legalisasi`,
dan `percetakan`. Akses `pelanggan` menunggu keputusan DPKP.

Permintaan:

```json
{"messages":[{"role":"user","content":"Apa syarat sertifikasi benih?"}]}
```

Batasan: 1–20 pesan, role hanya `user`/`assistant` (system prompt ditentukan
server), maksimal 4.000 karakter per pesan dan 16.000 karakter total, body
maksimal 64 KiB. Balasan:

```json
{"reply":"...","model":"deepseek-v4-flash"}
```

Status error: `400` validasi, `401` tanpa identitas, `403` peran tidak
diizinkan, `429` batas laju/kuota penyedia, `502` penyedia menolak atau tidak
dapat dihubungi, `503` konfigurasi belum tersedia.

## Batasan keamanan

- Jangan mengirim NIK, dokumen pribadi, atau data mentah pelanggan ke penyedia
  AI. Endpoint saat ini tidak memeriksa isi pesan secara otomatis.
- Jawaban AI tidak boleh dianggap keputusan resmi atau nasihat teknis final.
- Sebelum produksi, perlu keputusan tertulis tentang penyedia, retensi data,
  biaya, dan batas pemakaian, serta evaluasi kualitas model yang dipilih.

## Verifikasi

```powershell
cd backend
go test ./internal/ai ./internal/httpapi -count=1
```

Tes memakai `httptest` dan klien palsu; tidak memanggil penyedia sungguhan.
Setelah `AI_API_KEY` tersedia, uji manual endpoint dengan sesi login dan
bandingkan `deepseek-v4-flash` dengan `deepseek-v4-pro` memakai pertanyaan
yang sama.
