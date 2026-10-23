import Link from "next/link";
import { Layout } from "../components/layout";

const services = [
  { number: "01", title: "Repository Data Lokal", description: "Data pertanian lintas bidang dalam satu ruang kerja yang rapi dan mudah ditelusuri.", href: "/repository-data", tone: "blue" },
  { number: "02", title: "Registrasi Hortikultura", description: "Pengajuan kebun, verifikasi dokumen, inspeksi, dan penerbitan registrasi secara terpadu.", href: "/horticulture", tone: "green" },
  { number: "03", title: "Sertifikasi Benih BP3MBTP", description: "Alur sertifikasi tujuh aktor dari pemeriksaan lapangan hingga label benih resmi.", href: "/certification", tone: "gold" },
  { number: "04", title: "Registrasi & Penjualan Benih", description: "Kelola lot, stok, transaksi, masa berlaku, dan audit penjualan benih.", href: "/benih", tone: "mint" },
] as const;

export default function Home() {
  return (
    <Layout>
      <main className="landing">
        <section className="landing-hero">
          <div className="landing-shell landing-hero-grid">
            <div className="hero-copy">
              <p className="landing-kicker"><span /> Portal layanan pertanian DIY</p>
              <h1>Data tumbuh.<br />Layanan bergerak.</h1>
              <p className="landing-lead">Satu ruang digital untuk petani, pelaku usaha, laboratorium, dan perangkat daerah—dirancang agar setiap proses lebih jelas dan dapat dipercaya.</p>
              <div className="landing-actions">
                <Link href="#layanan" className="landing-button landing-button-primary">Jelajahi layanan <span aria-hidden="true">↗</span></Link>
                <Link href="/tentang" className="landing-text-link">Kenali SiTani DIY <span aria-hidden="true">→</span></Link>
              </div>
            </div>
            <div className="hero-visual" aria-label="Ilustrasi ekosistem data pertanian">
              <div className="orbit orbit-one" /><div className="orbit orbit-two" /><div className="hero-sun" />
              <div className="data-card data-card-main">
                <span className="data-label">Musim tanam 2026</span>
                <strong>Terhubung dari lahan<br />hingga sertifikasi.</strong>
                <div className="data-chart"><i /><i /><i /><i /><i /><i /></div>
              </div>
              <div className="data-card data-card-float"><span className="pulse-dot" /> Sistem aktif</div>
              <svg className="field-lines" viewBox="0 0 520 260" role="img" aria-label="Garis kontur lahan">
                <path d="M0 190C95 120 147 225 251 155s164-28 269-112" /><path d="M0 222c105-68 166 30 267-44s159-24 253-91" /><path d="M12 256c96-61 169 25 270-41s153-18 231-72" />
              </svg>
            </div>
          </div>
        </section>

        <section className="landing-intro"><div className="landing-shell intro-grid">
          <p className="section-index">01 / Tentang</p>
          <div><p className="intro-statement">Teknologi yang tenang di permukaan, kuat di balik layar.</p><p className="intro-detail">SiTani DIY menyatukan proses yang sebelumnya tersebar menjadi alur digital yang aman, transparan, dan mudah dipahami. Identitas diverifikasi melalui SSO Jogjaprov, sementara setiap perubahan penting meninggalkan jejak audit.</p></div>
        </div></section>

        <section className="services-showcase" id="layanan"><div className="landing-shell">
          <div className="services-heading"><p className="section-index light">02 / Layanan utama</p><h2>Empat layanan.<br />Satu ekosistem.</h2><p>Dari data dasar sampai dokumen resmi, setiap modul saling terhubung dan tetap mudah digunakan.</p></div>
          <div className="service-grid">{services.map((service) => (
            <Link className={`service-card service-${service.tone}`} href={service.href} key={service.number}>
              <span className="service-number">{service.number}</span><div className="service-icon" aria-hidden="true"><span /></div><h3>{service.title}</h3><p>{service.description}</p><span className="service-arrow" aria-hidden="true">↗</span>
            </Link>
          ))}</div>
        </div></section>

        <section className="impact-section"><div className="landing-shell">
          <div className="impact-head"><div><p className="section-index">03 / Ekosistem</p><h2>Dibangun untuk kerja nyata.</h2></div><p>Alur digital yang menyatukan pemohon, petugas lapangan, analis laboratorium, pejabat, legalisasi, dan percetakan.</p></div>
          <div className="impact-grid"><article><strong>4</strong><span>modul layanan utama</span></article><article><strong>7</strong><span>aktor sertifikasi benih</span></article><article><strong>2</strong><span>realm SSO Jogjaprov</span></article><article><strong>1</strong><span>sumber data terpadu</span></article></div>
        </div></section>

        <section className="landing-cta"><div className="landing-shell cta-inner">
          <p className="landing-kicker light"><span /> Siap digunakan</p><h2>Mulai dari kebutuhanmu.<br />Kami bantu alurnya.</h2>
          <div className="landing-actions"><Link href="/dashboard" className="landing-button landing-button-light">Buka dashboard <span aria-hidden="true">↗</span></Link><Link href="/bantuan" className="landing-text-link light-link">Lihat panduan <span aria-hidden="true">→</span></Link></div>
        </div><div className="landscape" aria-hidden="true"><i /><i /><i /></div></section>
      </main>
    </Layout>
  );
}
