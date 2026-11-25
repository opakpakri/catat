<script>
  import { Download, Upload, Tag, Trash2, LogOut, Plus, Moon, Sun, Monitor, Bell, BellOff, UserCog, ArrowRight } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import { authStore } from '$lib/stores/authStore.js';
  import { themeStore } from '$lib/stores/themeStore.js';
  import { notificationStore } from '$lib/stores/notificationStore.js';

  let { categories = [], onRefresh = () => {} } = $props();

  let newCatName = $state('');
  let newCatType = $state('EXPENSE');
  let msg = $state('');
  let err = $state('');

  async function addCategory(e) {
    e.preventDefault();
    err = '';
    msg = '';
    try {
      await api.post('/categories', {
        name: newCatName.trim(),
        type: newCatType,
        icon: 'tag',
        color_hex: newCatType === 'INCOME' ? '#3B82F6' : '#F05A4F'
      });
      newCatName = '';
      msg = 'Kategori berhasil ditambahkan.';
      onRefresh();
    } catch (e) {
      err = e.message;
    }
  }

  async function removeCategory(id) {
    await api.del(`/categories/${id}`);
    onRefresh();
  }

  async function downloadCSV() {
    const blob = await api.get('/export/csv');
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'catat_transaksi.csv';
    a.click();
    URL.revokeObjectURL(url);
  }

  async function downloadJSON() {
    const data = await api.get('/backup/json');
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'catat_backup.json';
    a.click();
    URL.revokeObjectURL(url);
  }

  async function handleRestore(e) {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const text = await file.text();
      const json = JSON.parse(text);
      const res = await api.post('/backup/json', json);
      msg = `Restore selesai (${res.counts.wallets} dompet, ${res.counts.categories} kategori, ${res.counts.events} acara).`;
      onRefresh();
    } catch (e) {
      err = e.message || 'Gagal restore JSON';
    }
  }

  async function toggleNotifications() {
    err = '';
    msg = '';
    if ($notificationStore.enabled) {
      notificationStore.disable();
      msg = 'Notifikasi pengingat acara dinonaktifkan.';
      return;
    }
    const res = await notificationStore.enable();
    if (res.ok) {
      msg = 'Notifikasi pengingat acara diaktifkan.';
      await notificationStore.sendTest();
    } else if (res.reason === 'denied') {
      err = 'Izin notifikasi diblokir oleh browser. Aktifkan izin notifikasi di pengaturan situs browser Anda.';
    } else if (res.reason === 'unsupported') {
      err = 'Browser Anda tidak mendukung fitur notifikasi.';
    }
  }
</script>

