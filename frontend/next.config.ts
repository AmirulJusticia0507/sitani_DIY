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
};

export default nextConfig;
