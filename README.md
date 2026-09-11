# git-merge-test

API Sistem Perpustakaan menggunakan Go dan standard library, sebagai kode awal latihan branch, merge, dan penyelesaian conflict untuk anak magang.

## Dokumentasi aplikasi

Gunakan Go 1.22 atau lebih baru. Dari folder project:

```sh
go run .
```

API tersedia di `http://localhost:8080`. Data disimpan di `data/books.json` pada folder kerja. Jalankan satu proses server untuk file ini; request dalam proses yang sama dilindungi dari perubahan bersamaan.

```sh
go test ./...
# Opsional: cek akses data bersamaan (membutuhkan dukungan race detector).
go test -race ./...
```

## API awal

| Request | Perilaku |
| --- | --- |
| `GET /` | Informasi API dan daftar routes |
| `GET /books` | Semua buku |
| `GET /books?id=1` | Satu buku berdasarkan ID |
| `GET /books?penulis=Leila%20S.%20Chudori` | Daftar buku dengan nama penulis yang sama, tanpa membedakan huruf besar/kecil |
| `POST /books` | Tambah buku, sukses `201`, ID duplikat `409` |
| `PUT /books` | Ubah sebagian field buku berdasarkan `id_buku` |
| `DELETE /books` | Hapus buku menggunakan JSON berisi `id_buku` |

Jika `id` dan `penulis` dikirim bersama, pencarian ID didahulukan. Data yang tidak ditemukan menghasilkan `404`. Input tidak valid menghasilkan `400` dengan JSON `{"message":"..."}`.

Saat POST, ketujuh field wajib hadir: `id_buku`, `judul`, `penulis`, `tahun`, `stok`, `penerbit`, dan `harga`. ID, tahun, dan harga harus bilangan bulat positif; stok harus bilangan bulat minimal `0`; judul, penulis, dan penerbit tidak boleh kosong atau hanya spasi. PUT wajib membawa `id_buku` dan memvalidasi setiap field yang dikirim, tanpa menghapus field yang tidak dikirim. DELETE wajib membawa ID positif.

Tiga buku awal mengikuti data sumber: belum memiliki `penerbit` dan `harga`. Gunakan POST atau PUT untuk menambahkan data tersebut saat latihan. Pencarian penerbit, pencarian tahun, dan batas harga Rp1.000 **belum diimplementasikan**; ketiganya adalah tugas peserta.

### Contoh cURL

Perintah berikut mengubah `data/books.json`. Jalankan berurutan saat server aktif; ID `4` harus belum digunakan.

```sh
curl -i http://localhost:8080/books
curl -i 'http://localhost:8080/books?id=1'
curl -i 'http://localhost:8080/books?penulis=leila%20s.%20chudori'

curl -i -X POST http://localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"id_buku":4,"judul":"Buku Baru","penulis":"Budi","tahun":2024,"stok":0,"penerbit":"Gramedia","harga":50000}'

curl -i -X PUT http://localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"id_buku":4,"stok":5}'

curl -i -X DELETE http://localhost:8080/books \
  -H 'Content-Type: application/json' \
  -d '{"id_buku":4}'
```

## TODO

Mentor commit dan push kode awal ke `main` sebelum latihan. Semua peserta mulai dari **commit main yang sama**, sebelum fitur pertama di-merge. Gunakan clone masing-masing agar pekerjaan tidak saling mengganggu.

```sh
git clone git@github.com:agripaa/git-merge-test.git
cd git-merge-test
git switch main
git pull --ff-only origin main
git status
git log -1 --oneline
go test ./...
```

Cocokkan hash commit awal dengan mentor. Setiap peserta membuat satu branch berikut dari titik awal tersebut.

### A — Pencarian penerbit

```sh
git switch -c feature/search-penerbit
```

- [ ] Tambahkan `GET /books?penerbit=Gramedia` di `getBooksHandler`.
- [ ] Cocokkan nama penerbit secara utuh tanpa membedakan huruf besar/kecil.
- [ ] Kembalikan daftar buku dan `200` jika ditemukan, atau JSON pesan dan `404` jika tidak ditemukan.
- [ ] Tambahkan route pada `homeHandler` dan data penerbit untuk pengujian.
- [ ] Uji hasil cocok, huruf besar/kecil, serta penerbit yang tidak ada; pertahankan pencarian ID dan penulis.

### B — Harga minimal Rp1.000

```sh
git switch -c feature/validasi-harga
```

