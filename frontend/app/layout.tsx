import type { Metadata, Viewport } from "next";
import { Poppins } from "next/font/google";
import "./globals.css";

// Tipografi mengikuti portal Jogjaprov (Poppins). Font di-host otomatis oleh
// Next.js agar tidak bergantung pada request ke Google saat runtime.
const poppins = Poppins({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800"],
  variable: "--font-poppins",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY",
    template: "%s — SiTani DIY",
  },
  description:
    "Platform terpadu repository data lokal dan layanan publik pertanian serta ketahanan pangan Daerah Istimewa Yogyakarta. Mengelola registrasi lahan hortikultura, sertifikasi benih BP3MBTP, dan data pangan.",
  openGraph: {
    title: "SiTani DIY — Portal Pertanian dan Ketahanan Pangan DIY",
    description: "Portal terpadu data dan layanan pertanian DPKP Pemda DIY.",
    type: "website",
  },
  icons: [
    { rel: "icon", type: "image/svg+xml", url: "/favicon.svg" },
    // iOS tidak memakai SVG untuk ikon layar utama; sediakan PNG 180x180.
    { rel: "apple-touch-icon", sizes: "180x180", url: "/apple-touch-icon.png" },
  ],
  applicationName: "SiTani DIY",
  appleWebApp: {
    capable: true,
    title: "SiTani DIY",
    statusBarStyle: "default",
  },
};

// theme-color tidak lagi dibaca dari `metadata` pada Next.js 16; tempatnya di
// `viewport`. Nilainya sama dengan theme_color pada app/manifest.ts agar bilah
// peramban dan splash screen PWA konsisten.
export const viewport: Viewport = {
  themeColor: "#0d4938",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id" className={poppins.variable}>
      <body>{children}</body>
    </html>
  );
}
