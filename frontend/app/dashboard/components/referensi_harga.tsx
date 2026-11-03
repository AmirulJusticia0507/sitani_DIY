import { useState, useCallback } from "react";
import { Card, Button } from "../../../components";

type ReferensiHargaProps = {
  api: (path: string, opts?: RequestInit) => Promise<Response>;
};

export function ReferensiHarga({ api }: ReferensiHargaProps) {
  const [harga, setHarga] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<string>("");

  const refreshHarga = useCallback(async () => {
    setLoading(true);
    setLastRefresh(new Date().toLocaleString("id-ID"));
    try {
      const res = await api("/v1/referensi/harga_benih", { method: "GET" });
      const data = await res.json();
      setHarga(data);
    } catch (err) {
      setHarga({ error: (err as Error).message });
    } finally {
      setLoading(false);
    }
  }, [api]);

  return (
    <Card className="mb-4">
      <div className="flex flex-col sm:flex-row items-center justify-between gap-2">
        <h4 className="text-sm font-bold text-sitani-green-700 uppercase tracking-wider mb-0">
          Referensi Harga Benih
        </h4>
        <Button
          variant="outline"
          size="sm"
          onClick={refreshHarga}
          disabled={loading}
        >
          {loading ? "Memuat…" : "Refresh"}
        </Button>
      </div>
      {loading && <p className="text-xs text-sitani-muted">Sedang memuat data dari Superset…</p>}
      {!loading && harga && harga.status !== "belum-configured" && harga.status !== "placeholder" && (
        <div className="mt-2 text-sm text-sitani-muted">
          <p>Tomat: {harga.data?.tomat || "-"}</p>
          <p>Rama: {harga.data?.rama || "-"}</p>
          <p>Cabai: {harga.data?.cabai || "-"}</p>
        </div>
      )}
      {!loading && (harga.status === "belum-configured" || harga.status === "placeholder") && (
        <p className="text-xs text-sitani-muted">
          <strong>Belum dikonfigurasi:</strong> set env SUPPSET_URL di backend untuk menghubungkan ke Superset pusat.
        </p>
      )}
      {loading && harga?.status === "error-superset" && (
        <p className="text-xs text-sitani-merah">Gagal mengambil data Superset: {harga.pesan}</p>
      )}
    </Card>
  );
}