"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";

import { masukHref } from "../../components/auth/masuk-url";

const apiURL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

type Principal = { name?: string; email: string; role: string };
type ImportItem = {
  id: number;
  nama_dataset: string;
  nama_file: string;
  jumlah_baris: number;
  diunggah_oleh: number;
  dibuat_pada: string;
};
type Detail = { import: ImportItem; headers: string[]; rows: Record<string, string>[] };

async function errorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string };
    return body.error ?? `Permintaan gagal (${response.status})`;
  } catch {
    return `Permintaan gagal (${response.status})`;
  }
}

export function RepositoryDataPanel() {
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [imports, setImports] = useState<ImportItem[]>([]);
  const [detail, setDetail] = useState<Detail | null>(null);
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  const loadImports = useCallback(async () => {
    const response = await fetch(`${apiURL}/v1/dataset/import`, { credentials: "include" });
    if (!response.ok) throw new Error(await errorMessage(response));
    const body = (await response.json()) as { imports: ImportItem[] };
    setImports(body.imports ?? []);
  }, []);

  useEffect(() => {
    let active = true;
    async function initialize() {
      try {
        const response = await fetch(`${apiURL}/auth/me`, { credentials: "include" });
        if (!response.ok) {
          if (active) setPrincipal(null);
          return;
        }
        const user = (await response.json()) as Principal;
        if (!active) return;
        setPrincipal(user);
        if (user.role === "admin" || user.role === "analis") await loadImports();
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : "Gagal memuat sesi atau daftar import.");
      } finally {
        if (active) setLoading(false);
      }
    }
    void initialize();
    return () => { active = false; };
  }, [loadImports]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!file) {
      setError("Pilih file CSV atau XLSX terlebih dahulu.");
      return;
    }
    setBusy(true);
    setError("");
    setMessage("");
    const form = new FormData();
    form.set("nama_dataset", name);
    form.set("file", file);
    try {
      const response = await fetch(`${apiURL}/v1/dataset/import`, {
        method: "POST",
        credentials: "include",
        body: form,
      });
      if (!response.ok) throw new Error(await errorMessage(response));
      const saved = (await response.json()) as ImportItem;
      setMessage(`Import #${saved.id} tersimpan — ${saved.jumlah_baris} baris.`);
      setName("");
      setFile(null);
      const input = document.getElementById("dataset-file") as HTMLInputElement | null;
      if (input) input.value = "";
      await loadImports();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Import gagal.");
    } finally {
      setBusy(false);
    }
  }

  async function openDetail(item: ImportItem) {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const response = await fetch(`${apiURL}/v1/dataset/import/${item.id}`, { credentials: "include" });
      if (!response.ok) throw new Error(await errorMessage(response));
      setDetail((await response.json()) as Detail);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Gagal memuat detail batch.");
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <section className="repository-content"><p role="status">Memeriksa sesi SSO…</p></section>;
  if (!principal) {
    const redirect = "/repository-data";
    return (
      <section className="repository-content">
        <h2>Masuk diperlukan</h2>
        <p>Silakan masuk dengan akun internal yang memiliki role admin atau analis.</p>
        <div className="repository-actions">
          <a href={masukHref(redirect, "satriya")}>Masuk dengan Satriya</a>
          <a href={masukHref(redirect, "jsp")}>Masuk dengan JSP</a>
        </div>
      </section>
    );
  }
  if (principal.role !== "admin" && principal.role !== "analis") {
    return (
      <section className="repository-content">
        <h2>Akses terbatas</h2>
        <p>Akun {principal.email} memiliki role {principal.role}. Import repository tersedia untuk admin dan analis.</p>
      </section>
    );
  }

  return (
    <section className="repository-content">
      <div className="repository-session">
        <span>Masuk sebagai <strong>{principal.name || principal.email}</strong> · {principal.role}</span>
        <button type="button" className="repository-button secondary" onClick={() => void loadImports()}>Muat ulang</button>
      </div>

      <div className="repository-grid">
        <section className="repository-card" aria-labelledby="upload-title">
          <p className="eyebrow">Impor data</p>
          <h2 id="upload-title">Unggah batch</h2>
          <p>CSV UTF-8 atau XLSX (sheet pertama). Maksimal 10 MiB dan 10.000 baris data.</p>
          <form className="repository-form" onSubmit={submit}>
            <label htmlFor="dataset-name">Nama dataset</label>
            <input id="dataset-name" required maxLength={120} value={name} onChange={(event) => setName(event.target.value)} placeholder="Contoh: Produksi pangan 2026" />
            <label htmlFor="dataset-file">File sumber</label>
            <input id="dataset-file" required type="file" accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" onChange={(event) => setFile(event.target.files?.[0] ?? null)} />
            <button className="repository-button" type="submit" disabled={busy}>{busy ? "Memproses…" : "Impor dataset"}</button>
          </form>
          <p className="repository-hint">Header duplikat/kosong dan data dengan kolom berlebih akan ditolak. Isi kolom disimpan sebagai nilai sumber tanpa normalisasi.</p>
        </section>

        <section className="repository-card" aria-labelledby="batches-title">
          <p className="eyebrow">Riwayat</p>
          <h2 id="batches-title">Batch import</h2>
          {imports.length === 0 ? <p>Belum ada batch import untuk akun ini.</p> : (
            <div className="repository-table-wrap">
              <table className="repository-table">
                <thead><tr><th>Dataset</th><th>File</th><th>Baris</th><th>Aksi</th></tr></thead>
                <tbody>{imports.map((item) => (
                  <tr key={item.id}>
                    <td>{item.nama_dataset}</td><td>{item.nama_file}</td><td>{item.jumlah_baris}</td>
                    <td><button className="repository-link-button" type="button" disabled={busy} onClick={() => void openDetail(item)}>Lihat data</button></td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          )}
        </section>
      </div>

      {error && <p className="repository-alert error" role="alert">{error}</p>}
      {message && <p className="repository-alert success" role="status">{message}</p>}

      {detail && (
        <section className="repository-card repository-detail" aria-labelledby="detail-title">
          <div className="repository-detail-heading">
            <div><p className="eyebrow">Detail batch #{detail.import.id}</p><h2 id="detail-title">{detail.import.nama_dataset}</h2></div>
            <button type="button" className="repository-button secondary" onClick={() => setDetail(null)}>Tutup</button>
          </div>
          <p>{detail.import.nama_file} · {detail.rows.length} baris</p>
          {detail.rows.length === 0 ? <p>Batch tidak memiliki baris data.</p> : (
            <div className="repository-table-wrap">
              <table className="repository-table">
                <thead><tr>{detail.headers.map((header) => <th key={header}>{header}</th>)}</tr></thead>
                <tbody>{detail.rows.map((row, index) => <tr key={`${detail.import.id}-${index}`}>
                  {detail.headers.map((header) => <td key={header}>{row[header] ?? ""}</td>)}
                </tr>)}</tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </section>
  );
}
