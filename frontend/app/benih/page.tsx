"use client";

import { useCallback, useState } from "react";

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

export default function BenihPage() {
  const [registrasi, setRegistrasi] = useState<Registrasi[]>([]);
  const [penjualan, setPenjualan] = useState<Penjualan[]>([]);
  const [stokDetail, setStokDetail] = useState<StokDetail | null>(null);
  const [laporanStok, setLaporanStok] = useState<Stok[] | null>(null);
  const [laporanKedaluwarsa, setLaporanKedaluwarsa] = useState<Registrasi[] | null>(null);
  const [pesan, setPesan] = useState("");
  const [galat, setGalat] = useState("");

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

  const bacaError = async (res: Response) => {
    try {
      const data = (await res.json()) as { error?: string };
      return data.error || `HTTP ${res.status}`;
    } catch {
      return `HTTP ${res.status}`;
    }
  };

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
    const form = new FormData(e.currentTarget);
    const body = JSON.stringify({
      permohonan_id: Number(form.get("permohonan_id")),
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
      setInfo(`Lot ${data.NomorLot} terdaftar sebagai ${data.NomorRegistrasi}`);
      await muatRegistrasi();
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

  return (
    <main style={{ padding: 24, fontFamily: "system-ui, sans-serif", maxWidth: 1100 }}>
      <p>
        <a href="/">← Kembali ke portal</a>
      </p>
      <h1>Registrasi &amp; Penjualan Benih</h1>
      <p>
        Registrasi lot, persediaan per lot, penjualan, pembatalan dengan jejak audit, dan laporan.
        Identitas memakai sesi SSO yang terverifikasi.
      </p>

      {galat && <p style={{ color: "brown" }}>Galat: {galat}</p>}
      {pesan && !galat && <p style={{ color: "green" }}>{pesan}</p>}

      <section style={{ marginTop: 24 }}>
        <h2>Registrasi Benih</h2>
        <p style={{ fontSize: 14 }}>
          Lot hanya dapat didaftarkan jika sertifikat sudah ditandatangani TTE dan label sudah
          dilegalisasi serta terkirim ke SIM Perbenihan.
        </p>
        <form onSubmit={buatRegistrasi} style={formGrid}>
          <input name="permohonan_id" type="number" min="1" placeholder="ID permohonan" required />
          <input name="komoditas" placeholder="Komoditas" required />
          <input name="varietas" placeholder="Varietas" required />
          <input name="produsen" placeholder="Produsen" required />
          <input name="sumber_benih" placeholder="Sumber benih" required />
          <input name="nomor_lot" placeholder="Nomor lot" required />
          <select name="kelas_benih" defaultValue="sebar" required>
            <option value="penjenis">Penjenis</option>
            <option value="dasar">Dasar</option>
            <option value="pokok">Pokok</option>
            <option value="sebar">Sebar</option>
          </select>
          <input name="jumlah" type="number" min="0.001" step="0.001" placeholder="Jumlah" required />
          <input name="satuan" placeholder="Satuan (kg)" required />
          <input name="tanggal_produksi" type="date" required />
          <input name="tanggal_kedaluwarsa" type="date" required />
          <select name="status_edar" defaultValue="aktif" required>
            <option value="aktif">Aktif</option>
            <option value="ditangguhkan">Ditangguhkan</option>
            <option value="dicabut">Dicabut</option>
          </select>
          <button type="submit">Daftarkan Lot</button>
        </form>

        <div style={{ marginTop: 12 }}>
          <button type="button" onClick={muatRegistrasi}>
            Muat Registrasi
          </button>
        </div>

        {registrasi.length > 0 && (
          <table style={tableStyle}>
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
                  <td>
                    {item.Komoditas} / {item.Varietas}
                  </td>
                  <td>{item.KelasBenih}</td>
                  <td>
                    {item.Jumlah} {item.Satuan}
                  </td>
                  <td>{item.TanggalKedaluwarsa}</td>
                  <td>{item.StatusEdar}</td>
                  <td style={{ whiteSpace: "nowrap" }}>
                    <button type="button" onClick={() => lihatStok(item.ID)}>
                      Stok
                    </button>{" "}
                    {item.StatusEdar === "aktif" ? (
                      <button type="button" onClick={() => ubahStatus(item.ID, "ditangguhkan")}>
                        Tangguhkan
                      </button>
                    ) : (
                      <button type="button" onClick={() => ubahStatus(item.ID, "aktif")}>
                        Aktifkan
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {stokDetail && (
        <section style={{ marginTop: 24, padding: 16, border: "1px solid #ddd" }}>
          <h2>Stok Lot {stokDetail.stok.NomorLot}</h2>
          <p>
            Tersedia: <b>{stokDetail.stok.JumlahTersedia}</b> {stokDetail.stok.Satuan} | Status:{" "}
            {stokDetail.stok.StatusEdar} | Kedaluwarsa: {stokDetail.stok.TanggalKedaluwarsa}
          </p>
          <table style={tableStyle}>
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
        </section>
      )}

      <section style={{ marginTop: 24 }}>
        <h2>Penjualan Benih</h2>
        <p style={{ fontSize: 14 }}>
          Penjualan hanya diterima untuk lot aktif, belum kedaluwarsa, serta memiliki sertifikat
          dan label yang sah. Stok dikurangi dalam transaksi database.
        </p>
        <form onSubmit={jualBenih} style={formGrid}>
          <input name="nomor_dokumen" placeholder="Nomor dokumen" required />
          <input name="registrasi_id" type="number" min="1" placeholder="ID registrasi" required />
          <input name="pembeli" placeholder="Pembeli" required />
          <input name="jumlah" type="number" min="0.001" step="0.001" placeholder="Jumlah" required />
          <input
            name="harga_satuan"
            type="number"
            min="0"
            step="0.01"
            placeholder="Harga satuan"
            required
          />
          <input
            name="tanggal_transaksi"
            type="date"
            defaultValue={tanggalHariIni()}
            required
          />
          <button type="submit">Catat Penjualan</button>
        </form>

        <div style={{ marginTop: 12 }}>
          <button type="button" onClick={muatPenjualan}>
            Muat Penjualan
          </button>
        </div>

        {penjualan.length > 0 && (
          <table style={tableStyle}>
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
                  <td>{item.Status}</td>
                  <td>
                    {item.Status === "selesai" ? (
                      <button type="button" onClick={() => batalkanPenjualan(item.ID)}>
                        Batalkan
                      </button>
                    ) : (
                      item.AlasanPembatalan
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section style={{ marginTop: 24 }}>
        <h2>Laporan</h2>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button type="button" onClick={muatLaporanStok}>
            Laporan Stok
          </button>
          <button type="button" onClick={muatLaporanKedaluwarsa}>
            Kedaluwarsa ≤ 90 Hari
          </button>
          <button type="button" onClick={muatPenjualan}>
            Laporan Penjualan
          </button>
        </div>

        {laporanStok && (
          <>
            <h3>Stok Seluruh Lot</h3>
            <table style={tableStyle}>
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
                    <td>
                      {item.Komoditas} / {item.Varietas}
                    </td>
                    <td>
                      {item.JumlahTersedia} {item.Satuan}
                    </td>
                    <td>{item.TanggalKedaluwarsa}</td>
                    <td>{item.StatusEdar}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </>
        )}

        {laporanKedaluwarsa && (
          <>
            <h3>Lot Mendekati Kedaluwarsa</h3>
            {laporanKedaluwarsa.length === 0 ? (
              <p>Tidak ada lot dalam rentang 90 hari.</p>
            ) : (
              <table style={tableStyle}>
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
                      <td>
                        {item.Komoditas} / {item.Varietas}
                      </td>
                      <td>{item.TanggalKedaluwarsa}</td>
                      <td>{item.StatusEdar}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </>
        )}
      </section>
    </main>
  );
}

const formGrid: React.CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(170px, 1fr))",
  gap: 8,
  alignItems: "center",
};

const tableStyle: React.CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  marginTop: 12,
  fontSize: 14,
};
