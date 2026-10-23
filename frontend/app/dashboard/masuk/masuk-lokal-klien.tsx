"use client";

import { useEffect, useState } from "react";

import { apiURL, devLoginTersedia } from "../../../components/dashboard/principal";

// Halaman login lokal untuk development/UAT. Formulir hanya ditampilkan bila
// backend mengaktifkan login lokal; di lingkungan lain halaman menjelaskan
// bahwa autentikasi memakai SSO.
export function MasukLokalKlien() {
  const [status, setStatus] = useState<"memeriksa" | "tersedia" | "tidak">("memeriksa");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [pesan, setPesan] = useState<string | null>(null);
  const [mengirim, setMengirim] = useState(false);

  useEffect(() => {
    let aktif = true;
    devLoginTersedia().then((tersedia) => {
      if (aktif) setStatus(tersedia ? "tersedia" : "tidak");
    });
    return () => {
      aktif = false;
    };
  }, []);

  async function kirim(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (mengirim) return;
    setMengirim(true);
    setPesan(null);
    try {
      const res = await fetch(`${apiURL}/auth/dev/login`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (res.ok) {
        window.location.assign("/dashboard");
        return;
      }
      const data = (await res.json().catch(() => null)) as { error?: string } | null;
      setPesan(data?.error ?? "Login gagal. Periksa email dan password.");
    } catch {
      setPesan("Backend tidak dapat dihubungi. Pastikan server API berjalan.");
    } finally {
      setMengirim(false);
    }
  }

  if (status === "memeriksa") {
    return <p className="dashboard-note">Memeriksa ketersediaan login…</p>;
  }

  if (status === "tidak") {
    return (
      <section className="dashboard-panel" aria-labelledby="sso-judul">
        <h2 id="sso-judul">Login lokal tidak tersedia</h2>
        <p>
          Autentikasi portal memakai SSO Jogjaprov: realm JSP untuk masyarakat dan Satriya
          untuk ASN. Login lokal hanya diaktifkan untuk development/UAT oleh administrator
          sistem.
        </p>
        <div className="dashboard-actions">
          <a href={`${apiURL}/auth/login/jsp?redirect=/dashboard`} className="btn btn-primary">
            Masuk JSP
          </a>
          <a href={`${apiURL}/auth/login/satriya?redirect=/dashboard`} className="btn btn-outline">
            Masuk Satriya
          </a>
        </div>
      </section>
    );
  }

  return (
    <section className="dashboard-panel masuk-lokal" aria-labelledby="masuk-lokal-judul">
      <p className="badge badge-warning">Lingkungan development</p>
      <h2 id="masuk-lokal-judul">Masuk dengan akun lokal</h2>
      <p>
        Formulir ini hanya aktif bila backend menjalankan login lokal development. Akun dibuat
        lewat perintah <code>go run ./cmd/seed-dev</code>; password tidak pernah disimpan di
        repositori.
      </p>
      <form onSubmit={kirim} className="masuk-lokal-form" noValidate>
        <label htmlFor="email">Email</label>
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="username"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <label htmlFor="password">Password</label>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        {pesan && (
          <p className="masuk-lokal-pesan" role="alert">
            {pesan}
          </p>
        )}
        <button type="submit" className="btn btn-primary" disabled={mengirim}>
          {mengirim ? "Memproses…" : "Masuk"}
        </button>
      </form>
    </section>
  );
}
