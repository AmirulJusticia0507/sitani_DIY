import Link from "next/link";
import "./repository-data.css";
import { RepositoryDataPanel } from "./repository-data-panel";
import { Layout } from "../../components/layout";

export default function RepositoryDataPage() {
  return (
    <Layout>
      <main className="repository-page">
        <header className="repository-header">
          <Link href="/" className="repository-back">
            ← Beranda
          </Link>
          <p className="eyebrow">DPKP Daerah Istimewa Yogyakarta</p>
          <h1>Repository Data Lokal</h1>
          <p className="repository-lead">
            Impor batch CSV atau XLSX, lalu tinjau kembali header dan baris yang tersimpan.
            Data sumber belum dinormalisasi ke model pertanian kanonis.
          </p>
        </header>
        <RepositoryDataPanel />
      </main>
    </Layout>
  );
}
