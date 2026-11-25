<script>
  import { goto } from '$app/navigation';
  import { Wallet, ShieldCheck, TrendingUp, CalendarCheck2, Mail, Lock, Eye, EyeOff, ArrowRight } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import { authStore } from '$lib/stores/authStore.js';

  let email = $state('');
  let password = $state('');
  let showPassword = $state(false);
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

<div class="min-h-screen flex bg-[#121212]">
  <!-- Left Brand Panel (Desktop) -->
  <div class="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-[#F05A4F] via-[#E54E43] to-[#C83B31] text-white flex-col justify-center px-14 xl:px-20">
    <div class="max-w-md space-y-6">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-white/20 backdrop-blur-sm flex items-center justify-center shrink-0">
          <Wallet class="w-5 h-5 text-white" />
        </div>
        <span class="text-sm font-semibold text-white/95">
          CATAT — Cellix Aplikasi Transaksi &amp; Arus Tunai
        </span>
      </div>

      <h2 class="text-4xl font-extrabold tracking-tight leading-[1.15] text-white">
        Arus kas Anda, rapi di mana saja.
      </h2>

      <ul class="space-y-3.5 pt-2 text-sm text-white/90">
        <li class="flex items-center gap-3">
          <TrendingUp class="w-4 h-4 text-white/85 shrink-0" />
          <span>Pantau pemasukan, pengeluaran &amp; saldo secara langsung</span>
        </li>
        <li class="flex items-center gap-3">
          <CalendarCheck2 class="w-4 h-4 text-white/85 shrink-0" />
          <span>Kelola kalender arus kas, anggaran, dan utang piutang</span>
        </li>
        <li class="flex items-center gap-3">
          <ShieldCheck class="w-4 h-4 text-white/85 shrink-0" />
          <span>Data setiap akun terisolasi dan aman</span>
        </li>
      </ul>
    </div>
  </div>

  <!-- Right Login Form Panel -->
  <div class="flex-1 flex items-center justify-center px-5 py-10 bg-[#121212]">
    <div class="w-full max-w-md bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-7 sm:p-8 shadow-xl">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 rounded-xl bg-[#F05A4F] flex items-center justify-center text-white shrink-0">
          <Wallet class="w-5 h-5" />
        </div>
        <div>
          <h1 class="text-lg font-extrabold tracking-tight text-[#F05A4F] leading-tight">Catat</h1>
          <p class="text-xs text-[#A1A1AA]">Cellix Aplikasi Transaksi &amp; Arus Tunai</p>
        </div>
      </div>

      <div class="mb-6">
        <h2 class="text-2xl font-extrabold tracking-tight text-[#F4F4F5]">Masuk ke Workspace</h2>
        <p class="text-xs text-[#A1A1AA] mt-1">
          Kelola transaksi, kalender arus kas, dan anggaran Anda.
        </p>
      </div>

      {#if error}
        <div class="mb-4 px-3 py-2 rounded-xl bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-4">
        <div>
          <label for="login-email" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Email</label>
          <div class="relative">
            <Mail class="w-4 h-4 text-[#A1A1AA] absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
            <input
              id="login-email"
              type="email"
              bind:value={email}
              placeholder="nama@perusahaan.com"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl pl-10 pr-3.5 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>
        </div>

        <div>
          <label for="login-password" class="block text-xs font-semibold text-[#F4F4F5] mb-1.5">Password</label>
          <div class="relative">
            <Lock class="w-4 h-4 text-[#A1A1AA] absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
            <input
              id="login-password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              placeholder="••••••••"
              minlength="6"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl pl-10 pr-10 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
            <button
              type="button"
              onclick={() => (showPassword = !showPassword)}
              class="absolute right-3 top-1/2 -translate-y-1/2 text-[#A1A1AA] hover:text-[#F4F4F5]"
              aria-label={showPassword ? 'Sembunyikan password' : 'Lihat password'}
            >
              {#if showPassword}
                <Eye class="w-4 h-4" />
              {:else}
                <EyeOff class="w-4 h-4" />
              {/if}
            </button>
          </div>
        </div>

        <button
          type="submit"
          disabled={loading}
          class="w-full py-3 rounded-xl bg-[#F05A4F] hover:bg-[#D9483E] text-white font-semibold text-sm flex items-center justify-center gap-2 transition-colors disabled:opacity-50"
        >
          <span>{loading ? 'Memproses...' : 'Masuk'}</span>
          {#if !loading}
            <ArrowRight class="w-4 h-4" />
          {/if}
        </button>
      </form>

      <p class="mt-6 text-center text-[11px] text-[#A1A1AA]">
        Pendaftaran akun baru hanya dapat dibuat <span class="text-[#F4F4F5] font-medium">oleh Admin.</span>
      </p>
    </div>
  </div>
</div>
