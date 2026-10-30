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

type StatusLogin =
  | { phase: "memeriksa" }
  | { phase: "siap"; devLogin: boolean; ssoBelumSiap: boolean }
  | { phase: "mengalihkan"; realm: RealmSSO }
  | { phase: "gagal"; realm: RealmSSO; pesan: string };

// Kartu login bertema SSO Pemda DIY: satu kartu putih di tengah, dua pilihan
// akun (JSP untuk masyarakat, Satriya untuk ASN), lalu satu tombol masuk.
// Backend baru dipanggil saat tombol ditekan sehingga klik selalu memberi
// umpan balik — termasuk ketika backend mati atau SSO belum dikonfigurasi.
export function MasukKlien({ redirect, realmAwal }: { redirect?: string; realmAwal?: RealmSSO }) {
  const tujuan = pathAman(redirect);
  const [realm, setRealm] = useState<RealmSSO>(realmAwal ?? "jsp");
  const [status, setStatus] = useState<StatusLogin>({ phase: "memeriksa" });

  useEffect(() => {
    let aktif = true;
    ssoBelumSiapEndpoint().then((belumSiap) => {
      if (!aktif) return;
      devLoginTersedia().then((devLogin) => {
        if (!aktif) return;
        setStatus({ phase: "siap", devLogin, ssoBelumSiap: belumSiap });
      });
    });
    return () => {
      aktif = false;
    };
  }, []);

  const lanjut = useCallback(
    async (pilihan: RealmSSO) => {
      setRealm(pilihan);
      setStatus({ phase: "mengalihkan", realm: pilihan });
      const target = ssoHref(apiURL, pilihan, tujuan);
      try {
        const cek = await fetch(target, {
          method: "GET",
          credentials: "include",
          redirect: "manual",
          cache: "no-store",
          headers: { Accept: "application/json" },
        });
        // Alur backend yang benar selalu mengalihkan ke issuer SSO: fetch manual
        // melaporkan status 0/opaqueredirect. Demikian pula Respons yang memakai
        // "inexact-relative" (redirect tak terekspos). Keduanya diteruskan.
        if (cek.type === "opaqueredirect" || cek.status === 0 || "inexact" in cek) {
          window.location.assign(target);
          return;
        }
        // Status json 4xx/5xx yang memuat pesan error backend ditampilkan apa
        // adanya (mis. realm belum dikonfigurasi, secret salah).
        const data = (await cek.json().catch(() => null)) as { error?: string } | null;
        if (data?.error) {
          setStatus({
            phase: "gagal",
            realm: pilihan,
            pesan: data.error,
          });
          return;
        }
        // Respons tanpa pesan error berarti fetch tidak melihat redirect (mis.
        // mode "manual" tidak menghasilkan opaqueredirect). Jangan biarkan
        // tombol diam: navigasi penuh ke endpoint backend agar server yang
        // memutuskan pengalihan ke issuer SSO.
        window.location.assign(target);
      } catch {
        // Backend tidak dapat dihubungi: beri pesan yang dapat ditindaklanjuti
        // alih-alih membiarkan tombol berputar.
        setStatus({
          phase: "gagal",
          realm: pilihan,
          pesan:
            "Server API SiTani tidak dapat dihubungi. Pastikan backend berjalan dan alamat API sudah benar.",
        });
      }
    },
    [tujuan],
  );

  const sibuk = status.phase === "memeriksa" || status.phase === "mengalihkan";
  const ssoBelumSiap = status.phase === "siap" ? status.ssoBelumSiap : false;
  const devLogin = status.phase === "siap" && status.devLogin;

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

      {status.phase !== "memeriksa" && ssoBelumSiap && (
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
          : status.phase === "memeriksa"
            ? "Menyiapkan…"
            : `Masuk dengan ${labelRealm(realm)}`}
      </button>

      <nav className="sso-tautan" aria-label="Bantuan masuk">
        <a href={BANTUAN_SSO_URL} target="_blank" rel="noreferrer">
          Panduan SSO/2FA
        </a>
        <Link href="/bantuan">Bantuan SiTani</Link>
        <Link href="/kebijakan-privasi">Kebijakan Privasi</Link>
      </nav>

      {status.phase !== "memeriksa" && devLogin && (
        <div className="sso-dev">
          <span className="badge badge-warning">Lingkungan development</span>
          <p>
            Backend ini mengaktifkan login lokal untuk UAT.{" "}
            <Link href={masukHref(tujuan === "/" ? "/dashboard" : tujuan)}>Gunakan login lokal</Link>{" "}
            atau buka{" "}
            <Link href={`/dashboard/masuk?redirect=${encodeURIComponent(tujuan === "/" ? "/dashboard" : tujuan)}`}>
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
