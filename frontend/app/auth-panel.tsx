"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { akhiriSesi, apiURL, devLoginTersedia } from "../components/dashboard/principal";
import { masukHref } from "../components/auth/masuk-url";

type Principal = {
  email: string;
  name: string;
  role: string;
  realm: string;
};

// Panjang sesi ringkas untuk panel portal. Semua data berasal dari /auth/me;
// tidak ada identitas yang diambil dari query atau localStorage.
export function AuthPanel() {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [loading, setLoading] = useState(true);
  const [devLogin, setDevLogin] = useState(false);

  useEffect(() => {
    let active = true;
    devLoginTersedia().then((tersedia) => {
      if (active) setDevLogin(tersedia);
    });
    fetch(`${apiURL}/auth/me`, { credentials: "include", cache: "no-store" })
      .then((response) => (response.ok ? response.json() : null))
      .then((data) => {
        if (active) setPrincipal(data);
      })
      .catch(() => {
        if (active) setPrincipal(null);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function logout() {
    // akhiriSesi mengakhiri sesi lokal dan SSO, lalu mengembalikan URL logout
    // penyedia identitas bila sesi SSO masih aktif.
    const tujuan = await akhiriSesi();
    if (tujuan) {
      window.location.assign(tujuan);
      return;
    }
    setPrincipal(null);
  }

  if (loading) {
    return <p className="auth-note">Memeriksa sesi SSO…</p>;
  }

  if (principal) {
    return (
      <div className="auth-panel">
        <div>
          <strong>{principal.name || principal.email}</strong>
          <span>{principal.role} · realm {principal.realm}</span>
        </div>
        <button type="button" onClick={logout}>Keluar</button>
      </div>
    );
  }

  return (
    <div className="auth-panel">
      <div>
        <strong>Masuk melalui SSO Jogjaprov</strong>
        <span>Gunakan JSP untuk masyarakat atau Satriya untuk ASN.</span>
      </div>
      <div className="auth-actions">
        <Link href={masukHref("/", "jsp")}>Masuk dengan JSP</Link>
        <Link href={masukHref("/", "satriya")}>Masuk dengan Satriya</Link>
        {devLogin && <Link href="/dashboard/masuk">Masuk lokal (dev)</Link>}
      </div>
    </div>
  );
}
