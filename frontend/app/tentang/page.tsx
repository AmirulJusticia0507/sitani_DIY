import Link from "next/link";

export default function TentangPage() {
  return (
    <article className="info-page">
      <Link href="/" className="info-back">← Beranda</Link>
      <p className="eyebrow">Mengenal portal</p>
      <h1>Tentang SiTani DIY</h1>
      <p className="info-lead">
        SiTani DIY adalah portal terpadu untuk mendukung layanan publik dan pengelolaan
        data pertanian serta ketahanan pangan di Daerah Istimewa Yogyakarta.
      </p>

      <section className="info-card">
        <h2>Tujuan</h2>
        <p>
          Memudahkan masyarakat, pelaku usaha, petani, dan perangkat daerah mengakses
          layanan pertanian, sekaligus menjadi fondasi repository data lokal DPKP DIY.
        </p>
      </section>

      <section className="info-card">
        <h2>Layanan portal</h2>
        <ul>
          <li>Repository data lokal dan sarana impor dataset.</li>
          <li>Registrasi lahan usaha hortikultura.</li>
          <li>Sertifikasi dan alur perbenihan BP3MBTP.</li>
          <li>Registrasi lot, persediaan, dan penjualan benih.</li>
        </ul>
      </section>

      <section className="info-card">
        <h2>Status pengembangan</h2>
        <p>
          Portal masih dalam tahap pengembangan dan integrasi. Ketersediaan halaman
          atau alur lokal tidak dengan sendirinya menandakan layanan pemerintah eksternal
          seperti SSO, SPL/Cakra, Sadewa, atau SIM Perbenihan telah aktif di produksi.
          Integrasi tersebut memerlukan konfigurasi resmi, pengujian bersama, dan UAT.
        </p>
      </section>

      <p className="info-updated">Informasi ini terakhir ditinjau 23 Oktober 2026.</p>
    </article>
  );
}
