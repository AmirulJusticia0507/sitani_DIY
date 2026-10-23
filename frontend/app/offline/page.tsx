import Link from "next/link";

export default function OfflinePage() {
  return (
    <main className="info-page">
      <p className="eyebrow">SiTani DIY</p>
      <h1>Anda sedang offline</h1>
      <p className="info-lead">
        Portal memerlukan koneksi untuk memuat layanan dan data. Sambungkan kembali ke internet,
        lalu coba buka halaman ini lagi.
      </p>
      <Link href="/" className="btn btn-primary">Coba lagi</Link>
    </main>
  );
}
