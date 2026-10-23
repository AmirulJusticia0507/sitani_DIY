import type { Metadata } from "next";

import { MasukLokalKlien } from "./masuk-lokal-klien";

export const metadata: Metadata = {
  title: { absolute: "Masuk — SiTani DIY" },
  robots: { index: false, follow: false },
};

export default function MasukPage() {
  return (
    <div className="dashboard-page">
      <div className="dashboard-page-head">
        <p className="eyebrow">Akun</p>
        <h1>Masuk ke SiTani DIY</h1>
        <p>Gunakan akun JSP/Satriya, atau akun lokal bila login development diaktifkan.</p>
      </div>
      <MasukLokalKlien />
    </div>
  );
}
