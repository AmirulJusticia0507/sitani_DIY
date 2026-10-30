"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Autocomplete } from "../../components/autocomplete";
import { Badge } from "../../components/badge";
import { Button } from "../../components/button";
import { Card } from "../../components/card";
import { JENIS_BENIH_OPTIONS, VARIETAS_BENIH } from "./jenis-benih";

// Saran autocomplete jenis benih: satu entri per komoditas dengan label grup
// komoditasnya, sehingga daftar saran tampil berkelompok.
const OPSI_JENIS_BENIH = JENIS_BENIH_OPTIONS.flatMap((kelompok) =>
  kelompok.jenis.map((jenis) => ({ nilai: jenis, grup: kelompok.grup })),
);

const samakan = (v: string) => v.trim().toLowerCase();

// Jenis benih wajib berupa salah satu komoditas pada daftar, supaya data
// permohonan dapat direkap per komoditas. Komoditas baru ditambahkan ke
// frontend/app/certification/jenis-benih.ts setelah daftar resmi dikonfirmasi.
function jenisDikenal(nilai: string): boolean {
  const isian = samakan(nilai);
  return OPSI_JENIS_BENIH.some((o) => samakan(o.nilai) === isian);
}

// Saran autocomplete varietas: disaring mengikuti jenis benih yang dipilih
// (Autocomplete menerima `induk` untuk penyaringan bertingkat).
const OPSI_VARIETAS = Object.entries(VARIETAS_BENIH).flatMap(([jenis, varietas]) =>
  varietas.map((v) => ({ nilai: v, induk: jenis })),
);

// Form pengajuan permohonan sertifikasi benih (Fase 1).
//
// Hanya Fase 1 yang berbentuk form pelanggan. Fase 2 (pengujian lab mutu) dan
// Fase 3 (sertifikat, TTE, label) dikerjakan petugas internal lewat aksi pada
// permohonan yang sudah diajukan, bukan isian pelanggan — lihat ringkasan
// fase di halaman induk.
//
// Alur tombol:
//   Simpan Draft   -> POST /v1/permohonan/draft          (boleh isian belum lengkap)
//   Edit Draft     -> kembali ke isian draft yang tersimpan bila input terkunci
//   Ajukan Permohonan -> POST /v1/permohonan/ajukan (isian wajib lengkap,
//                       nomor SPB-YYYYMMDD-XXXX dibuat server)
//
// Selama draft atau pengajuan berjalan sudah ada, membuat permohonan baru
// ditolak backend (409) sehingga pelanggan menyelesaikan yang sedang berjalan.

export interface DataPermohonan {
  ID: number;
  NomorPermohonan: string;
  Status: string;
  JenisBenih: string;
  Varietas: string;
  LuasLahan: number;
  LokasiKebun: string;
  PembayaranAwalStatus: string;
  DiajukanPada: string | null;
}

type Mode = "baru" | "draft" | "terkirim";

interface NilaiForm {
  jenisBenih: string;
  varietas: string;
  luasLahan: string;
  lokasiKebun: string;
}

const KOSONG: NilaiForm = { jenisBenih: "", varietas: "", luasLahan: "", lokasiKebun: "" };

const dariPermohonan = (p: DataPermohonan): NilaiForm => ({
  jenisBenih: p.JenisBenih ?? "",
  varietas: p.Varietas ?? "",
  luasLahan: p.LuasLahan ? String(p.LuasLahan) : "",
  lokasiKebun: p.LokasiKebun ?? "",
});

// wajibIsi menandai isian yang belum lengkap untuk pengajuan resmi.
function wajibIsi(nilai: NilaiForm): string[] {
  const kurang: string[] = [];
  if (!nilai.jenisBenih.trim()) kurang.push("Jenis benih");
  if (!nilai.varietas.trim()) kurang.push("Varietas");
  const luas = Number(nilai.luasLahan);
  if (!nilai.luasLahan.trim() || Number.isNaN(luas) || luas <= 0) kurang.push("Luas lahan");
  if (!nilai.lokasiKebun.trim()) kurang.push("Lokasi kebun");
  return kurang;
}

export function labelStatusPermohonan(status: string, nomor: string): string {
  switch (status) {
    case "draft":
      return "Draft (belum diajukan)";
    case "diajukan":
      return nomor ? `Diajukan — ${nomor}` : "Diajukan";
    case "dibayar":
      return nomor ? `Dibayar — ${nomor}` : "Dibayar";
    case "diverifikasi":
      return nomor ? `Terverifikasi — ${nomor}` : "Terverifikasi";
    case "diinspeksi":
      return nomor ? `Buku induk tercatat — ${nomor}` : "Buku induk tercatat";
    case "divalidasi":
      return nomor ? `Selesai — ${nomor}` : "Selesai";
    case "ditolak":
      return nomor ? `Ditolak — ${nomor}` : "Ditolak";
    case "dibatalkan":
      return nomor ? `Dibatalkan — ${nomor}` : "Dibatalkan";
    default:
      return nomor || status;
  }
}

