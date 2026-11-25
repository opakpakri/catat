<script>
  import { formatIDR } from '../../utils/formatCurrency.js';

  let { summary = { income: 0, expense: 0, total: 0 }, budgets = [] } = $props();
</script>

<div class="p-4 space-y-4">
  <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA]">Ringkasan Periode</h3>
    <div class="flex justify-between items-center py-1.5 border-b border-[#2E2E34]">
      <span class="text-sm text-[#A1A1AA]">Total Pemasukan</span>
      <span class="text-sm font-semibold text-[#3B82F6] tabular-nums">{formatIDR(summary.income)}</span>
    </div>
    <div class="flex justify-between items-center py-1.5 border-b border-[#2E2E34]">
      <span class="text-sm text-[#A1A1AA]">Total Pengeluaran</span>
      <span class="text-sm font-semibold text-[#EF4444] tabular-nums">{formatIDR(summary.expense)}</span>
    </div>
    <div class="flex justify-between items-center pt-1">
      <span class="text-sm font-semibold text-[#F4F4F5]">Arus Kas Bersih</span>
      <span class="text-base font-bold text-[#F8FAFC] tabular-nums">{formatIDR(summary.total)}</span>
    </div>
  </div>

  <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA] mb-3">
      Progres Anggaran Aktif ({budgets.length})
    </h3>
    {#if budgets.length === 0}
      <p class="text-xs text-[#A1A1AA]/70">Belum ada anggaran bulan ini. Atur anggaran di menu Stats.</p>
    {:else}
      <div class="space-y-3.5">
        {#each budgets as b}
          {@const pct = Math.min(100, Math.round((b.spent_amount / b.amount_limit) * 100))}
          {@const isOver = b.spent_amount > b.amount_limit}
          {@const isWarn = pct >= 80 && !isOver}
          <div>
            <div class="flex justify-between text-xs mb-1">
              <span class="font-medium text-[#F4F4F5]">{b.category_name}</span>
              <span class="tabular-nums {isOver ? 'text-[#EF4444] font-semibold' : 'text-[#A1A1AA]'}">
                {formatIDR(b.spent_amount)} / {formatIDR(b.amount_limit)}
              </span>
            </div>
            <div class="h-2 rounded-full bg-[#121212] overflow-hidden">
              <div
                class="h-full rounded-full transition-all {isOver
                  ? 'bg-[#EF4444]'
                  : isWarn
                    ? 'bg-[#F59E0B]'
                    : 'bg-[#10B981]'}"
                style="width: {pct}%"
              ></div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
