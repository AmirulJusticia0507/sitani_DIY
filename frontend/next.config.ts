import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  poweredByHeader: false,
  // Portal dev dibuka lewat 127.0.0.1:3002 (lihat DEV_FRONTEND_HOST di .env),
  // sedangkan Next.js 16 memblokir aset dev dari host selain localhost. Tanpa
  // entri ini HMR ditolak (ERR_INVALID_HTTP_RESPONSE) dan JavaScript halaman
  // tidak selesai dimuat, sehingga tombol masuk tertahan di "Menyiapkan…".
  // Hanya berlaku di `next dev`; build produksi tidak berubah.
  allowedDevOrigins: ["127.0.0.1"],
  async headers() {
    return [
      {
        // Header dasar portal. X-Frame-Options DENY aman karena SiTani tidak
        // pernah ditanam sebagai iframe di portal lain.
        source: "/(.*)",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "DENY" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
      {
        // Service worker harus selalu diambil ulang agar pembaruan cache PWA
        // tidak tertahan versi lama. Masa simpan aset `/_next/static/` yang
        // ber-hash tidak terpengaruh aturan ini.
        source: "/sw.js",
        headers: [
          { key: "Content-Type", value: "application/javascript; charset=utf-8" },
          { key: "Cache-Control", value: "no-cache, no-store, must-revalidate" },
        ],
      },
    ];
  },
};

export default nextConfig;
