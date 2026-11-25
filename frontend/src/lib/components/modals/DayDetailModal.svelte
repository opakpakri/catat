<script>
  import { X, Plus, CheckCircle2, Circle, Trash2, Calendar } from 'lucide-svelte';
  import { api } from '../../api.js';
  import { formatIDR } from '../../utils/formatCurrency.js';

  let {
    open = false,
    dateStr = '',
    transactions = [],
    events = [],
    onClose = () => {},
    onAddNew = () => {},
    onRefresh = () => {}
  } = $props();

  let dayTxs = $derived(
    transactions.filter((t) => (t.transaction_date || '').slice(0, 10) === dateStr)
  );
  let dayEvents = $derived(
    events.filter((ev) => ev.event_date === dateStr)
  );

  async function toggleEvent(ev) {
    if (ev.is_completed) return;
    await api.patch(`/events/${ev.id}/toggle`);
    onRefresh();
  }

  async function deleteEvent(id) {
    await api.del(`/events/${id}`);
    onRefresh();
  }

  async function deleteTx(id) {
    await api.del(`/transactions/${id}`);
    onRefresh();
  }
</script>

{#if open}
  <div class="fixed inset-0 z-40 flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
    <div class="w-full max-w-md bg-[#1E1E22] border border-[#2E2E34] rounded-t-2xl sm:rounded-2xl p-5 max-h-[85vh] flex flex-col">
      <div class="flex items-center justify-between pb-3 border-b border-[#2E2E34]">
        <div class="flex items-center gap-2">
          <Calendar class="w-4 h-4 text-[#F05A4F]" />
          <h3 class="text-base font-semibold text-[#F4F4F5]">{dateStr}</h3>
        </div>
        <div class="flex items-center gap-2">
          <button
            onclick={() => onAddNew(dateStr)}
            class="flex items-center gap-1 px-2.5 py-1 rounded-lg bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-medium"
          >
            <Plus class="w-3.5 h-3.5" />
            Tambah
          </button>
          <button onclick={onClose} class="p-1 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5]">
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <div class="overflow-y-auto py-3 space-y-4">
        <!-- Events / Reminders -->
        <div>
          <h4 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA] mb-2">
            Reminder & Acara ({dayEvents.length})
          </h4>
          {#if dayEvents.length === 0}
            <p class="text-xs text-[#A1A1AA]/70 italic">Tidak ada acara pada tanggal ini.</p>
          {:else}
            <div class="space-y-2">
              {#each dayEvents as ev}
                <div class="flex items-center justify-between gap-2 bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5">
                  <button
                    onclick={() => toggleEvent(ev)}
                    disabled={ev.is_completed}
                    class="flex items-center gap-2.5 text-left flex-1 min-w-0 {ev.is_completed ? 'cursor-default' : 'cursor-pointer'}"
                  >
                    {#if ev.is_completed}
                      <CheckCircle2 class="w-4 h-4 text-[#10B981] shrink-0" />
                    {:else}
                      <Circle class="w-4 h-4 text-[#A1A1AA] shrink-0" />
                    {/if}
                    <div class="min-w-0">
                      <p class="text-sm font-medium truncate {ev.is_completed ? 'line-through text-[#A1A1AA]' : 'text-[#F4F4F5]'}">
                        {ev.title}
                      </p>
                      {#if ev.event_time || ev.description}
                        <p class="text-xs text-[#A1A1AA] truncate">
                          {ev.event_time ? ev.event_time + ' • ' : ''}{ev.description}
                        </p>
                      {/if}
                    </div>
                  </button>
                  <button onclick={() => deleteEvent(ev.id)} class="p-1 text-[#A1A1AA] hover:text-[#EF4444]">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Transactions -->
        <div>
          <h4 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA] mb-2">
            Transaksi ({dayTxs.length})
          </h4>
          {#if dayTxs.length === 0}
            <p class="text-xs text-[#A1A1AA]/70 italic">Belum ada transaksi pada tanggal ini.</p>
          {:else}
            <div class="space-y-2">
              {#each dayTxs as t}
                <div class="flex items-center justify-between bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5">
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-[#F4F4F5] truncate">
                      {t.type === 'TRANSFER' ? `Transfer → ${t.target_wallet_name}` : (t.category_name || 'Tanpa Kategori')}
                    </p>
                    <p class="text-xs text-[#A1A1AA] truncate">
                      {t.wallet_name}{t.note ? ` • ${t.note}` : ''}
                    </p>
                  </div>
                  <div class="flex items-center gap-2 shrink-0">
                    <span
                      class="text-sm font-semibold tabular-nums {t.type === 'INCOME'
                        ? 'text-[#3B82F6]'
                        : t.type === 'EXPENSE'
                          ? 'text-[#EF4444]'
                          : 'text-[#F8FAFC]'}"
                    >
                      {t.type === 'INCOME' ? '+' : t.type === 'EXPENSE' ? '-' : ''}{formatIDR(t.amount)}
                    </span>
                    <button onclick={() => deleteTx(t.id)} class="p-1 text-[#A1A1AA] hover:text-[#EF4444]">
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
