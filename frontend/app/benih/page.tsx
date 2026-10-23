"use client";

import { useCallback, useEffect, useState } from "react";
import { Layout } from "../../components/layout";
import { Card } from "../../components/card";
import { Badge } from "../../components/badge";
import { Button } from "../../components/button";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ||
  process.env.NEXT_PUBLIC_BACKEND_URL ||
  "http://localhost:8080";

interface Registrasi {
  ID: number;
  NomorRegistrasi: string;
  NomorLot: string;
  Komoditas: string;
  Varietas: string;
  Produsen: string;
  KelasBenih: string;
  Jumlah: number;
  Satuan: string;
  TanggalKedaluwarsa: string;
  StatusEdar: string;
  NomorSertifikat: string;
  KodeLabel: string;
}

interface Stok {
  RegistrasiID: number;
  NomorLot: string;
  Komoditas: string;
  Varietas: string;
  JumlahTersedia: number;
  Satuan: string;
  TanggalKedaluwarsa: string;
  StatusEdar: string;
}

interface Mutasi {
  ID: number;
  Jenis: string;
  Jumlah: number;
  SaldoSetelah: number;
  ReferensiTipe: string;
  Catatan: string;
  CreatedAt: string;
}

interface StokDetail {
  registrasi: Registrasi;
  stok: Stok;
  mutasi: Mutasi[] | null;
}

// Pilihan permohonan dari GET /v1/permohonan-siap-registrasi: hanya permohonan
// yang sertifikatnya sudah TTE, labelnya sudah legalisasi dan terkirim ke SIM.
interface PermohonanSiap {
  PermohonanID: number;
  NomorPermohonan: string;
  JenisBenih: string;
  Varietas: string;
  Status: string;
}

interface Penjualan {
  ID: number;
  NomorDokumen: string;
  NomorLot: string;
  Pembeli: string;
  Jumlah: number;
  HargaSatuan: number;
  TotalHarga: number;
  TanggalTransaksi: string;
  Status: string;
  AlasanPembatalan: string;
}

const rupiah = new Intl.NumberFormat("id-ID", {
  style: "currency",
  currency: "IDR",
  maximumFractionDigits: 0,
});

const tanggalHariIni = () => new Date().toISOString().slice(0, 10);

const labelPermohonan = (item: PermohonanSiap) =>
  `${item.NomorPermohonan} — ${item.JenisBenih}${item.Varietas ? ` (${item.Varietas})` : ""}`;

