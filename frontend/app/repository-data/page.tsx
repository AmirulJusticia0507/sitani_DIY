import { RepositoryDataPanel } from "./repository-data-panel";
import { Layout } from "../../components/layout";

export default function RepositoryDataPage() {
  return (
    <Layout>
      <main className="layanan-page">
        <div className="layanan-shell">
          <a className="layanan-back" href="/">
            <span aria-hidden="true">←</span> Beranda
          </a>
          <div className="layanan-head">
            <p className="layanan-eyebrow">DPKP Daerah Istimewa Yogyakarta</p>
            <h1>Repository Data Lokal</h1>
            <p className="layanan-lead">
              Impor batch CSV atau XLSX, lalu tinjau kembali header dan baris yang tersimpan.
              Data sumber belum dinormalisasi ke model pertanian kanonis.
            </p>
          </div>
          <div className="layanan-body">
            <RepositoryDataPanel />
          </div>
        </div>
      </main>
    </Layout>
  );
}