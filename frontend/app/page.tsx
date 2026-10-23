const modules = [
  {
    title: "Repository Data Lokal",
    description: "Pusat data pertanian lintas bidang untuk pelaporan dan visualisasi terpadu.",
  },
  {
    title: "Registrasi Hortikultura",
    description: "Alur permohonan, verifikasi, inspeksi, evaluasi, dan penerbitan registrasi.",
  },
];

export default function Home() {
  return (
    <main>
      <section className="hero">
        <p className="eyebrow">DPKP Daerah Istimewa Yogyakarta</p>
        <h1>Portal Pertanian dan Ketahanan Pangan DIY</h1>
        <p className="lead">
          Satu pintu layanan dan data pertanian untuk masyarakat, pelaku usaha, serta perangkat daerah.
        </p>
        <span className="status">Fondasi aplikasi sedang dikembangkan</span>
      </section>

      <section className="modules" aria-labelledby="modules-heading">
        <div>
          <p className="eyebrow">Tahap pertama</p>
          <h2 id="modules-heading">Layanan utama</h2>
        </div>
        <div className="grid">
          {modules.map((module) => (
            <article key={module.title}>
              <h3>{module.title}</h3>
              <p>{module.description}</p>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
