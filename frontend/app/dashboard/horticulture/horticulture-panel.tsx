"use client";

import { FormEvent, useCallback, useEffect, useId, useRef, useState } from "react";

import { masukHref } from "../../../components/auth/masuk-url";

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";
type Principal = { name?: string; email: string; role: string };
type Application = { id: number; nomor_permohonan: string; nama_organisasi: string; status: string; dibuat_pada: string; jadwal_inspeksi?: { id: number; waktu: string; lokasi: string; status: string }[]; inspeksi?: { id: number; temuan: string; rekomendasi: string }[]; penilaian?: { id: number; jenis_penilaian: string; rekomendasi: string; nama_tanaman: string; varietas: string; skor_vigoritas:number; skor_kesehatan:number; skor_kebenaran_varietas:number; catatan:string }[] };
type Upload = { id: number; nama: string };
type Attachment = { id: number; jenis: string; nama: string };
type CropRow = { nama: string; varietas: string; jumlah: string; sumber: string };

const attachmentTypes = [
  ["proposal", "Proposal permohonan"],
  ["organization", "Struktur organisasi"],
  ["members", "Daftar anggota"],
  ["farm_form", "Form registrasi kebun"],
  ["farm_map", "Peta kebun"],
  ["other", "Dokumen pendukung lain"],
] as const;

const jenisLampiran = (value: string) =>
  attachmentTypes.find(([jenis]) => jenis === value)?.[1] ?? value;

// Empat langkah form pengajuan. Tampilan hanya memecah bagian; data tetap dikirim
// sebagai satu permintaan di akhir, sehingga validasi dan otorisasi tetap milik
// backend.
const langkahPengajuan = [
  { judul: "Dokumen pendukung", ringkas: "Lampiran privat" },
  { judul: "Pemohon", ringkas: "Kelompok tani" },
  { judul: "Lokasi lahan", ringkas: "Alamat & luas" },
  { judul: "Komoditas", ringkas: "Tanaman & benih" },
] as const;

// Daftar kabupaten/kota DIY untuk pilihan cepat; isian tetap berupa kolom teks
// biasa agar nilai lain tidak tertolak.
const kabupatenDIY = ["Kota Yogyakarta", "Sleman", "Bantul", "Gunungkidul", "Kulon Progo"];

const cropKosong = (): CropRow => ({ nama: "", varietas: "", jumlah: "", sumber: "" });

async function errorMessage(response: Response): Promise<string> {
  try { return ((await response.json()) as { error?: string }).error ?? `Permintaan gagal (${response.status})`; }
  catch { return `Permintaan gagal (${response.status})`; }
}

