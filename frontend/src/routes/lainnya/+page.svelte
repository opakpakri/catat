<script>
  import { appDataStore } from '$lib/stores/appDataStore.js';
  import MoreView from '$lib/components/views/MoreView.svelte';
</script>

<svelte:head>
  <title>Lainnya &amp; Pengaturan — Catat</title>
</svelte:head>

<MoreView
  categories={$appDataStore.categories}
  onRefresh={() => appDataStore.refreshAll()}
/>
