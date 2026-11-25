<script>
  import { goto } from '$app/navigation';
  import { Wallet, ShieldCheck } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import { authStore } from '$lib/stores/authStore.js';

  let email = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  async function handleSubmit(e) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      const res = await api.post('/auth/login', { email, password });
      authStore.setAuth(res.token, res.user);
      if (res.user?.role === 'ADMIN') {
        goto('/pengguna', { replaceState: true });
      } else {
        goto('/transaksi', { replaceState: true });
      }
    } catch (err) {
      error = err.message || 'Autentikasi gagal';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen bg-[#121212] flex flex-col items-center justify-center px-5 py-10">
  <div class="w-full max-w-sm bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-6 shadow-xl">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-11 h-11 rounded-xl bg-[#F05A4F] flex items-center justify-center text-white">
        <Wallet class="w-6 h-6" />
      </div>
      <div>
        <h1 class="text-xl font-bold tracking-tight text-[#F4F4F5]">Catat</h1>
        <p class="text-xs text-[#A1A1AA]">Cellix Aplikasi Transaksi &amp; Arus Tunai</p>
      </div>
    </div>

    {#if error}
      <div class="mb-4 px-3 py-2 rounded-lg bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">
        {error}
      </div>
    {/if}

    <form onsubmit={handleSubmit} class="space-y-3.5">
      <div>
        <label for="login-email" class="block text-xs text-[#A1A1AA] mb-1">Email</label>
        <input
          id="login-email"
          type="email"
          bind:value={email}
          placeholder="nama@email.com"
          required
          class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
        />
      </div>

      <div>
        <label for="login-password" class="block text-xs text-[#A1A1AA] mb-1">Password</label>
        <input
          id="login-password"
          type="password"
          bind:value={password}
          placeholder="Masukkan password"
          minlength="6"
          required
          class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        class="w-full py-3 rounded-xl bg-[#F05A4F] hover:bg-[#D9483E] text-white font-semibold text-sm transition-colors disabled:opacity-50"
      >
        {loading ? 'Memproses...' : 'Masuk ke Catat'}
      </button>
    </form>

    <div class="mt-5 pt-4 border-t border-[#2E2E34] flex items-center gap-2 text-[11px] text-[#A1A1AA]">
      <ShieldCheck class="w-4 h-4 text-[#10B981] shrink-0" />
      <span>Pembuatan akun baru hanya dilakukan oleh Admin.</span>
    </div>
  </div>
</div>
