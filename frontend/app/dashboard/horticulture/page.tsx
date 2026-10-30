import { HorticulturePanel } from "./horticulture-panel";

// Halaman modul berada di bawah app/dashboard sehingga memakai kerangka area
// kerja (sidebar peran + status sesi) dari app/dashboard/layout.tsx. Elemen
// utama sudah disediakan shell (main#isi-dashboard), jadi di sini cukup div.
export default function HorticulturePage() {
  return (
    <div className="layanan-page">
      <div className="layanan-shell">
        <a className="layanan-back" href="/dashboard">
          <span aria-hidden="true">←</span> Ringkasan dashboard
        </a>
        <div className="layanan-head">
          <p className="layanan-eyebrow">Layanan hortikultura</p>
          <h1>Registrasi Lahan Usaha</h1>
          <p className="layanan-lead">
            Ajukan registrasi kebun, lengkapi data tanaman, dan lampirkan berkas pendukung secara privat.
          </p>
        </div>
        <div className="layanan-body">
          <HorticulturePanel />
        </div>
      </div>
    </div>
  );
}
