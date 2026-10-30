"use client";

import { useEffect, useState, useCallback } from "react";
import { Layout } from "../../components/layout";
import { Card } from "../../components/card";
import { Badge } from "../../components/badge";
import { Button } from "../../components/button";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ||
  process.env.NEXT_PUBLIC_BACKEND_URL ||
  "http://localhost:8080";

type Role =
  | "pelanggan"
  | "admin"
  | "pengawas"
  | "analis"
  | "penandatangan"
  | "legalisasi"
  | "percetakan";

// Peran yang dapat dipilih dari halaman ini, berurutan sesuai alur kerja
// (pemohon → laboratorium → penandatangan). Urutan tampil = urutan step bar.
const PERAN: ReadonlyArray<{ id: Role; label: string; keterangan: string }> = [
  { id: "pelanggan", label: "Pelanggan", keterangan: "Mengajukan & memantau" },
  { id: "analis", label: "Analis Lab", keterangan: "Menguji mutu sampel" },
  { id: "penandatangan", label: "Penandatangan", keterangan: "Menandatangani hasil" },
];

// Ikon stepper mengikuti pola Material Tailwind (user / settings / home).
const IKON_PELANGGAN = (
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" aria-hidden="true" className="w-5 h-5">
    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
  </svg>
);
const IKON_ANALIS = (
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" aria-hidden="true" className="w-5 h-5">
    <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12a7.5 7.5 0 0015 0m-15 0a7.5 7.5 0 1115 0m-15 0H3m16.5 0H21m-1.5 0H12m-8.457 3.077l1.41-.513m14.095-5.13l1.41-.513M5.106 17.785l1.15-.964m11.49-9.642l1.149-.964M7.501 19.795l.75-1.3m7.5-12.99l.75-1.3m-6.063 16.658l.26-1.477m2.605-14.772l.26-1.477m0 17.726l-.26-1.477M10.698 4.614l-.26-1.477M16.5 19.794l-.75-1.299M7.5 4.205L12 12m6.894 5.785l-1.149-.964M6.256 7.178l-1.15-.964m15.352 8.864l-1.41-.513M4.954 9.435l-1.41-.514M12.002 12l-3.75 6.495" />
  </svg>
);
const IKON_PENANDATANGAN = (
  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" aria-hidden="true" className="w-5 h-5">
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0012 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75z" />
  </svg>
);

interface Permohonan {
  ID: number;
  NomorPermohonan: string;
  PelangganID: number;
  JenisBenih: string;
  Varietas: string;
  LuasLahan: number;
  LokasiKebun: string;
  Status: string;
  PembayaranAwalStatus: string;
}

interface Sampel {
  ID: number;
  KodeSampel: string;
  Status: string;
}

