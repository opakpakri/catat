<script>
  import '../app.css';
  import '$lib/stores/themeStore.js';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api.js';
  import { authStore } from '$lib/stores/authStore.js';
  import { periodStore } from '$lib/stores/periodStore.js';
  import { notificationStore } from '$lib/stores/notificationStore.js';
  import { appDataStore } from '$lib/stores/appDataStore.js';
  import { startProductTour, hasCompletedTour } from '$lib/utils/productTour.js';

  import SidebarNav from '$lib/components/layout/SidebarNav.svelte';
  import BottomNav from '$lib/components/layout/BottomNav.svelte';
  import TopHeader from '$lib/components/layout/TopHeader.svelte';
  import QuickAddModal from '$lib/components/modals/QuickAddModal.svelte';

  let { children } = $props();

  const currentPath = $derived(page.url.pathname);
  const isAuthPage = $derived(currentPath === '/login' || currentPath === '/register');

  // Auth route guard
  $effect(() => {
    const token = $authStore.token;
    if (!token && !isAuthPage) {
      goto('/login', { replaceState: true });
    } else if (token && (isAuthPage || currentPath === '/')) {
      const target = $authStore.user?.role === 'ADMIN' ? '/pengguna' : '/transaksi';
      goto(target, { replaceState: true });
    }
  });

  // Fetch data only when token or period/search actually changes
  let lastFetchKey = '';
  $effect(() => {
    const token = $authStore.token;
    const m = $periodStore.month;
    const y = $periodStore.year;
    const s = $periodStore.search;
    if (!token) {
      lastFetchKey = '';
      return;
    }
    const key = `${token}|${m}|${y}|${s}`;
    if (key !== lastFetchKey) {
      lastFetchKey = key;
      appDataStore.refreshAll();
    }
  });

  // Sync user profile & role once per token
  let lastSyncedToken = null;
  $effect(() => {
    const token = $authStore.token;
    if (!token) {
      lastSyncedToken = null;
      return;
    }
    if (token !== lastSyncedToken) {
      lastSyncedToken = token;
      api.get('/auth/me').then((res) => {
        if (res?.user) authStore.setUser(res.user);
      }).catch(() => {});
    }
  });

  // Push notification scheduler
  $effect(() => {
    notificationStore.init();
    const timer = setInterval(() => {
      notificationStore.checkUpcomingEvents($appDataStore.calendarEvents);
    }, 30000);
    return () => clearInterval(timer);
  });

  $effect(() => {
    if ($authStore.token && $appDataStore.calendarEvents.length > 0) {
      notificationStore.checkUpcomingEvents($appDataStore.calendarEvents);
    }
  });

  // Auto-start product tour once for new users
  let tourAutoTriggered = false;
  $effect(() => {
    if ($authStore.token && !isAuthPage && !tourAutoTriggered && !hasCompletedTour()) {
      tourAutoTriggered = true;
      startProductTour(currentPath);
    }
  });

  const modal = appDataStore.modal;
</script>

{#if !$authStore.token || isAuthPage}
  {@render children()}
{:else}
  <div class="min-h-screen bg-[#121212] text-[#F4F4F5] flex">
    <SidebarNav />

    <main class="flex-1 min-w-0 pb-20 md:pb-8">
      {#if currentPath.startsWith('/transaksi') || currentPath.startsWith('/statistik')}
        <TopHeader showPeriod={true} onStartTour={() => startProductTour(currentPath)} />
      {:else if currentPath.startsWith('/akun')}
        <TopHeader showPeriod={false} title="Akun & Utang Piutang" onStartTour={() => startProductTour(currentPath)} />
      {:else if currentPath.startsWith('/pengguna')}
        <TopHeader showPeriod={false} title="Manajemen Pengguna" onStartTour={() => startProductTour(currentPath)} />
      {:else}
        <TopHeader showPeriod={false} title="Lainnya & Pengaturan" onStartTour={() => startProductTour(currentPath)} />
      {/if}

      <div class="max-w-lg md:max-w-4xl mx-auto border-x border-[#2E2E34]/50 md:border-x-0 relative">
        {@render children()}
      </div>
    </main>

    <BottomNav />

    <QuickAddModal
      open={$modal.quickAddOpen}
      initialMode={$modal.quickAddMode}
      initialDate={$modal.quickAddDate}
      wallets={$appDataStore.wallets}
      categories={$appDataStore.categories}
      editingTx={$modal.editingTx}
      onClose={() => appDataStore.modal.closeQuickAdd()}
      onSaved={() => appDataStore.refreshAll()}
    />
  </div>
{/if}
