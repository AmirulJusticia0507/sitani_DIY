import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SiTani DIY",
  description: "Portal Pertanian dan Ketahanan Pangan Daerah Istimewa Yogyakarta",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id">
      <body>{children}</body>
    </html>
  );
}
