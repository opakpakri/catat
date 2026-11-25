<script>
  import { Search, Plus, Pencil, Trash2, User, Shield, ChevronLeft, ChevronRight, X } from 'lucide-svelte';
  import { api } from '$lib/api.js';

  let users = $state([]);
  let loading = $state(false);
  let searchInput = $state('');
  let searchQuery = $state('');
  let currentPage = $state(1);
  const pageSize = 10;

  // Modal state (Create / Edit)
  let modalOpen = $state(false);
  let editingUser = $state(null);
  let fullName = $state('');
  let email = $state('');
  let password = $state('');
  let role = $state('USER');
  let currency = $state('IDR');
  let saving = $state(false);

  let msg = $state('');
  let err = $state('');

  async function loadUsers() {
    loading = true;
    try {
      const res = await api.get('/users');
      users = res.users || [];
    } catch (e) {
      err = e.message || 'Gagal memuat daftar pengguna';
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadUsers();
  });

  let filteredUsers = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return users;
    return users.filter(
      (u) =>
        (u.full_name || '').toLowerCase().includes(q) ||
        (u.email || '').toLowerCase().includes(q)
    );
  });

  let totalPages = $derived(Math.max(1, Math.ceil(filteredUsers.length / pageSize)));
  let safePage = $derived(Math.min(currentPage, totalPages));
  let pagedUsers = $derived(
    filteredUsers.slice((safePage - 1) * pageSize, safePage * pageSize)
  );
  let startIdx = $derived(filteredUsers.length === 0 ? 0 : (safePage - 1) * pageSize + 1);
  let endIdx = $derived(Math.min(safePage * pageSize, filteredUsers.length));

  function handleSearch(e) {
    e.preventDefault();
    searchQuery = searchInput;
    currentPage = 1;
  }

  function openCreateModal() {
    editingUser = null;
    fullName = '';
    email = '';
    password = '';
    role = 'USER';
    currency = 'IDR';
    err = '';
    modalOpen = true;
  }

  function openEditModal(u) {
    editingUser = u;
    fullName = u.full_name || '';
    email = u.email || '';
    password = '';
    role = u.role || 'USER';
    currency = u.currency || 'IDR';
    err = '';
    modalOpen = true;
  }

  async function saveUser(e) {
    e.preventDefault();
    err = '';
    msg = '';
    saving = true;
    try {
      if (editingUser) {
        await api.put(`/users/${editingUser.id}`, {
          full_name: fullName.trim(),
          email: email.trim(),
          password: password.trim(),
          role,
          currency
        });
        msg = `Akun ${fullName.trim()} berhasil diperbarui.`;
      } else {
        await api.post('/users', {
          full_name: fullName.trim(),
          email: email.trim(),
          password,
          role,
          currency
        });
        msg = `Akun pengguna ${fullName.trim()} berhasil dibuat.`;
      }
      modalOpen = false;
      loadUsers();
    } catch (e) {
      err = e.message || 'Gagal menyimpan akun pengguna';
    } finally {
      saving = false;
    }
  }

  async function deleteUser(u) {
    if (!confirm(`Hapus akun pengguna "${u.full_name}" (${u.email})? Seluruh data miliknya akan terhapus.`)) return;
    err = '';
    msg = '';
    try {
      await api.del(`/users/${u.id}`);
      msg = `Akun ${u.full_name} berhasil dihapus.`;
      loadUsers();
    } catch (e) {
      err = e.message || 'Gagal menghapus akun pengguna';
    }
  }

  function formatDate(iso) {
    if (!iso) return '-';
    return new Date(iso).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric'
    });
  }
</script>

