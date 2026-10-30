"use client";

import { useEffect, useId, useMemo, useRef, useState } from "react";

// Input teks dengan saran pilihan (combobox ARIA). Pengguna dapat memilih dari
// daftar atau mengetik nilai bebas — cocok untuk data yang daftarnya praktis
// tetapi tidak menutup kemungkinan nilai baru.
//
// Pilihan grup (opsional) menampilkan saran berkelompok; tanpa grup, semua
// saran tampil sebagai satu daftar. Status ala ARIA dipertahankan karena
// <datalist> bawaan browser tidak dapat dikendalikan gayanya dan tidak
// terbaca seragam oleh pembaca layar.
//
// Saran bertingkat: lewat properti `induk`, daftar saran disaring mengikuti
// nilai kolom induk. Sebelum kolom induk terisi, input turunan dinonaktifkan
// agar pengguna memilih urutan yang benar lebih dulu.

export interface AutocompleteOpsi {
  nilai: string;
  grup?: string;
  induk?: string;
}

const samakan = (v: string) => v.trim().toLowerCase();

export function Autocomplete({
  id,
  label,
  nilai,
  onUbah,
  opsi,
  placeholder,
  disabled = false,
  wajib = false,
  bantuan,
  nama,
  bolehNilaiBebas = true,
  pesanKosong,
  onPilih,
  induk,
  indukLabel,
  className = "",
}: {
  id?: string;
  label: string;
  nilai: string;
  onUbah: (v: string) => void;
  opsi: ReadonlyArray<AutocompleteOpsi>;
  placeholder?: string;
  disabled?: boolean;
  wajib?: boolean;
  bantuan?: string;
  nama?: string;
  bolehNilaiBebas?: boolean;
  pesanKosong?: string;
  onPilih?: (v: string) => void;
  /** Nama kolom induk; bila diisi, saran disaring menurut nilai kolom itu. */
  induk?: string;
  /** Label kolom induk untuk pesan "pilih ... terlebih dahulu". */
  indukLabel?: string;
  className?: string;
}) {
  const idDasar = useId();
  const idInput = id ?? `autocomplete-${idDasar}`;
  const idDaftar = `${idInput}-daftar`;
  const idBantuan = `${idInput}-bantuan`;

  const [terbuka, setTerbuka] = useState(false);
  const [aktif, setAktif] = useState(0);
  const [disentuh, setDisentuh] = useState(false);
  const wadah = useRef<HTMLDivElement>(null);

  // Kebijakan kolom turunan: tanpa induk terisi, input dinonaktifkan supaya
  // pengguna memilih kolom induk lebih dulu. Induk yang diketik bebas (tidak
  // ada di daftar) tetap membuka kolom turunan sehingga nilai baru dapat diisi.
  const opsiInduk = useMemo(
    () => (induk ? [...new Set(opsi.map((o) => o.induk ?? ""))].filter(Boolean) : []),
    [induk, opsi],
  );
  const indukKosong = induk !== undefined && !induk.trim();
  const indukBebas = induk !== undefined && !!induk.trim() && !opsiInduk.includes(induk.trim());
  const terkunciInduk = indukKosong || indukBebas;

  // Saran difilter tanpa membedakan huruf besar/kecil, baik pada awal kata
  // maupun di tengah kata (mis. "benih" menemukan "Bawang Merah"). Bila ada
  // induk, hanya nilai milik induk aktif yang ditawarkan.
  const disaring = useMemo(() => {
    const kata = samakan(nilai);
    // Induk yang diketik bebas (di luar daftar) tidak punya saran varietas;
    // input tetap terbuka sebagai isian manual.
    if (induk !== undefined && indukBebas) return [];
    const sesuaiInduk = induk
      ? opsi.filter((o) => samakan(o.induk ?? "") === samakan(induk))
      : opsi;
    if (!kata) return sesuaiInduk;
    return sesuaiInduk.filter((o) => samakan(o.nilai).includes(kata));
  }, [nilai, opsi, induk, indukBebas]);

  const kelompok = useMemo(() => {
    const hasil: Array<{ grup: string; isi: AutocompleteOpsi[] }> = [];
    for (const o of disaring) {
      const nama = o.grup ?? "";
      const terakhir = hasil[hasil.length - 1];
      if (terakhir && terakhir.grup === nama) {
        terakhir.isi.push(o);
      } else {
        hasil.push({ grup: nama, isi: [o] });
      }
    }
    return hasil;
  }, [disaring]);

  // Tutup daftar saat fokus keluar dari komponen.
  useEffect(() => {
    if (!terbuka) return;
    const onKlik = (e: MouseEvent) => {
      if (wadah.current && !wadah.current.contains(e.target as Node)) setTerbuka(false);
    };
    document.addEventListener("mousedown", onKlik);
    return () => document.removeEventListener("mousedown", onKlik);
  }, [terbuka]);

  const pilih = (v: string) => {
    onUbah(v);
    onPilih?.(v);
    setTerbuka(false);
    setDisentuh(true);
  };

  const idOpsi = (i: number) => `${idInput}-opsi-${i}`;

  const galat =
    wajib && bolehNilaiBebas && disentuh && !nilai.trim()
      ? pesanKosong ?? `${label} wajib diisi.`
      : "";

  // Kolom turunan: nonaktif selama induknya kosong, dan dibuka sebagai isian
  // bebas saat induk diketik di luar daftar.
  const dinonaktifkan = disabled || indukKosong;
  const bisaKetikanBebas = bolehNilaiBebas && (induk === undefined || !terkunciInduk);
  const pesanPlaceholder = indukKosong
    ? `Pilih ${indukLabel?.toLowerCase() ?? "kolom sebelumnya"} terlebih dahulu`
    : placeholder;

  return (
    <div className={`form-group ${className}`.trim()} ref={wadah}>
      <label className="form-label" htmlFor={idInput}>
        {label}
        {wajib && (
          <span className="autocomplete-wajib" aria-hidden="true">
            {" "}
            *
          </span>
        )}
      </label>
      <div className="autocomplete">
        <input
          id={idInput}
          name={nama}
          type="text"
          role="combobox"
          className="form-input autocomplete-input"
          value={nilai}
          placeholder={pesanPlaceholder}
          disabled={dinonaktifkan}
          autoComplete="off"
          aria-autocomplete="list"
          aria-expanded={terbuka}
          aria-controls={idDaftar}
          aria-activedescendant={terbuka && disaring.length > 0 ? idOpsi(aktif) : undefined}
          aria-invalid={galat ? true : undefined}
          aria-describedby={bantuan || galat ? [bantuan && idBantuan, galat && `${idInput}-galat`].filter(Boolean).join(" ") : undefined}
          onChange={(e) => {
            onUbah(e.target.value);
            // Nilai yang tidak ada di daftar saran dianggap isian bebas, jadi
            // daftar tetap terbuka untuk menawarkan pilihan resmi.
            setTerbuka(true);
            setAktif(0);
          }}
          onFocus={() => {
            setTerbuka(bisaKetikanBebas || disaring.length > 0);
            setAktif(0);
          }}
          onBlur={() => setDisentuh(true)}
          onKeyDown={(e) => {
            if (e.key === "ArrowDown") {
              e.preventDefault();
              setTerbuka(true);
              setAktif((i) => Math.min(disaring.length - 1, i + 1));
            } else if (e.key === "ArrowUp") {
              e.preventDefault();
              setTerbuka(true);
              setAktif((i) => Math.max(0, i - 1));
            } else if (e.key === "Enter") {
              if (terbuka && disaring[aktif]) {
                e.preventDefault();
                pilih(disaring[aktif].nilai);
              }
            } else if (e.key === "Escape") {
              setTerbuka(false);
            } else if (e.key === "Tab") {
              setTerbuka(false);
            }
          }}
        />
        <button
          type="button"
          className="autocomplete-tombol"
          tabIndex={-1}
          aria-label={terbuka ? `Tutup saran ${label}` : `Buka saran ${label}`}
          aria-expanded={terbuka}
          disabled={dinonaktifkan}
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => {
            setTerbuka((t) => !t);
            setAktif(0);
          }}
        >
          <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false">
            <path
              d="M6 9l6 6 6-6"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
        {terbuka && !dinonaktifkan && !terkunciInduk && disaring.length > 0 && (
          <ul className="autocomplete-daftar" id={idDaftar} role="listbox" aria-label={`Saran ${label}`}>
            {kelompok.map((k) => (
              <li key={k.grup || "_tanpa_grup"} role="presentation">
                {k.grup && <div className="autocomplete-grup">{k.grup}</div>}
                <ul role="group" aria-label={k.grup || undefined}>
                  {k.isi.map((o) => {
                    const i = disaring.indexOf(o);
                    return (
                      <li
                        key={`${k.grup}-${o.nilai}`}
                        id={idOpsi(i)}
                        role="option"
                        aria-selected={o.nilai === nilai}
                        className={`autocomplete-opsi${i === aktif ? " is-aktif" : ""}`}
                        onMouseEnter={() => setAktif(i)}
                        onMouseDown={(e) => e.preventDefault()}
                        onClick={() => pilih(o.nilai)}
                      >
                        {o.nilai}
                      </li>
                    );
                  })}
                </ul>
              </li>
            ))}
          </ul>
        )}
      </div>
      {galat ? (
        <p id={`${idInput}-galat`} className="autocomplete-galat" role="alert">
          {galat}
        </p>
      ) : null}
      {indukKosong ? (
        <p className="text-xs text-sitani-muted mt-1">
          Pilih {indukLabel?.toLowerCase() ?? "kolom sebelumnya"} terlebih dahulu agar pilihan
          {label.toLowerCase()} muncul.
        </p>
      ) : indukBebas ? (
        <p className="text-xs text-sitani-muted mt-1">
          {indukLabel ?? "Kolom sebelumnya"} di luar daftar — ketik {label.toLowerCase()} secara
          manual.
        </p>
      ) : (
        bantuan && (
          <p id={idBantuan} className="text-xs text-sitani-muted mt-1">
            {bantuan}
          </p>
        )
      )}
    </div>
  );
}
