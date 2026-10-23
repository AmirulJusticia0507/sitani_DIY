"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import {
  apiURL,
  devLoginTersedia,
  getRealmLabel,
  getRoleLabel,
  type DashboardPrincipal,
} from "../../components/dashboard/principal";
import { getDashboardNav, getRoleAccess, isRole } from "../../components/dashboard/role-access";

// Ringkasan membaca sesi di klien karena cookie SSO HttpOnly hanya berlaku
// untuk backend. Semua hitungan di sini berasal dari sesi nyata; tidak ada
// angka jumlah permohonan yang dikarang sebelum endpoint statistik tersedia.
export function RingkasanKlien() {
  const [principal, setPrincipal] = useState<DashboardPrincipal | null>(null);
  const [status, setStatus] = useState<"memuat" | "masuk" | "tamu">("memuat");
  const [devLogin, setDevLogin] = useState(false);

  useEffect(() => {
    let active = true;
    devLoginTersedia()
      .then((tersedia) => {
        if (active) setDevLogin(tersedia);
        return fetch(`${apiURL}/auth/me`, { credentials: "include", cache: "no-store" });
      })
      .then((res) => (res && res.ok ? res.json() : null))
      .then((data: DashboardPrincipal | null) => {
        if (!active) return;
        if (data && typeof data.role === "string") {
          setPrincipal(data);
          setStatus("masuk");
        } else {
          setStatus("tamu");
        }
      })
      .catch(() => {
        if (active) setStatus("tamu");
      });
    return () => {
      active = false;
    };
  }, []);

  if (status === "memuat") return <p className="dashboard-note">Memeriksa sesi…</p>;

  if (status === "tamu" || !principal) {
    return (
      <section className="dashboard-panel" aria-labelledby="masuk-judul">
        <h2 id="masuk-judul">Masuk untuk membuka area kerja</h2>
        <p>
          Dashboard menampilkan modul sesuai peran Anda. Pengguna masyarakat memakai akun JSP,
          sedangkan ASN memakai akun Satriya.
        </p>
        <div className="dashboard-actions">
          {devLogin && (
            <Link href="/dashboard/masuk" className="btn btn-primary">Masuk lokal (dev)</Link>
          )}
          <a href={`${apiURL}/auth/login/jsp?redirect=/dashboard`} className="btn btn-primary">Masuk JSP</a>
          <a href={`${apiURL}/auth/login/satriya?redirect=/dashboard`} className="btn btn-outline">Masuk Satriya</a>
        </div>
        <p className="dashboard-note">
          {devLogin
            ? "Login lokal aktif di backend ini untuk development/UAT. Akun dibuat lewat perintah seed-dev; jangan dipakai di produksi."
            : "Login terintegrasi bergantung pada aktivasi client SSO oleh pengelola Jogjaprov. Selama aktivasi belum selesai, tautan masuk dapat berakhir dengan pesan kesalahan dari backend."}
        </p>
      </section>
    );
  }

  const nav = getDashboardNav(principal.role);
  const modulPortal = nav.filter((item) => item.key !== "ringkasan" && item.key !== "profil");
  const akses = getRoleAccess(principal.role);

  return (
    <div className="dashboard-stack">
      <section className="dashboard-panel" aria-labelledby="sesi-judul">
        <div className="dashboard-panel-head">
          <div>
            <p className="eyebrow">Sesi aktif</p>
            <h2 id="sesi-judul">{principal.name || principal.email}</h2>
          </div>
          <span className={`badge ${isRole(principal.role) ? "badge-success" : "badge-warning"}`}>
            {getRoleLabel(principal)}
          </span>
        </div>
        <dl className="dashboard-data">
          <div>
            <dt>Email</dt>
            <dd>{principal.email}</dd>
          </div>
          <div>
            <dt>Username</dt>
            <dd>{principal.username || "-"}</dd>
          </div>
          <div>
            <dt>Realm SSO</dt>
            <dd>{getRealmLabel(principal)}</dd>
          </div>
          <div>
            <dt>Modul tersedia</dt>
            <dd>{akses.size} modul</dd>
          </div>
        </dl>
        {!isRole(principal.role) && (
          <p className="dashboard-callout" role="status">
            Peran <strong>{principal.role}</strong> belum dipetakan ke modul dashboard. Sampaikan
            peran seharusnya kepada administrator agar akses dapat disesuaikan di backend.
          </p>
        )}
      </section>

      <section aria-labelledby="modul-judul">
        <div className="dashboard-section-head">
          <h2 id="modul-judul">Modul kerja Anda</h2>
          <p>Menu berikut mengikuti peran yang diberikan backend. Otorisasi tetap diperiksa server pada setiap aksi.</p>
        </div>
        {modulPortal.length === 0 ? (
          <p className="dashboard-note">Belum ada modul portal untuk peran ini.</p>
        ) : (
          <ul className="dashboard-tiles">
            {modulPortal.map((item) => (
              <li key={item.key}>
                <Link href={item.href} className="dashboard-tile">
                  <span className="dashboard-tile-label">{item.label}</span>
                  <span className="dashboard-tile-desc">{item.deskripsi}</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="dashboard-panel" aria-labelledby="langkah-judul">
        <h2 id="langkah-judul">Langkah berikutnya</h2>
        <ol className="dashboard-steps">
          <li>
            Lengkapi <Link href="/dashboard/profil">profil dan akses</Link> untuk memastikan peran dan
            realm SSO Anda sesuai tugas.
          </li>
          <li>
            Buka modul kerja utama sesuai kebutuhan. Setiap pengajuan mencatat audit trail di backend.
          </li>
          <li>
            Jika terjadi kendala, lihat <Link href="/bantuan">Bantuan</Link> dan hubungi administrator
            melalui saluran resmi.
          </li>
        </ol>
      </section>
    </div>
  );
}