<div class="p-4 sm:p-6 space-y-4">
  <!-- Header Row -->
  <div data-tour="admin-banner" class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
    <div>
      <h2 class="text-xl font-bold tracking-tight text-[#F4F4F5]">Manajemen Pengguna</h2>
      <p class="text-xs text-[#A1A1AA] mt-0.5">
        Menu khusus Admin untuk membuat dan mengelola akun pengguna.
      </p>
    </div>

    <button
      data-tour="admin-add-btn"
      type="button"
      onclick={openCreateModal}
      class="inline-flex items-center justify-center gap-1.5 px-4 py-2.5 rounded-xl bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-semibold shadow-sm transition-colors shrink-0"
    >
      <Plus class="w-4 h-4" />
      <span>Tambah Pengguna</span>
    </button>
  </div>

  <!-- Search Bar -->
  <form onsubmit={handleSearch} class="flex items-center gap-2 max-w-md">
    <div class="relative flex-1">
      <Search class="w-4 h-4 text-[#A1A1AA] absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
      <input
        type="text"
        bind:value={searchInput}
        oninput={() => {
          searchQuery = searchInput;
          currentPage = 1;
        }}
        placeholder="Cari nama atau email..."
        class="w-full bg-[#1E1E22] border border-[#2E2E34] rounded-xl pl-9 pr-3 py-2 text-xs text-[#F4F4F5] placeholder-[#A1A1AA] focus:outline-none focus:border-[#F05A4F]"
      />
    </div>
    <button
      type="submit"
      class="px-4 py-2 rounded-xl bg-[#1E1E22] border border-[#2E2E34] text-xs font-semibold text-[#F4F4F5] hover:border-[#F05A4F] transition-colors"
    >
      Cari
    </button>
  </form>

  {#if msg}
    <div class="px-3 py-2 rounded-xl bg-[#10B981]/15 border border-[#10B981]/40 text-xs text-[#34D399]">
      {msg}
    </div>
  {/if}
  {#if err && !modalOpen}
    <div class="px-3 py-2 rounded-xl bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">
      {err}
    </div>
  {/if}

  <!-- Table Card -->
  <div data-tour="admin-list" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="border-b border-[#2E2E34] text-[10px] font-semibold uppercase tracking-wider text-[#A1A1AA]">
            <th class="py-3.5 px-4">Nama</th>
            <th class="py-3.5 px-4">Email</th>
            <th class="py-3.5 px-4">Mata Uang</th>
            <th class="py-3.5 px-4">Role</th>
            <th class="py-3.5 px-4">Dibuat</th>
            <th class="py-3.5 px-4 text-right">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#2E2E34] text-xs">
          {#if loading && users.length === 0}
            <tr>
              <td colspan="6" class="py-10 text-center text-[#A1A1AA]">Memuat daftar pengguna...</td>
            </tr>
          {:else if pagedUsers.length === 0}
            <tr>
              <td colspan="6" class="py-10 text-center text-[#A1A1AA]">Tidak ada data pengguna ditemukan.</td>
            </tr>
          {:else}
            {#each pagedUsers as u}
              <tr class="hover:bg-[#2E2E34]/40 transition-colors">
                <td class="py-3.5 px-4 font-semibold text-[#F4F4F5] whitespace-nowrap">
                  {u.full_name}
                </td>
                <td class="py-3.5 px-4 text-[#A1A1AA] whitespace-nowrap">
                  {u.email}
                </td>
                <td class="py-3.5 px-4 whitespace-nowrap">
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-[#3B82F6]/15 text-[#3B82F6]">
                    {u.currency || 'IDR'}
                  </span>
                </td>
                <td class="py-3.5 px-4 whitespace-nowrap">
                  {#if u.role === 'ADMIN'}
                    <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-[#F59E0B]/15 text-[#F59E0B] border border-[#F59E0B]/30">
                      <Shield class="w-3 h-3" />
                      ADMIN
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-[#F4F4F5]/10 text-[#F4F4F5] border border-[#2E2E34]">
                      <User class="w-3 h-3" />
                      USER
                    </span>
                  {/if}
                </td>
                <td class="py-3.5 px-4 text-[#A1A1AA] whitespace-nowrap tabular-nums">
                  {formatDate(u.created_at)}
                </td>
                <td class="py-3.5 px-4 text-right whitespace-nowrap">
                  <div class="inline-flex items-center gap-1">
                    <button
                      type="button"
                      onclick={() => openEditModal(u)}
                      class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60 transition-colors"
                      title="Edit pengguna"
                      aria-label="Edit pengguna"
                    >
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                    {#if u.role !== 'ADMIN'}
                      <button
                        type="button"
                        onclick={() => deleteUser(u)}
                        class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#EF4444] hover:bg-[#EF4444]/10 transition-colors"
                        title="Hapus pengguna"
                        aria-label="Hapus pengguna"
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    <!-- Pagination Footer -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 px-4 py-3 border-t border-[#2E2E34] text-xs text-[#A1A1AA]">
      <span>
        Menampilkan {startIdx}–{endIdx} dari {filteredUsers.length} data
      </span>

      <div class="flex items-center gap-1.5">
        <button
          type="button"
          disabled={safePage <= 1}
          onclick={() => (currentPage = Math.max(1, safePage - 1))}
          class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg bg-[#121212] border border-[#2E2E34] text-[#A1A1AA] hover:text-[#F4F4F5] disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <ChevronLeft class="w-3.5 h-3.5" />
          <span>Sebelumnya</span>
        </button>

        <span class="w-7 h-7 rounded-lg bg-[#F05A4F] text-white font-semibold flex items-center justify-center tabular-nums">
          {safePage}
        </span>

        <button
          type="button"
          disabled={safePage >= totalPages}
          onclick={() => (currentPage = Math.min(totalPages, safePage + 1))}
          class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg bg-[#121212] border border-[#2E2E34] text-[#A1A1AA] hover:text-[#F4F4F5] disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <span>Berikutnya</span>
          <ChevronRight class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  </div>
</div>

<!-- Modal Tambah / Edit Pengguna -->
{#if modalOpen}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm p-0 sm:p-4">
    <div class="w-full max-w-lg bg-[#1E1E22] border border-[#2E2E34] rounded-t-2xl sm:rounded-2xl shadow-2xl overflow-hidden">
      <div class="flex items-center justify-between px-6 py-4 border-b border-[#2E2E34]">
        <h3 class="text-base font-bold text-[#F4F4F5]">
          {editingUser ? 'Edit Pengguna' : 'Tambah Pengguna Baru'}
        </h3>
        <button
          type="button"
          onclick={() => (modalOpen = false)}
          class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5]"
          aria-label="Tutup"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <form onsubmit={saveUser} class="p-6 space-y-4">
        {#if err}
          <div class="px-3 py-2 rounded-lg bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">
            {err}
          </div>
        {/if}

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label for="u-name" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Nama Lengkap</label>
            <input
              id="u-name"
              type="text"
              bind:value={fullName}
              placeholder="Nama User"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>

          <div>
            <label for="u-role" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Role</label>
            <select
              id="u-role"
              bind:value={role}
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            >
              <option value="USER">User (Pengguna)</option>
              <option value="ADMIN">Admin</option>
            </select>
          </div>

          <div>
            <label for="u-email" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Email Login</label>
            <input
              id="u-email"
              type="email"
              bind:value={email}
              placeholder="Email User"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>

          <div>
            <label for="u-pass" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">
              {editingUser ? 'Password Baru (Opsional)' : 'Password (min. 6)'}
            </label>
            <input
              id="u-pass"
              type="password"
              bind:value={password}
              placeholder={editingUser ? 'Kosongkan jika tetap' : '••••••••'}
              minlength={editingUser && !password ? undefined : 6}
              required={!editingUser}
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>
        </div>

        <div>
          <label for="u-curr" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Mata Uang Default</label>
          <select
            id="u-curr"
            bind:value={currency}
            class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
          >
            <option value="IDR">IDR — Rupiah Indonesia</option>
            <option value="USD">USD — US Dollar</option>
            <option value="SGD">SGD — Singapore Dollar</option>
            <option value="MYR">MYR — Malaysian Ringgit</option>
          </select>
        </div>

        <div class="flex items-center justify-end gap-2.5 pt-2">
          <button
            type="button"
            onclick={() => (modalOpen = false)}
            class="px-4 py-2.5 rounded-xl bg-[#121212] border border-[#2E2E34] text-xs font-semibold text-[#A1A1AA] hover:text-[#F4F4F5] transition-colors"
          >
            Batal
          </button>
          <button
            type="submit"
            disabled={saving}
            class="px-5 py-2.5 rounded-xl bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-semibold transition-colors disabled:opacity-50"
          >
            {saving ? 'Menyimpan...' : editingUser ? 'Perbarui Pengguna' : 'Simpan Pengguna'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
