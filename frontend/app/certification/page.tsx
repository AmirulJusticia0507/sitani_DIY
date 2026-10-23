"use client";

import { useEffect, useState, useCallback } from "react";

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

  const roleLabel: Record<Role, string> = {
    pelanggan: "Pelanggan",
    admin: "Admin",
    pengawas: "Pengawas / PBT",
    analis: "Analis Lab",
    penandatangan: "Penandatangan TTE",
    legalisasi: "Legalisasi",
    percetakan: "Percetakan",
  };

  return (
    <main style={{ padding: 24, fontFamily: "system-ui, sans-serif" }}>
      <h1>Sertifikasi Benih BP3MBTP</h1>
      <p>Alur 7 aktor berdasarkan diagram XML (Fase 1–3).</p>

      <label style={{ marginRight: 8 }}>
        Peran:{" "}
        <select
          value={role}
          onChange={(e) => setRole(e.currentTarget.value as Role)}
        >
          {Object.entries(roleLabel).map(([k, v]) => (
            <option key={k} value={k as Role}>
              {v}
            </option>
          ))}
        </select>
      </label>
      {error && <p style={{ color: "brown" }}>{error}</p>}
      {statusMsg && <p style={{ color: "green" }}>{statusMsg}</p>}

      {permohonan ? (
        <>
          <h2>Permohonan {permohonan.NomorPermohonan}</h2>
          <p>
            Status: <b>{permohonan.Status}</b> | Pembayaran awal:
            {" "}{permohonan.PembayaranAwalStatus}
          </p>

          <section style={{ marginTop: 16 }}>
            <h3>Fase 1 — Verifikasi Lapangan</h3>
            {!permohonan.PembayaranAwalStatus.startsWith("lunas") && (
              <button onClick={() => bayarAwal(permohonan.ID)}>
                Bayar Awal
              </button>
            )}
            <button onClick={() => catatInspeksi(permohonan.ID, 1, "lulus")}>
              Inspeksi Fase 1 (Lulus)
            </button>
            <button onClick={() => catatInspeksi(permohonan.ID, 2, "lulus")}>
              Inspeksi Fase 2 (Lulus)
            </button>
            <button onClick={() => catatInspeksi(permohonan.ID, 3, "lulus")}>
              Inspeksi Fase 3 (Lulus)
            </button>
            <button onClick={() => catatInspeksi(permohonan.ID, 4, "lulus")}>
              Inspeksi Fase 4 (Lulus)
            </button>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                const form = new FormData(e.currentTarget as HTMLFormElement);
                catatBukuInduk(permohonan.ID, String(form.get("nomor")));
              }}
            >
              <input name="nomor" placeholder="Nomor Buku Induk" />
              <button type="submit">Catat Buku Induk</button>
            </form>
          </section>

          <section style={{ marginTop: 16 }}>
            <h3>Fase 2 — Pengujian Lab Mutu</h3>
            <button onClick={() => ambilSampel(permohonan.ID)}>Ambil Sampel</button>
            {sampel && (
              <>
                <p>Sampel: {sampel.KodeSampel} ({sampel.Status})</p>
                <button onClick={() => prosesMutu(sampel.ID)}>
                  Proses Hasil Mutu (Auto Lulus/Tidak Lulus)
                </button>
              </>
            )}
          </section>

          <section style={{ marginTop: 16 }}>
            <h3>Fase 3 — Sertifikasi & Penerbitan Label</h3>
            <p>
              (Jalur lulus: Sertifikat → TTE → Draft Label → SIM Perbenihan →
              Percetakan → Legalisasi → Penyerahan. Jalur tidak lulus: kirim
              laporan lengkap, proses selesai.)
            </p>
          </section>
        </>
      ) : (
        <section style={{ marginTop: 16 }}>
          <h3>Fase 1 — Ajukan Permohonan</h3>
          <form onSubmit={buatPermohonan}>
            <input
              name="jenis_benih"
              placeholder="Jenis benih"
              required
              style={{ marginRight: 8 }}
            />
            <input
              name="varietas"
              placeholder="Varietas"
              required
              style={{ marginRight: 8 }}
            />
            <input
              name="luas_lahan"
              type="number"
              placeholder="Luas lahan"
              required
              style={{ marginRight: 8 }}
            />
            <input
              name="lokasi_kebun"
              placeholder="Lokasi kebun"
              required
            />
            <button type="submit">Ajukan</button>
          </form>
        </section>
      )}
    </main>
  );
}