export default function BenihPage() {
  const [registrasi, setRegistrasi] = useState<Registrasi[]>([]);
  const [penjualan, setPenjualan] = useState<Penjualan[]>([]);
  const [stokDetail, setStokDetail] = useState<StokDetail | null>(null);
  const [laporanStok, setLaporanStok] = useState<Stok[] | null>(null);
  const [laporanKedaluwarsa, setLaporanKedaluwarsa] = useState<Registrasi[] | null>(null);
  const [permohonanSiap, setPermohonanSiap] = useState<PermohonanSiap[] | null>(null);
  const [permohonanID, setPermohonanID] = useState("");
  const [pesan, setPesan] = useState("");
  const [galat, setGalat] = useState("");
  const [memuatPermohonan, setMemuatPermohonan] = useState(false);

  const api = useCallback((path: string, opts: RequestInit = {}) => {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(opts.headers as Record<string, string>),
    };
    return fetch(`${API_BASE}${path}`, {
      ...opts,
      credentials: "include",
      headers,
    });
  }, []);

  const setInfo = useCallback((msg: string, err = "") => {
    setPesan(msg);
    setGalat(err);
  }, []);

  const bacaError = useCallback(async (res: Response) => {
    try {
      const data = (await res.json()) as { error?: string };
      return data.error || `HTTP ${res.status}`;
    } catch {
      return `HTTP ${res.status}`;
    }
  }, []);

  // Form registrasi lot memilih permohonan, bukan mengetik ID: ID internal
  // tidak bermakna bagi pengguna dan rawan salah input.
  const muatPermohonanSiap = useCallback(async () => {
    setMemuatPermohonan(true);
    try {
      const res = await api("/v1/permohonan-siap-registrasi");
      if (!res.ok) throw new Error(await bacaError(res));
      const items = (await res.json()) as PermohonanSiap[];
      setPermohonanSiap(items);
      // Pertahankan pilihan sebelumnya selama masih tersedia.
      setPermohonanID((sebelumnya) =>
        items.some((item) => String(item.PermohonanID) === sebelumnya)
          ? sebelumnya
          : items.length === 1
            ? String(items[0].PermohonanID)
            : "",
      );
    } catch (err) {
      setPermohonanSiap([]);
      setInfo("", err instanceof Error ? err.message : String(err));
    } finally {
      setMemuatPermohonan(false);
    }
  }, [api, bacaError, setInfo]);

  useEffect(() => {
    void muatPermohonanSiap();
  }, [muatPermohonanSiap]);

  const muatRegistrasi = useCallback(async () => {
    try {
      const res = await api("/v1/registrasi-benih");
      if (!res.ok) throw new Error(await bacaError(res));
      setRegistrasi(await res.json());
      setInfo("Daftar registrasi dimuat");
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  }, [api, setInfo]);

  const muatPenjualan = useCallback(async () => {
    try {
      const res = await api("/v1/penjualan-benih");
      if (!res.ok) throw new Error(await bacaError(res));
      setPenjualan(await res.json());
      setInfo("Daftar penjualan dimuat");
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  }, [api, setInfo]);

  const buatRegistrasi = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formEl = e.currentTarget;
    if (!permohonanID) {
      setInfo("", "Pilih permohonan yang akan diregistrasi terlebih dahulu");
      return;
    }
    const permohonanTerpilih = permohonanSiap?.find((item) => String(item.PermohonanID) === permohonanID);
    const form = new FormData(formEl);
    const body = JSON.stringify({
      permohonan_id: Number(permohonanID),
      komoditas: form.get("komoditas"),
      varietas: form.get("varietas"),
      produsen: form.get("produsen"),
      sumber_benih: form.get("sumber_benih"),
      nomor_lot: form.get("nomor_lot"),
      kelas_benih: form.get("kelas_benih"),
      jumlah: Number(form.get("jumlah")),
      satuan: form.get("satuan"),
      tanggal_produksi: form.get("tanggal_produksi"),
      tanggal_kedaluwarsa: form.get("tanggal_kedaluwarsa"),
      status_edar: form.get("status_edar"),
    });
    try {
      const res = await api("/v1/registrasi-benih", { method: "POST", body });
      if (!res.ok) throw new Error(await bacaError(res));
      const data = (await res.json()) as Registrasi;
      setInfo(
        `Lot ${data.NomorLot} terdaftar sebagai ${data.NomorRegistrasi}` +
          (permohonanTerpilih ? ` untuk permohonan ${permohonanTerpilih.NomorPermohonan}` : ""),
      );
      formEl.reset();
      setPermohonanID("");
      await Promise.all([muatRegistrasi(), muatPermohonanSiap()]);
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const lihatStok = async (id: number) => {
    try {
      const res = await api(`/v1/registrasi-benih/${id}/stok`);
      if (!res.ok) throw new Error(await bacaError(res));
      setStokDetail(await res.json());
      setInfo("Riwayat stok dimuat");
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const ubahStatus = async (id: number, status: string) => {
    try {
      const res = await api(`/v1/registrasi-benih/${id}/status`, {
        method: "POST",
        body: JSON.stringify({ status_edar: status }),
      });
      if (!res.ok) throw new Error(await bacaError(res));
      setInfo(`Status edar diubah menjadi ${status}`);
      await muatRegistrasi();
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const jualBenih = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const body = JSON.stringify({
      nomor_dokumen: form.get("nomor_dokumen"),
      registrasi_id: Number(form.get("registrasi_id")),
      pembeli: form.get("pembeli"),
      jumlah: Number(form.get("jumlah")),
      harga_satuan: Number(form.get("harga_satuan")),
      tanggal_transaksi: form.get("tanggal_transaksi"),
    });
    try {
      const res = await api("/v1/penjualan-benih", { method: "POST", body });
      if (!res.ok) throw new Error(await bacaError(res));
      const data = (await res.json()) as Penjualan;
      setInfo(`Penjualan ${data.NomorDokumen} tercatat (${rupiah.format(data.TotalHarga)})`);
      await muatPenjualan();
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const batalkanPenjualan = async (id: number) => {
    const alasan = window.prompt("Alasan pembatalan penjualan:");
    if (!alasan) return;
    try {
      const res = await api(`/v1/penjualan-benih/${id}/batal`, {
        method: "POST",
        body: JSON.stringify({ alasan }),
      });
      if (!res.ok) throw new Error(await bacaError(res));
      setInfo("Penjualan dibatalkan dan stok dikembalikan");
      await muatPenjualan();
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const muatLaporanStok = async () => {
    try {
      const res = await api("/v1/laporan/stok-benih");
      if (!res.ok) throw new Error(await bacaError(res));
      setLaporanStok(await res.json());
      setInfo("Laporan stok dimuat");
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const muatLaporanKedaluwarsa = async () => {
    try {
      const res = await api("/v1/laporan/kedaluwarsa-benih?hari=90");
      if (!res.ok) throw new Error(await bacaError(res));
      setLaporanKedaluwarsa(await res.json());
      setInfo("Laporan kedaluwarsa dimuat");
    } catch (err) {
      setInfo("", err instanceof Error ? err.message : String(err));
    }
  };

  const statusBadge = (status: string) => {
    const map: Record<string, string> = {
      aktif: "success",
      ditangguhkan: "warning",
      dicabut: "danger",
      selesai: "success",
    };
    return <Badge variant={(map[status] ?? "info") as any}>{status}</Badge>;
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div className="container">
          <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4 mb-2">
            <div>
              <p className="eyebrow">Layanan benih</p>
              <h1 className="text-2xl font-bold text-sitani-green-800">
                Registrasi &amp; Penjualan Benih
              </h1>
              <p className="text-sitani-muted text-sm">
                Registrasi lot, persediaan per lot, penjualan, pembatalan dengan jejak audit, dan laporan.
                Identitas memakai sesi SSO yang terverifikasi.
              </p>
            </div>
          </div>
        </div>

        {galat && (
          <div className="container alert alert-error" role="alert">
            Galat: {galat}
          </div>
        )}
        {pesan && !galat && (
          <div className="container alert alert-success" role="status">
            {pesan}
          </div>
        )}

        <div className="container">
          <section>
            <h2 className="text-xl font-bold text-sitani-green-800 mb-3">Registrasi Benih</h2>
            <p className="text-sitani-muted text-sm mb-4">
              Pilih permohonan yang sertifikatnya sudah ditandatangani TTE serta labelnya sudah
              dilegalisasi dan terkirim ke SIM Perbenihan. Nomor permohonan dibuat sistem saat
              permohonan diajukan, bukan diketik di sini.
            </p>
            <Card>
              <form onSubmit={buatRegistrasi} className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                <div className="form-group sm:col-span-2 lg:col-span-1">
                  <label className="form-label" htmlFor="permohonan-select">
                    Permohonan
                  </label>
                  <select
                    id="permohonan-select"
                    name="permohonan"
                    className="form-input"
                    value={permohonanID}
                    onChange={(e) => setPermohonanID(e.target.value)}
                    disabled={memuatPermohonan || (permohonanSiap?.length ?? 0) === 0}
                    required
                  >
                    <option value="">
                      {memuatPermohonan
                        ? "Memuat permohonan…"
                        : (permohonanSiap?.length ?? 0) === 0
                          ? "Belum ada permohonan siap registrasi"
                          : "Pilih permohonan"}
                    </option>
                    {(permohonanSiap ?? []).map((item) => (
                      <option key={item.PermohonanID} value={String(item.PermohonanID)}>
                        {labelPermohonan(item)}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">Komoditas</label>
                  <input name="komoditas" className="form-input" placeholder="Komoditas" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Varietas</label>
                  <input name="varietas" className="form-input" placeholder="Varietas" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Produsen</label>
                  <input name="produsen" className="form-input" placeholder="Produsen" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Sumber benih</label>
                  <input name="sumber_benih" className="form-input" placeholder="Sumber benih" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Nomor lot</label>
                  <input name="nomor_lot" className="form-input" placeholder="Nomor lot" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Kelas benih</label>
                  <select name="kelas_benih" defaultValue="sebar" className="form-input" required>
                    <option value="penjenis">Penjenis</option>
                    <option value="dasar">Dasar</option>
                    <option value="pokok">Pokok</option>
                    <option value="sebar">Sebar</option>
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">Jumlah</label>
                  <input
                    name="jumlah"
                    type="number"
                    min="0.001"
                    step="0.001"
                    className="form-input"
                    placeholder="Jumlah"
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Satuan</label>
                  <input name="satuan" className="form-input" placeholder="kg" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Tanggal produksi</label>
                  <input name="tanggal_produksi" type="date" className="form-input" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Tanggal kedaluwarsa</label>
                  <input name="tanggal_kedaluwarsa" type="date" className="form-input" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Status edar</label>
                  <select name="status_edar" defaultValue="aktif" className="form-input" required>
                    <option value="aktif">Aktif</option>
                    <option value="ditangguhkan">Ditangguhkan</option>
                    <option value="dicabut">Dicabut</option>
                  </select>
                </div>
                <div className="form-group sm:col-span-2 lg:col-span-3">
                  <Button type="submit" variant="primary" className="w-auto">
                    Daftarkan Lot
                  </Button>
                </div>
              </form>
            </Card>

            <div className="mt-4 flex justify-between items-center">
              <div className="flex gap-3">
                <Button variant="outline" size="sm" onClick={muatRegistrasi}>
                  Muat Registrasi
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={muatPermohonanSiap}
                  disabled={memuatPermohonan}
                >
                  Muat Permohonan
                </Button>
              </div>
            </div>

            {registrasi.length > 0 && (
              <div className="mt-6">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Registrasi</th>
                      <th>Lot</th>
                      <th>Komoditas</th>
                      <th>Kelas</th>
                      <th>Jumlah</th>
                      <th>Kedaluwarsa</th>
                      <th>Status</th>
                      <th>Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {registrasi.map((item) => (
                      <tr key={item.ID}>
                        <td>{item.NomorRegistrasi}</td>
                        <td>{item.NomorLot}</td>
                        <td>{item.Komoditas} / {item.Varietas}</td>
                        <td>{item.KelasBenih}</td>
                        <td>{item.Jumlah} {item.Satuan}</td>
                        <td>{item.TanggalKedaluwarsa}</td>
                        <td>{statusBadge(item.StatusEdar)}</td>
                        <td style={{ whiteSpace: "nowrap" }}>
                          <Button size="sm" variant="outline" onClick={() => lihatStok(item.ID)}>
                            Stok
                          </Button>{" "}
                          {item.StatusEdar === "aktif" ? (
                            <Button size="sm" variant="secondary" onClick={() => ubahStatus(item.ID, "ditangguhkan")}>
                              Tangguhkan
                            </Button>
                          ) : (
                            <Button size="sm" variant="secondary" onClick={() => ubahStatus(item.ID, "aktif")}>
                              Aktifkan
                            </Button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {stokDetail && (
            <section className="mt-8">
              <Card>
                <h2 className="text-xl font-bold text-sitani-green-800 mb-2">
                  Stok Lot {stokDetail.stok.NomorLot}
                </h2>
                <p className="text-sitani-muted text-sm mb-4">
                  Tersedia:{" "}
                  <strong>{stokDetail.stok.JumlahTersedia} {stokDetail.stok.Satuan}</strong>{" "}
                  | Status:{stokDetail.stok.StatusEdar} | Kedaluwarsa:{" "}
                  {stokDetail.stok.TanggalKedaluwarsa}
                </p>
                <table className="table">
                  <thead>
                    <tr>
                      <th>Waktu</th>
                      <th>Jenis</th>
                      <th>Jumlah</th>
                      <th>Saldo</th>
                      <th>Referensi</th>
                      <th>Catatan</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(stokDetail.mutasi || []).map((mutasi) => (
                      <tr key={mutasi.ID}>
                        <td>{mutasi.CreatedAt}</td>
                        <td>{mutasi.Jenis}</td>
                        <td>{mutasi.Jumlah}</td>
                        <td>{mutasi.SaldoSetelah}</td>
                        <td>{mutasi.ReferensiTipe}</td>
                        <td>{mutasi.Catatan}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </Card>
            </section>
          )}

          <section className="mt-8">
            <h2 className="text-xl font-bold text-sitani-green-800 mb-3">Penjualan Benih</h2>
            <p className="text-sitani-muted text-sm mb-4">
              Penjualan hanya diterima untuk lot aktif, belum kedaluwarsa, serta memiliki sertifikat
              dan label yang sah. Stok dikurangi dalam transaksi database.
            </p>
            <Card>
              <form onSubmit={jualBenih} className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                <div className="form-group">
                  <label className="form-label">Nomor dokumen</label>
                  <input name="nomor_dokumen" className="form-input" placeholder="Nomor dokumen" required />
                </div>
                <div className="form-group">
                  <label className="form-label">ID registrasi</label>
                  <input
                    name="registrasi_id"
                    type="number"
                    min="1"
                    className="form-input"
                    placeholder="ID registrasi"
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Pembeli</label>
                  <input name="pembeli" className="form-input" placeholder="Pembeli" required />
                </div>
                <div className="form-group">
                  <label className="form-label">Jumlah</label>
                  <input
                    name="jumlah"
                    type="number"
                    min="0.001"
                    step="0.001"
                    className="form-input"
                    placeholder="Jumlah"
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Harga satuan</label>
                  <input
                    name="harga_satuan"
                    type="number"
                    min="0"
                    step="0.01"
                    className="form-input"
                    placeholder="Harga satuan"
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Tanggal transaksi</label>
                  <input
                    name="tanggal_transaksi"
                    type="date"
                    defaultValue={tanggalHariIni()}
                    className="form-input"
                    required
                  />
                </div>
                <div className="form-group sm:col-span-2 lg:col-span-3">
                  <Button type="submit" variant="primary" className="w-auto">
                    Catat Penjualan
                  </Button>
                </div>
              </form>
            </Card>

            <div className="mt-4 flex justify-between items-center">
              <Button variant="outline" size="sm" onClick={muatPenjualan}>
                Muat Penjualan
              </Button>
            </div>

            {penjualan.length > 0 && (
              <div className="mt-6">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Dokumen</th>
                      <th>Lot</th>
                      <th>Pembeli</th>
                      <th>Jumlah</th>
                      <th>Total</th>
                      <th>Tanggal</th>
                      <th>Status</th>
                      <th>Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {penjualan.map((item) => (
                      <tr key={item.ID}>
                        <td>{item.NomorDokumen}</td>
                        <td>{item.NomorLot}</td>
                        <td>{item.Pembeli}</td>
                        <td>{item.Jumlah}</td>
                        <td>{rupiah.format(item.TotalHarga)}</td>
                        <td>{item.TanggalTransaksi}</td>
                        <td>{statusBadge(item.Status)}</td>
                        <td style={{ whiteSpace: "nowrap" }}>
                          {item.Status === "selesai" ? (
                            <Button size="sm" variant="secondary" onClick={() => batalkanPenjualan(item.ID)}>
                              Batalkan
                            </Button>
                          ) : (
                            item.AlasanPembatalan
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="mt-8">
            <h2 className="text-xl font-bold text-sitani-green-800 mb-3">Laporan</h2>
            <div className="flex flex-wrap gap-3 mb-4">
              <Button variant="outline" size="sm" onClick={muatLaporanStok}>
                Laporan Stok
              </Button>
              <Button variant="outline" size="sm" onClick={muatLaporanKedaluwarsa}>
                Kedaluwarsa ≤ 90 Hari
              </Button>
              <Button variant="outline" size="sm" onClick={muatPenjualan}>
                Laporan Penjualan
              </Button>
            </div>

            {laporanStok && (
              <>
                <h3 className="text-sm font-bold text-sitani-green-700 mb-2">Stok Seluruh Lot</h3>
                <table className="table">
                  <thead>
                    <tr>
                      <th>Lot</th>
                      <th>Komoditas</th>
                      <th>Tersedia</th>
                      <th>Kedaluwarsa</th>
                      <th>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {laporanStok.map((item) => (
                      <tr key={item.RegistrasiID}>
                        <td>{item.NomorLot}</td>
                        <td>{item.Komoditas} / {item.Varietas}</td>
                        <td>{item.JumlahTersedia} {item.Satuan}</td>
                        <td>{item.TanggalKedaluwarsa}</td>
                        <td>{statusBadge(item.StatusEdar)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            )}

            {laporanKedaluwarsa && (
              <div className="mt-6">
                <h3 className="text-sm font-bold text-sitani-green-700 mb-2">
                  Lot Mendekati Kedaluwarsa
                </h3>
                {laporanKedaluwarsa.length === 0 ? (
                  <p className="text-sitani-muted text-sm">Tidak ada lot dalam rentang 90 hari.</p>
                ) : (
                  <table className="table">
                    <thead>
                      <tr>
                        <th>Lot</th>
                        <th>Komoditas</th>
                        <th>Kedaluwarsa</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {laporanKedaluwarsa.map((item) => (
                        <tr key={item.ID}>
                          <td>{item.NomorLot}</td>
                          <td>{item.Komoditas} / {item.Varietas}</td>
                          <td>{item.TanggalKedaluwarsa}</td>
                          <td>{statusBadge(item.StatusEdar)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            )}
          </section>
        </div>
      </div>
    </Layout>
  );
}
