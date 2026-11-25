<script>
  import { buildCalendarGrid } from '../../utils/calendarGrid.js';
  import { formatIDR } from '../../utils/formatCurrency.js';

  let { month = 1, year = 2026, days = [], events = [], onSelectDate = () => {} } = $props();

  const weekDays = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
  const todayStr = new Date().toISOString().slice(0, 10);

  let grid = $derived(buildCalendarGrid(year, month));

  let dayMap = $derived.by(() => {
    const m = new Map();
    for (const d of days) m.set(d.date, d);
    return m;
  });

  let eventMap = $derived.by(() => {
    const m = new Map();
    for (const ev of events) {
      if (!m.has(ev.event_date)) m.set(ev.event_date, []);
      m.get(ev.event_date).push(ev);
    }
    return m;
  });
</script>

<div class="bg-[#121212]">
  <!-- Weekday Header -->
  <div class="grid grid-cols-7 border-b border-[#2E2E34] bg-[#1E1E22]/50">
    {#each weekDays as wd, idx}
      <div
        class="py-2 text-center text-[11px] font-semibold uppercase tracking-wider {idx === 0
          ? 'text-[#EF4444]'
          : idx === 6
            ? 'text-[#3B82F6]'
            : 'text-[#A1A1AA]'}"
      >
        {wd}
      </div>
    {/each}
  </div>

  <!-- Calendar Grid -->
  <div class="grid grid-cols-7 border-l border-[#2E2E34]">
    {#each grid as cell}
      {@const agg = dayMap.get(cell.dateStr)}
      {@const cellEvents = eventMap.get(cell.dateStr) || []}
      {@const isToday = cell.dateStr === todayStr}
      <button
        onclick={() => onSelectDate(cell.dateStr)}
        class="min-h-[96px] sm:min-h-[112px] p-1.5 border-r border-b border-[#2E2E34] flex flex-col justify-between text-left transition-colors hover:bg-[#1E1E22] {cell.inCurrentMonth
          ? 'bg-[#121212]'
          : 'bg-[#121212]/40 opacity-45'}"
      >
        <div class="flex items-center justify-between w-full">
          <span
            class="text-[11px] font-medium px-1.5 py-0.5 rounded-full {isToday
              ? 'bg-[#F05A4F] text-white font-bold'
              : 'text-[#A1A1AA]'}"
          >
            {cell.day}
          </span>
          {#if cellEvents.length > 0}
            <span class="px-1 py-0.2 rounded text-[9px] font-semibold bg-[#10B981]/20 text-[#34D399]">
              {cellEvents.length}
            </span>
          {/if}
        </div>

        <!-- Event pill preview -->
        {#if cellEvents.length > 0}
          <div class="w-full truncate text-[9px] px-1 py-0.5 rounded bg-[#10B981]/15 text-[#34D399] my-0.5">
            {cellEvents[0].title}
          </div>
        {/if}

        <!-- Daily Income & Expense numbers -->
        <div class="w-full flex flex-col items-end leading-tight tabular-nums">
          {#if agg?.income > 0}
            <span class="text-[9px] sm:text-[10px] font-medium text-[#3B82F6] truncate max-w-full">
              {formatIDR(agg.income, true)}
            </span>
          {/if}
          {#if agg?.expense > 0}
            <span class="text-[9px] sm:text-[10px] font-medium text-[#EF4444] truncate max-w-full">
              -{formatIDR(agg.expense, true)}
            </span>
          {/if}
        </div>
      </button>
    {/each}
  </div>
</div>
