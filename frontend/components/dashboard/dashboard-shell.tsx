"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";

import { masukHref } from "../auth/masuk-url";
import { Brand } from "../brand";
import { akhiriSesi, apiURL, devLoginTersedia, getRealmLabel, getRoleLabel, type DashboardPrincipal } from "./principal";
import { getDashboardNav, getRoleAccess, isRole } from "./role-access";
type Status = "memuat" | "masuk" | "tamu";

// Kunci penyimpanan preferensi tampilan sidebar di layar lebar. Hanya berisi
// "terbuka"/"tertutup"; tidak ada data sesi atau pribadi.
export const KUNCI_SIDEBAR = "sitani.dashboard.sidebar";

// Batas lebar: di atas terlipat rail ikon berlaku, di bawahnya panel lipat.
const SEMPIT_SIDEBAR = "(max-width: 900px)";
const LANGSING_SIDEBAR = "(min-width: 901px)";

// Ikon nav digambar sebagai SVG inline (stroke currentColor) agar tidak
// bergantung pada font ikon atau emoji yang tampil berbeda antarsistem.
const IKON_NAV: Record<string, React.ReactNode> = {
  ringkasan: (
    <>
      <rect x="3" y="3" width="7" height="7" rx="1.5" />
      <rect x="14" y="3" width="7" height="7" rx="1.5" />
      <rect x="3" y="14" width="7" height="7" rx="1.5" />
      <rect x="14" y="14" width="7" height="7" rx="1.5" />
    </>
  ),
  "repository-data": (
    <>
      <ellipse cx="12" cy="5" rx="8" ry="3" />
      <path d="M4 5v6c0 1.7 3.6 3 8 3s8-1.3 8-3V5" />
      <path d="M4 11v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6" />
    </>
  ),
  hortikultura: (
    <>
      <path d="M12 21V11" />
      <path d="M12 11c0-4 3-7 7-7 0 4-3 7-7 7Z" />
      <path d="M12 14c0-3-2.5-5.5-5.5-5.5 0 3 2.5 5.5 5.5 5.5Z" />
    </>
  ),
  sertifikasi: (
    <>
      <circle cx="12" cy="9" r="5" />
      <path d="m8.5 13.5-1.5 8 5-2.5 5 2.5-1.5-8" />
    </>
  ),
  laboratorium: (
    <>
      <path d="M9 3h6" />
      <path d="M10 3v6.5L5.5 17a2 2 0 0 0 1.7 3h9.6a2 2 0 0 0 1.7-3L14 9.5V3" />
      <path d="M8.5 14h7" />
    </>
  ),
  penerbitan: (
    <>
      <path d="M6 3h9l4 4v14H6z" />
      <path d="M14 3v5h5" />
      <path d="M9 13h6M9 17h4" />
    </>
  ),
  "registrasi-benih": (
    <>
      <path d="M3 7h18M3 12h18M3 17h18" />
      <circle cx="8" cy="7" r="1.6" />
      <circle cx="15" cy="12" r="1.6" />
      <circle cx="10" cy="17" r="1.6" />
    </>
  ),
  "penjualan-benih": (
    <>
      <path d="M3 4h2.5l2.2 10.5h9.6L20 7H6" />
      <circle cx="9" cy="19" r="1.6" />
      <circle cx="17" cy="19" r="1.6" />
    </>
  ),
  laporan: (
    <>
      <path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
    </>
  ),
  dokumen: (
    <>
      <path d="M4 6a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2Z" />
    </>
  ),
  profil: (
    <>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21c0-4 3.6-6.5 8-6.5s8 2.5 8 6.5" />
    </>
  ),
};

function IkonNav({ kunci }: { kunci: string }) {
  return (
    <svg
      className="dashboard-nav-ikon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      {IKON_NAV[kunci] ?? IKON_NAV.ringkasan}
    </svg>
  );
}

