"use client";

import { useEffect, useState } from "react";

import {
  apiURL,
  getRealmLabel,
  getRoleLabel,
  type DashboardPrincipal,
} from "../../../components/dashboard/principal";
import { getDashboardNav, isRole } from "../../../components/dashboard/role-access";

export function ProfilKlien() {
  const [principal, setPrincipal] = useState<DashboardPrincipal | null>(null);
  const [status, setStatus] = useState<"memuat" | "masuk" | "tamu">("memuat");

  useEffect(() => {
    let active = true;
    fetch(`${apiURL}/auth/me`, { credentials: "include", cache: "no-store" })
      .then((res) => (res.ok ? res.json() : null))
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
      <section className="dashboard-panel">
        <h2>Sesi belum aktif</h2>
        <p>Halaman profil menampilkan identitas dari sesi SSO. Masuk terlebih dahulu melalui JSP atau Satriya.</p>
        <div className="dashboard-actions">
          <a href={`${apiURL}/auth/login/jsp?redirect=/dashboard/profil`} className="btn btn-primary">Masuk JSP</a>
          <a href={`${apiURL}/auth/login/satriya?redirect=/dashboard/profil`} className="btn btn-outline">Masuk Satriya</a>
        </div>
      </section>
    );
  }

  const modul = getDashboardNav(principal.role);

  return (
    <div className="dashboard-stack">
      <section className="dashboard-panel" aria-labelledby="identitas-judul">
        <div className="dashboard-panel-head">
          <div>
            <p className="eyebrow">Identitas sesi</p>
            <h2 id="identitas-judul">{principal.name || principal.email}</h2>
          </div>
          <span className={`badge ${isRole(principal.role) ? "badge-success" : "badge-warning"}`}>
            {getRoleLabel(principal)}
          </span>
        </div>
        <dl className="dashboard-data">
          <div>
            <dt>Nama</dt>
            <dd>{principal.name || "-"}</dd>
          </div>
          <div>
            <dt>Email</dt>
            <dd>{principal.email}</dd>
          </div>
          <div>
            <dt>Username</dt>
            <dd>{principal.username || "-"}</dd>
          </div>
          <div>
            <dt>Peran</dt>
            <dd>{principal.role}</dd>
          </div>
          <div>
            <dt>Realm SSO</dt>
            <dd>{getRealmLabel(principal)}</dd>
          </div>
          <div>
            <dt>ID pengguna</dt>
            <dd>{principal.user_id ?? "-"}</dd>
          </div>
        </dl>
      </section>

      <section className="dashboard-panel" aria-labelledby="akses-judul">
        <h2 id="akses-judul">Modul yang dapat dibuka</h2>
        <ul className="dashboard-access">
          {modul.map((item) => (
            <li key={item.key}>
              <strong>{item.label}</strong>
              <span>{item.deskripsi}</span>
            </li>
          ))}
        </ul>
        <p className="dashboard-note">
          Daftar ini berasal dari pemetaan peran di antarmuka. Backend tetap memeriksa kewenangan
          pada setiap permintaan, sehingga penolakan akses dapat terjadi bila peran di server berbeda.
        </p>
      </section>

      <section className="dashboard-panel" aria-labelledby="perubahan-judul">
        <h2 id="perubahan-judul">Perubahan data diri</h2>
        <p>
          Perubahan nama, email, atau peran mengikuti data akun pada penyedia identitas (JSP/Satriya)
          dan pengelolaan pengguna di backend. Portal belum menyediakan formulir ubah data mandiri;
          ajukan perubahan melalui administrator atau kanal resmi DPKP DIY.
        </p>
      </section>
    </div>
  );
}
