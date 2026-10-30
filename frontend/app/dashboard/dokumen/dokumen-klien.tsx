"use client";

import { useCallback, useEffect, useState } from "react";

import { masukHref } from "../../../components/auth/masuk-url";
import {
  apiURL,
  devLoginTersedia,
  fetchSession,
  type DashboardPrincipal,
} from "../../../components/dashboard/principal";

// Baris sesuai respons `GET /v1/dokumen` (field struct db.Dokumen tanpa tag JSON).
type Dokumen = {
  ID: number;
  PemilikID: number;
  Tujuan: string;
  ObjectKey: string;
  NamaAsli: string;
  Jenis: string;
  Ukuran: number;
  CreatedAt: string;
};

type Status = "memuat" | "masuk" | "tamu" | "galat";

function formatUkuran(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "-";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatWaktu(nilai: string): string {
  const waktu = new Date(nilai);
  if (Number.isNaN(waktu.getTime())) return "-";
  return waktu.toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
}

export function DokumenKlien() {
  const [status, setStatus] = useState<Status>("memuat");
  const [principal, setPrincipal] = useState<DashboardPrincipal | null>(null);
  const [dokumen, setDokumen] = useState<Dokumen[]>([]);
  const [pesan, setPesan] = useState<string | null>(null);
  const [devLogin, setDevLogin] = useState(false);

  const muat = useCallback(async (aktif: () => boolean) => {
    setStatus("memuat");
    setPesan(null);
    const sesi = await fetchSession();
    if (!aktif()) return;
    if (sesi.status !== "masuk") {
      setStatus("tamu");
      return;
    }
    setPrincipal(sesi.principal);
    try {
      const res = await fetch(`${apiURL}/v1/dokumen`, { credentials: "include", cache: "no-store" });
      if (!aktif()) return;
      if (!res.ok) {
        const data = (await res.json().catch(() => null)) as { error?: string } | null;
        setPesan(data?.error ?? `Daftar dokumen gagal dimuat (HTTP ${res.status}).`);
        setStatus("galat");
        return;
      }
      const data = (await res.json()) as { dokumen?: Dokumen[] };
      setDokumen(Array.isArray(data.dokumen) ? data.dokumen : []);
      setStatus("masuk");
    } catch {
      if (!aktif()) return;
      setPesan("Backend tidak dapat dihubungi. Pastikan server API berjalan.");
      setStatus("galat");
    }
  }, []);

  useEffect(() => {
    let aktif = true;
    void muat(() => aktif);
    devLoginTersedia().then((tersedia) => {
      if (aktif) setDevLogin(tersedia);
    });
    return () => {
      aktif = false;
    };
  }, [muat]);

  if (status === "memuat") return <p className="dashboard-note">Memeriksa sesi dan memuat dokumen…</p>;

  if (status === "tamu") {
    return (
      <section className="dashboard-panel">
        <h2>Sesi belum aktif</h2>
        <p>Daftar dokumen hanya tersedia untuk pengguna yang sudah masuk.</p>
        <div className="dashboard-actions">
          {devLogin && (
            <a href="/dashboard/masuk" className="btn btn-outline">
              Masuk lokal (dev)
            </a>
          )}
          <a href={masukHref("/dashboard/dokumen", "jsp")} className="btn btn-primary">
            Masuk JSP
          </a>
          <a href={masukHref("/dashboard/dokumen", "satriya")} className="btn btn-outline">
            Masuk Satriya
          </a>
        </div>
      </section>
    );
  }

  return (
    <section className="dashboard-panel" aria-labelledby="daftar-dokumen-judul">
      <div className="dashboard-panel-head">
        <div>
          <p className="eyebrow">Lampiran tersimpan</p>
          <h2 id="daftar-dokumen-judul">
            {dokumen.length} dokumen
            {principal ? ` — ${principal.name || principal.email}` : ""}
          </h2>
        </div>
        <button type="button" className="btn btn-outline" onClick={() => void muat(() => true)}>
          Muat ulang
        </button>
      </div>

      {pesan && (
        <p className="masuk-lokal-pesan" role="alert">
          {pesan}
        </p>
      )}

      {dokumen.length === 0 ? (
        <p className="dashboard-note">
          Belum ada lampiran. Dokumen terunggah saat Anda melengkapi permohonan.
        </p>
      ) : (
        <div className="dashboard-table-wrapper">
          <table className="dashboard-table">
          <thead>
            <tr>
              <th scope="col">Nama berkas</th>
              <th scope="col">Tujuan</th>
              <th scope="col">Jenis</th>
              <th scope="col">Ukuran</th>
              <th scope="col">Diunggah</th>
              <th scope="col">Unduh</th>
            </tr>
          </thead>
          <tbody>
            {dokumen.map((item) => (
              <tr key={item.ID}>
                <td>{item.NamaAsli || "-"}</td>
                <td>{item.Tujuan || "-"}</td>
                <td>{item.Jenis || "-"}</td>
                <td>{formatUkuran(item.Ukuran)}</td>
                <td>{formatWaktu(item.CreatedAt)}</td>
                <td>
                  <a
                    className="btn btn-outline"
                    href={`${apiURL}/v1/dokumen/${item.ID}/unduh`}
                    target="_blank"
                    rel="noreferrer"
                  >
                    Unduh
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
          </table>
        </div>
      )}
      <p className="dashboard-note">
        Tautan unduh dibuat backend dengan masa berlaku singkat dan hanya diberikan bila Anda
        berhak atas berkas tersebut.
      </p>
    </section>
  );
}
