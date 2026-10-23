import { HorticulturePanel } from "./horticulture-panel";
import { Layout } from "../../components/layout";

export default function HorticulturePage() {
  return (
    <Layout>
      <main className="horticulture-page">
        <header className="repository-header">
          <a className="repository-back" href="/">
            ← Beranda
          </a>
          <p className="eyebrow">Layanan hortikultura</p>
          <h1>Registrasi Lahan Usaha</h1>
          <p className="repository-lead">
            Ajukan registrasi kebun, lengkapi data tanaman, dan lampirkan berkas pendukung secara privat.
          </p>
        </header>
        <HorticulturePanel />
      </main>
    </Layout>
  );
}
