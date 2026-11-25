<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/authStore.js';
  import { appDataStore } from '$lib/stores/appDataStore.js';
  import { startProductTour } from '$lib/utils/productTour.js';
  import { ReceiptText, BarChart3, Wallet, Ellipsis, Plus, LogOut, HelpCircle, User, UserCog } from 'lucide-svelte';

  const user = $derived($authStore.user);
  const isAdmin = $derived(user?.role === 'ADMIN');
  const currentPath = $derived(page.url.pathname);

  const navItems = $derived(
    [
      { href: '/transaksi', label: 'Transaksi', icon: ReceiptText },
      { href: '/statistik', label: 'Statistik', icon: BarChart3 },
      { href: '/akun', label: 'Akun', icon: Wallet },
      { href: '/pengguna', label: 'Pengguna', icon: UserCog, adminOnly: true },
      { href: '/lainnya', label: 'Lainnya', icon: Ellipsis }
    ].filter((item) => !item.adminOnly || isAdmin)
  );

  function isActive(href) {
    return currentPath === href || currentPath.startsWith(href + '/');
  }

  function handleLogout() {
    authStore.logout();
    goto('/login', { replaceState: true });
  }
</script>

<!-- Desktop Sticky Sidebar (Coral-Red Gradient with White Active Pill, matching Cermat layout) -->
<aside
  class="hidden w-60 shrink-0 flex-col bg-gradient-to-b from-[#F05A4F] via-[#D9483E] to-[#B91C1C] text-white shadow-sm md:sticky md:top-0 md:flex md:h-screen"
>
  <!-- Branding Header -->
  <div class="flex items-center justify-between border-b border-white/15 px-4 py-4 select-text">
    <a href="/" class="flex items-center gap-2.5">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white/20 text-white backdrop-blur-sm">
        <Wallet class="h-5 w-5" />
      </div>
      <div class="min-w-0 select-text">
        <p class="text-base font-extrabold tracking-tight text-white leading-tight">Catat</p>
        <p class="text-[10px] font-medium text-red-100 leading-snug">
          Cellix Aplikasi Transaksi &amp; Arus Tunai
        </p>
      </div>
    </a>
  </div>

  <!-- Active Tenant User Card -->
  <div class="px-3 pt-3">
    <div class="rounded-xl border border-white/15 bg-white/10 p-3 backdrop-blur-sm">
      <div class="flex items-center gap-2">
        <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-white/20 text-white">
          <User class="h-3.5 w-3.5" />
        </div>
        <div class="min-w-0">
          <p class="truncate text-xs font-bold text-white">
            {user?.full_name || 'Catat Workspace'}
          </p>
          <p class="truncate text-[11px] text-red-100">
            {user?.email || 'Pengguna'}
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Main Navigation Items -->
  <nav data-tour="sidebar-nav" class="flex-1 space-y-1 overflow-y-auto px-3 py-3">
    {#each navItems as item}
      {@const active = isActive(item.href)}
      {@const Icon = item.icon}
      <a
        href={item.href}
        class="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm font-semibold transition {active
          ? 'bg-white text-[#D9483E] shadow-sm'
          : 'text-red-100 hover:bg-white/10 hover:text-white'}"
      >
        <Icon class="h-4 w-4 shrink-0" />
        <span class="truncate">{item.label}</span>
      </a>
    {/each}

    <div class="pt-2">
      <button
        type="button"
        data-tour="sidebar-quick-add"
        onclick={() => appDataStore.modal.openQuickAdd('', null)}
        class="flex w-full items-center gap-2.5 rounded-xl border border-white/15 bg-white/15 px-3 py-2 text-sm font-semibold text-white transition hover:bg-white/25"
      >
        <Plus class="h-4 w-4 shrink-0" />
        <span class="truncate">Catat Baru</span>
      </button>
    </div>
  </nav>

  <!-- Panduan & Logout Footer -->
  <div class="space-y-1 border-t border-white/15 p-3">
    <button
      type="button"
      onclick={() => startProductTour(currentPath)}
      class="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm font-semibold text-red-100 transition hover:bg-white/10 hover:text-white"
    >
      <HelpCircle class="h-4 w-4 shrink-0" />
      <span>Panduan Halaman</span>
    </button>
    <button
      type="button"
      onclick={handleLogout}
      class="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm font-semibold text-red-100 transition hover:bg-black/20 hover:text-white"
    >
      <LogOut class="h-4 w-4 shrink-0" />
      <span>Keluar</span>
    </button>
  </div>
</aside>
