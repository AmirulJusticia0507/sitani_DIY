import Link from "next/link";
import { Card } from "../components/card";
import { Layout } from "../components/layout";

interface Module {
  title: string;
  description: string;
  href: string;
  icon: string;
  color: string;
}

const modules: Module[] = [
  {
    title: "Repository Data Lokal",
    description:
      "Pusat data pertanian lintas bidang untuk pelaporan dan visualisasi terpadu. Impor batch CSV atau XLSX, lalu tinjau kembali header dan baris yang tersimpan.",
    href: "/repository-data",
    icon: "📊",
    color: "sitani-green-500",
  },
  {
    title: "Registrasi Hortikultura",
    description:
      "Ajukan registrasi kebun, lengkapi data tanaman, dan lampirkan berkas pendukung secara privat melalui alur pengajuan, verifikasi, dan inspeksi.",
    href: "/horticulture",
    icon: "🌿",
    color: "sitani-green-600",
  },
  {
    title: "Sertifikasi Benih BP3MBTP",
    description:
      "Alur 7 aktor: verifikasi lapangan, pengujian lab mutu, penerbitan sertifikat dengan TTE, draft label, hingga penyerahan dan pembayaran akhir.",
    href: "/certification",
    icon: "🌱",
    color: "sitani-emerald-500",
  },
  {
    title: "Registrasi & Penjualan Benih",
    description:
      "Registrasi lot, persediaan per lot, penjualan, pembatalan dengan jejak audit, dan laporan stok serta kedaluwarsa.",
    href: "/benih",
    icon: "🌾",
    color: "sitani-earth-600",
  },
];

export default function Home() {
  return (
    <Layout>
    <div className="space-y-12">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-[linear-gradient(135deg,#0d4938_0%,#1f6b49_100%)] -z-10" />
        <div className="container">
          <div className="hero-content">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-green-500/10 text-green-600 font-bold text-sm uppercase tracking-wider mb-4 border border-green-200">
              DPKP Daerah Istimewa Yogyakarta
            </div>
            <h1 className="text-4xl md:text-5xl font-extrabold text-white leading-tight mb-4">
              Portal Pertanian dan Ketahanan Pangan <span className="text-green-400">DIY</span>
            </h1>
            <p className="text-green-100 text-lg md:text-xl max-w-2xl mb-8 leading-relaxed">
              Satu pintu layanan dan data pertanian untuk masyarakat, pelaku usaha,
              serta perangkat daerah. Terpadu, terukur, dan terpercaya.
            </p>
            <div className="flex flex-wrap gap-3 justify-center md:justify-start">
              <Link href="#layanan" className="btn btn-primary">
                Mulai Layanan
              </Link>
              <Link href="#tentang" className="btn btn-outline hero-secondary-link">
                Tentang Portal
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Status ringkas */}
      <section className="bg-white border border-sitani-border/50">
        <div className="container">
          <div className="grid gap-4 md:gap-6 lg:gap-8">
            <div className="p-6">
              <div className="text-green-600 font-bold text-2xl">4 Modul</div>
              <div className="text-sitani-muted text-sm">Layanan utama tersedia</div>
            </div>
            <div className="p-6">
              <div className="text-green-600 font-bold text-2xl">7 Aktor</div>
              <div className="text-sitani-muted text-sm">Alur sertifikasi benih BP3MBTP</div>
            </div>
            <div className="p-6">
              <div className="text-green-600 font-bold text-2xl">SSO Terverifikasi</div>
              <div className="text-sitani-muted text-sm">JSP &amp; Satriya realm</div>
            </div>
          </div>
        </div>
      </section>

      {/* Layanan utama */}
      <section className="section" id="layanan">
        <div className="container">
          <div className="section-head">
            <span className="eyebrow">Tahap pertama</span>
            <h2>Layanan Utama</h2>
            <p>Pilih modul sesuai kebutuhan layanan publik pertanian dan ketahanan pangan.</p>
          </div>

          <div className="grid-cols grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            {modules.map((module) => (
              <Card key={module.title} hover>
                <div className="text-3xl mb-3" aria-hidden="true">
                  {module.icon}
                </div>
                <h3 className="text-xl mb-2">{module.title}</h3>
                <p className="text-sitani-muted text-sm mb-4 leading-relaxed">
                  {module.description}
                </p>
                <Link href={module.href} className="btn btn-outline btn-sm text-sitani-green-700 border-sitani-green-300">
                  Buka layanan →
                </Link>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Info SSO */}
      <section className="bg-sitani-green-800 text-green-100 border border-green-700/30" id="tentang">
        <div className="container">
          <div className="grid gap-6 md:grid-cols-2 items-center">
            <div>
              <h2 className="text-2xl font-bold mb-2">Akses Aman melalui SSO</h2>
              <p className="text-green-200 text-sm leading-relaxed">
                Semua layanan memerlukan verifikasi identitas melalui SSO Pemda DIY.
                Masyarakat menggunakan realm <strong>JSP</strong>, sedangkan ASN dan
                aparatur pemerintah menggunakan realm <strong>Satriya</strong>.
              </p>
            </div>
            <div className="flex justify-end">
              <div className="p-6 bg-green-700/30 rounded-xl border border-green-500/20">
                <div className="text-xs text-green-300 uppercase tracking-wider mb-1">Status sistem</div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" aria-hidden="true"></span>
                  <span className="text-sm font-medium">Fondasi layanan tersedia</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
    </Layout>
  );
}
