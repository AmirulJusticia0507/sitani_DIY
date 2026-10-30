"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";

import { masukHref } from "../../components/auth/masuk-url";

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";
type Principal = { name?: string; email: string; role: string };
type Application = { id: number; nomor_permohonan: string; nama_organisasi: string; status: string; dibuat_pada: string; jadwal_inspeksi?: { id: number; waktu: string; lokasi: string; status: string }[]; inspeksi?: { id: number; temuan: string; rekomendasi: string }[]; penilaian?: { id: number; jenis_penilaian: string; rekomendasi: string; nama_tanaman: string; varietas: string; skor_vigoritas:number; skor_kesehatan:number; skor_kebenaran_varietas:number; catatan:string }[] };
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
  const [selected, setSelected] = useState<Application | null>(null);
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
        if (["pelanggan", "admin", "pengawas", "verifikator", "inspektor", "pbt", "kasi_pmbtp"].includes(user.role)) await loadApplications();
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
    const redirect = "/horticulture";
    return <section className="repository-content"><div className="repository-card"><h2>Masuk diperlukan</h2><p>Masuk dengan akun yang telah terverifikasi melalui SSO.</p><div className="repository-actions"><a href={masukHref(redirect, "jsp")}>Masuk dengan JSP</a><a href={masukHref(redirect, "satriya")}>Masuk dengan Satriya</a></div></div></section>;
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
    {(principal.role === "pelanggan" || principal.role === "admin" || principal.role === "pengawas" || principal.role === "verifikator" || principal.role === "inspektor" || principal.role === "pbt" || principal.role === "kasi_pmbtp") && <section className="repository-card horticulture-history"><h2>Riwayat permohonan</h2>
      {applications.length === 0 ? <p>Belum ada permohonan.</p> : <div className="repository-table-wrap"><table className="repository-table"><thead><tr><th>Nomor</th><th>Organisasi</th><th>Status</th><th>Tanggal</th><th>Aksi</th></tr></thead><tbody>{applications.map((item) => <tr key={item.id}><td>{item.nomor_permohonan}</td><td>{item.nama_organisasi}</td><td>{item.status}</td><td>{new Date(item.dibuat_pada).toLocaleDateString("id-ID")}</td><td><button className="repository-link-button" type="button" onClick={() => void openApplication(item)}>Buka detail</button></td></tr>)}</tbody></table></div>}
    </section>}
    {selected && <section className="repository-card horticulture-history"><h2>Alur permohonan {selected.nomor_permohonan}</h2><p>Status: <strong>{selected.status}</strong></p>
      {principal.role === "verifikator" || principal.role === "admin" ? <div className="horticulture-actions"><button className="repository-button" disabled={busy} onClick={() => void action(`/v1/horticulture/applications/${selected.id}/verification`,"POST",{keputusan:"diverifikasi"},"Proposal dinyatakan lengkap dan terverifikasi.")}>Verifikasi lengkap</button><button className="repository-button secondary" disabled={busy} onClick={() => { const catatan=window.prompt("Catatan dokumen yang harus diperbaiki:"); if(catatan?.trim()) void action(`/v1/horticulture/applications/${selected.id}/verification`,"POST",{keputusan:"perlu_perbaikan",catatan},"Permintaan perbaikan dicatat."); }}>Minta perbaikan</button></div> : null}
      {principal.role === "inspektor" || principal.role === "admin" ? <>
        <h3>Jadwalkan inspeksi lapangan</h3><form className="horticulture-inline" onSubmit={(event) => { event.preventDefault(); const v=new FormData(event.currentTarget); void action(`/v1/horticulture/applications/${selected.id}/inspection-schedules`,"POST",{waktu:new Date(String(v.get("waktu"))).toISOString(),lokasi:v.get("lokasi"),catatan:v.get("catatan")},"Jadwal inspeksi tersimpan."); }}><label>Waktu<input required name="waktu" type="datetime-local" /></label><label>Lokasi<input required name="lokasi" maxLength={255} /></label><label>Catatan<input name="catatan" /></label><button className="repository-button" disabled={busy}>Simpan jadwal</button></form>
        {selected.jadwal_inspeksi?.filter((schedule) => schedule.status === "terjadwal").map((schedule) => <form className="horticulture-inline" key={schedule.id} onSubmit={(event) => { event.preventDefault(); const v=new FormData(event.currentTarget); const checks=Object.fromEntries(gapChecks.map((key)=>[key,v.get(key)==="on"])); void action(`/v1/horticulture/inspection-schedules/${schedule.id}/complete`,"POST",{checklist_gap:checks,temuan:v.get("temuan"),rekomendasi:v.get("rekomendasi")},"Form audit GAP disimpan."); }}><h3>Audit GAP · {new Date(schedule.waktu).toLocaleString("id-ID")} · {schedule.lokasi}</h3><fieldset><legend>Checklist GAP (centang bila memenuhi)</legend>{gapChecks.map((key)=><label className="horticulture-check" key={key}><input type="checkbox" name={key}/>{gapLabels[key]}</label>)}</fieldset><label>Temuan<textarea name="temuan" required rows={3}/></label><label>Rekomendasi<textarea name="rekomendasi" required rows={3}/></label><button className="repository-button" disabled={busy}>Kirim hasil inspeksi</button></form>)}
      </> : null}
      {principal.role === "pbt" || principal.role === "kasi_pmbtp" || principal.role === "admin" ? <><h3>Penilaian pohon induk / PIT / RIP / PI / RI</h3><form className="horticulture-inline" onSubmit={(event)=>{event.preventDefault();const v=new FormData(event.currentTarget);void action(`/v1/horticulture/applications/${selected.id}/evaluations`,"POST",{jenis_penilaian:v.get("jenis"),nama_tanaman:v.get("tanaman"),varietas:v.get("varietas"),skor_vigoritas:Number(v.get("vigoritas")),skor_kesehatan:Number(v.get("kesehatan")),skor_kebenaran_varietas:Number(v.get("kebenaran")),rekomendasi:v.get("rekomendasi"),catatan:v.get("catatan")},"Penilaian tersimpan dalam audit trail.");}}><label>Jenis penilaian<select name="jenis">{["pohon_induk","pit","rip","pi","ri"].map((value)=><option key={value} value={value}>{value.toUpperCase()}</option>)}</select></label><label>Nama tanaman<input name="tanaman" required/></label><label>Varietas<input name="varietas" required/></label>{[["vigoritas","Vigoritas"],["kesehatan","Kesehatan"],["kebenaran","Kebenaran varietas"]].map(([name,label])=><label key={name}>{label} (1–5)<input type="number" name={name} min="1" max="5" required/></label>)}<label>Rekomendasi<select name="rekomendasi"><option value="layak">Layak</option><option value="perlu_perbaikan">Perlu perbaikan</option><option value="tidak_layak">Tidak layak</option></select></label><label>Catatan penilaian<textarea name="catatan" required rows={3}/></label><button className="repository-button" disabled={busy}>Simpan penilaian</button></form></> : null}
      {(selected.inspeksi?.length ?? 0)>0 && <><h3>Riwayat inspeksi GAP</h3><ul>{selected.inspeksi?.map((inspection)=><li key={inspection.id}>{inspection.temuan} — {inspection.rekomendasi}</li>)}</ul></>}
      {(selected.penilaian?.length ?? 0)>0 && <><h3>Riwayat penilaian</h3><ul>{selected.penilaian?.map((evaluation)=><li key={evaluation.id}>{evaluation.jenis_penilaian.toUpperCase()} · {evaluation.nama_tanaman} / {evaluation.varietas} — {evaluation.rekomendasi} (vigoritas {evaluation.skor_vigoritas}/5, kesehatan {evaluation.skor_kesehatan}/5, varietas {evaluation.skor_kebenaran_varietas}/5). {evaluation.catatan}</li>)}</ul></>}
      <button className="repository-button secondary" type="button" onClick={()=>setSelected(null)}>Tutup detail</button>
    </section>}
    {error && <p className="repository-alert error" role="alert">{error}</p>}{message && <p className="repository-alert success" role="status">{message}</p>}
  </section>;
}

const gapChecks = ["legalitas_lahan","riwayat_lahan","kondisi_lahan","sumber_air","drainase","konservasi_tanah","benih_bibit","pemupukan","pengendalian_hama","panen_pascapanen","pencatatan_budidaya"];
const gapLabels: Record<string,string> = { legalitas_lahan:"Legalitas dan batas lahan", riwayat_lahan:"Riwayat penggunaan lahan", kondisi_lahan:"Kondisi dan kesesuaian lahan", sumber_air:"Kualitas dan ketersediaan sumber air", drainase:"Drainase dan pengelolaan air", konservasi_tanah:"Konservasi tanah", benih_bibit:"Mutu benih/bibit", pemupukan:"Pemupukan dan penggunaan input", pengendalian_hama:"Pengendalian hama dan penyakit", panen_pascapanen:"Panen dan pascapanen", pencatatan_budidaya:"Pencatatan praktik budidaya" };