export default function CertificationPage() {
  const [role, setRole] = useState<Role>("pelanggan");
  const [permohonan, setPermohonan] = useState<Permohonan | null>(null);
  const [sampel, setSampel] = useState<Sampel | null>(null);
  const [statusMsg, setStatusMsg] = useState("");
  const [error, setError] = useState("");

  const api = useCallback(
    (path: string, opts: RequestInit = {}) => {
      const headers: Record<string, string> = {
        "X-User-ID": "1",
        "X-User-Role": role,
        "Content-Type": "application/json",
        ...(opts.headers as Record<string, string>),
      };
      return fetch(`${API_BASE}${path}`, { ...opts, headers });
    },
    [role],
  );

  const setInfo = useCallback((msg: string, err: string | null = null) => {
    setStatusMsg(msg);
    setError(err || "");
  }, []);

  useEffect(() => {
    setInfo("");
  }, [role, setInfo]);

  const buatPermohonan = async (e: React.FormEvent) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget as HTMLFormElement);
    const body = JSON.stringify({
      jenis_benih: form.get("jenis_benih"),
      varietas: form.get("varietas"),
      luas_lahan: parseFloat(String(form.get("luas_lahan"))),
      lokasi_kebun: form.get("lokasi_kebun"),
    });
    try {
      const res = await api("/v1/permohonan", {
        method: "POST",
        body,
      });
      if (!res.ok) throw new Error(await res.text());
      const data: Permohonan = await res.json();
      setPermohonan(data);
      setInfo("Permohonan dibuat: " + data.NomorPermohonan);
    } catch (err) {
      setError(String(err));
    }
  };

  const bayarAwal = async (id: number) => {
    try {
      const res = await api(`/v1/permohonan/${id}/pembayaran/awal`, {
        method: "POST",
        body: JSON.stringify({ jumlah: 1000000 }),
      });
      if (!res.ok) throw new Error(await res.text());
      setInfo("Pembayaran awal dicatat");
    } catch (err) {
      setError(String(err));
    }
  };

  const catatInspeksi = async (id: number, fase: number, hasil: string) => {
    try {
      const res = await api(`/v1/permohonan/${id}/inspeksi`, {
        method: "POST",
        body: JSON.stringify({ fase, hasil }),
      });
      if (!res.ok) throw new Error(await res.text());
      setInfo(`Inspeksi fase ${fase} dicatat: ${hasil}`);
    } catch (err) {
      setError(String(err));
    }
  };

  const catatBukuInduk = async (id: number, nomor: string) => {
    try {
      const res = await api(`/v1/permohonan/${id}/buku-induk`, {
        method: "POST",
        body: JSON.stringify({ nomor_buku_induk: nomor }),
      });
      if (!res.ok) throw new Error(await res.text());
      setInfo("Buku induk dicatat");
    } catch (err) {
      setError(String(err));
    }
  };

  const ambilSampel = async (id: number) => {
    try {
      const res = await api("/v1/sampel", {
        method: "POST",
        body: JSON.stringify({ permohonan_id: id, jumlah_sampel: 2 }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data: Sampel = await res.json();
      setSampel(data);
      setInfo(`Sampel ambil: ${data.KodeSampel || data.ID}`);
    } catch (err) {
      setError(String(err));
    }
  };

  const prosesMutu = async (sampelId: number) => {
    try {
      const res = await api("/v1/proses-mutu", {
        method: "POST",
        body: JSON.stringify({ sampel_id: sampelId }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data: { hasil: string } = await res.json();
      setInfo("Hasil mutu: " + data.hasil);
    } catch (err) {
      setError(String(err));
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div className="container">
          <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4 mb-2">
            <div>
              <p className="eyebrow">Layanan sertifikasi benih</p>
              <h1 className="text-2xl font-bold text-sitani-green-800">Sertifikasi Benih BP3MBTP</h1>
              <p className="text-sitani-muted text-sm">
                Alur 7 aktor berdasarkan diagram XML (Fase 1–3).
              </p>
            </div>
            <p className="text-xs text-sitani-muted">
              Peran aktif: <strong className="text-sitani-green-800">{PERAN.find((p) => p.id === role)?.label}</strong>
            </p>
          </div>

          {/* Stepper peran mengikuti struktur Material Tailwind stepper: garis penuh di
              belakang, lingkaran 40px berisi ikon, label di bawah setiap langkah, tombol
              Prev/Next. Langkah aktif/selesai hijau solid, langkah pending abu. */}
          <div className="stepper-mt" role="group" aria-label="Peran dalam alur sertifikasi">
            <div className="stepper-mt-jalur" aria-hidden="true">
              <span
                className="stepper-mt-garis"
                aria-hidden="true"
              />
              <span
                className="stepper-mt-garis-isi"
                style={{
                  width: `${(PERAN.findIndex((p) => p.id === role) / (PERAN.length - 1)) * 100}%`,
                }}
              />
              {PERAN.map((item) => {
                const idx = PERAN.findIndex((p) => p.id === item.id);
                const idxAktif = PERAN.findIndex((p) => p.id === role);
                const aktif = role === item.id;
                const selesai = idx < idxAktif;
                const ikon = item.id === "pelanggan" ? IKON_PELANGGAN : item.id === "analis" ? IKON_ANALIS : IKON_PENANDATANGAN;
                return (
                  <button
                    key={item.id}
                    type="button"
                    className={`stepper-mt-tombol${aktif ? " is-aktif" : ""}${selesai ? " is-selesai" : ""}${!aktif && !selesai ? " is-pending" : ""}`}
                    aria-current={aktif ? "step" : undefined}
                    onClick={() => setRole(item.id)}
                  >
                    {ikon}
                    <span className="stepper-mt-label">
                      <span className="stepper-mt-label-judul">{item.label}</span>
                      <span className="stepper-mt-label-detail">{item.keterangan}</span>
                    </span>
                  </button>
                );
              })}
            </div>

            <div className="stepper-mt-footer">
              <Button
                variant="outline"
                size="sm"
                disabled={role === PERAN[0].id}
                onClick={() => setRole(PERAN[Math.max(0, PERAN.findIndex((p) => p.id === role) - 1)].id)}
              >
                Prev
              </Button>
              <Button
                variant="primary"
                size="sm"
                disabled={role === PERAN[PERAN.length - 1].id}
                onClick={() => setRole(PERAN[Math.min(PERAN.length - 1, PERAN.findIndex((p) => p.id === role) + 1)].id)}
              >
                Next
              </Button>
            </div>
          </div>
        </div>

        {error && (
          <div className="container alert alert-error" role="alert">
            {error}
          </div>
        )}
        {statusMsg && (
          <div className="container alert alert-success" role="status">
            {statusMsg}
          </div>
        )}

        <div className="container">
          {permohonan ? (
            <>
              <div className="flex items-center justify-between gap-4">
                <div>
                  <h2 className="text-xl font-bold text-sitani-green-800">
                    Permohonan {permohonan.NomorPermohonan}
                  </h2>
                  <p className="text-sitani-muted text-sm">
                    Status:{" "}
                    <Badge variant="success">{permohonan.Status}</Badge>{" "}
                    | Pembayaran awal:{permohonan.PembayaranAwalStatus}
                  </p>
                </div>
              </div>

              <section className="grid grid-cols-2 md:grid-cols-3 gap-4">
                <Card>
                  <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-3">
                    Fase 1 — Verifikasi Lapangan
                  </h3>
                  {!permohonan.PembayaranAwalStatus.startsWith("lunas") && (
                    <Button variant="secondary" size="sm" onClick={() => bayarAwal(permohonan.ID)}>
                      Bayar Awal
                    </Button>
                  )}
                  <Button size="sm" onClick={() => catatInspeksi(permohonan.ID, 1, "lulus")}>
                    Inspeksi Fase 1 (Lulus)
                  </Button>
                  <Button size="sm" onClick={() => catatInspeksi(permohonan.ID, 2, "lulus")}>
                    Inspeksi Fase 2 (Lulus)
                  </Button>
                  <Button size="sm" onClick={() => catatInspeksi(permohonan.ID, 3, "lulus")}>
                    Inspeksi Fase 3 (Lulus)
                  </Button>
                  <Button size="sm" onClick={() => catatInspeksi(permohonan.ID, 4, "lulus")}>
                    Inspeksi Fase 4 (Lulus)
                  </Button>
                  <form
                    className="mt-4"
                    onSubmit={(e) => {
                      e.preventDefault();
                      const form = new FormData(e.currentTarget as HTMLFormElement);
                      catatBukuInduk(permohonan.ID, String(form.get("nomor")));
                    }}
                  >
                    <label className="form-label">Nomor Buku Induk</label>
                    <input
                      name="nomor"
                      className="form-input"
                      placeholder="Nomor Buku Induk"
                      required
                    />
                    <Button type="submit" variant="primary" size="sm" className="mt-2 w-full">
                      Catat
                    </Button>
                  </form>
                </Card>

                <Card>
                  <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-3">
                    Fase 2 — Pengujian Lab Mutu
                  </h3>
                  <Button variant="secondary" size="sm" onClick={() => ambilSampel(permohonan.ID)}>
                    Ambil Sampel
                  </Button>
                  {sampel && (
                    <>
                      <p className="text-sm text-sitani-muted mb-3">
                        Sampel: <strong>{sampel.KodeSampel}</strong> ({sampel.Status})
                      </p>
                      <Button size="sm" onClick={() => prosesMutu(sampel.ID)}>
                        Proses Hasil Mutu
                      </Button>
                    </>
                  )}
                </Card>

                <Card>
                  <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-3">
                    Fase 3 — Sertifikasi & Label
                  </h3>
                  <p className="text-xs text-sitani-muted leading-relaxed mb-4">
                    Jalur lulus: Sertifikat → TTE → Draft Label → SIM Perbenihan →
                    Percetakan → Legalisasi → Penyerahan.
                  </p>
                  <p className="text-xs text-sitani-muted">
                    Jalur tidak lulus: kirim laporan lengkap, proses selesai.
                  </p>
                </Card>
              </section>
            </>
          ) : (
            <section>
              <h3 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-3">
                Fase 1 — Ajukan Permohonan
              </h3>
              <Card>
                <form onSubmit={buatPermohonan} className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="form-group">
                    <label className="form-label">Jenis benih</label>
                    <input name="jenis_benih" className="form-input" required />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Varietas</label>
                    <input name="varietas" className="form-input" required />
                  </div>
                  <div className="form-group sm:col-span-2">
                    <label className="form-label">Luas lahan (ha)</label>
                    <input name="luas_lahan" type="number" className="form-input" required />
                  </div>
                  <div className="form-group sm:col-span-2">
                    <label className="form-label">Lokasi kebun</label>
                    <input name="lokasi_kebun" className="form-input" required />
                  </div>
                  <div className="form-group sm:col-span-2">
                    <Button type="submit" variant="primary" className="w-auto">
                      Ajukan Permohonan
                    </Button>
                  </div>
                </form>
              </Card>
            </section>
          )}
        </div>
      </div>
    </Layout>
  );
}
