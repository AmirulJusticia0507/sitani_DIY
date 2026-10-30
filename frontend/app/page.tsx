import Link from "next/link";
import { Layout } from "../components/layout";
import { masukHref } from "../components/auth/masuk-url";

// Kartu layanan kini mengarah ke halaman login dengan tujuan modulnya: modul
// berada di area kerja /dashboard dan memerlukan sesi SSO terverifikasi.
const modules = [
  { title: "Repository Data Lokal", description: "Pusat data pertanian lintas bidang untuk pelaporan dan visualisasi terpadu.", href: masukHref("/dashboard/repository-data"), icon: "chart", label: "Data & pelaporan" },
  { title: "Registrasi Hortikultura", description: "Pengajuan kebun, verifikasi dokumen, inspeksi, dan penerbitan registrasi.", href: masukHref("/dashboard/horticulture"), icon: "leaf", label: "Layanan kebun" },
  { title: "Sertifikasi Benih BP3MBTP", description: "Alur tujuh aktor dari verifikasi lapangan sampai sertifikat dan label resmi.", href: masukHref("/dashboard/certification"), icon: "seed", label: "Sertifikasi" },
  { title: "Registrasi & Penjualan Benih", description: "Kelola registrasi lot, persediaan, transaksi, masa berlaku, dan audit.", href: masukHref("/dashboard/benih"), icon: "grain", label: "Benih & stok" },
] as const;

function ModuleIcon({ type }: { type: string }) {
  return (
    <svg viewBox="0 0 48 48" aria-hidden="true">
      {type === "chart" ? <><path d="M9 37V23h7v14M20 37V12h7v25M31 37V18h7v19" /><path d="M7 40h34" /></> : null}
      {type === "leaf" ? <><path d="M38 9C21 10 10 19 11 34c15 2 26-8 27-25Z" /><path d="M12 36c6-8 12-13 22-21" /></> : null}
      {type === "seed" ? <><path d="M24 39V20" /><path d="M24 25c-8 0-13-5-13-12 8 0 13 5 13 12ZM24 31c8 0 13-5 13-12-8 0-13 5-13 12Z" /></> : null}
      {type === "grain" ? <><path d="M24 40V9" /><path d="M24 17c-6 0-9-3-9-8 6 0 9 3 9 8ZM24 25c6 0 9-3 9-8-6 0-9 3-9 8ZM24 33c-6 0-9-3-9-8 6 0 9 3 9 8Z" /></> : null}
    </svg>
  );
}

export default function Home() {
  return (
    <Layout>
      <main className="portal-home">
        <section className="portal-hero">
          <div className="portal-shell portal-hero-grid">
            <div className="portal-hero-copy">
              <p className="portal-eyebrow"><span /> DPKP Daerah Istimewa Yogyakarta</p>
              <h1>Portal Pertanian dan Ketahanan Pangan <em>DIY</em></h1>
              <p>Satu pintu layanan dan data pertanian untuk masyarakat, pelaku usaha, serta perangkat daerah. Lebih sederhana, terukur, dan terpercaya.</p>
              <div className="portal-actions">
                <Link href={masukHref("/dashboard")} className="portal-primary-action">Masuk <span aria-hidden="true">↗</span></Link>
                <Link href="#layanan" className="portal-secondary-action">Lihat layanan <span aria-hidden="true">→</span></Link>
              </div>
            </div>
            <div className="portal-hero-art" aria-label="Ringkasan layanan SiTani DIY">
              <div className="portal-orbit" />
              <div className="portal-art-card">
                <span>Ekosistem SiTani</span>
                <strong>Layanan publik yang saling terhubung.</strong>
                <div className="portal-mini-chart"><i /><i /><i /><i /><i /></div>
              </div>
              <div className="portal-art-badge"><i /> Sistem aktif</div>
              <svg viewBox="0 0 520 160" aria-hidden="true"><path d="M0 125c84-68 153 19 238-41s157-16 282-73" /><path d="M0 152c102-61 169 20 265-40s151-11 255-63" /></svg>
            </div>
          </div>
        </section>

        <section className="portal-stats" aria-label="Ringkasan portal">
          <div className="portal-shell portal-stats-grid">
            <article><strong>4</strong><span>modul layanan utama</span></article>
            <article><strong>7</strong><span>aktor sertifikasi benih</span></article>
            <article><strong>2</strong><span>realm SSO Jogjaprov</span></article>
            <div className="portal-stats-note"><span className="portal-live-dot" /> Fondasi layanan tersedia</div>
          </div>
        </section>

        <section className="portal-services" id="layanan">
          <div className="portal-shell">
            <header className="portal-section-head">
              <div><p className="portal-eyebrow"><span /> Layanan utama</p><h2>Satu portal untuk setiap kebutuhan.</h2></div>
              <p>Pilih layanan yang dibutuhkan. Setiap modul dirancang sebagai bagian dari alur data yang sama.</p>
            </header>
            <div className="portal-module-grid">
              {modules.map((module, index) => (
                <Link href={module.href} className="portal-module-card" key={module.title}>
                  <div className="portal-module-top"><span>0{index + 1}</span><span>{module.label}</span></div>
                  <div className="portal-module-icon"><ModuleIcon type={module.icon} /></div>
                  <h3>{module.title}</h3><p>{module.description}</p>
                  <span className="portal-card-link">Buka layanan <i aria-hidden="true">↗</i></span>
                </Link>
              ))}
            </div>
          </div>
        </section>

        <section className="portal-sso" id="tentang">
          <div className="portal-shell portal-sso-grid">
            <div><p className="portal-eyebrow portal-eyebrow-light"><span /> Akses aman</p><h2>Satu identitas untuk layanan yang terhubung.</h2></div>
            <div className="portal-sso-copy"><p>Identitas pengguna diverifikasi melalui SSO Pemda DIY. Masuk sebagai JSP untuk masyarakat dan pelaku usaha, atau sebagai Satriya untuk ASN dan aparatur pemerintah.</p><Link href={masukHref("/dashboard")}>Masuk ke dashboard <span aria-hidden="true">→</span></Link></div>
          </div>
          <div className="portal-landscape" aria-hidden="true"><i /><i /></div>
        </section>
      </main>
    </Layout>
  );
}
