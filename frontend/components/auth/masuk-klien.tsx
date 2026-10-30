"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import { apiURL, devLoginTersedia } from "../dashboard/principal";
import {
  BANTUAN_SSO_URL,
  labelRealm,
  masukHref,
  pathAman,
  ssoHref,
  type RealmSSO,
} from "./masuk-url";

// Status halaman. Tidak ada fase "memeriksa" yang memblokir tombol: tombol
// selalu siap ditekan, dan hasil pemeriksaan backend hanya menambah keterangan.
type StatusLogin =
  | { phase: "siap" }
  | { phase: "mengalihkan"; realm: RealmSSO }
  | { phase: "gagal"; realm: RealmSSO; pesan: string };

type Keterangan = {
  // devLogin: login lokal dev tersedia saat pemeriksaan selesai.
  devLogin: boolean;
  // ssoBelumSiap: /auth/me menjawab 503 (konfigurasi SSO kosong).
  ssoBelumSiap: boolean;
};

// Kartu login bertema SSO Pemda DIY: satu kartu putih di tengah, dua pilihan
// akun (JSP untuk masyarakat, Satriya untuk ASN), lalu satu tombol masuk.
//
// Tombol memakai navigasi penuh (window.location.assign) ke endpoint backend.
// Browser mengikuti 302 ke issuer SSO secara native — termasuk melintasi origin
// — sehingga tidak terkena CORS seperti fetch. Sebelum berpindah, endpoint
// diperiksa tanpa mengikuti redirect agar balasan 4xx/5xx berpesan JSON bisa
// ditampilkan di kartu; galat pada langkah itu tidak pernah menghalangi.
export function MasukKlien({ redirect, realmAwal }: { redirect?: string; realmAwal?: RealmSSO }) {
  const tujuan = pathAman(redirect);
  const [realm, setRealm] = useState<RealmSSO>(realmAwal ?? "jsp");
  const [status, setStatus] = useState<StatusLogin>({ phase: "siap" });
  const [keterangan, setKeterangan] = useState<Keterangan>({
    devLogin: false,
    ssoBelumSiap: false,
  });

  useEffect(() => {
    let aktif = true;
    Promise.all([ssoBelumSiapEndpoint(), devLoginTersedia()]).then(([belumSiap, devLogin]) => {
      if (!aktif) return;
      setKeterangan({ devLogin, ssoBelumSiap: belumSiap });
    });
    return () => {
      aktif = false;
    };
  }, []);

  const lanjut = useCallback(
    async (pilihan: RealmSSO) => {
      setRealm(pilihan);
      const target = ssoHref(apiURL, pilihan, tujuan);
      setStatus({ phase: "mengalihkan", realm: pilihan });

      const pesanGalat = await pesanGalatBackend(target);
      if (pesanGalat) {
        setStatus({ phase: "gagal", realm: pilihan, pesan: pesanGalat });
        return;
      }
      // Navigasi penuh: browser mengikuti 302 backend ke issuer SSO sendiri,
      // tanpa CORS dan tanpa menyentuh token di sisi klien.
      window.location.assign(target);
    },
    [tujuan],
  );

  const sibuk = status.phase === "mengalihkan";
  const ssoBelumSiap = keterangan.ssoBelumSiap;
  const devLogin = keterangan.devLogin;

  return (
    <section className="sso-card" aria-labelledby="masuk-judul">
      <div className="sso-card-head">
        <span className="sso-logo" aria-hidden="true">🌾</span>
        <h1 id="masuk-judul">Single Sign On</h1>
        <p>Selamat Datang di SiTani DIY</p>
      </div>

      <p className="sso-tips-judul">Masuk ke akun Anda</p>
      <p className="sso-tips">
        Tips Nyaman Karena Aman
        <span>Jaga kerahasiaan akun Anda dan tidak membagikan password kepada orang lain.</span>
        <span>Akses portal hanya melalui alamat resmi DPKP DIY.</span>
      </p>

      <fieldset className="sso-realm" disabled={sibuk}>
        <legend>Pilih jenis akun</legend>
        <label className={`sso-realm-opsi${realm === "jsp" ? " is-terpilih" : ""}`}>
          <input
            type="radio"
            name="realm"
            value="jsp"
            checked={realm === "jsp"}
            onChange={() => setRealm("jsp")}
          />
          <span>
            <strong>JSP</strong>
            <small>Masyarakat / pelaku usaha</small>
          </span>
        </label>
        <label className={`sso-realm-opsi${realm === "satriya" ? " is-terpilih" : ""}`}>
          <input
            type="radio"
            name="realm"
            value="satriya"
            checked={realm === "satriya"}
            onChange={() => setRealm("satriya")}
          />
          <span>
            <strong>Satriya</strong>
            <small>ASN Pemda DIY</small>
          </span>
        </label>
      </fieldset>

      {status.phase === "gagal" && (
        <p className="sso-pesan sso-pesan-galat" role="alert">
          <strong>Login {labelRealm(status.realm)} belum dapat dilanjutkan.</strong> {status.pesan}
        </p>
      )}

      {ssoBelumSiap && (
        <p className="sso-pesan sso-pesan-info" role="status">
          SSO Jogjaprov belum aktif di environment ini. Pengalihan tetap dicoba; bila gagal, gunakan
          kanal bantuan resmi.
        </p>
      )}

      <button
        type="button"
        className="sso-tombol"
        onClick={() => lanjut(realm)}
        disabled={sibuk}
      >
        {status.phase === "mengalihkan"
          ? `Mengalihkan ke ${labelRealm(status.realm)}…`
          : `Masuk dengan ${labelRealm(realm)}`}
      </button>

      <nav className="sso-tautan" aria-label="Bantuan masuk">
        <a href={BANTUAN_SSO_URL} target="_blank" rel="noreferrer">
          Panduan SSO/2FA
        </a>
        <Link href="/bantuan">Bantuan SiTani</Link>
        <Link href="/kebijakan-privasi">Kebijakan Privasi</Link>
      </nav>

      {devLogin && (
        <div className="sso-dev">
          <span className="badge badge-warning">Lingkungan development</span>
          <p>
            Backend ini mengaktifkan login lokal untuk UAT.{" "}
            <Link href={masukHref(tujuan === "/" ? "/dashboard" : tujuan)}>
              formulir email &amp; password dev
            </Link>
            .
          </p>
        </div>
      )}

      <p className="sso-kembali">
        <Link href={tujuan === "/" ? "/" : tujuan}>&larr; Kembali</Link>
      </p>
    </section>
  );
}