interface Props {
  // api dipakai bersama halaman induk agar header identitas konsisten.
  api: (path: string, opts?: RequestInit) => Promise<Response>;
  // awal diisi bila halaman sudah memuat draft/pengajuan aktif.
  awal: DataPermohonan | null;
  // selesai dipanggil setelah draft disimpan atau permohonan diajukan.
  onBerubah: (p: DataPermohonan) => void;
  // aktif = true bila peran aktif adalah pelanggan.
  aktif: boolean;
}

export function FormPermohonan({ api, awal, onBerubah, aktif }: Props) {
  const [permohonan, setPermohonan] = useState<DataPermohonan | null>(awal);
  const [nilai, setNilai] = useState<NilaiForm>(awal ? dariPermohonan(awal) : KOSONG);
  const [mode, setMode] = useState<Mode>(awal?.Status === "draft" ? "draft" : awal ? "terkirim" : "baru");
  const [sibuk, setSibuk] = useState<"draft" | "ajukan" | null>(null);
  const [pesan, setPesan] = useState("");
  const [galat, setGalat] = useState("");
  const [statusMuat, setStatusMuat] = useState<"memuat" | "selesai">(awal ? "selesai" : "memuat");

  // Terapkan data dari server: isian, mode, dan status dimutakhirkan bersama
  // sehingga mode tidak pernah bergantung pada render sebelumnya.
  const terapkan = useCallback(
    (data: DataPermohonan | null) => {
      setPermohonan(data);
      setNilai(data ? dariPermohonan(data) : KOSONG);
      setMode(data ? (data.Status === "draft" ? "draft" : "terkirim") : "baru");
      if (data) onBerubah(data);
    },
    [onBerubah],
  );

  // Muat draft/pengajuan berjalan milik pelanggan saat halaman dibuka tanpa
  // data awal, agar tombol "Edit Draft" langsung tersedia. 204 berarti belum
  // ada draft maupun pengajuan.
  //
  // dimuat menahan pemuatan ulang bila efek dijalankan dua kali (React
  // StrictMode di mode pengembangan): satu permintaan saja per muat halaman.
  const dimuat = useRef(false);
  useEffect(() => {
    if (awal || !aktif) {
      setStatusMuat("selesai");
      return;
    }
    if (dimuat.current) return;
    dimuat.current = true;
    (async () => {
      try {
        const res = await api("/v1/permohonan/aktif");
        if (res.status === 204) {
          setStatusMuat("selesai");
          return;
        }
        if (!res.ok) throw new Error(await bacaGalat(res));
        const data = (await res.json()) as DataPermohonan;
        terapkan(data);
      } catch (err) {
        setGalat(err instanceof Error ? err.message : String(err));
      } finally {
        setStatusMuat("selesai");
      }
    })();
  }, [api, awal, aktif, terapkan]);

  const ubah = useCallback((kunci: keyof NilaiForm, v: string) => {
    setNilai((lama) => ({ ...lama, [kunci]: v }));
  }, []);

  const kurang = useMemo(() => wajibIsi(nilai), [nilai]);
  // Jenis benih yang diketik di luar daftar saran: isian tetap disimpan sebagai
  // draft, tetapi tidak boleh diajukan sampai dipilih dari daftar.
  const jenisDiLuarDaftar = nilai.jenisBenih.trim() !== "" && !jenisDikenal(nilai.jenisBenih);
  const sudahDiajukan = permohonan !== null && permohonan.Status !== "draft";
  const terkunci = sudahDiajukan || sibuk !== null;
  const bisaDitulisi = !terkunci;

  const bodySimpan = useCallback(
    () =>
      JSON.stringify({
        draft_id: permohonan?.Status === "draft" ? permohonan.ID : 0,
        jenis_benih: nilai.jenisBenih,
        varietas: nilai.varietas,
        luas_lahan: nilai.luasLahan.trim() === "" ? 0 : Number(nilai.luasLahan),
        lokasi_kebun: nilai.lokasiKebun,
      }),
    [nilai, permohonan],
  );

  const simpanDraft = useCallback(async () => {
    setSibuk("draft");
    setPesan("");
    setGalat("");
    try {
      const res = await api("/v1/permohonan/draft", { method: "POST", body: bodySimpan() });
      if (!res.ok) throw new Error(await bacaGalat(res));
      const data = (await res.json()) as DataPermohonan;
      terapkan(data);
      setPesan("Draft tersimpan. Anda dapat melanjutkan mengisi dan mengajukan permohonan kapan saja.");
    } catch (err) {
      setGalat(err instanceof Error ? err.message : String(err));
    } finally {
      setSibuk(null);
    }
  }, [api, bodySimpan, terapkan]);

  // ajukan dipakai untuk draft tersimpan maupun permohonan baru. Backend
  // memakai draft_id dari body; nilai 0 berarti membuat draft lalu langsung
  // mengajukannya dalam satu langkah sehingga pelanggan tidak perlu menekan
  // "Simpan Draft" lebih dulu.
  const ajukan = useCallback(async () => {
    if (kurang.length > 0) {
      setPesan("");
      setGalat(`Lengkapi isian berikut sebelum mengajukan: ${kurang.join(", ")}.`);
      return;
    }
    if (jenisDiLuarDaftar) {
      setPesan("");
      setGalat(
        `Jenis benih "${nilai.jenisBenih.trim()}" belum ada di daftar komoditas. Pilih salah satu saran jenis benih sebelum mengajukan.`,
      );
      return;
    }
    setSibuk("ajukan");
    setPesan("");
    setGalat("");
    try {
      const res = await api("/v1/permohonan/ajukan", { method: "POST", body: bodySimpan() });
      if (!res.ok) throw new Error(await bacaGalat(res));
      const data = (await res.json()) as DataPermohonan;
      terapkan(data);
      setPesan(`Permohonan diajukan dengan nomor ${data.NomorPermohonan}. Langkah Analis Lab dan Penandatangan kini terbuka.`);
    } catch (err) {
      setGalat(err instanceof Error ? err.message : String(err));
    } finally {
      setSibuk(null);
    }
  }, [api, bodySimpan, jenisDiLuarDaftar, kurang, nilai.jenisBenih, terapkan]);

  const editDraft = useCallback(() => {
    if (!permohonan || permohonan.Status !== "draft") {
      setPesan("");
      setGalat("Permohonan yang sudah diajukan tidak dapat diubah lagi. Hubungi petugas bila perlu perbaikan.");
      return;
    }
    setMode("draft");
    setNilai(dariPermohonan(permohonan));
    setPesan("Mode edit draft: ubah isian lalu simpan kembali atau ajukan.");
    setGalat("");
  }, [permohonan]);

  if (!aktif) {
    return (
      <Card>
        <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-3">
          Fase 1 — Pengajuan Permohonan
        </h3>
        <p className="text-sm text-sitani-muted">
          Pengajuan permohonan hanya dapat dilakukan oleh peran Pelanggan. Langkah Fase 1–3 di atas
          akan aktif setelah ada permohonan yang diajukan.
        </p>
        {permohonan && (
          <p className="text-sm mt-3">
            Permohonan berikutnya diproses:{" "}
            <strong>{labelStatusPermohonan(permohonan.Status, permohonan.NomorPermohonan)}</strong>
          </p>
        )}
      </Card>
    );
  }

  return (
    <Card>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-3">
        <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider">
          Fase 1 — Pengajuan Permohonan
        </h3>
        {permohonan && (
          <Badge variant={permohonan.Status === "draft" ? "warning" : "success"}>
            {labelStatusPermohonan(permohonan.Status, permohonan.NomorPermohonan)}
          </Badge>
        )}
      </div>

      <p className="text-sm text-sitani-muted mb-4">
        Isi data permohonan sertifikasi. Anda dapat menyimpan sebagai <strong>draft</strong> lebih
        dulu; nomor permohonan baru dibuat saat permohonan diajukan. Langkah Analis Lab dan
        Penandatangan terbuka setelah permohonan diajukan.
      </p>
      <p className="text-xs text-sitani-muted mb-4">
        Form ini adalah form pelanggan untuk <strong>Fase 1</strong>. Fase 2 (uji lab mutu) dan
        Fase 3 (sertifikat, TTE, label) tidak punya form pengajuan pelanggan — keduanya dikerjakan
        petugas internal atas permohonan yang sudah diajukan.
      </p>

      {statusMuat === "memuat" && (
        <p className="text-sm text-sitani-muted mb-3" role="status">
          Memuat draft permohonan…
        </p>
      )}

      <form
        className="grid grid-cols-1 sm:grid-cols-2 gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          void ajukan();
        }}
      >
        <Autocomplete
          id="form-jenis-benih"
          nama="jenis_benih"
          label="Jenis benih"
          nilai={nilai.jenisBenih}
          onUbah={(v) => ubah("jenisBenih", v)}
          opsi={OPSI_JENIS_BENIH}
          disabled={!bisaDitulisi}
          wajib
          pesanKosong="Jenis benih wajib diisi."
          placeholder="Ketik atau pilih jenis benih, mis. Padi"
          bantuan="Pilih dari saran atau ketik nama komoditas yang belum tercantum."
          onPilih={() => ubah("varietas", "")}
        />
        <Autocomplete
          id="form-varietas"
          nama="varietas"
          label="Varietas"
          nilai={nilai.varietas}
          onUbah={(v) => ubah("varietas", v)}
          opsi={OPSI_VARIETAS}
          induk={nilai.jenisBenih}
          indukLabel="Jenis benih"
          disabled={!bisaDitulisi}
          wajib
          pesanKosong="Varietas wajib diisi."
          placeholder="Ketik atau pilih varietas"
          bantuan="Saran varietas mengikuti jenis benih. Ketik langsung bila varietas belum tercantum."
        />
        <div className="form-group">
          <label className="form-label" htmlFor="luas-lahan">
            Luas lahan (ha)
          </label>
          <input
            id="luas-lahan"
            name="luas_lahan"
            type="number"
            min="0.0001"
            step="0.0001"
            className="form-input"
            value={nilai.luasLahan}
            onChange={(e) => ubah("luasLahan", e.target.value)}
            disabled={!bisaDitulisi}
            placeholder="Contoh: 0.5"
          />
        </div>
        <div className="form-group">
          <label className="form-label" htmlFor="lokasi-kebun">
            Lokasi kebun
          </label>
          <input
            id="lokasi-kebun"
            name="lokasi_kebun"
            className="form-input"
            value={nilai.lokasiKebun}
            onChange={(e) => ubah("lokasiKebun", e.target.value)}
            disabled={!bisaDitulisi}
            placeholder="Contoh: Sleman"
          />
        </div>

        {mode !== "terkirim" && kurang.length > 0 && (
          <p className="text-xs text-sitani-muted sm:col-span-2">
            Belum lengkap: {kurang.join(", ")}. Draft tetap dapat disimpan.
          </p>
        )}

        {mode !== "terkirim" && jenisDiLuarDaftar && (
          <p className="text-xs text-sitani-muted sm:col-span-2" id="peringatan-jenis-benih">
            Jenis benih “{nilai.jenisBenih.trim()}” belum ada di daftar komoditas. Pilih salah satu
            saran agar permohonan dapat diajukan; draft tetap dapat disimpan.
          </p>
        )}

        <div className="sm:col-span-2 flex flex-wrap gap-3">
          {(mode === "baru" || mode === "draft") && (
            <>
              <Button type="button" variant="outline" onClick={simpanDraft} disabled={sibuk !== null}>
                {sibuk === "draft" ? "Menyimpan…" : "Simpan Draft"}
              </Button>
              <Button type="submit" variant="primary" disabled={sibuk !== null}>
                {sibuk === "ajukan" ? "Mengajukan…" : "Ajukan Permohonan"}
              </Button>
              {mode === "draft" && (
                <Button type="button" variant="secondary" onClick={editDraft} disabled={sibuk !== null}>
                  Edit Draft
                </Button>
              )}
            </>
          )}
          {mode === "terkirim" && (
            <span className="text-xs text-sitani-muted self-center">
              Permohonan sudah diajukan dengan nomor{" "}
              <strong>{permohonan?.NomorPermohonan || "-"}</strong>. Isian dikunci; hubungi petugas
              untuk perubahan atau pembatalan.
            </span>
          )}
        </div>
      </form>

      {mode === "terkirim" && (
        <p className="text-xs text-sitani-muted mt-3">
          Permohonan sudah diajukan sehingga isian aplikasi dikunci. Bila perlu perbaikan atau
          pembatalan, hubungi petugas layanan; riwayat pengajuan tercatat pada audit trail dan
          tidak diubah dari halaman ini.
        </p>
      )}

      {pesan && (
        <p className="alert alert-success mt-4" role="status">
          {pesan}
        </p>
      )}
      {galat && (
        <p className="alert alert-error mt-4" role="alert">
          {galat}
        </p>
      )}
    </Card>
  );
}

async function bacaGalat(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    return data.error || `HTTP ${res.status}`;
  } catch {
    return `HTTP ${res.status}`;
  }
}