// Kerangka dashboard: header area kerja, sidebar modul sesuai peran, dan footer
// ringkas yang terpisah dari layout portal publik. Tombol menu menyembunyikan
// atau menampilkan sidebar di semua ukuran layar; pilihan di layar lebar
// diingat, sedangkan pada layar kecil sidebar selalu tertutup saat halaman
// pertama dibuka dan ditutup setelah navigasi agar konten tidak tertutup.
export function DashboardShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [principal, setPrincipal] = useState<DashboardPrincipal | null>(null);
  const [status, setStatus] = useState<Status>("memuat");
  // Dua keadaan sidebar untuk dua jenis layar:
  // - `terlipat` (layar lebar): sidebar menyempit menjadi rail ikon; menu tetap
  //   terlihat dan dapat dijangkau lewat title/aria-label tiap ikon.
  // - `panelTerbuka` (layar ≤900px): sidebar adalah panel lipat di atas konten.
  // Nilai awalnya sama di server dan klien agar tidak ada hydration mismatch;
  // efek di bawah menyesuaikannya dengan preferensi dan lebar layar.
  const [terlipat, setTerlipat] = useState(false);
  const [panelTerbuka, setPanelTerbuka] = useState(false);
  const [devLogin, setDevLogin] = useState(false);

  useEffect(() => {
    let tersimpan: string | null = null;
    try {
      tersimpan = window.localStorage.getItem(KUNCI_SIDEBAR);
    } catch {
      // Mode privat atau penyimpanan diblokir: pakai tampilan bawaan.
    }
    // Layar sempit: panel selalu mulai tertutup agar konten tidak memakai
    // seluruh layar. Layar lebar: rail tertutup hanya bila terakhir dipilih.
    setTerlipat(window.matchMedia(LANGSING_SIDEBAR).matches && tersimpan === "tertutup");
  }, []);

  // Di layar kecil panel ditutup setiap navigasi agar konten tidak tertutup.
  // Di layar lebar pilihan pengguna dibiarkan agar tidak berubah saat berpindah
  // modul.
  useEffect(() => {
    if (window.matchMedia(SEMPIT_SIDEBAR).matches) setPanelTerbuka(false);
  }, [pathname]);

  // Satu tombol menu dengan dua arti: di layar sempit membuka/menutup panel,
  // di layar lebar melipat/membuka rail ikon. Preferensi rail diingat.
  function ubahTampilanSidebar() {
    if (window.matchMedia(SEMPIT_SIDEBAR).matches) {
      setPanelTerbuka((terbuka) => !terbuka);
      return;
    }
    const lipat = !terlipat;
    setTerlipat(lipat);
    try {
      window.localStorage.setItem(KUNCI_SIDEBAR, lipat ? "tertutup" : "terbuka");
    } catch {
      // Preferensi tidak dapat disimpan; perubahan tetap berlaku di sesi ini.
    }
  }

  const sempit = panelTerbuka;
  const sidebarTerlihat = sempit || !terlipat;

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
  const tombolLabel = sempit
    ? panelTerbuka
      ? "Sembunyikan menu dashboard"
      : "Buka menu dashboard"
    : terlipat
      ? "Lebarkan menu dashboard"
      : "Lipat menu dashboard";

  return (
    <div className="dashboard-root">
      <a href="#isi-dashboard" className="skip-link">Lewati ke isi dashboard</a>

      <header className="dashboard-header">
        <div className="dashboard-header-inner">
          <button
            type="button"
            className="dashboard-toggle"
            // Nama tombol dibaca pembaca layar dan dipakai tes e2e.
            aria-label={tombolLabel}
            aria-expanded={sidebarTerlihat}
            aria-controls="sidebar-dashboard"
            title={tombolLabel}
            onClick={ubahTampilanSidebar}
          >
            <span aria-hidden="true" className="dashboard-toggle-ikon">
              {sempit ? (panelTerbuka ? "☰" : "»") : terlipat ? "»" : "«"}
            </span>
            <span className="dashboard-toggle-text">{tombolLabel}</span>
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
          className={`dashboard-sidebar${sempit ? (panelTerbuka ? " is-open" : " is-panel-tertutup") : ""}${!sempit && terlipat ? " is-terlipat" : ""}`}
          aria-label="Navigasi dashboard"
          data-terlipat={!sempit && terlipat ? "true" : undefined}
          // Panel lipat di layar sempit yang tertutup dikeluarkan dari urutan
          // fokus. Rail tertutup di layar lebar tetap dapat difokus (ikon +
          // tooltip), jadi tidak diberi `inert`.
          inert={sempit && !panelTerbuka ? true : undefined}
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
                        // Ikon rail tidak memiliki teks tampak; label aksesibel
                        // tetap utuh. Judul memunculkan keterangan modul.
                        aria-label={item.label}
                        title={item.deskripsi}
                      >
                        <IkonNav kunci={item.key} />
                        <span className="dashboard-nav-label">{item.label}</span>
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
  if (pathname.startsWith("/dashboard/dokumen")) return "dokumen";
  // Modul layanan berada di bawah /dashboard agar memakai sidebar area kerja.
  if (pathname.startsWith("/dashboard/certification")) {
    const akses = getRoleAccess(role);
    if (akses.has("sertifikasi")) return "sertifikasi";
    if (akses.has("laboratorium")) return "laboratorium";
    return "penerbitan";
  }
  if (pathname.startsWith("/dashboard/benih")) {
    // Modul benih dipakai bersama oleh registrasi, penjualan, dan laporan;
    // pilih kunci pertama yang tersedia untuk peran tersebut.
    const akses = getRoleAccess(role);
    if (akses.has("registrasi-benih")) return "registrasi-benih";
    if (akses.has("penjualan-benih")) return "penjualan-benih";
    return "laporan";
  }
  if (pathname.startsWith("/dashboard/horticulture")) return "hortikultura";
  if (pathname.startsWith("/dashboard/repository-data")) return "repository-data";
  if (pathname.startsWith("/dashboard")) return "ringkasan";
  return null;
}
