<script>
  import { appDataStore } from '$lib/stores/appDataStore.js';
  import AccountsView from '$lib/components/views/AccountsView.svelte';
</script>

<svelte:head>
  <title>Akun &amp; Utang Piutang — Catat</title>
</svelte:head>

<AccountsView
  wallets={$appDataStore.wallets}
  netWorth={$appDataStore.netWorth}
  onRefresh={() => appDataStore.refreshAll()}
/>
