import { RepositoryDataPanel } from "./repository-data-panel";

// Halaman modul berada di bawah app/dashboard sehingga memakai kerangka area
// kerja (sidebar peran + status sesi) dari app/dashboard/layout.tsx. Elemen
// utama sudah disediakan shell (main#isi-dashboard), jadi di sini cukup div.
export default function RepositoryDataPage() {
  return (
    <div className="layanan-page">
      <div className="layanan-shell">
        <a className="layanan-back" href="/dashboard">
          <span aria-hidden="true">←</span> Ringkasan dashboard
        </a>
        <div className="layanan-head">
          <p className="layanan-eyebrow">DPKP Daerah Istimewa Yogyakarta</p>
          <h1>Repository Data Lokal</h1>
          <p className="layanan-lead">
            Impor batch CSV atau XLSX, lalu tinjau kembali header dan baris yang tersimpan.
            Data sumber belum dinormalisasi ke model pertanian kanonis.
          </p>
        </div>
        <div className="layanan-body">
          <RepositoryDataPanel />
        </div>
      </div>
    </div>
  );
}
