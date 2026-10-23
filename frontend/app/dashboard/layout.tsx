import { DashboardShell } from "../../components/dashboard/dashboard-shell";

// Layout dashboard dipasang di level `app/dashboard` (bukan route group) agar
// hanya rute /dashboard* yang memakai header area kerja, sidebar, dan footer
// ringkas. Halaman publik tetap memakai `components/layout.tsx`. Sesi dibaca
// di klien karena cookie SSO bersifat HttpOnly dan hanya dikirim ke backend.
export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return <DashboardShell>{children}</DashboardShell>;
}
