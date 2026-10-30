// Pilihan jenis/komoditas benih untuk form permohonan sertifikasi BP3MBTP.
//
// Daftar ini adalah pilihan praktis untuk mempermudah pengisian, bukan daftar
// resmi yang mengikat: backend hanya mewajibkan jenis benih terisi (maks. 100
// karakter), sehingga komoditas di luar daftar tetap dapat diajukan lewat opsi
// "Lainnya". Validasi daftar resmi komoditas dan kelas benih masih menunggu
// konfirmasi BP3MBTP (lihat todo_list.md).
export const JENIS_BENIH_LAINNYA = "Lainnya";

export const JENIS_BENIH_OPTIONS: ReadonlyArray<{
  grup: string;
  jenis: ReadonlyArray<string>;
}> = [
  {
    grup: "Tanaman pangan",
    jenis: [
      "Padi",
      "Jagung",
      "Kedelai",
      "Kacang Tanah",
      "Kacang Hijau",
      "Kacang Tunggak",
      "Sorgum",
      "Gandum",
      "Ubi Kayu",
      "Ubi Jalar",
      "Talas",
      "Ganyong",
      "Garut",
    ],
  },
  {
    grup: "Hortikultura — sayuran",
    jenis: [
      "Cabai",
      "Tomat",
      "Terung",
      "Bawang Merah",
      "Bawang Putih",
      "Bawang Daun",
      "Kubis",
      "Sawi",
      "Kangkung",
      "Bayam",
      "Mentimun",
      "Pare",
      "Labu",
      "Buncis",
      "Kacang Panjang",
      "Wortel",
      "Kentang",
      "Kacang Kapri",
    ],
  },
  {
    grup: "Hortikultura — buah",
    jenis: [
      "Mangga",
      "Pisang",
      "Salak",
      "Jambu",
      "Pepaya",
      "Durian",
      "Rambutan",
      "Melon",
      "Semangka",
      "Stroberi",
    ],
  },
  {
    grup: "Perkebunan",
    jenis: ["Kelapa", "Kakao", "Kopi", "Tebu", "Tembakau", "Karet"],
  },
];

// JENIS_BENIH_DATAR dikirim sebagai daftar datar ke elemen <datalist> agar
// pengguna tetap dapat mengetik nama komoditas yang tidak tercantum.
export const JENIS_BENIH_DATAR: ReadonlyArray<string> = JENIS_BENIH_OPTIONS.flatMap(
  (kelompok) => kelompok.jenis,
);
