"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

type Principal = {
  name?: string;
  email: string;
  role: string;
  realm?: string;
};

const NAV_LINKS = [
  { label: "Beranda", href: "/" },
  { label: "Repository Data Lokal", href: "/repository-data" },
  { label: "Registrasi Hortikultura", href: "/horticulture" },
  { label: "Sertifikasi Benih BP3MBTP", href: "/certification" },
  { label: "Registrasi & Penjualan Benih", href: "/benih" },
] as const;

const FOOTER_LINKS = [
  { label: "Tentang", href: "/tentang" },
  { label: "Kebijakan Privasi", href: "/kebijakan-privasi" },
  { label: "Bantuan", href: "/bantuan" },
] as const;

export function Layout({ children }: { children: React.ReactNode }) {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    if ("serviceWorker" in navigator && window.isSecureContext) {
      void navigator.serviceWorker.register("/sw.js", { scope: "/" }).catch(() => {
        // PWA enhancement is optional; core portal remains usable if registration fails.
      });
    }
    fetch(`${apiURL}/auth/me`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!active) return;
        setPrincipal(data ?? null);
      })
      .catch(() => {
        if (!active) return;
        setPrincipal(null);
      })
      .finally(() => {
        if (!active) return;
        setLoading(false);
      });
    return () => { active = false; };
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col">
        <header className="sitani-header">
          <div className="container">
            <div className="header-inner">
              <Link href="/" className="brand">
                <span className="brand-mark">🌾</span>
                <span className="brand-name">SiTani DIY</span>
              </Link>
              <div className="auth-loading">Memeriksa sesi…</div>
            </div>
          </div>
        </header>
        <div className="container page-body"><p>Memuat…</p></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sitani-header">
        <div className="container">
          <div className="header-inner">
            <Link href="/" className="brand">
              <span className="brand-mark">🌾</span>
              <span className="brand-name">SiTani DIY</span>
            </Link>
            <nav className="nav-list">
              {NAV_LINKS.map((link) => (
                <Link key={link.label} href={link.href} className="nav-link">
                  {link.label}
                </Link>
              ))}
            </nav>
            <div className="auth-panel">
              {principal ? (
                <div className="user-info">
                  <span>{principal.name || principal.email}</span>
                  <span className="realm-badge">{principal.role} · {principal.realm ?? "-"}</span>
                  <button className="btn btn-outline btn-sm" onClick={handleLogout}>Keluar</button>
                </div>
              ) : (
                <div className="auth-actions">
                  <a href={`${apiURL}/auth/login/jsp?redirect=/`} className="btn btn-outline btn-sm">Masuk JSP</a>
                  <a href={`${apiURL}/auth/login/satriya?redirect=/`} className="btn btn-outline btn-sm">Masuk Satriya</a>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      <div className="page-body">{children}</div>

      <footer className="sitani-footer">
        <div className="container">
          <p>Portal Pertanian dan Ketahanan Pangan DIY &mdash; Dinas Pertanian dan Ketahanan Pangan Pemda DIY.</p>
          <p className="footer-note">Sistem pengembangan &bull; SSO JSP/Satriya &bull; SPL / Cakra API</p>
          <nav className="footer-links" aria-label="Informasi portal">
            {FOOTER_LINKS.map((link) => (
              <Link key={link.href} href={link.href}>{link.label}</Link>
            ))}
          </nav>
        </div>
      </footer>
    </div>
  );

  async function handleLogout() {
    await fetch(`${apiURL}/auth/logout`, { method: "POST", credentials: "include" });
    setPrincipal(null);
  }
}
