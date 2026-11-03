// Next.js 16 menolak `next start` ketika `output: "standalone"` (lihat
// next.config.ts). Server produksi yang dipakai image Docker adalah
// `.next/standalone/server.js`, tetapi folder `public` dan `.next/static` tidak
// disalin otomatis ke sana. Skrip ini menyalin keduanya lalu menjalankan server
// yang sama, sehingga `npm start` konsisten dengan container.
//
// Pengaturan port memakai variabel lingkungan seperti server standalone biasa:
//   PORT=3100 HOSTNAME=127.0.0.1 npm start
import { spawnSync } from "node:child_process";
import { cpSync, existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const standalone = path.join(root, ".next", "standalone");

if (!existsSync(standalone)) {
  console.error("Folder .next/standalone tidak ditemukan. Jalankan `npm run build` lebih dulu.");
  process.exit(1);
}

// copyFile/dereference bawaan cpSync mengikuti symlink; cukup untuk aset statis.
cpSync(path.join(root, "public"), path.join(standalone, "public"), { recursive: true });
cpSync(path.join(root, ".next", "static"), path.join(standalone, ".next", "static"), { recursive: true });

const hasil = spawnSync(process.execPath, [path.join(standalone, "server.js")], {
  stdio: "inherit",
  env: process.env,
});
process.exit(hasil.status ?? 1);
