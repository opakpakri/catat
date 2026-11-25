import { driver } from "driver.js";
import "driver.js/dist/driver.css";

const TOUR_DONE_KEY = "catat_tour_done_v1";

export function hasCompletedTour() {
  try {
    return localStorage.getItem(TOUR_DONE_KEY) === "true";
  } catch {
    return true;
  }
}

export function markTourCompleted() {
  try {
    localStorage.setItem(TOUR_DONE_KEY, "true");
  } catch {
    // ignore
  }
}

function getStepsForPage(activeTab, isDesktop) {
  const navSelector = isDesktop
    ? '[data-tour="sidebar-nav"]'
    : '[data-tour="bottom-nav"]';

  if (activeTab === "stats" || activeTab?.startsWith("/statistik")) {
    return [
      {
        popover: {
          title: "Panduan Halaman Statistik",
          description:
            "Halaman ini membantu Anda menganalisis pengeluaran per kategori, tren arus kas bulanan, dan batas anggaran.",
          align: "center",
        },
      },
      {
        element: '[data-tour="period-picker"]',
        popover: {
          title: "1. Pilih Periode Statistik",
          description:
            "Ganti bulan atau tahun untuk melihat laporan statistik pada periode tertentu.",
          side: "bottom",
          align: "start",
        },
      },
      {
        element: '[data-tour="stats-donut"]',
        popover: {
          title: "2. Pengeluaran per Kategori",
          description:
            "Visualisasi persentase dan nominal pengeluaran Anda di setiap kategori.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="stats-trend"]',
        popover: {
          title: "3. Tren Arus Kas (6 Bulan)",
          description:
            "Bandingkan grafik batang Pendapatan (biru) dan Pengeluaran (merah) selama 6 bulan terakhir.",
          side: "top",
          align: "center",
        },
      },
      {
        element: '[data-tour="stats-budgets"]',
        popover: {
          title: "4. Anggaran (Budgeting)",
          description:
            "Tetapkan batas pengeluaran per kategori. Indikator otomatis berubah menjadi Aman, Mendekati Limit, atau Overlimit, dan Anda bisa menghapus anggaran kapan saja.",
          side: "top",
          align: "center",
        },
      },
    ];
  }

  if (activeTab === "accounts" || activeTab?.startsWith("/akun")) {
    return [
      {
        popover: {
          title: "Panduan Halaman Akun",
          description:
            "Kelola seluruh sumber dana (Dompet) dan catatan Utang/Piutang yang terhubung langsung ke saldo dompet Anda.",
          align: "center",
        },
      },
      {
        element: '[data-tour="accounts-networth"]',
        popover: {
          title: "1. Total Kekayaan Bersih",
          description:
            "Akumulasi saldo dari seluruh dompet aktif yang Anda miliki.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="accounts-subtabs"]',
        popover: {
          title: "2. Tab Dompet & Utang Piutang",
          description:
            "Beralih antara daftar <b>Dompet</b> (Tunai, Bank, E-Wallet, Investasi) dan catatan <b>Utang & Piutang</b>.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="accounts-content"]',
        popover: {
          title: "3. Kelola Saldo & Cicilan",
          description:
            "Tambah dompet baru atau catat utang/piutang dengan tanggal jatuh tempo dan dompet terkait. Pembayaran cicilan otomatis menyesuaikan saldo dompet.",
          side: "top",
          align: "center",
        },
      },
    ];
  }

  if (activeTab?.startsWith("/pengguna")) {
    return [
      {
        popover: {
          title: "Panduan Halaman Pengguna",
          description:
            "Halaman khusus Admin untuk membuat dan mengelola akun pengguna aplikasi Catat.",
          align: "center",
        },
      },
      {
        element: '[data-tour="admin-banner"]',
        popover: {
          title: "1. Total Pengguna Terdaftar",
          description:
            "Menampilkan jumlah akun pengguna yang saat ini aktif di sistem.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="admin-add-btn"]',
        popover: {
          title: "2. Tambah Pengguna Baru",
          description:
            "Buat akun pengguna baru dengan nama, email, dan password awal. Dompet dan kategori default otomatis disiapkan.",
          side: "bottom",
          align: "end",
        },
      },
      {
        element: '[data-tour="admin-list"]',
        popover: {
          title: "3. Daftar Akun Pengguna",
          description:
            "Lihat daftar akun pengguna yang sudah dibuat atau hapus akun yang tidak lagi digunakan.",
          side: "top",
          align: "center",
        },
      },
    ];
  }

  if (activeTab === "more" || activeTab?.startsWith("/lainnya")) {
    return [
      {
        popover: {
          title: "Panduan Halaman Lainnya",
          description:
            "Atur preferensi tampilan, aktifkan notifikasi pengingat, ekspor data, serta kelola kategori kustom Anda.",
          align: "center",
        },
      },
      {
        element: '[data-tour="more-theme"]',
        popover: {
          title: "1. Tampilan Aplikasi",
          description:
            "Pilih tema <b>Terang</b>, <b>Gelap</b>, atau ikuti pengaturan <b>Sistem</b>.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="more-notif"]',
        popover: {
          title: "2. Push Notifikasi Acara",
          description:
            "Aktifkan izin notifikasi browser agar pengingat acara dan jatuh tempo muncul otomatis.",
          side: "bottom",
          align: "center",
        },
      },
      {
        element: '[data-tour="more-export"]',
        popover: {
          title: "3. Ekspor & Backup Data",
          description:
            "Unduh riwayat transaksi ke <b>CSV</b> atau simpan & pulihkan seluruh data dengan <b>Backup JSON</b>.",
          side: "top",
          align: "center",
        },
      },
      {
        element: '[data-tour="more-categories"]',
        popover: {
          title: "4. Manajemen Kategori",
          description:
            "Tambah atau hapus kategori Pendapatan dan Pengeluaran sesuai kebutuhan Anda.",
          side: "top",
          align: "center",
        },
      },
    ];
  }

  // Default: 'trans' page
  return [
    {
      popover: {
        title: "Panduan Halaman Transaksi",
        description:
          "Halaman utama untuk mencatat pemasukan, pengeluaran, transfer antar dompet, serta memantau kalender dan catatan acara.",
        align: "center",
      },
    },
    {
      element: '[data-tour="period-picker"]',
      popover: {
        title: "1. Pilih Periode Bulan & Tahun",
        description:
          "Klik bulan/tahun di sini untuk membuka pemilih kalender cepat atau tekan panah kiri/kanan untuk berpindah periode.",
        side: "bottom",
        align: "start",
      },
    },
    {
      element: '[data-tour="summary-bar"]',
      popover: {
        title: "2. Ringkasan Saldo & Arus Kas",
        description:
          "Pantau total Saldo Bersih seluruh dompet Anda, status kesehatan keuangan, serta total Pendapatan dan Pengeluaran bulan aktif.",
        side: "bottom",
        align: "center",
      },
    },
    {
      element: '[data-tour="trans-subtabs"]',
      popover: {
        title: "3. Empat Tampilan Transaksi",
        description:
          "<b>Daily:</b> daftar harian.<br/><b>Monthly:</b> kalender pemasukan, pengeluaran & acara.<br/><b>Total:</b> ringkasan & progres anggaran.<br/><b>Note:</b> daftar acara & reminder.",
        side: "bottom",
        align: "center",
      },
    },
    {
      element: isDesktop
        ? '[data-tour="sidebar-quick-add"]'
        : '[data-tour="fab-quick-add"]',
      popover: {
        title: "4. Catat Baru (+)",
        description:
          "Gunakan tombol ini untuk mencatat <b>Pendapatan</b>, <b>Pengeluaran</b>, <b>Transfer Antar Dompet</b>, atau <b>Acara / Reminder</b> baru.",
        side: isDesktop ? "right" : "top",
        align: "center",
      },
    },
    {
      element: navSelector,
      popover: {
        title: "5. Navigasi Halaman Lainnya",
        description:
          "Pindah ke halaman <b>Statistik</b>, <b>Akun</b>, atau <b>Lainnya</b>, lalu tekan <b>Panduan Halaman</b> untuk melihat tur di halaman tersebut.",
        side: isDesktop ? "right" : "top",
        align: "center",
      },
    },
  ];
}

export function startProductTour(activeTab = "trans") {
  if (typeof window === "undefined") return;

  setTimeout(() => {
    const isDesktop = window.innerWidth >= 768;
    const steps = getStepsForPage(activeTab, isDesktop).filter(
      (s) => !s.element || document.querySelector(s.element),
    );

    const driverObj = driver({
      showProgress: true,
      animate: true,
      allowClose: true,
      overlayOpacity: 0.65,
      stagePadding: 6,
      stageRadius: 12,
      nextBtnText: "Lanjut →",
      prevBtnText: "← Kembali",
      doneBtnText: "Selesai",
      progressText: "{{current}} dari {{total}}",
      popoverClass: "catat-driver-popover",
      onDestroyed: () => {
        markTourCompleted();
      },
      steps,
    });

    driverObj.drive();
  }, 60);
}
