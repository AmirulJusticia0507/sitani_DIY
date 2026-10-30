// Pilihan jenis/komoditas benih untuk form permohonan sertifikasi BP3MBTP.
//
// Daftar ini adalah pilihan praktis untuk mempermudah pengisian, bukan daftar
// resmi yang mengikat: backend hanya mewajibkan jenis benih terisi (maks. 100
// karakter), sehingga komoditas di luar daftar tetap dapat diajukan dengan
// mengetiknya langsung. Validasi daftar resmi komoditas dan kelas benih masih
// menunggu konfirmasi BP3MBTP (lihat todo_list.md).
//
// Varietas di bawah juga bersifat contoh populer di DIY, bukan daftar varietas
// yang terdaftar/berlaku. Penandatangan tetap memeriksa keaslian varietas saat
// verifikasi; nilai bebas (mis. varietas lokal atau galur baru) dapat diketik.
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

// VARIETAS_BENIH memetakan jenis benih -> saran varietas. Jenis yang tidak
// tercantum di sini tetap dapat diajukan; kolom varietas menjadi isian bebas.
export const VARIETAS_BENIH: Readonly<Record<string, ReadonlyArray<string>>> = {
  Padi: [
    "Inpari 32 HDB",
    "Inpari 33",
    "Inpari 42",
    "Inpari 48 Blas",
    "Ciherang",
    "IR64",
    "Mekongga",
    "Memberamo",
    "Sintanur",
    "Segreng Handayani",
    "Rojolele",
    "Pandan Wangi",
  ],
  Jagung: [
    "Bisi 18",
    "Bisi 2",
    "Pioneer 36",
    "NK 212",
    "NK 7328",
    "HJ 21",
    "Sukmaraga",
    "Lamuru",
    "Bisma",
  ],
  Kedelai: ["Anjasmoro", "Grobogan", "Bromo", "Wilis", "Dena 1", "Detam 1", "Devon 1"],
  "Kacang Tanah": ["Kancil", "Jerapah", "Hypoma 1", "Domba", "Talam 1", "Simpai"],
  "Kacang Hijau": ["Vima 1", "Vima 2", "Kutilang", "Murai", "Sriti"],
  "Kacang Tunggak": ["Lokal DIY", "KT 1"],
  Sorgum: ["Numbu", "Keller", "Super 1", "Super 2"],
  Gandum: ["Dewata", "Selayar", "Nias", "Jayawijaya"],
  "Ubi Kayu": ["UJ 5", "Malang 4", "Darul Hidayah", "Lokal Gunungkidul"],
  "Ubi Jalar": ["Cilembu", "Sukuh", "Antin 1", "Beta 1", "Kuning Lokal"],
  Talas: ["Talas Bogor", "Talas Bentul", "Talas Lokal DIY"],
  Ganyong: ["Ganyong Lokal"],
  Garut: ["Garut Lokal"],
  Cabai: [
    "Lado F1",
    "Cakra Putih",
    "Seloka",
    "Baja F1",
    "Laris",
    "Hot Beauty",
    "Gada F1",
    "Rawat",
  ],
  Tomat: ["Servo F1", "Tymoti F1", "Permata", "Karina", "Mutiara", "Ranti Lokal"],
  Terung: ["Mustang F1", "Kenari", "Antaboga", "Gelatik", "Lokal DIY"],
  "Bawang Merah": ["Bima Brebes", "Tajuk", "Maja Cipanas", "Bauji", "Lokal Imogiri"],
  "Bawang Putih": ["Lumbu Hijau", "Lumbu Kuning", "Tawangmangu", "Lokal Tawangmangu"],
  "Bawang Daun": ["Lokal DIY", "Baja F1", "Widuri"],
  Kubis: ["Green Nova", "KK Cross", "Grand 11", "Delima"],
  Sawi: ["Tosakan", "Shinta", "Mustard Hijau", "Lokal DIY"],
  Kangkung: ["Kangkung Darat", "Kangkung Air", "Bisi 1"],
  Bayam: ["Bayam Merah", "Bayam Putih", "Bisi Bayam"],
  Mentimun: ["Mercy F1", "Wulan", "Sinta", "Lokal DIY"],
  Pare: ["Horti 2", "Pare Hijau Besar", "Lokal DIY"],
  Labu: ["Labu Madu", "Labu Kuning", "Butternut"],
  Buncis: ["Pandora", "Gypsy", "Nagara", "Lokal DIY"],
  "Kacang Panjang": ["Kanton Tavi", "Nagara", "Parade Tavi", "Lokal DIY"],
  Wortel: ["Imperator", "Chantenay", "Berlian", "Lokal DIY"],
  Kentang: ["Granola", "Atlantik", "Median", "Lokal Wonosobo"],
  "Kacang Kapri": ["Kapri Putih", "Kapri Biji Keriput", "Lokal DIY"],
  Mangga: ["Arumanis", "Gedong Gincana", "Harumanis", "Podang", "Madu"],
  Pisang: ["Cavendish", "Ambon", "Emas", "Kepok", "Raja Nangka"],
  Salak: ["Pondoh", "Nglumut", "Gading"],
  Jambu: ["Jambu Kristal", "Jambu Biji Merah", "Jambu Air Citra"],
  Pepaya: ["California", "Callina", "Arnusa", "Lokal DIY"],
  Durian: ["Montong", "Musang King", "Bawor", "Lokal DIY"],
  Rambutan: ["Binjai", "Rapiah", "Sibatuk"],
  Melon: ["Sky Rocket", "Glamour", "Intanon", "Sweet Rinda"],
  Semangka: ["Inul", "Bintang", "Amelia", "Lokal Kulon Progo"],
  Stroberi: ["California", "Rosalinda", "Mencir"],
  Kelapa: ["Kelapa Dalam", "Kelapa Genjah", "Hibrida"],
  Kakao: ["Forastero", "Criollo", "Trinitario"],
  Kopi: ["Arabika", "Robusta", "Liberika"],
  Tebu: ["PS 862", "PS 881", "VMC 76-16"],
  Tembakau: ["Temanggung", "Virginia", "Local DIY"],
  Karet: ["PB 260", "GT 1", "RRIC 100"],
};

// JENIS_BENIH_DATAR dikirim sebagai daftar datar ke <datalist> untuk isian di
// luar halaman ini; form permohonan memakai <Autocomplete> berkelompok.
export const JENIS_BENIH_DATAR: ReadonlyArray<string> = JENIS_BENIH_OPTIONS.flatMap(
  (kelompok) => kelompok.jenis,
);
