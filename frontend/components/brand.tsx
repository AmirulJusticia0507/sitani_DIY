import Link from "next/link";

// Merek tunggal agar header publik dan header dashboard tetap konsisten.
export function Brand({ href = "/", variant = "light" }: { href?: string; variant?: "light" | "dark" }) {
  return (
    <Link href={href} className={`brand brand-${variant}`} aria-label="SiTani DIY — beranda portal">
      <span className="brand-mark" aria-hidden="true">🌾</span>
      <span className="brand-name">SiTani DIY</span>
    </Link>
  );
}
