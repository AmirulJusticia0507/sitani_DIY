import type { Metadata } from "next";

import { DokumenKlien } from "./dokumen-klien";

export const metadata: Metadata = {
  title: { absolute: "Dokumen — SiTani DIY" },
  robots: { index: false, follow: false },
};

export default function DokumenPage() {
  return (
    <div className="dashboard-page">
      <div className="dashboard-page-head">
        <p className="eyebrow">Lampiran</p>
        <h1>Dokumen</h1>
        <p>
          Daftar lampiran sesuai hak akses Anda. Unggahan dilakukan dari halaman permohonan; unduhan
          memakai tautan bertanda tangan berumur pendek dari backend.
        </p>
      </div>
      <DokumenKlien />
    </div>
  );
}