<div class="p-4 space-y-4">
  <!-- User Profile Card -->
  <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 flex items-center justify-between">
    <div>
      <p class="text-sm font-semibold text-[#F4F4F5]">{$authStore.user?.full_name}</p>
      <p class="text-xs text-[#A1A1AA]">{$authStore.user?.email}</p>
    </div>
    <button
      onclick={() => authStore.logout()}
      class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#EF4444]/15 text-[#EF4444] text-xs font-semibold hover:bg-[#EF4444]/25"
    >
      <LogOut class="w-3.5 h-3.5" />
      Keluar
    </button>
  </div>

  {#if msg}
    <div class="px-3 py-2 rounded-xl bg-[#10B981]/15 border border-[#10B981]/40 text-xs text-[#34D399]">{msg}</div>
  {/if}
  {#if err}
    <div class="px-3 py-2 rounded-xl bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">{err}</div>
  {/if}

  <!-- Tampilan Aplikasi (Terang / Gelap / Sistem) -->
  <div data-tour="more-theme" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
    <div>
      <p class="text-sm font-semibold text-[#F4F4F5]">Tampilan Aplikasi</p>
      <p class="text-xs text-[#A1A1AA]">Pilih tema terang atau gelap sesuai kenyamanan Anda.</p>
    </div>
    <div class="grid grid-cols-3 gap-1 bg-[#121212] p-1 rounded-xl border border-[#2E2E34]">
      <button
        type="button"
        onclick={() => themeStore.setTheme('light')}
        class="flex items-center justify-center gap-1.5 py-2 px-2 rounded-lg text-xs font-medium transition-all {$themeStore === 'light'
          ? 'bg-[#1E1E22] text-[#3B82F6] shadow-sm border border-[#2E2E34]'
          : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
      >
        <Sun class="w-3.5 h-3.5" />
        <span>Terang</span>
      </button>
      <button
        type="button"
        onclick={() => themeStore.setTheme('dark')}
        class="flex items-center justify-center gap-1.5 py-2 px-2 rounded-lg text-xs font-medium transition-all {$themeStore === 'dark'
          ? 'bg-[#1E1E22] text-[#F05A4F] shadow-sm border border-[#2E2E34]'
          : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
      >
        <Moon class="w-3.5 h-3.5" />
        <span>Gelap</span>
      </button>
      <button
        type="button"
        onclick={() => themeStore.setTheme('system')}
        class="flex items-center justify-center gap-1.5 py-2 px-2 rounded-lg text-xs font-medium transition-all {$themeStore === 'system'
          ? 'bg-[#1E1E22] text-[#10B981] shadow-sm border border-[#2E2E34]'
          : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
      >
        <Monitor class="w-3.5 h-3.5" />
        <span>Sistem</span>
      </button>
    </div>
  </div>

  <!-- Shortcut Manajemen Pengguna (Khusus Admin) -->
  {#if $authStore.user?.role === 'ADMIN'}
    <a
      href="/pengguna"
      class="bg-[#1E1E22] border border-[#2E2E34] hover:border-[#F05A4F] rounded-2xl p-4 flex items-center justify-between gap-3 transition-colors group"
    >
      <div class="flex items-center gap-3 min-w-0">
        <div class="w-10 h-10 rounded-xl bg-[#F05A4F]/15 text-[#F05A4F] flex items-center justify-center shrink-0">
          <UserCog class="w-5 h-5" />
        </div>
        <div class="min-w-0">
          <p class="text-sm font-semibold text-[#F4F4F5]">Manajemen Pengguna</p>
          <p class="text-xs text-[#A1A1AA] truncate">
            Buat, edit, dan kelola akun pengguna (khusus Admin).
          </p>
        </div>
      </div>
      <ArrowRight class="w-4 h-4 text-[#A1A1AA] group-hover:text-[#F05A4F] shrink-0 transition-colors" />
    </a>
  {/if}

  <!-- Push Notification Reminder Acara -->
  <div data-tour="more-notif" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
    <div class="flex items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-xl bg-[#121212] border border-[#2E2E34] flex items-center justify-center shrink-0">
          {#if $notificationStore.enabled}
            <Bell class="w-4 h-4 text-[#10B981]" />
          {:else}
            <BellOff class="w-4 h-4 text-[#A1A1AA]" />
          {/if}
        </div>
        <div>
          <p class="text-sm font-semibold text-[#F4F4F5]">Push Notifikasi Acara</p>
          <p class="text-xs text-[#A1A1AA]">
            {#if !$notificationStore.supported}
              Tidak didukung di browser ini
            {:else if $notificationStore.permission === 'denied'}
              Diblokir oleh browser (izinkan di pengaturan situs)
            {:else if $notificationStore.enabled}
              Aktif — pengingat muncul otomatis sesuai jadwal
            {:else}
              Aktifkan pengingat untuk acara & jatuh tempo
            {/if}
          </p>
        </div>
      </div>

      <button
        type="button"
        role="switch"
        aria-checked={$notificationStore.enabled}
        aria-label="Toggle Push Notifikasi"
        disabled={!$notificationStore.supported}
        onclick={toggleNotifications}
        class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 disabled:opacity-40 {$notificationStore.enabled
          ? 'bg-[#10B981]'
          : 'bg-[#71717A]/40'}"
      >
        <span
          class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition duration-200 {$notificationStore.enabled
            ? 'translate-x-5'
            : 'translate-x-0'}"
        ></span>
      </button>
    </div>

    {#if $notificationStore.enabled}
      <button
        type="button"
        onclick={() => notificationStore.sendTest()}
        class="w-full py-2 rounded-xl bg-[#121212] border border-[#2E2E34] text-xs font-medium text-[#F4F4F5] hover:border-[#10B981] transition-colors"
      >
        Kirim Tes Notifikasi Sekarang
      </button>
    {/if}
  </div>

  <!-- Export & Backup -->
  <div data-tour="more-export" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA]">Ekspor & Backup Data</h3>
    <div class="grid grid-cols-2 gap-2.5">
      <button
        onclick={downloadCSV}
        class="flex items-center justify-center gap-2 py-2.5 rounded-xl bg-[#121212] border border-[#2E2E34] text-xs font-medium text-[#F4F4F5] hover:border-[#F05A4F]"
      >
        <Download class="w-4 h-4 text-[#3B82F6]" />
        Ekspor CSV
      </button>
      <button
        onclick={downloadJSON}
        class="flex items-center justify-center gap-2 py-2.5 rounded-xl bg-[#121212] border border-[#2E2E34] text-xs font-medium text-[#F4F4F5] hover:border-[#F05A4F]"
      >
        <Download class="w-4 h-4 text-[#10B981]" />
        Backup JSON
      </button>
    </div>
    <label class="flex items-center justify-center gap-2 py-2.5 rounded-xl bg-[#121212] border border-dashed border-[#2E2E34] text-xs font-medium text-[#A1A1AA] hover:text-[#F4F4F5] cursor-pointer">
      <Upload class="w-4 h-4" />
      Restore dari File Backup JSON
      <input type="file" accept=".json" onchange={handleRestore} class="hidden" />
    </label>
  </div>

  <!-- Category Management -->
  <div data-tour="more-categories" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
    <div class="flex items-center gap-2">
      <Tag class="w-4 h-4 text-[#F05A4F]" />
      <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA]">Manajemen Kategori</h3>
    </div>

    <form onsubmit={addCategory} class="flex gap-2">
      <select bind:value={newCatType} class="bg-[#121212] border border-[#2E2E34] rounded-xl px-2.5 py-2 text-xs text-[#F4F4F5]">
        <option value="EXPENSE">Pengeluaran</option>
        <option value="INCOME">Pendapatan</option>
      </select>
      <input
        type="text"
        bind:value={newCatName}
        placeholder="Nama kategori baru"
        required
        class="flex-1 bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-xs text-[#F4F4F5]"
      />
      <button type="submit" class="px-3 py-2 rounded-xl bg-[#F05A4F] text-white text-xs font-semibold flex items-center gap-1">
        <Plus class="w-3.5 h-3.5" />
      </button>
    </form>

    <div class="max-h-56 overflow-y-auto divide-y divide-[#2E2E34]/60">
      {#each categories as cat}
        <div class="flex items-center justify-between py-2">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full" style="background-color: {cat.color_hex}"></span>
            <span class="text-xs text-[#F4F4F5]">{cat.name}</span>
            <span class="text-[10px] px-1.5 py-0.2 rounded bg-[#121212] text-[#A1A1AA]">
              {cat.type === 'INCOME' ? 'Pendapatan' : 'Pengeluaran'}
            </span>
          </div>
          <button onclick={() => removeCategory(cat.id)} class="p-1 text-[#A1A1AA] hover:text-[#EF4444]">
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>
      {/each}
    </div>
  </div>
</div>
