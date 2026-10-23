import Link from "next/link";

export default function KebijakanPrivasiPage() {
  return (
    <article className="info-page">
      <Link href="/" className="info-back">← Beranda</Link>
      <p className="eyebrow">Informasi pengguna</p>
      <h1>Kebijakan Privasi</h1>
      <p className="info-lead">
        Dokumen ini menjelaskan secara umum data yang dapat diproses saat menggunakan
        SiTani DIY. Rincian final perlu disahkan oleh DPKP DIY dan pengelola layanan terkait
        sebelum portal digunakan untuk produksi.
      </p>

      <section className="info-card">
        <h2>Data yang dapat diproses</h2>
        <ul>
          <li>Informasi identitas dari proses masuk SSO, seperti subject, nama, email, dan realm, bila tersedia dan diizinkan oleh penyedia identitas.</li>
          <li>Informasi yang Anda masukkan pada permohonan, registrasi, atau transaksi layanan.</li>
          <li>Dokumen dan lampiran yang Anda unggah untuk mendukung permohonan.</li>
          <li>Data teknis minimum untuk keamanan, audit, dan diagnosis layanan.</li>
        </ul>
      </section>

      <section className="info-card">
        <h2>Tujuan dan akses</h2>
        <p>
          Data digunakan untuk menyediakan layanan yang diminta, memverifikasi identitas,
          mengelola alur dan audit permohonan, menjaga keamanan, serta memenuhi kewajiban
          penyelenggara. Akses seharusnya dibatasi sesuai kewenangan tugas. Dokumen permohonan
          tidak dimaksudkan untuk menjadi informasi publik.
        </p>
      </section>

      <section className="info-card">
        <h2>Penyimpanan dan pembagian</h2>
        <p>
          Dokumen desain menetapkan penyimpanan lampiran privat. Lama penyimpanan, lokasi
          pemrosesan, pencadangan, dasar hukum, dan daftar penerima data masih perlu ditetapkan
          secara resmi. Data hanya boleh dibagikan kepada pihak berwenang atau layanan integrasi
          yang diperlukan sesuai dasar dan kewenangan yang berlaku.
        </p>
      </section>

      <section className="info-card">
        <h2>Keamanan dan hak pengguna</h2>
        <p>
          Sistem dirancang menggunakan autentikasi SSO, kontrol akses, dan pembatasan akses
          dokumen. Pengguna dapat menghubungi pengelola resmi untuk meminta informasi atau
          menyampaikan koreksi terkait data pribadi. Saluran kontak resmi dan prosedur permintaan
          akan dipublikasikan setelah ditetapkan oleh DPKP DIY.
        </p>
      </section>

      <section className="info-card">
        <h2>Cookie dan teknologi serupa</h2>
        <p>
          Proses autentikasi menggunakan cookie sesi yang diperlukan untuk mempertahankan
          status masuk. Detail masa berlaku, konfigurasi lingkungan, serta teknologi lain yang
          digunakan harus dikonfirmasi dalam kebijakan operasional sebelum produksi.
        </p>
      </section>

      <p className="info-updated">Status: draf informasi untuk tahap pengembangan. Ditinjau 23 Oktober 2026.</p>
    </article>
  );
}