export function HorticulturePanel() {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [applications, setApplications] = useState<Application[]>([]);
  const [selected, setSelected] = useState<Application | null>(null);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  // Form pemohon: nilai dipisah per kolom agar pilihan kabupaten, pratinjau
  // lampiran, dan ringkasan langkah bisa ditampilkan/asalkan tanpa membaca DOM.
  const [jenisAktif, setJenisAktif] = useState<string>("proposal");
  const [berkasTerpilih, setBerkasTerpilih] = useState<string>("");
  const [organisasi, setOrganisasi] = useState("");
  const [kontak, setKontak] = useState("");
  const [telepon, setTelepon] = useState("");
  const [alamat, setAlamat] = useState("");
  const [kalurahan, setKalurahan] = useState("");
  const [kapanewon, setKapanewon] = useState("");
  const [kabupaten, setKabupaten] = useState("");
  const [luas, setLuas] = useState("");
  const [crops, setCrops] = useState<CropRow[]>([cropKosong()]);

  const loadApplications = useCallback(async () => {
    const response = await fetch(`${apiURL}/v1/horticulture/applications`, { credentials: "include" });
    if (!response.ok) throw new Error(await errorMessage(response));
    const body = (await response.json()) as { permohonan: Application[] };
    setApplications(body.permohonan ?? []);
  }, []);

  useEffect(() => {
    let active = true;
    async function initialize() {
      try {
        const response = await fetch(`${apiURL}/auth/me`, { credentials: "include" });
        if (!response.ok) return;
        const user = (await response.json()) as Principal;
        if (!active) return;
        setPrincipal(user);
        if (["pelanggan", "admin", "superadmin", "pengawas", "verifikator", "inspektor", "pbt", "kasi_pmbtp"].includes(user.role)) await loadApplications();
      } catch (cause) { if (active) setError(cause instanceof Error ? cause.message : "Gagal memuat data."); }
      finally { if (active) setLoading(false); }
    }
    void initialize();
    return () => { active = false; };
  }, [loadApplications]);

  async function uploadFiles(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    const file = new FormData(formElement).get("file");
    if (!(file instanceof File) || file.size === 0) { setError("Pilih berkas terlebih dahulu."); return; }
    setBusy(true); setError(""); setMessage("");
    const upload = new FormData();
    upload.set("file", file);
    upload.set("tujuan", "permohonan_registrasi_hortikultura");
    try {
      const response = await fetch(`${apiURL}/v1/dokumen`, { method: "POST", credentials: "include", body: upload });
      if (!response.ok) throw new Error(await errorMessage(response));
      const saved = (await response.json()) as Upload;
      setAttachments((current) => [...current, { id: saved.id, nama: saved.nama, jenis: jenisAktif }]);
      formElement.reset();
      setBerkasTerpilih("");
      setMessage(`Lampiran “${saved.nama}” berhasil diunggah secara privat.`);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Unggah lampiran gagal."); }
    finally { setBusy(false); }
  }

  async function submitApplication(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (attachments.length === 0) { setError("Unggah minimal satu lampiran sebelum mengirim permohonan."); return; }
    const tanaman = crops.flatMap((row) => {
      const nama = row.nama.trim();
      const varietas = row.varietas.trim();
      if (!nama && !varietas) return [];
      return [{ nama, varietas, jumlah_tanaman: Number(row.jumlah || 0), sumber_benih: row.sumber.trim() }];
    });
    setBusy(true); setError(""); setMessage("");
    try {
      const response = await fetch(`${apiURL}/v1/horticulture/applications`, {
        method: "POST", credentials: "include", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nama_organisasi: organisasi, nama_kontak: kontak, telepon,
          alamat, kalurahan, kapanewon,
          kabupaten_kota: kabupaten, luas_lahan_ha: Number(luas), tanaman,
          dokumen: attachments.map(({ id, jenis }) => ({ id, jenis })),
        }),
      });
      if (!response.ok) throw new Error(await errorMessage(response));
      const saved = (await response.json()) as Application;
      setMessage(`Permohonan ${saved.nomor_permohonan} berhasil diajukan.`);
      setAttachments([]);
      setOrganisasi(""); setKontak(""); setTelepon(""); setAlamat("");
      setKalurahan(""); setKapanewon(""); setKabupaten(""); setLuas("");
      setCrops([cropKosong()]);
      await loadApplications();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Pengajuan gagal."); }
    finally { setBusy(false); }
  }

  async function openApplication(application: Application) {
    try {
      const response = await fetch(`${apiURL}/v1/horticulture/applications/${application.id}`, { credentials: "include" });
      if (!response.ok) throw new Error(await errorMessage(response));
      setSelected((await response.json()) as Application);
      setError("");
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Gagal membuka permohonan."); }
  }

  async function action(url: string, method: string, body: unknown, success: string) {
    setBusy(true); setError(""); setMessage("");
    try {
      const response = await fetch(`${apiURL}${url}`, { method, credentials: "include", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
      if (!response.ok) throw new Error(await errorMessage(response));
      setMessage(success);
      await loadApplications();
      if (selected) await openApplication(selected);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Aksi gagal."); }
    finally { setBusy(false); }
  }

  if (loading) return <section className="repository-content"><p role="status">Memeriksa sesi SSO…</p></section>;
  if (!principal) {
    const redirect = "/dashboard/horticulture";
    return <section className="repository-content"><div className="repository-card"><h2>Masuk diperlukan</h2><p>Masuk dengan akun yang telah terverifikasi melalui SSO.</p><div className="repository-actions"><a href={masukHref(redirect, "jsp")}>Masuk dengan JSP</a><a href={masukHref(redirect, "satriya")}>Masuk dengan Satriya</a></div></div></section>;
  }

  const bolehMengajukan = principal.role === "pelanggan" || principal.role === "admin" || principal.role === "superadmin";
  const bolehMelihatRiwayat = ["pelanggan", "admin", "superadmin", "pengawas", "verifikator", "inspektor", "pbt", "kasi_pmbtp"].includes(principal.role);
  const komoditasTerisi = crops.filter((row) => row.nama.trim() || row.varietas.trim()).length;

  return <section className="repository-content">
    <p className="repository-session">Masuk sebagai <strong>{principal.name || principal.email}</strong> · {principal.role}</p>
    {bolehMengajukan ? <>
      <section className="repository-card hortikultura-pengajuan" aria-labelledby="form-title">
        <header className="pengajuan-head">
          <div>
            <p className="eyebrow">Pengajuan baru</p>
            <h2 id="form-title">Registrasi kebun hortikultura</h2>
            <p>Nomor registrasi diberikan setelah permohonan tercatat. Pengajuan ini belum merupakan persetujuan registrasi.</p>
          </div>
          <ol className="pengajuan-langkah" aria-label="Langkah pengajuan">
            {langkahPengajuan.map((langkah, index) => (
              <li key={langkah.judul}>
                <span aria-hidden="true">{index + 1}</span>
                <span>
                  <strong>{langkah.judul}</strong>
                  <small>{langkah.ringkas}</small>
                </span>
              </li>
            ))}
          </ol>
        </header>

        {error && <p className="repository-alert error" role="alert">{error}</p>}
        {message && <p className="repository-alert success" role="status">{message}</p>}

        <form className="hortikultura-form" onSubmit={submitApplication}>
          <fieldset className="pengajuan-bagian">
            <legend>
              <span className="pengajuan-nomor" aria-hidden="true">1</span>
              Dokumen pendukung
            </legend>
            <p className="pengajuan-hint">
              Unggah proposal, struktur organisasi, daftar anggota, form registrasi kebun, dan peta kebun.
              Berkas disimpan privat dan hanya dapat dibuka pengguna berhak.
            </p>
            <Pengunggah
              jenis={jenisAktif}
              onJenis={setJenisAktif}
              berkas={berkasTerpilih}
              onBerkas={setBerkasTerpilih}
              busy={busy}
              onSubmit={uploadFiles}
            />
            <div className="lampiran">
              <div className="lampiran-judul">
                <h3>Lampiran siap dikirim</h3>
                <span className={attachments.length > 0 ? "lampiran-hitung is-ada" : "lampiran-hitung"}>
                  {attachments.length} berkas
                </span>
              </div>
              {attachments.length === 0
                ? <p className="lampiran-kosong">Belum ada lampiran. Unggah minimal satu berkas untuk melanjutkan.</p>
                : <ul>{attachments.map((item) => <li key={item.id}>
                    <span className="lampiran-ikon" aria-hidden="true">{ekstensi(item.nama)}</span>
                    <span className="lampiran-nama">
                      <strong>{item.nama}</strong>
                      <small>{jenisLampiran(item.jenis)}</small>
                    </span>
                    <button className="repository-link-button" type="button" onClick={() => setAttachments((current) => current.filter((doc) => doc.id !== item.id))}>
                      Hapus
                    </button>
                  </li>)}</ul>}
            </div>
          </fieldset>

          <fieldset className="pengajuan-bagian">
            <legend>
              <span className="pengajuan-nomor" aria-hidden="true">2</span>
              Data pemohon
            </legend>
            <div className="hortikultura-grid">
              <label className="kolom-lebar">
                <span>Nama kelompok tani / organisasi</span>
                <input name="organisasi" required maxLength={200} value={organisasi} onChange={(e) => setOrganisasi(e.target.value)} placeholder="Contoh: Kelompok Tani Ngudi Rejeki" />
              </label>
              <label>
                <span>Nama penanggung jawab</span>
                <input name="kontak" required maxLength={150} value={kontak} onChange={(e) => setKontak(e.target.value)} placeholder="Nama lengkap" />
              </label>
              <label>
                <span>Nomor telepon</span>
                <input name="telepon" required type="tel" maxLength={40} value={telepon} onChange={(e) => setTelepon(e.target.value)} placeholder="08xxxxxxxxxx" inputMode="tel" />
              </label>
            </div>
          </fieldset>

          <fieldset className="pengajuan-bagian">
            <legend>
              <span className="pengajuan-nomor" aria-hidden="true">3</span>
              Lokasi lahan
            </legend>
            <div className="hortikultura-grid">
              <label className="kolom-lebar">
                <span>Alamat lengkap</span>
                <textarea name="alamat" required maxLength={500} rows={3} value={alamat} onChange={(e) => setAlamat(e.target.value)} placeholder="Jalan, RT/RW, patokan lokasi kebun" />
              </label>
              <label>
                <span>Kalurahan</span>
                <input name="kalurahan" required maxLength={120} value={kalurahan} onChange={(e) => setKalurahan(e.target.value)} />
              </label>
              <label>
                <span>Kapanewon</span>
                <input name="kapanewon" required maxLength={120} value={kapanewon} onChange={(e) => setKapanewon(e.target.value)} />
              </label>
              <label>
                <span>Kabupaten / Kota</span>
                <input name="kabupaten" required maxLength={120} list="kabupaten-diy" value={kabupaten} onChange={(e) => setKabupaten(e.target.value)} />
                <datalist id="kabupaten-diy">
                  {kabupatenDIY.map((nama) => <option key={nama} value={nama} />)}
                </datalist>
              </label>
              <label>
                <span>Luas lahan (ha)</span>
                <input name="luas" required type="number" min="0.0001" max="100000" step="0.0001" value={luas} onChange={(e) => setLuas(e.target.value)} placeholder="0,2500" inputMode="decimal" />
              </label>
            </div>
          </fieldset>

          <fieldset className="pengajuan-bagian">
            <legend>
              <span className="pengajuan-nomor" aria-hidden="true">4</span>
              Detail tanaman
            </legend>
            <p className="pengajuan-hint">Isi minimal satu komoditas. Kolom jumlah dan sumber benih boleh dikosongkan bila belum diketahui.</p>
            <div className="komoditas-daftar">
              {crops.map((row, index) => (
                <article className="komoditas-kartu" key={index} aria-label={`Komoditas ${index + 1}`}>
                  <header>
                    <h3>Komoditas {index + 1}</h3>
                    {crops.length > 1 && (
                      <button className="repository-link-button" type="button" onClick={() => setCrops((current) => current.filter((_, i) => i !== index))}>
                        Hapus
                      </button>
                    )}
                  </header>
                  <div className="hortikultura-grid">
                    <label>
                      <span>Komoditas</span>
                      <input name={`tanaman_${index}_nama`} required={index === 0} maxLength={120} value={row.nama} onChange={(e) => ubahCrop(index, "nama", e.target.value, setCrops)} placeholder="Cabai, tomat, bawang merah…" />
                    </label>
                    <label>
                      <span>Varietas</span>
                      <input name={`tanaman_${index}_varietas`} required={index === 0} maxLength={120} value={row.varietas} onChange={(e) => ubahCrop(index, "varietas", e.target.value, setCrops)} />
                    </label>
                    <label>
                      <span>Jumlah tanaman</span>
                      <input name={`tanaman_${index}_jumlah`} type="number" min="0" max="100000000" step="1" value={row.jumlah} onChange={(e) => ubahCrop(index, "jumlah", e.target.value, setCrops)} inputMode="numeric" />
                    </label>
                    <label>
                      <span>Sumber benih</span>
                      <input name={`tanaman_${index}_sumber`} maxLength={200} value={row.sumber} onChange={(e) => ubahCrop(index, "sumber", e.target.value, setCrops)} placeholder="Balai benih, penangkar, swadaya…" />
                    </label>
                  </div>
                </article>
              ))}
            </div>
            <button className="tombol-tambah" type="button" onClick={() => setCrops((current) => [...current, cropKosong()])}>
              + Tambah komoditas
            </button>
          </fieldset>

          <div className="pengajuan-kirim">
            <p>
              <strong>{attachments.length} lampiran</strong> · <strong>{komoditasTerisi} komoditas</strong> terisi.
              Pastikan data sudah benar sebelum dikirim.
            </p>
            <button className="repository-button" type="submit" disabled={busy}>
              {busy ? "Mengirim…" : "Kirim permohonan"}
            </button>
          </div>
        </form>
      </section>
    </> : <section className="repository-card"><h2>Akses pemohon</h2><p>Pengajuan baru tersedia bagi pemohon atau admin.</p></section>}

    {bolehMelihatRiwayat && <section className="repository-card horticulture-history"><h2>Riwayat permohonan</h2>
      {applications.length === 0 ? <p>Belum ada permohonan.</p> : <div className="repository-table-wrap"><table className="repository-table"><thead><tr><th>Nomor</th><th>Organisasi</th><th>Status</th><th>Tanggal</th><th>Aksi</th></tr></thead><tbody>{applications.map((item) => <tr key={item.id}><td>{item.nomor_permohonan}</td><td>{item.nama_organisasi}</td><td>{item.status}</td><td>{new Date(item.dibuat_pada).toLocaleDateString("id-ID")}</td><td><button className="repository-link-button" type="button" onClick={() => void openApplication(item)}>Buka detail</button></td></tr>)}</tbody></table></div>}
    </section>}
    {selected && <section className="repository-card horticulture-history"><h2>Alur permohonan {selected.nomor_permohonan}</h2><p>Status: <strong>{selected.status}</strong></p>
      {principal.role === "verifikator" || principal.role === "admin" || principal.role === "superadmin" ? <div className="horticulture-actions"><button className="repository-button" disabled={busy} onClick={() => void action(`/v1/horticulture/applications/${selected.id}/verification`,"POST",{keputusan:"diverifikasi"},"Proposal dinyatakan lengkap dan terverifikasi.")}>Verifikasi lengkap</button><button className="repository-button secondary" disabled={busy} onClick={() => { const catatan=window.prompt("Catatan dokumen yang harus diperbaiki:"); if(catatan?.trim()) void action(`/v1/horticulture/applications/${selected.id}/verification`,"POST",{keputusan:"perlu_perbaikan",catatan},"Permintaan perbaikan dicatat."); }}>Minta perbaikan</button></div> : null}
      {principal.role === "inspektor" || principal.role === "admin" || principal.role === "superadmin" ? <>
        <h3>Jadwalkan inspeksi lapangan</h3><form className="horticulture-inline" onSubmit={(event) => { event.preventDefault(); const v=new FormData(event.currentTarget); void action(`/v1/horticulture/applications/${selected.id}/inspection-schedules`,"POST",{waktu:new Date(String(v.get("waktu"))).toISOString(),lokasi:v.get("lokasi"),catatan:v.get("catatan")},"Jadwal inspeksi tersimpan."); }}><label>Waktu<input required name="waktu" type="datetime-local" /></label><label>Lokasi<input required name="lokasi" maxLength={255} /></label><label>Catatan<input name="catatan" /></label><button className="repository-button" disabled={busy}>Simpan jadwal</button></form>
        {selected.jadwal_inspeksi?.filter((schedule) => schedule.status === "terjadwal").map((schedule) => <form className="horticulture-inline" key={schedule.id} onSubmit={(event) => { event.preventDefault(); const v=new FormData(event.currentTarget); const checks=Object.fromEntries(gapChecks.map((key)=>[key,v.get(key)==="on"])); void action(`/v1/horticulture/inspection-schedules/${schedule.id}/complete`,"POST",{checklist_gap:checks,temuan:v.get("temuan"),rekomendasi:v.get("rekomendasi")},"Form audit GAP disimpan."); }}><h3>Audit GAP · {new Date(schedule.waktu).toLocaleString("id-ID")} · {schedule.lokasi}</h3><fieldset><legend>Checklist GAP (centang bila memenuhi)</legend>{gapChecks.map((key)=><label className="horticulture-check" key={key}><input type="checkbox" name={key}/>{gapLabels[key]}</label>)}</fieldset><label>Temuan<textarea name="temuan" required rows={3}/></label><label>Rekomendasi<textarea name="rekomendasi" required rows={3}/></label><button className="repository-button" disabled={busy}>Kirim hasil inspeksi</button></form>)}
      </> : null}
      {principal.role === "pbt" || principal.role === "kasi_pmbtp" || principal.role === "admin" || principal.role === "superadmin" ? <><h3>Penilaian pohon induk / PIT / RIP / PI / RI</h3><form className="horticulture-inline" onSubmit={(event)=>{event.preventDefault();const v=new FormData(event.currentTarget);void action(`/v1/horticulture/applications/${selected.id}/evaluations`,"POST",{jenis_penilaian:v.get("jenis"),nama_tanaman:v.get("tanaman"),varietas:v.get("varietas"),skor_vigoritas:Number(v.get("vigoritas")),skor_kesehatan:Number(v.get("kesehatan")),skor_kebenaran_varietas:Number(v.get("kebenaran")),rekomendasi:v.get("rekomendasi"),catatan:v.get("catatan")},"Penilaian tersimpan dalam audit trail.");}}><label>Jenis penilaian<select name="jenis">{["pohon_induk","pit","rip","pi","ri"].map((value)=><option key={value} value={value}>{value.toUpperCase()}</option>)}</select></label><label>Nama tanaman<input name="tanaman" required/></label><label>Varietas<input name="varietas" required/></label>{[["vigoritas","Vigoritas"],["kesehatan","Kesehatan"],["kebenaran","Kebenaran varietas"]].map(([name,label])=><label key={name}>{label} (1–5)<input type="number" name={name} min="1" max="5" required/></label>)}<label>Rekomendasi<select name="rekomendasi"><option value="layak">Layak</option><option value="perlu_perbaikan">Perlu perbaikan</option><option value="tidak_layak">Tidak layak</option></select></label><label>Catatan penilaian<textarea name="catatan" required rows={3}/></label><button className="repository-button" disabled={busy}>Simpan penilaian</button></form></> : null}
      {(selected.inspeksi?.length ?? 0)>0 && <><h3>Riwayat inspeksi GAP</h3><ul>{selected.inspeksi?.map((inspection)=><li key={inspection.id}>{inspection.temuan} — {inspection.rekomendasi}</li>)}</ul></>}
      {(selected.penilaian?.length ?? 0)>0 && <><h3>Riwayat penilaian</h3><ul>{selected.penilaian?.map((evaluation)=><li key={evaluation.id}>{evaluation.jenis_penilaian.toUpperCase()} · {evaluation.nama_tanaman} / {evaluation.varietas} — {evaluation.rekomendasi} (vigoritas {evaluation.skor_vigoritas}/5, kesehatan {evaluation.skor_kesehatan}/5, varietas {evaluation.skor_kebenaran_varietas}/5). {evaluation.catatan}</li>)}</ul></>}
      <button className="repository-button secondary" type="button" onClick={()=>setSelected(null)}>Tutup detail</button>
    </section>}
    {error && <p className="repository-alert error" role="alert">{error}</p>}{message && <p className="repository-alert success" role="status">{message}</p>}
  </section>;
}

// ubahCrop mengganti satu kolom pada kartu komoditas tanpa mengubah urutan
// (nama kolom input diturunkan dari indeks array).
function ubahCrop(index: number, kolom: keyof CropRow, nilai: string, setCrops: (fn: (current: CropRow[]) => CropRow[]) => void) {
  setCrops((current) => current.map((row, i) => (i === index ? { ...row, [kolom]: nilai } : row)));
}

function ekstensi(nama: string): string {
  const bagian = nama.split(".");
  const akhir = bagian.length > 1 ? bagian.pop() ?? "" : "";
  return akhir.slice(0, 4).toUpperCase() || "FILE";
}

// Pengunggah menampilkan jenis lampiran, tombol unggah, pratinjau nama berkas,
// dan batas format/ukuran dalam satu blok agar tidak perlu menebak.
function Pengunggah({ jenis, onJenis, berkas, onBerkas, busy, onSubmit }: {
  jenis: string;
  onJenis: (value: string) => void;
  berkas: string;
  onBerkas: (value: string) => void;
  busy: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  const id = useId();
  const berkasRef = useRef<HTMLInputElement>(null);

  return <form className="unggah" onSubmit={onSubmit}>
    <div className="unggah-jenis" role="group" aria-labelledby={`${id}-jenis`}>
      <span className="unggah-label" id={`${id}-jenis`}>Jenis lampiran</span>
      <div className="unggah-pil">
        {attachmentTypes.map(([value, label]) => (
          <button
            key={value}
            type="button"
            className={jenis === value ? "pil is-aktif" : "pil"}
            aria-pressed={jenis === value}
            onClick={() => onJenis(value)}
          >
            {label}
          </button>
        ))}
      </div>
    </div>

    <input type="hidden" name="jenis" value={jenis} readOnly />
    <div className="unggah-drop">
      <input
        ref={berkasRef}
        id={`${id}-berkas`}
        className="unggah-input"
        name="file"
        type="file"
        required
        accept=".pdf,.png,.jpg,.jpeg,.zip,.csv,.xls,.xlsx,.doc,.docx,application/pdf,image/png,image/jpeg,application/zip,text/csv"
        onChange={(event) => onBerkas(event.currentTarget.files?.[0]?.name ?? "")}
      />
      <span className="unggah-ikon" aria-hidden="true">📎</span>
      <span className="unggah-teks">
        <strong>{berkas || "Belum ada berkas dipilih"}</strong>
        <small>PDF, gambar, spreadsheet, dokumen, atau ZIP. Ukuran maksimum mengikuti konfigurasi layanan.</small>
      </span>
      <button className="repository-button secondary" type="button" onClick={() => berkasRef.current?.click()}>
        {berkas ? "Ganti berkas" : "Pilih berkas"}
      </button>
      <button className="repository-button" type="submit" disabled={busy || !berkas}>
        {busy ? "Mengunggah…" : "Unggah"}
      </button>
    </div>
  </form>;
}

const gapChecks = ["legalitas_lahan","riwayat_lahan","kondisi_lahan","sumber_air","drainase","konservasi_tanah","benih_bibit","pemupukan","pengendalian_hama","panen_pascapanen","pencatatan_budidaya"];
const gapLabels: Record<string,string> = { legalitas_lahan:"Legalitas dan batas lahan", riwayat_lahan:"Riwayat penggunaan lahan", kondisi_lahan:"Kondisi dan kesesuaian lahan", sumber_air:"Kualitas dan ketersediaan sumber air", drainase:"Drainase dan pengelolaan air", konservasi_tanah:"Konservasi tanah", benih_bibit:"Mutu benih/bibit", pemupukan:"Pemupukan dan penggunaan input", pengendalian_hama:"Pengendalian hama dan penyakit", panen_pascapanen:"Panen dan pascapanen", pencatatan_budidaya:"Pencatatan praktik budidaya" };
