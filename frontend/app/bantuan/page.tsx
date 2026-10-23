import Link from "next/link";

const questions = [
  {
    question: "Bagaimana cara masuk ke portal?",
    answer:
      "Pilih Masuk JSP untuk masyarakat atau Masuk Satriya untuk ASN. Proses masuk memerlukan layanan SSO dan konfigurasi resmi. Jika belum dapat masuk, pastikan akun Anda aktif dan hubungi pengelola SSO melalui saluran resmi instansi.",
  },
  {
    question: "Saya sudah masuk, tetapi tidak dapat membuka layanan tertentu.",
    answer:
      "Setiap modul membatasi tindakan berdasarkan peran dan kewenangan. Hubungi administrator layanan atau unit DPKP yang menangani modul tersebut untuk pemeriksaan akses.",
  },
  {
    question: "Bagaimana mengunggah dokumen permohonan?",
    answer:
      "Buka Registrasi Hortikultura, lengkapi informasi lampiran, lalu unggah dokumen melalui formulir. Unggah lampiran saja belum berarti permohonan telah terkirim; kirim formulir permohonan setelah semua data dan berkas siap.",
  },
  {
    question: "Apa yang harus dilakukan jika terjadi kesalahan atau layanan tidak tersedia?",
    answer:
      "Catat nama modul, waktu kejadian, dan pesan kesalahan tanpa mengirim token, kata sandi, NIK, atau isi dokumen melalui kanal yang tidak resmi. Laporkan kepada administrator atau helpdesk DPKP DIY setelah saluran kontak resmi diumumkan.",
  },
  {
    question: "Apakah portal bisa dipasang sebagai aplikasi?",
    answer:
      "Ya, browser yang kompatibel dapat menawarkan opsi pemasangan saat situs disajikan melalui HTTPS. Pemasangan tidak membuat semua layanan tersedia offline; pengajuan dan data tetap memerlukan koneksi ke layanan backend.",
  },
];

export default function BantuanPage() {
  return (
    <article className="info-page">
      <Link href="/" className="info-back">← Beranda</Link>
      <p className="eyebrow">Pusat bantuan</p>
      <h1>Bantuan SiTani DIY</h1>
      <p className="info-lead">
        Panduan ringkas untuk masuk, menggunakan modul layanan, dan melaporkan kendala.
      </p>

      <div className="help-list">
        {questions.map(({ question, answer }) => (
          <details className="info-card help-item" key={question}>
            <summary>{question}</summary>
            <p>{answer}</p>
          </details>
        ))}
      </div>

      <section className="info-card">
        <h2>Kontak pengelola</h2>
        <p>
          Kontak bantuan resmi akan ditampilkan setelah ditetapkan oleh DPKP DIY. Untuk saat ini,
          gunakan saluran layanan resmi instansi dan jangan mengirimkan kata sandi, token SSO,
          NIK, atau dokumen pribadi melalui kanal publik.
        </p>
      </section>
    </article>
  );
}
