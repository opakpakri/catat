<script>
  import { Plus } from 'lucide-svelte';
  import { authStore } from '$lib/stores/authStore.js';
  import { periodStore } from '$lib/stores/periodStore.js';
  import { appDataStore } from '$lib/stores/appDataStore.js';

  import SummaryBar from '$lib/components/layout/SummaryBar.svelte';
  import DailyList from '$lib/components/views/DailyList.svelte';
  import MonthlyCalendar from '$lib/components/views/MonthlyCalendar.svelte';
  import TotalSummary from '$lib/components/views/TotalSummary.svelte';
  import NoteEventsList from '$lib/components/views/NoteEventsList.svelte';
  import DayDetailModal from '$lib/components/modals/DayDetailModal.svelte';

  let transSubTab = $state('daily'); // 'daily' | 'monthly' | 'total' | 'note'
  let dayDetailOpen = $state(false);
  let selectedDateStr = $state('');

  function openDayDetail(dateStr) {
    selectedDateStr = dateStr;
    dayDetailOpen = true;
  }
</script>

<svelte:head>
  <title>Transaksi — Catat</title>
</svelte:head>

<SummaryBar
  income={$appDataStore.summary.income}
  expense={$appDataStore.summary.expense}
  total={$appDataStore.summary.total}
  netWorth={$appDataStore.netWorth}
  userName={$authStore.user?.full_name || ''}
/>

<!-- Sub-Tab Navigation (Daily | Monthly | Total | Note) -->
<div data-tour="trans-subtabs" class="grid grid-cols-4 bg-[#1E1E22] border-b border-[#2E2E34]">
  <button
    onclick={() => (transSubTab = 'daily')}
    class="py-2.5 text-xs font-semibold transition-colors {transSubTab === 'daily'
      ? 'border-b-2 border-[#F05A4F] text-white'
      : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
  >
    Daily
  </button>
  <button
    onclick={() => (transSubTab = 'monthly')}
    class="py-2.5 text-xs font-semibold transition-colors {transSubTab === 'monthly'
      ? 'border-b-2 border-[#F05A4F] text-white'
      : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
  >
    Monthly
  </button>
  <button
    onclick={() => (transSubTab = 'total')}
    class="py-2.5 text-xs font-semibold transition-colors {transSubTab === 'total'
      ? 'border-b-2 border-[#F05A4F] text-white'
      : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
  >
    Total
  </button>
  <button
    onclick={() => (transSubTab = 'note')}
    class="py-2.5 text-xs font-semibold transition-colors {transSubTab === 'note'
      ? 'border-b-2 border-[#F05A4F] text-white'
      : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
  >
    Note
  </button>
</div>

{#if transSubTab === 'daily'}
  <DailyList
    transactions={$appDataStore.transactions}
    onEdit={(t) => appDataStore.modal.openQuickAdd('', t)}
    onRefresh={() => appDataStore.refreshAll()}
  />
{:else if transSubTab === 'monthly'}
  <MonthlyCalendar
    month={$periodStore.month}
    year={$periodStore.year}
    days={$appDataStore.calendarDays}
    events={$appDataStore.calendarEvents}
    onSelectDate={(d) => openDayDetail(d)}
  />
{:else if transSubTab === 'total'}
  <TotalSummary summary={$appDataStore.summary} budgets={$appDataStore.budgets} />
{:else}
  <NoteEventsList
    events={$appDataStore.calendarEvents}
    search={$periodStore.search}
    onAddEvent={() => appDataStore.modal.openQuickAdd('', null, 'EVENT')}
    onRefresh={() => appDataStore.refreshAll()}
  />
{/if}

<!-- Floating Action Button (+) -->
<button
  data-tour="fab-quick-add"
  onclick={() => appDataStore.modal.openQuickAdd('', null)}
  class="fixed bottom-20 md:bottom-6 right-5 md:right-8 z-20 w-14 h-14 rounded-full bg-[#F05A4F] hover:bg-[#D9483E] text-white shadow-lg shadow-[#F05A4F]/30 flex items-center justify-center active:scale-95 transition-all"
  aria-label="Catat transaksi atau acara"
>
  <Plus class="w-7 h-7" />
</button>

<DayDetailModal
  open={dayDetailOpen}
  dateStr={selectedDateStr}
  transactions={$appDataStore.transactions}
  events={$appDataStore.calendarEvents}
  onClose={() => (dayDetailOpen = false)}
  onAddNew={(d) => {
    dayDetailOpen = false;
    appDataStore.modal.openQuickAdd(d, null);
  }}
  onRefresh={() => appDataStore.refreshAll()}
/>