// Endpoint session (/auth/me) menjawab 503 selama konfigurasi SSO kosong.
// Pemeriksaan ini hanya untuk memberi keterangan di antarmuka; keputusan
// autentikasi tetap sepenuhnya di backend.
async function ssoBelumSiapEndpoint(): Promise<boolean> {
  try {
    const res = await fetch(`${apiURL}/auth/me`, { credentials: "include", cache: "no-store" });
    return res.status === 503;
  } catch {
    return true;
  }
}

// pesanGalatBackend memeriksa endpoint login tanpa mengikuti pengalihan dan
// mengembalikan pesan galat backend bila ada. Alur normal (302 ke issuer SSO,
// atau balasan yang tak terbaca karena CORS) menghasilkan null: navigasi penuh
// yang akan menanganinya.
async function pesanGalatBackend(target: string): Promise<string | null> {
  try {
    const res = await fetch(target, {
      method: "GET",
      credentials: "include",
      redirect: "manual",
      cache: "no-store",
      headers: { Accept: "application/json" },
    });
    if (res.type === "opaqueredirect" || res.status === 0 || res.ok) return null;
    const data = (await res.json().catch(() => null)) as { error?: unknown } | null;
    return typeof data?.error === "string" && data.error !== ""
      ? data.error
      : `Server API menjawab ${res.status}.`;
  } catch {
    return null;
  }
}
