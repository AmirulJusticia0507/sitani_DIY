"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { masukHref } from "./auth/masuk-url";
import { akhiriSesi } from "./dashboard/principal";

const apiURL =
  process.env.NEXT_PUBLIC_API_BASE ||
  process.env.NEXT_PUBLIC_BACKEND_URL ||
  "http://localhost:8080";

type Principal = {
  name?: string;
  email: string;
  role: string;
  realm?: string;
};

// Navigasi portal publik hanya memuat halaman informasi. Modul layanan
// (repository data, hortikultura, sertifikasi, benih) berada di area kerja
// /dashboard dan dibuka setelah masuk melalui SSO, sehingga sidebar publik
// tidak lagi menautkannya langsung.
type NavIcon = "home" | "layout" | "chart" | "leaf" | "seed" | "grain";

const NAV_LINKS: ReadonlyArray<{ label: string; short: string; href: string; icon: NavIcon }> = [
  { label: "Beranda", short: "Beranda", href: "/", icon: "home" },
  { label: "Masuk area kerja", short: "Masuk", href: "/masuk", icon: "layout" },
];

const FOOTER_LINKS = [
  { label: "Tentang", href: "/tentang" },
  { label: "Kebijakan Privasi", href: "/kebijakan-privasi" },
  { label: "Bantuan", href: "/bantuan" },
] as const;

const COLLAPSE_KEY = "sitani:sidebar-collapsed";

// Ikon garis sederhana agar sidebar collapsed tetap terbaca tanpa teks.
function NavIcon({ name }: { name: NavIcon }) {
  const paths: Record<NavIcon, React.ReactNode> = {
    home: <><path d="M4 11.5 12 4l8 7.5" /><path d="M6 10v9h12v-9" /></>,
    layout: <><path d="M4 5h16v14H4z" /><path d="M9 5v14" /></>,
    chart: <><path d="M5 19v-6M10 19V7M15 19v-8M20 19v-4" /><path d="M3 21h18" /></>,
    leaf: <><path d="M20 4c-9 .5-14 5-13.5 12C14 16.5 19 12 20 4Z" /><path d="M5 19c3-4 6-7 11-11" /></>,
    seed: <><path d="M12 20v-9" /><path d="M12 13c-4 0-6.5-2.5-6.5-6 4 0 6.5 2.5 6.5 6Z" /><path d="M12 16c4 0 6.5-2.5 6.5-6-4 0-6.5 2.5-6.5 6Z" /></>,
    grain: <><path d="M12 20V5" /><path d="M12 9C9 9 7.5 7.5 7.5 5 10.5 5 12 6.5 12 9Z" /><path d="M12 13c3 0 4.5-1.5 4.5-4-3 0-4.5 1.5-4.5 4Z" /><path d="M12 17c-3 0-4.5-1.5-4.5-4 3 0 4.5 1.5 4.5 4Z" /></>,
  };
  return (
    <svg className="sidebar-icon" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      {paths[name]}
    </svg>
  );
}

function aktifUntuk(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(`${href}/`);
}

