<script>
  import { Plus, Trash2 } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import { formatIDR } from '$lib/utils/formatCurrency.js';

  let { month = 1, year = 2026, categories = [], onRefresh = () => {} } = $props();

  let report = $state({ summary: { income: 0, expense: 0, total: 0 }, expense_by_category: [], monthly_trend: [] });
  let budgets = $state([]);
  let showBudgetForm = $state(false);
  let selectedCatId = $state('');
  let budgetLimit = $state('');

  const MONTH_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];

  async function loadStats() {
    try {
      const [rep, bdg] = await Promise.all([
        api.get(`/reports/summary?month=${month}&year=${year}`),
        api.get(`/budgets?month=${month}&year=${year}`)
      ]);
      report = rep;
      budgets = bdg.budgets || [];
    } catch {
      // handled gracefully
    }
  }

  $effect(() => {
    month;
    year;
    loadStats();
  });

  let expenseCategories = $derived(categories.filter((c) => c.type === 'EXPENSE'));

  let donutGradient = $derived.by(() => {
    const items = report.expense_by_category || [];
    if (items.length === 0) return 'conic-gradient(#2E2E34 0% 100%)';
    let acc = 0;
    const stops = items.map((item) => {
      const start = acc;
      acc += item.percentage;
      return `${item.color_hex || '#F05A4F'} ${start.toFixed(2)}% ${acc.toFixed(2)}%`;
    });
    return `conic-gradient(${stops.join(', ')})`;
  });

  let maxTrend = $derived.by(() => {
    let m = 1;
    for (const pt of report.monthly_trend || []) {
      m = Math.max(m, pt.income, pt.expense);
    }
    return m;
  });

  async function saveBudget(e) {
    e.preventDefault();
    if (!selectedCatId || !Number(budgetLimit)) return;
    await api.post('/budgets', {
      category_id: Number(selectedCatId),
      amount_limit: Number(budgetLimit),
      period_month: month,
      period_year: year
    });
    showBudgetForm = false;
    budgetLimit = '';
    loadStats();
    onRefresh();
  }

  async function deleteBudget(id) {
    if (!confirm('Hapus anggaran ini?')) return;
    await api.del(`/budgets/${id}`);
    loadStats();
    onRefresh();
  }
</script>

