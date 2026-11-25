<script>
  import { periodStore } from '$lib/stores/periodStore.js';
  import { appDataStore } from '$lib/stores/appDataStore.js';
  import StatsView from '$lib/components/views/StatsView.svelte';
</script>

<svelte:head>
  <title>Statistik — Catat</title>
</svelte:head>

<StatsView
  month={$periodStore.month}
  year={$periodStore.year}
  categories={$appDataStore.categories}
  onRefresh={() => appDataStore.refreshAll()}
/>
