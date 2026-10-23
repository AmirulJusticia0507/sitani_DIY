import type { Metadata } from "next";

import { RingkasanKlien } from "./ringkasan-klien";

export const metadata: Metadata = {
  title: { absolute: "Dashboard — SiTani DIY" },
  description: "Area kerja modul pertanian dan ketahanan pangan DIY sesuai peran pengguna.",
  robots: { index: false, follow: false },
};

export default function DashboardPage() {
  return (
    <div className="dashboard-page">
      <div className="dashboard-page-head">
        <p className="eyebrow">Area kerja</p>
        <h1>Dashboard SiTani DIY</h1>
        <p>
          Ringkasan sesi dan modul kerja yang tersedia untuk peran Anda. Pengelolaan permohonan,
          sertifikasi, benih, dan repository data dilakukan di modul masing-masing.
        </p>
      </div>
      <RingkasanKlien />
    </div>
  );
}
