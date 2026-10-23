import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY",
  description:
    "Platform terpadu repository data lokal dan layanan publik pertanian serta ketahanan pangan Daerah Istimewa Yogyakarta. Mengelola registrasi lahan hortikultura, sertifikasi benih BP3MBTP, dan data pangan.",
  openGraph: {
    title: "SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY",
    description: "Portal terpadu data dan layanan pertanian DPKP Pemda DIY.",
    type: "website",
  },
  icons: [
    { rel: "icon", type: "image/svg+xml", url: "/favicon.svg" },
    { rel: "apple-touch-icon", url: "/favicon.svg" },
  ],
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body>{children}</body>
    </html>
  );
}
