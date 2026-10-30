import { HorticulturePanel } from "./horticulture-panel";
import { Layout } from "../../components/layout";

export default function HorticulturePage() {
  return (
    <Layout>
      <main className="layanan-page">
        <div className="layanan-shell">
          <a className="layanan-back" href="/">
            <span aria-hidden="true">←</span> Beranda
          </a>
          <div className="layanan-head">
            <p className="layanan-eyebrow">Layanan hortikultura</p>
            <h1>Registrasi Lahan Usaha</h1>
            <p className="layanan-lead">
              Ajukan registrasi kebun, lengkapi data tanaman, dan lampirkan berkas pendukung secara privat.
            </p>
          </div>
          <div className="layanan-body">
            <HorticulturePanel />
          </div>
        </div>
      </main>
    </Layout>
  );
}