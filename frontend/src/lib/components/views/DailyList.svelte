<script>
  import { Trash2, Edit3, ArrowLeftRight } from 'lucide-svelte';
  import { formatIDR, formatDayHeader } from '../../utils/formatCurrency.js';
  import { api } from '../../api.js';

  let { transactions = [], onEdit = () => {}, onRefresh = () => {} } = $props();

  let groupedDays = $derived.by(() => {
    const map = new Map();
    for (const t of transactions) {
      const dateKey = (t.transaction_date || '').slice(0, 10);
      if (!map.has(dateKey)) {
        map.set(dateKey, { dateKey, income: 0, expense: 0, items: [] });
      }
      const grp = map.get(dateKey);
      if (t.type === 'INCOME') grp.income += Number(t.amount) || 0;
      else if (t.type === 'EXPENSE') grp.expense += Number(t.amount) || 0;
      grp.items.push(t);
    }
    return Array.from(map.values()).sort((a, b) => b.dateKey.localeCompare(a.dateKey));
  });

  async function removeTx(id) {
    if (!confirm('Hapus transaksi ini? Saldo dompet akan dikembalikan.')) return;
    await api.del(`/transactions/${id}`);
    onRefresh();
  }
</script>

{#if groupedDays.length === 0}
  <div class="py-16 text-center px-6">
    <p class="text-sm text-[#A1A1AA]">Belum ada transaksi pada periode ini.</p>
    <p class="text-xs text-[#A1A1AA]/70 mt-1">Tekan tombol + di kanan bawah untuk mencatat transaksi baru.</p>
  </div>
{:else}
  <div class="divide-y divide-[#2E2E34]">
    {#each groupedDays as grp}
      {@const hdr = formatDayHeader(grp.dateKey)}
      <div class="bg-[#121212]">
        <!-- Day Header -->
        <div class="flex items-center justify-between px-4 py-2 bg-[#1E1E22]/60 border-b border-[#2E2E34]/60">
          <div class="flex items-center gap-2">
            <span class="text-lg font-bold text-[#F4F4F5] tabular-nums">{hdr.dayNum}</span>
            <span class="text-[11px] font-medium px-1.5 py-0.5 rounded bg-[#2E2E34] text-[#F4F4F5]">
              {hdr.dayName}
            </span>
            <span class="text-xs text-[#A1A1AA]">{hdr.monthYear}</span>
          </div>
          <div class="flex items-center gap-4 text-xs font-semibold tabular-nums">
            <span class="text-[#3B82F6]">{formatIDR(grp.income)}</span>
            <span class="text-[#EF4444]">{grp.expense > 0 ? '-' : ''}{formatIDR(grp.expense)}</span>
          </div>
        </div>

        <!-- Day Transactions -->
        <div class="divide-y divide-[#2E2E34]/40">
          {#each grp.items as t}
            <div class="flex items-center justify-between px-4 py-3 hover:bg-[#1E1E22]/40 transition-colors">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 text-xs font-bold"
                  style="background-color: {(t.category_color || '#64748B')}22; color: {t.category_color || '#A1A1AA'}"
                >
                  {#if t.type === 'TRANSFER'}
                    <ArrowLeftRight class="w-4 h-4 text-[#F8FAFC]" />
                  {:else}
                    {(t.category_name || 'T').slice(0, 2).toUpperCase()}
                  {/if}
                </div>
                <div class="min-w-0">
                  <p class="text-sm font-medium text-[#F4F4F5] truncate">
                    {t.type === 'TRANSFER'
                      ? `${t.wallet_name} → ${t.target_wallet_name}`
                      : (t.category_name || 'Tanpa Kategori')}
                  </p>
                  <p class="text-xs text-[#A1A1AA] truncate">
                    {t.wallet_name}{t.note ? ` • ${t.note}` : ''}
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <span
                  class="text-sm font-semibold tabular-nums {t.type === 'INCOME'
                    ? 'text-[#3B82F6]'
                    : t.type === 'EXPENSE'
                      ? 'text-[#EF4444]'
                      : 'text-[#F8FAFC]'}"
                >
                  {t.type === 'EXPENSE' ? '-' : ''}{formatIDR(t.amount)}
                </span>
                <button
                  onclick={() => onEdit(t)}
                  class="p-1.5 text-[#A1A1AA] hover:text-[#F4F4F5] rounded-lg"
                  title="Edit"
                >
                  <Edit3 class="w-3.5 h-3.5" />
                </button>
                <button
                  onclick={() => removeTx(t.id)}
                  class="p-1.5 text-[#A1A1AA] hover:text-[#EF4444] rounded-lg"
                  title="Hapus"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/each}
  </div>
{/if}
