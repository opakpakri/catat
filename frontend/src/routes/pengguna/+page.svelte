<script>
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/authStore.js';
  import UsersView from '$lib/components/views/UsersView.svelte';

  $effect(() => {
    if ($authStore.token && $authStore.user?.role !== 'ADMIN') {
      goto('/transaksi', { replaceState: true });
    }
  });
</script>

<svelte:head>
  <title>Pengguna — Catat</title>
</svelte:head>

{#if $authStore.user?.role === 'ADMIN'}
  <UsersView />
{/if}
