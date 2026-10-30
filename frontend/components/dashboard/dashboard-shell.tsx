"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";

import { masukHref } from "../auth/masuk-url";
import { Brand } from "../brand";
import { akhiriSesi, apiURL, devLoginTersedia, getRealmLabel, getRoleLabel, type DashboardPrincipal } from "./principal";
import { getDashboardNav, getRoleAccess, isRole } from "./role-access";
type Status = "memuat" | "masuk" | "tamu";

// Kerangka dashboard: header area kerja, sidebar modul sesuai peran, dan footer
// ringkas yang terpisah dari layout portal publik. Pada layar kecil sidebar
// ditutup setelah navigasi agar konten tidak tertutup.
export function DashboardShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [principal, setPrincipal] = useState<DashboardPrincipal | null>(null);
  const [status, setStatus] = useState<Status>("memuat");
  const [sidebarTerbuka, setSidebarTerbuka] = useState(false);
  const [devLogin, setDevLogin] = useState(false);

  useEffect(() => {
    let active = true;
    // Login lokal diperiksa lebih dulu dan dipakai sebagai fallback saat SSO
    // belum tersedia: /auth/me menjawab 503 tanpa sesi, sehingga tanpa status
    // ini pengguna dev akan diberi tautan SSO yang pasti gagal.
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
          setPrincipal(null);
          setStatus("tamu");
        }
      })
      .catch(() => {
        if (!active) return;
        setPrincipal(null);
        setStatus("tamu");
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    setSidebarTerbuka(false);
  }, [pathname]);

  async function handleKeluar() {
    // Sesi lokal dan SSO memakai cookie berbeda; akhiriSesi mengakhiri keduanya
    // lalu mengembalikan URL logout penyedia identitas bila ada.
    const tujuan = await akhiriSesi();
    if (tujuan) {
      window.location.assign(tujuan);
      return;
    }
    window.location.assign("/dashboard");
  }

  const nav = getDashboardNav(principal?.role);
  const roleDikenal = principal ? isRole(principal.role) : false;
  const modulAktif = getModulAktif(pathname, principal?.role);

  return (
    <div className="dashboard-root">
      <a href="#isi-dashboard" className="skip-link">Lewati ke isi dashboard</a>

      <header className="dashboard-header">
        <div className="dashboard-header-inner">
          <button
            type="button"
            className="dashboard-toggle"
            aria-label="Buka menu dashboard"
            aria-expanded={sidebarTerbuka}
            aria-controls="sidebar-dashboard"
            onClick={() => setSidebarTerbuka((v) => !v)}
          >
            <span aria-hidden="true">☰</span>
            <span className="dashboard-toggle-text">Menu</span>
          </button>

          <Brand href="/dashboard" variant="dark" />

          <div className="dashboard-header-actions">
            <Link href="/" className="dashboard-link-publik">Portal publik</Link>
            {status === "memuat" && <span className="dashboard-header-note">Memeriksa sesi…</span>}
            {status === "masuk" && principal && (
              <>
                <span className="dashboard-user" title={principal.email}>
                  <span className="dashboard-user-name">{principal.name || principal.email}</span>
                  <span className="dashboard-user-role">{getRoleLabel(principal)}</span>
                </span>
                <button type="button" className="btn btn-outline btn-sm" onClick={handleKeluar}>
                  Keluar
                </button>
              </>
            )}
            {status === "tamu" && (
              <div className="auth-actions">
                {devLogin && (
                  <Link href="/dashboard/masuk" className="btn btn-outline btn-sm">
                    Masuk lokal
                  </Link>
                )}
                <a href={masukHref("/dashboard", "jsp")} className="btn btn-outline btn-sm">
                  Masuk JSP
                </a>
                <a href={masukHref("/dashboard", "satriya")} className="btn btn-outline btn-sm">
                  Masuk Satriya
                </a>
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="dashboard-body">
        <aside
          id="sidebar-dashboard"
          className={`dashboard-sidebar${sidebarTerbuka ? " is-open" : ""}`}
          aria-label="Navigasi dashboard"
        >
          {status === "memuat" && <p className="dashboard-sidebar-note">Memuat menu…</p>}

          {status === "masuk" && principal && !roleDikenal && (
            <p className="dashboard-sidebar-note dashboard-warning">
              Peran <strong>{principal.role}</strong> belum dipetakan ke modul dashboard. Hubungi
              administrator untuk penyesuaian akses.
            </p>
          )}

          {status !== "memuat" && (
            <nav className="dashboard-nav" aria-label="Modul dashboard">
              <ul>
                {nav.map((item) => {
                  const aktif = modulAktif === item.key || (item.key === "ringkasan" && pathname === "/dashboard");
                  return (
                    <li key={item.key}>
                      <Link
                        href={item.href}
                        className={`dashboard-nav-link${aktif ? " is-active" : ""}`}
                        aria-current={aktif ? "page" : undefined}
                        title={item.deskripsi}
                      >
                        {item.label}
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </nav>
          )}

          {status === "tamu" && (
            <p className="dashboard-sidebar-note">
              Masuk untuk melihat modul kerja sesuai peran Anda. Sebagian modul tetap dapat dibuka
              dari portal publik.
            </p>
          )}
        </aside>

        <main id="isi-dashboard" className="dashboard-main" tabIndex={-1}>
          {children}
        </main>
      </div>

      <footer className="dashboard-footer">
        <div className="dashboard-footer-inner">
          <p>
            Area kerja SiTani DIY — Dinas Pertanian dan Ketahanan Pangan DIY.
            {principal ? ` Sesi: ${getRealmLabel(principal)}.` : ""}
          </p>
          <nav className="dashboard-footer-links" aria-label="Informasi portal">
            <Link href="/tentang">Tentang</Link>
            <Link href="/kebijakan-privasi">Kebijakan Privasi</Link>
            <Link href="/bantuan">Bantuan</Link>
            <Link href="/dashboard/profil">Profil &amp; akses</Link>
          </nav>
        </div>
      </footer>
    </div>
  );
}

// getModulAktif menyelaraskan path yang dibuka (termasuk modul portal yang
// ditautkan dari sidebar) dengan kunci modul untuk penanda navigasi aktif.
function getModulAktif(pathname: string, role?: string): string | null {
  if (pathname.startsWith("/dashboard/profil")) return "profil";
  if (pathname.startsWith("/dashboard")) return "ringkasan";
  if (pathname.startsWith("/repository-data")) return "repository-data";
  if (pathname.startsWith("/horticulture")) return "hortikultura";
  if (pathname.startsWith("/benih")) {
    // Modul benih dipakai bersama oleh registrasi, penjualan, dan laporan;
    // pilih kunci pertama yang tersedia untuk peran tersebut.
    const akses = getRoleAccess(role);
    if (akses.has("registrasi-benih")) return "registrasi-benih";
    if (akses.has("penjualan-benih")) return "penjualan-benih";
    return "laporan";
  }
  if (pathname.startsWith("/certification")) {
    const akses = getRoleAccess(role);
    if (akses.has("sertifikasi")) return "sertifikasi";
    if (akses.has("laboratorium")) return "laboratorium";
    return "penerbitan";
  }
  if (pathname.startsWith("/dashboard/dokumen")) return "dokumen";
  if (pathname.startsWith("/v1/dokumen")) return "dokumen";
  return null;
}