// Kerangka portal publik: sidebar kiri yang dapat di-collapse ke rail ikon,
// bukan header horizontal. Pada layar sempit sidebar menjadi drawer dengan
// overlay; state collapsed disimpan di localStorage agar bertahan antarhalaman.
export function Layout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [loading, setLoading] = useState(true);
  const [rinci, setRinci] = useState(false);
  const [terbuka, setTerbuka] = useState(false);
  const tombolRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if ("serviceWorker" in navigator && window.isSecureContext) {
      void navigator.serviceWorker.register("/sw.js", { scope: "/" }).catch(() => {
        // PWA enhancement is optional; core portal remains usable if registration fails.
      });
    }
    fetch(`${apiURL}/auth/me`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        setPrincipal(data ?? null);
      })
      .catch(() => {
        setPrincipal(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  // Preferensi collapsed hanya dibaca di klien; markup awal memakai sidebar
  // penuh agar tidak ada pergeseran yang membingungkan pembaca layar.
  useEffect(() => {
    try {
      if (window.localStorage.getItem(COLLAPSE_KEY) === "1") setRinci(true);
    } catch {
      // localStorage dapat diblokir (mode privat ketat); default tetap penuh.
    }
  }, []);

  useEffect(() => {
    setTerbuka(false);
  }, [pathname]);

  useEffect(() => {
    if (!terbuka) return;
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setTerbuka(false);
        tombolRef.current?.focus();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [terbuka]);

  function toggleRinci() {
    const berikutnya = !rinci;
    setRinci(berikutnya);
    try {
      window.localStorage.setItem(COLLAPSE_KEY, berikutnya ? "1" : "0");
    } catch {
      // Abaikan kegagalan penyimpanan preferensi.
    }
  }

  async function handleLogout() {
    // Sesi lokal dan SSO diakhiri sekaligus. Bila backend memberi URL logout
    // penyedia identitas, pengguna dibawa ke sana agar sesi SSO ikut berakhir.
    const tujuan = await akhiriSesi();
    if (tujuan) {
      window.location.assign(tujuan);
      return;
    }
    setPrincipal(null);
  }

  return (
    <div className={`portal-root${rinci ? " is-rinci" : ""}${terbuka ? " is-terbuka" : ""}`}>
      <a href="#isi-portal" className="skip-link">Lewati ke isi portal</a>

      <button
        ref={tombolRef}
        type="button"
        className="sidebar-buka"
        aria-expanded={terbuka}
        aria-controls="sidebar-portal"
        onClick={() => setTerbuka((v) => !v)}
      >
        <span aria-hidden="true">☰</span>
        <span className="sidebar-buka-teks">Menu</span>
      </button>

      <div
        className="sidebar-latar"
        aria-hidden="true"
        onClick={() => setTerbuka(false)}
      />

      <aside id="sidebar-portal" className="sidebar" aria-label="Navigasi portal">
        <div className="sidebar-atas">
          <Link href="/" className="sidebar-brand" aria-label="SiTani DIY — beranda portal">
            <span className="sidebar-mark" aria-hidden="true">🌾</span>
            <span className="sidebar-teks sidebar-nama">SiTani DIY</span>
          </Link>
          <button
            type="button"
            className="sidebar-tutup"
            aria-label="Tutup menu"
            onClick={() => setTerbuka(false)}
          >
            ×
          </button>
        </div>

        <button
          type="button"
          className="sidebar-collapse"
          aria-expanded={!rinci}
          aria-controls="sidebar-portal"
          aria-label={rinci ? "Perluas sidebar" : "Ciutkan sidebar"}
          title={rinci ? "Perluas sidebar" : "Ciutkan sidebar"}
          onClick={toggleRinci}
        >
          <span className="sidebar-collapse-ikon" aria-hidden="true">{rinci ? "»" : "«"}</span>
          <span className="sidebar-teks">Ciutkan</span>
        </button>

        <nav className="sidebar-nav" aria-label="Navigasi utama">
          <ul>
            {NAV_LINKS.map((link) => {
              const aktif = aktifUntuk(pathname, link.href);
              return (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className={`sidebar-link${aktif ? " is-active" : ""}`}
                    aria-current={aktif ? "page" : undefined}
                    // Nama aksesibel tetap label lengkap meski di rail/drawer
                    // layar sempit yang tampil hanya label singkat.
                    aria-label={link.label}
                    onClick={() => setTerbuka(false)}
                  >
                    <NavIcon name={link.icon} />
                    <span className="sidebar-teks">
                      <span className="sidebar-label">{link.label}</span>
                      <span className="sidebar-label-singkat">{link.short}</span>
                    </span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="sidebar-bawah">
          {loading ? (
            <p className="sidebar-teks sidebar-catatan">Memeriksa sesi…</p>
          ) : principal ? (
            <div className="sidebar-akun">
              <span className="sidebar-avatar" aria-hidden="true">
                {(principal.name || principal.email).slice(0, 1).toUpperCase()}
              </span>
              <span className="sidebar-teks sidebar-akun-teks">
                <span className="sidebar-akun-nama">{principal.name || principal.email}</span>
                <span className="sidebar-akun-peran">{principal.role} · {principal.realm ?? "-"}</span>
              </span>
              <button type="button" className="sidebar-keluar" onClick={handleLogout}>Keluar</button>
            </div>
          ) : (
            <div className="sidebar-masuk">
              <Link href={masukHref("/dashboard", "jsp")} className="sidebar-link masuk-jsp" title="Masuk JSP">
                <span className="sidebar-icon sidebar-icon-teks" aria-hidden="true">JSP</span>
                <span className="sidebar-teks">Masuk JSP</span>
              </Link>
              <Link href={masukHref("/dashboard", "satriya")} className="sidebar-link" title="Masuk Satriya">
                <span className="sidebar-icon sidebar-icon-teks" aria-hidden="true">SR</span>
                <span className="sidebar-teks">Masuk Satriya</span>
              </Link>
            </div>
          )}
        </div>
      </aside>

      <div className="portal-isi">
        <main id="isi-portal" className="page-body" tabIndex={-1}>{children}</main>

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
    </div>
  );
}