<div class="p-4 space-y-4">
  <!-- Donut Chart Card -->
  <div data-tour="stats-donut" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA] mb-4">
      Pengeluaran per Kategori
    </h3>

    {#if (report.expense_by_category || []).length === 0}
      <p class="text-xs text-[#A1A1AA]/70 text-center py-8">Belum ada pengeluaran pada bulan ini.</p>
    {:else}
      <div class="flex flex-col sm:flex-row items-center gap-6">
        <div
          class="relative w-36 h-36 rounded-full shrink-0 flex items-center justify-center"
          style="background: {donutGradient}"
        >
          <div class="w-24 h-24 rounded-full bg-[#1E1E22] flex flex-col items-center justify-center px-2 text-center">
            <span class="text-[10px] text-[#A1A1AA]">Total Keluar</span>
            <span class="text-xs font-bold text-[#F8FAFC] tabular-nums truncate max-w-full">
              {formatIDR(report.summary.expense, true)}
            </span>
          </div>
        </div>

        <div class="flex-1 w-full space-y-2">
          {#each report.expense_by_category as item}
            <div class="flex items-center justify-between text-xs">
              <div class="flex items-center gap-2 min-w-0">
                <span class="w-2.5 h-2.5 rounded-full shrink-0" style="background-color: {item.color_hex}"></span>
                <span class="text-[#F4F4F5] truncate">{item.name}</span>
              </div>
              <div class="flex items-center gap-2 shrink-0 tabular-nums">
                <span class="text-[#A1A1AA]">{item.percentage.toFixed(1)}%</span>
                <span class="font-semibold text-[#F4F4F5]">{formatIDR(item.total)}</span>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <!-- Bar Chart Monthly Trend -->
  <div data-tour="stats-trend" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA] mb-4">
      Tren Arus Kas (6 Bulan Terakhir)
    </h3>
    {#if (report.monthly_trend || []).length === 0}
      <p class="text-xs text-[#A1A1AA]/70 text-center py-6">Belum ada data tren.</p>
    {:else}
      <div class="flex items-end justify-around gap-2 h-32 pt-2">
        {#each report.monthly_trend as pt}
          {@const incH = Math.round((pt.income / maxTrend) * 100)}
          {@const expH = Math.round((pt.expense / maxTrend) * 100)}
          <div class="flex flex-col items-center gap-1.5 flex-1">
            <div class="flex items-end gap-1 h-24 w-full justify-center">
              <div
                class="w-3 rounded-t bg-[#3B82F6] transition-all"
                style="height: {Math.max(4, incH)}%"
                title="Pendapatan: {formatIDR(pt.income)}"
              ></div>
              <div
                class="w-3 rounded-t bg-[#EF4444] transition-all"
                style="height: {Math.max(4, expH)}%"
                title="Pengeluaran: {formatIDR(pt.expense)}"
              ></div>
            </div>
            <span class="text-[10px] text-[#A1A1AA]">{MONTH_SHORT[pt.month - 1]}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Budgeting Card -->
  <div data-tour="stats-budgets" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA]">
        Anggaran (Budgeting)
      </h3>
      <button
        onclick={() => {
          showBudgetForm = !showBudgetForm;
          if (!selectedCatId && expenseCategories[0]) selectedCatId = String(expenseCategories[0].id);
        }}
        class="flex items-center gap-1 px-2.5 py-1 rounded-lg bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-medium"
      >
        <Plus class="w-3.5 h-3.5" />
        Atur Budget
      </button>
    </div>

    {#if showBudgetForm}
      <form onsubmit={saveBudget} class="bg-[#121212] border border-[#2E2E34] rounded-xl p-3 mb-4 space-y-2.5">
        <select
          bind:value={selectedCatId}
          class="w-full bg-[#1E1E22] border border-[#2E2E34] rounded-lg px-3 py-2 text-xs text-[#F4F4F5]"
        >
          {#each expenseCategories as cat}
            <option value={String(cat.id)}>{cat.name}</option>
          {/each}
        </select>
        <input
          type="number"
          min="1"
          bind:value={budgetLimit}
          placeholder="Batas maksimal (IDR)"
          required
          class="w-full bg-[#1E1E22] border border-[#2E2E34] rounded-lg px-3 py-2 text-xs text-[#F4F4F5]"
        />
        <button type="submit" class="w-full py-2 rounded-lg bg-[#10B981] text-white text-xs font-semibold">
          Simpan Anggaran
        </button>
      </form>
    {/if}

    {#if budgets.length === 0}
      <p class="text-xs text-[#A1A1AA]/70">Belum ada batas anggaran bulan ini.</p>
    {:else}
      <div class="space-y-3.5">
        {#each budgets as b}
          {@const pct = Math.min(100, Math.round((b.spent_amount / b.amount_limit) * 100))}
          {@const status = b.spent_amount > b.amount_limit ? 'Overlimit' : pct >= 80 ? 'Mendekati Limit' : 'Aman'}
          <div>
            <div class="flex justify-between items-center text-xs mb-1">
              <span class="font-medium text-[#F4F4F5]">{b.category_name}</span>
              <div class="flex items-center gap-2">
                <span class="px-1.5 py-0.5 rounded text-[10px] font-medium {status === 'Overlimit'
                  ? 'bg-[#EF4444]/20 text-[#EF4444]'
                  : status === 'Mendekati Limit'
                    ? 'bg-[#F59E0B]/20 text-[#F59E0B]'
                    : 'bg-[#10B981]/20 text-[#34D399]'}">
                  {status}
                </span>
                <button
                  onclick={() => deleteBudget(b.id)}
                  class="p-1 text-[#A1A1AA] hover:text-[#EF4444] rounded"
                  title="Hapus anggaran"
                  aria-label="Hapus anggaran"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <div class="flex justify-between text-[11px] text-[#A1A1AA] tabular-nums mb-1">
              <span>Terpakai: {formatIDR(b.spent_amount)}</span>
              <span>Limit: {formatIDR(b.amount_limit)}</span>
            </div>
            <div class="h-2 rounded-full bg-[#121212] overflow-hidden">
              <div
                class="h-full rounded-full {status === 'Overlimit'
                  ? 'bg-[#EF4444]'
                  : status === 'Mendekati Limit'
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
