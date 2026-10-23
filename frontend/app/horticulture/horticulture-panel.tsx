"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";
type Principal = { name?: string; email: string; role: string };
type Application = { id: number; nomor_permohonan: string; nama_organisasi: string; status: string; dibuat_pada: string };
type Upload = { id: number; nama: string };
type Attachment = { id: number; jenis: string; nama: string };
const attachmentTypes = [
  ["proposal", "Proposal permohonan"],
  ["organization", "Struktur organisasi"],
  ["members", "Daftar anggota"],
  ["farm_form", "Form registrasi kebun"],
  ["farm_map", "Peta kebun"],
  ["other", "Dokumen pendukung lain"],
] as const;

async function errorMessage(response: Response): Promise<string> {
  try { return ((await response.json()) as { error?: string }).error ?? `Permintaan gagal (${response.status})`; }
  catch { return `Permintaan gagal (${response.status})`; }
}

export function HorticulturePanel() {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [applications, setApplications] = useState<Application[]>([]);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

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
        if (["pelanggan", "admin", "pengawas"].includes(user.role)) await loadApplications();
      } catch (cause) { if (active) setError(cause instanceof Error ? cause.message : "Gagal memuat data."); }
      finally { if (active) setLoading(false); }
    }
    void initialize();
    return () => { active = false; };
  }, [loadApplications]);

  async function uploadFiles(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    const file = form.get("file");
    if (!(file instanceof File) || file.size === 0) { setError("Pilih berkas terlebih dahulu."); return; }
    setBusy(true); setError(""); setMessage("");
    const upload = new FormData();
    upload.set("file", file);
    upload.set("tujuan", "permohonan_registrasi_hortikultura");
    try {
      const response = await fetch(`${apiURL}/v1/dokumen`, { method: "POST", credentials: "include", body: upload });
      if (!response.ok) throw new Error(await errorMessage(response));
      const saved = (await response.json()) as Upload;
      setAttachments((current) => [...current, { id: saved.id, nama: saved.nama, jenis: String(form.get("jenis")) }]);
      formElement.reset();
      setMessage(`Lampiran “${saved.nama}” berhasil diunggah secara privat.`);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Unggah lampiran gagal."); }
    finally { setBusy(false); }
  }

  async function submitApplication(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formElement = event.currentTarget;
    if (attachments.length === 0) { setError("Unggah minimal satu lampiran sebelum mengirim permohonan."); return; }
    const values = new FormData(formElement);
    const tanaman = [0, 1, 2].flatMap((index) => {
      const nama = String(values.get(`tanaman_${index}_nama`) ?? "").trim();
      const varietas = String(values.get(`tanaman_${index}_varietas`) ?? "").trim();
      if (!nama && !varietas) return [];
      return [{ nama, varietas, jumlah_tanaman: Number(values.get(`tanaman_${index}_jumlah`) || 0), sumber_benih: String(values.get(`tanaman_${index}_sumber`) ?? "").trim() }];
    });
    setBusy(true); setError(""); setMessage("");
    try {
      const response = await fetch(`${apiURL}/v1/horticulture/applications`, {
        method: "POST", credentials: "include", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nama_organisasi: values.get("organisasi"), nama_kontak: values.get("kontak"), telepon: values.get("telepon"),
          alamat: values.get("alamat"), kalurahan: values.get("kalurahan"), kapanewon: values.get("kapanewon"),
          kabupaten_kota: values.get("kabupaten"), luas_lahan_ha: Number(values.get("luas")), tanaman,
          dokumen: attachments.map(({ id, jenis }) => ({ id, jenis })),
        }),
      });
      if (!response.ok) throw new Error(await errorMessage(response));
      const saved = (await response.json()) as Application;
      setMessage(`Permohonan ${saved.nomor_permohonan} berhasil diajukan.`);
      setAttachments([]);
      formElement.reset();
      await loadApplications();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Pengajuan gagal."); }
    finally { setBusy(false); }
  }

  if (loading) return <section className="repository-content"><p role="status">Memeriksa sesi SSO…</p></section>;
  if (!principal) {
    const redirect = encodeURIComponent("/horticulture");
    return <section className="repository-content"><div className="repository-card"><h2>Masuk diperlukan</h2><p>Masuk dengan akun yang telah terverifikasi melalui SSO.</p><div className="repository-actions"><a href={`${apiURL}/auth/login/jsp?redirect=${redirect}`}>Masuk dengan JSP</a><a href={`${apiURL}/auth/login/satriya?redirect=${redirect}`}>Masuk dengan Satriya</a></div></div></section>;
  }

  return <section className="repository-content">
    <p className="repository-session">Masuk sebagai <strong>{principal.name || principal.email}</strong> · {principal.role}</p>
    {principal.role === "pelanggan" || principal.role === "admin" ? <>
      <section className="repository-card" aria-labelledby="form-title">
        <p className="eyebrow">Pengajuan baru</p><h2 id="form-title">Data pemohon dan lahan</h2>
        <p>Nomor registrasi akan diberikan setelah permohonan tercatat. Pengajuan ini belum merupakan persetujuan registrasi.</p>
        <form className="horticulture-upload" onSubmit={uploadFiles}>
          <fieldset><legend>Lampiran dokumen privat</legend>
            <p>Unggah proposal, struktur organisasi, daftar anggota, form registrasi kebun, dan peta kebun. Format PDF/gambar/spreadsheet/dokumen/ZIP; ukuran maksimum mengikuti konfigurasi layanan.</p>
            <label>Jenis lampiran<select name="jenis" defaultValue="proposal">{attachmentTypes.map(([value,label]) => <option key={value} value={value}>{label}</option>)}</select></label>
            <label>Berkas<input name="file" type="file" required accept=".pdf,.png,.jpg,.jpeg,.zip,.csv,.xls,.xlsx,.doc,.docx,application/pdf,image/png,image/jpeg,application/zip,text/csv" /></label>
            <button className="repository-button secondary" type="submit" disabled={busy}>Unggah lampiran</button>
            {attachments.length > 0 && <ul>{attachments.map((item) => <li key={item.id}>{item.nama} — {attachmentTypes.find(([value]) => value === item.jenis)?.[1] ?? item.jenis} <button className="repository-link-button" type="button" onClick={() => setAttachments((current) => current.filter((doc) => doc.id !== item.id))}>hapus dari pengajuan</button></li>)}</ul>}
          </fieldset>
        </form>
        <form className="horticulture-form" onSubmit={submitApplication}>
          <label>Nama kelompok tani / organisasi<input name="organisasi" required maxLength={200} /></label>
          <label>Nama penanggung jawab<input name="kontak" required maxLength={150} /></label>
          <label>Nomor telepon<input name="telepon" required type="tel" maxLength={40} /></label>
          <label>Alamat lengkap<textarea name="alamat" required maxLength={500} rows={3} /></label>
          <label>Kalurahan<input name="kalurahan" required maxLength={120} /></label>
          <label>Kapanewon<input name="kapanewon" required maxLength={120} /></label>
          <label>Kabupaten / Kota<input name="kabupaten" required maxLength={120} /></label>
          <label>Luas lahan (ha)<input name="luas" required type="number" min="0.0001" max="100000" step="0.0001" /></label>
          <fieldset><legend>Detail tanaman (minimal satu komoditas)</legend>{[0, 1, 2].map((index) => <div className="horticulture-crop" key={index}>
            <label>Komoditas {index + 1}<input name={`tanaman_${index}_nama`} required={index === 0} maxLength={120} /></label>
            <label>Varietas<input name={`tanaman_${index}_varietas`} required={index === 0} maxLength={120} /></label>
            <label>Jumlah tanaman<input name={`tanaman_${index}_jumlah`} type="number" min="0" max="100000000" step="1" /></label>
            <label>Sumber benih<input name={`tanaman_${index}_sumber`} maxLength={200} /></label>
          </div>)}</fieldset>
          {error && <p className="repository-alert error" role="alert">{error}</p>}{message && <p className="repository-alert success" role="status">{message}</p>}
          <button className="repository-button" type="submit" disabled={busy}>{busy ? "Menyimpan…" : "Kirim permohonan"}</button>
        </form>
      </section>
    </> : <section className="repository-card"><h2>Akses pemohon</h2><p>Pengajuan baru tersedia bagi pemohon atau admin.</p></section>}
    {(principal.role === "pelanggan" || principal.role === "admin" || principal.role === "pengawas") && <section className="repository-card horticulture-history"><h2>Riwayat permohonan</h2>
      {applications.length === 0 ? <p>Belum ada permohonan.</p> : <div className="repository-table-wrap"><table className="repository-table"><thead><tr><th>Nomor</th><th>Organisasi</th><th>Status</th><th>Tanggal</th></tr></thead><tbody>{applications.map((item) => <tr key={item.id}><td>{item.nomor_permohonan}</td><td>{item.nama_organisasi}</td><td>{item.status}</td><td>{new Date(item.dibuat_pada).toLocaleDateString("id-ID")}</td></tr>)}</tbody></table></div>}
    </section>}
  </section>;
}
