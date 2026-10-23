import type { Metadata } from "next";

import { ProfilKlien } from "./profil-klien";

export const metadata: Metadata = {
  title: { absolute: "Profil & Akses — SiTani DIY" },
  robots: { index: false, follow: false },
};

export default function ProfilPage() {
  return (
    <div className="dashboard-page">
      <div className="dashboard-page-head">
        <p className="eyebrow">Akun</p>
        <h1>Profil &amp; Akses</h1>
        <p>Identitas dari sesi SSO dan modul dashboard yang dibuka untuk peran Anda.</p>
      </div>
      <ProfilKlien />
    </div>
  );
}
