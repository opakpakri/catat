<script>
  import { ChevronLeft, ChevronRight, Search, X, HelpCircle } from 'lucide-svelte';
  import { periodStore } from '../../stores/periodStore.js';
  import { formatPeriodLabel } from '../../utils/formatCurrency.js';

  let { title = '', showPeriod = true, onSearchChange = () => {}, onStartTour = () => {} } = $props();
  let searchOpen = $state(false);
  let searchInput = $state('');
  let pickerOpen = $state(false);
  let pickerYear = $state(new Date().getFullYear());

  const MONTHS_GRID = [
    'JAN', 'FEB', 'MAR', 'APR',
    'MAY', 'JUN', 'JUL', 'AUG',
    'SEP', 'OCT', 'NOV', 'DEC'
  ];

  function togglePicker() {
    if (!pickerOpen) {
      pickerYear = $periodStore.year;
    }
    pickerOpen = !pickerOpen;
  }

  function selectMonth(monthIndex) {
    periodStore.setPeriod(monthIndex + 1, pickerYear);
    pickerOpen = false;
  }

  function selectThisMonth() {
    periodStore.resetToNow();
    pickerOpen = false;
  }

  function toggleSearch() {
    searchOpen = !searchOpen;
    if (!searchOpen) {
      searchInput = '';
      periodStore.setSearch('');
      onSearchChange('');
    }
  }

  function handleInput(e) {
    searchInput = e.target.value;
    periodStore.setSearch(searchInput);
    onSearchChange(searchInput);
  }
</script>

<header class="sticky top-0 z-30 bg-[#1E1E22] border-b border-[#2E2E34] w-full">
  <div class="flex items-center justify-between h-14 md:h-16 px-4 md:px-6">
    {#if showPeriod}
      <div data-tour="period-picker" class="flex items-center gap-1">
        <button
          onclick={() => periodStore.prevMonth()}
          class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60 transition-colors"
          aria-label="Bulan sebelumnya"
        >
          <ChevronLeft class="w-5 h-5" />
        </button>
        <button
          onclick={togglePicker}
          class="px-2 py-1 rounded-lg font-semibold text-base tracking-tight text-[#F4F4F5] hover:text-[#F05A4F] hover:bg-[#2E2E34]/40 transition-colors"
          title="Pilih Bulan & Tahun"
        >
          {formatPeriodLabel($periodStore.month, $periodStore.year)}
        </button>
        <button
          onclick={() => periodStore.nextMonth()}
          class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60 transition-colors"
          aria-label="Bulan berikutnya"
        >
          <ChevronRight class="w-5 h-5" />
        </button>
      </div>
    {:else}
      <h1 class="font-semibold text-base text-[#F4F4F5]">{title}</h1>
    {/if}

    <div class="flex items-center gap-1.5">
      <button
        onclick={onStartTour}
        class="md:hidden p-2 rounded-lg text-[#A1A1AA] hover:text-[#F05A4F] hover:bg-[#2E2E34]/60 transition-colors"
        title="Panduan Fitur (Product Tour)"
        aria-label="Panduan Fitur"
      >
        <HelpCircle class="w-5 h-5" />
      </button>
      {#if showPeriod}
        <button
          onclick={toggleSearch}
          class="p-2 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60 transition-colors"
          aria-label="Cari transaksi"
        >
          {#if searchOpen}
            <X class="w-5 h-5 text-[#F05A4F]" />
          {:else}
            <Search class="w-5 h-5" />
          {/if}
        </button>
      {/if}
    </div>
  </div>

  {#if searchOpen && showPeriod}
    <div class="px-4 pb-3">
      <input
        type="text"
        value={searchInput}
        oninput={handleInput}
        placeholder="Cari catatan, kategori, atau dompet..."
        class="w-full bg-[#121212] border border-[#2E2E34] rounded-lg px-3 py-2 text-sm text-[#F4F4F5] placeholder-[#A1A1AA] focus:outline-none focus:border-[#F05A4F]"
      />
    </div>
  {/if}
</header>

{#if pickerOpen && showPeriod}
  <!-- Backdrop -->
  <button
    type="button"
    onclick={() => (pickerOpen = false)}
    class="fixed inset-0 z-40 bg-black/50 cursor-default"
    aria-label="Tutup pemilih tanggal"
  ></button>

  <!-- Month/Year Picker Card -->
  <div class="fixed left-1/2 md:left-[calc(50%+7.5rem)] -translate-x-1/2 top-14 z-50 w-[calc(100%-2rem)] max-w-sm bg-[#1E1E22] border border-[#2E2E34] rounded-2xl shadow-2xl overflow-hidden">
    <!-- Card Header -->
    <div class="flex items-center justify-between px-4 py-3 bg-[#2E2E34]/40 border-b border-[#2E2E34]">
      <span class="text-sm font-semibold text-[#F4F4F5]">Date</span>
      <div class="flex items-center gap-3">
        <button
          type="button"
          onclick={selectThisMonth}
          class="text-xs font-bold tracking-wide text-[#F4F4F5] hover:text-[#F05A4F] transition-colors"
        >
          THIS MONTH
        </button>
        <button
          type="button"
          onclick={() => (pickerOpen = false)}
          class="p-1 text-[#A1A1AA] hover:text-[#F4F4F5]"
          aria-label="Tutup"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Year Navigation -->
    <div class="flex items-center justify-between px-6 py-3 border-b border-[#2E2E34]/60">
      <button
        type="button"
        onclick={() => (pickerYear -= 1)}
        class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60"
        aria-label="Tahun sebelumnya"
      >
        <ChevronLeft class="w-4 h-4" />
      </button>
      <span class="text-sm font-semibold text-[#F4F4F5] tabular-nums">{pickerYear}</span>
      <button
        type="button"
        onclick={() => (pickerYear += 1)}
        class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5] hover:bg-[#2E2E34]/60"
        aria-label="Tahun berikutnya"
      >
        <ChevronRight class="w-4 h-4" />
      </button>
    </div>

    <!-- 4x3 Month Grid -->
    <div class="grid grid-cols-4 gap-1 p-3">
      {#each MONTHS_GRID as mLabel, idx}
        {@const isSelected = pickerYear === $periodStore.year && idx + 1 === $periodStore.month}
        <button
          type="button"
          onclick={() => selectMonth(idx)}
          class="py-3 rounded-xl text-xs font-semibold tracking-wide transition-colors {isSelected
            ? 'text-[#F05A4F] bg-[#F05A4F]/10'
            : 'text-[#F4F4F5] hover:bg-[#2E2E34]/60'}"
        >
          {mLabel}
        </button>
      {/each}
    </div>
  </div>
{/if}