- [ ] Tambahkan aturan `harga >= 1000` pada validasi bersama untuk POST dan PUT.
- [ ] Pertahankan validasi tipe, field wajib pada POST, dan update sebagian pada PUT.
- [ ] Kembalikan `400` dengan pesan yang menjelaskan batas minimal harga.
- [ ] Uji harga `-1`, `0`, `500`, `999`, `1000`, dan harga berupa string; hanya bilangan bulat minimal `1000` diterima.
- [ ] Pastikan PUT tanpa field harga tetap dapat memperbarui stok, termasuk buku awal yang belum mempunyai harga.

### C — Pencarian tahun

```sh
git switch -c feature/search-tahun
```

- [ ] Tambahkan `GET /books?tahun=2020` di `getBooksHandler`.
- [ ] Tahun pencarian harus bilangan bulat positif; input tidak valid menghasilkan `400`.
- [ ] Kembalikan daftar buku dan `200` jika ditemukan, atau JSON pesan dan `404` jika tidak ditemukan.
- [ ] Tambahkan route pada `homeHandler` dan uji tahun cocok, tidak ditemukan, serta input salah.

Untuk hasil integrasi, gunakan urutan prioritas `id`, `penulis`, `penerbit`, lalu `tahun` jika lebih dari satu parameter diberikan. Setiap peserta menambahkan pemeriksaan kecil untuk fiturnya pada test yang tersedia.

### Commit dan kirim branch

Setelah implementasi, jalankan perintah berikut di branch masing-masing. Sesuaikan pesan commit dengan fitur; tambahkan `data/books.json` hanya jika perubahan data memang bagian dari tugas.

```sh
gofmt -w main.go main_test.go
go test ./...
git diff
git add main.go main_test.go
git commit -m "feat: add book search by publisher"
git push -u origin HEAD
```

## Merge oleh mentor

Branch peserta berasal dari clone berbeda, sehingga ambil branch remote terlebih dahulu. Pastikan working tree bersih sebelum mulai.

```sh
git switch main
git pull --ff-only origin main
git fetch origin
git merge --no-ff origin/feature/search-penerbit
git merge --no-ff origin/feature/validasi-harga
git merge --no-ff origin/feature/search-tahun
```

Jalankan merge satu per satu. Jika conflict muncul, selesaikan dan commit sebelum melanjutkan merge berikutnya. Perubahan dalam file yang sama **bisa** menyebabkan conflict, tetapi tidak selalu; Git dapat menggabungkan perubahan pada bagian yang berbeda.

## Latihan conflict yang disengaja

Setelah integrasi fitur selesai, buat **dua branch baru dari commit yang sama** di satu clone mentor/peserta. Gunakan nama branch berikut hanya jika belum ada.

```sh
git switch main
git branch latihan/pesan-a
git branch latihan/pesan-b
git switch latihan/pesan-a
```

Pada branch A, ubah nilai `message` dalam `homeHandler` menjadi `API Perpustakaan Digital`, lalu:

```sh
git add main.go
git commit -m "feat: rename library API"
git switch latihan/pesan-b
```

Pada branch B, ubah **baris message yang sama** menjadi `API Sistem Perpustakaan - Latihan Git Merge`, lalu:

```sh
git add main.go
git commit -m "feat: describe merge exercise"
git switch main
git merge --no-ff latihan/pesan-a
git merge --no-ff latihan/pesan-b
```

Kedua branch mengubah baris yang sama dengan nilai berbeda, sehingga merge kedua menghasilkan conflict. Tugas peserta adalah mempertahankan dua kebutuhan: nama `Perpustakaan Digital` dan keterangan `Latihan Git Merge`.

- [ ] Jalankan `git status` dan `git diff`; temukan conflict pada `main.go`.
- [ ] Baca `<<<<<<< HEAD` sebagai versi branch aktif, `=======` sebagai pemisah, dan `>>>>>>> ...` sebagai versi branch yang digabungkan.
- [ ] Gabungkan kebutuhan kedua perubahan dan hapus seluruh marker conflict. Jangan sekadar mengambil semua versi ours/theirs.
- [ ] Pastikan fitur pencarian, validasi, POST, PUT, dan DELETE tetap berjalan setelah penyelesaian.

```sh
gofmt -w main.go main_test.go
go test ./...
git add main.go
git commit -m "fix: resolve home message merge conflict"
git log --oneline --graph --all
git push origin main
```

Sebelum push, uji juga semua route melalui server yang sudah dijalankan ulang. Target akhir: seluruh fitur dari ketiga branch tetap tersedia dan hasil penyelesaian conflict memenuhi kedua kebutuhan.
