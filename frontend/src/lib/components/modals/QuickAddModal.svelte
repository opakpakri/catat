<script>
  import { untrack } from 'svelte';
  import { X, ArrowDownLeft, ArrowUpRight, ArrowLeftRight, CalendarPlus } from 'lucide-svelte';
  import { api } from '../../api.js';

  let {
    open = false,
    initialMode = 'EXPENSE',
    initialDate = '',
    wallets = [],
    categories = [],
    editingTx = null,
    onClose = () => {},
    onSaved = () => {}
  } = $props();

  let mode = $state('EXPENSE'); // INCOME | EXPENSE | TRANSFER | EVENT
  let walletId = $state('');
  let targetWalletId = $state('');
  let categoryId = $state('');
  let amount = $state('');
  let txDateOnly = $state('');
  let txTimeOnly = $state('');
  let note = $state('');

  // Event fields
  let eventTitle = $state('');
  let eventDate = $state('');
  let eventTime = $state('');
  let eventColor = $state('#10B981');

  let loading = $state(false);
  let error = $state('');

  function toLocalDate(d = new Date()) {
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  }

  function toLocalTime(d = new Date()) {
    const pad = (n) => String(n).padStart(2, '0');
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  $effect(() => {
    const isOpen = open;
    const txToEdit = editingTx;
    if (!isOpen) return;

    untrack(() => {
      error = '';
      if (txToEdit) {
        const d = new Date(txToEdit.transaction_date);
        mode = txToEdit.type;
        walletId = String(txToEdit.wallet_id);
        targetWalletId = txToEdit.target_wallet_id ? String(txToEdit.target_wallet_id) : '';
        categoryId = txToEdit.category_id ? String(txToEdit.category_id) : '';
        amount = String(txToEdit.amount);
        txDateOnly = toLocalDate(d);
        txTimeOnly = toLocalTime(d);
        note = txToEdit.note || '';
      } else {
        mode = initialMode || 'EXPENSE';
        walletId = wallets[0]?.id ? String(wallets[0].id) : '';
        targetWalletId = wallets[1]?.id ? String(wallets[1].id) : '';
        const filtered = categories.filter((c) => c.type === (mode === 'INCOME' ? 'INCOME' : 'EXPENSE'));
        categoryId = filtered[0]?.id ? String(filtered[0].id) : '';
        amount = '';
        txDateOnly = initialDate || toLocalDate();
        txTimeOnly = toLocalTime();
        note = '';
        eventTitle = '';
        eventDate = initialDate || toLocalDate();
        eventTime = toLocalTime();
        eventColor = '#10B981';
      }
    });
  });

  let filteredCategories = $derived(
    categories.filter((c) => c.type === (mode === 'INCOME' ? 'INCOME' : 'EXPENSE'))
  );

  function switchMode(nextMode) {
    if (editingTx && nextMode === 'EVENT') return;
    mode = nextMode;
    if (nextMode === 'INCOME' || nextMode === 'EXPENSE') {
      const list = categories.filter((c) => c.type === nextMode);
      categoryId = list[0]?.id ? String(list[0].id) : '';
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      if (mode === 'EVENT') {
        if (!eventTitle.trim() || !eventDate) {
          throw new Error('Judul acara dan tanggal wajib diisi');
        }
        await api.post('/events', {
          title: eventTitle.trim(),
          description: note.trim(),
          event_date: eventDate,
          event_time: eventTime || null,
          color_hex: eventColor
        });
      } else {
        const numAmount = Number(amount);
        if (!numAmount || numAmount <= 0) {
          throw new Error('Nominal harus lebih dari 0');
        }
        if (!walletId) {
          throw new Error('Pilih dompet terlebih dahulu');
        }
        if (!txDateOnly) {
          throw new Error('Tanggal transaksi wajib diisi');
        }
        const combinedDate = new Date(`${txDateOnly}T${txTimeOnly || '00:00'}:00`);
        const payload = {
          wallet_id: Number(walletId),
          target_wallet_id: mode === 'TRANSFER' && targetWalletId ? Number(targetWalletId) : null,
          category_id: mode !== 'TRANSFER' && categoryId ? Number(categoryId) : null,
          type: mode,
          amount: numAmount,
          transaction_date: combinedDate.toISOString(),
          note: note.trim()
        };
        if (editingTx) {
          await api.put(`/transactions/${editingTx.id}`, payload);
        } else {
          await api.post('/transactions', payload);
        }
      }
      onSaved();
      onClose();
    } catch (err) {
      error = err.message || 'Gagal menyimpan data';
    } finally {
      loading = false;
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/70 backdrop-blur-sm">
    <div class="w-full max-w-md bg-[#1E1E22] border border-[#2E2E34] rounded-t-2xl sm:rounded-2xl p-5 max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-base font-semibold text-[#F4F4F5]">
          {editingTx ? 'Edit Transaksi' : 'Catat Baru'}
        </h3>
        <button onclick={onClose} class="p-1.5 rounded-lg text-[#A1A1AA] hover:text-[#F4F4F5]" aria-label="Tutup">
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Mode Tabs: Income, Expense, Transfer, Acara -->
      <div class="grid {editingTx ? 'grid-cols-3' : 'grid-cols-4'} gap-1.5 bg-[#121212] p-1 rounded-xl border border-[#2E2E34] mb-4">
        <button
          type="button"
          onclick={() => switchMode('INCOME')}
          class="flex items-center justify-center gap-1 py-2 px-1 rounded-lg text-[11px] sm:text-xs font-medium transition-colors {mode === 'INCOME'
            ? 'bg-[#3B82F6] text-white'
            : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
        >
          <ArrowDownLeft class="w-3.5 h-3.5 shrink-0" />
          <span>Pendapatan</span>
        </button>
        <button
          type="button"
          onclick={() => switchMode('EXPENSE')}
          class="flex items-center justify-center gap-1 py-2 px-1 rounded-lg text-[11px] sm:text-xs font-medium transition-colors {mode === 'EXPENSE'
            ? 'bg-[#F05A4F] text-white'
            : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
        >
          <ArrowUpRight class="w-3.5 h-3.5 shrink-0" />
          <span>Pengeluaran</span>
        </button>
        <button
          type="button"
          onclick={() => switchMode('TRANSFER')}
          class="flex items-center justify-center gap-1 py-2 px-1 rounded-lg text-[11px] sm:text-xs font-medium transition-colors {mode === 'TRANSFER'
            ? 'bg-[#F8FAFC] text-[#121212]'
            : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
        >
          <ArrowLeftRight class="w-3.5 h-3.5 shrink-0" />
          <span>Transfer</span>
        </button>
        {#if !editingTx}
          <button
            type="button"
            onclick={() => switchMode('EVENT')}
            class="flex items-center justify-center gap-1 py-2 px-1 rounded-lg text-[11px] sm:text-xs font-medium transition-colors {mode === 'EVENT'
              ? 'bg-[#10B981] text-white'
              : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
          >
            <CalendarPlus class="w-3.5 h-3.5 shrink-0" />
            <span>Acara</span>
          </button>
        {/if}
      </div>

      {#if error}
        <div class="mb-3 px-3 py-2 rounded-lg bg-[#EF4444]/15 border border-[#EF4444]/40 text-xs text-[#EF4444]">
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-3.5">
        {#if mode === 'EVENT'}
          <div>
            <label for="ev-title" class="block text-xs text-[#A1A1AA] mb-1">Judul Acara / Reminder</label>
            <input
              id="ev-title"
              type="text"
              bind:value={eventTitle}
              placeholder="Contoh: Bayar Listrik, Meeting Klien"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#10B981]"
            />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label for="ev-date" class="block text-xs text-[#A1A1AA] mb-1">Tanggal</label>
              <input
                id="ev-date"
                type="date"
                bind:value={eventDate}
                required
                class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] [color-scheme:dark] focus:outline-none focus:border-[#10B981]"
              />
            </div>
            <div>
              <label for="ev-time" class="block text-xs text-[#A1A1AA] mb-1">Waktu (Opsional)</label>
              <input
                id="ev-time"
                type="time"
                bind:value={eventTime}
                class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] [color-scheme:dark] focus:outline-none focus:border-[#10B981]"
              />
            </div>
          </div>
          <div>
            <label for="ev-note" class="block text-xs text-[#A1A1AA] mb-1">Catatan</label>
            <input
              id="ev-note"
              type="text"
              bind:value={note}
              placeholder="Detail pengingat..."
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#10B981]"
            />
          </div>
        {:else}
          <div>
            <label for="tx-amount" class="block text-xs text-[#A1A1AA] mb-1">Nominal (IDR)</label>
            <input
              id="tx-amount"
              type="number"
              step="1"
              min="1"
              bind:value={amount}
              placeholder="0"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-lg font-semibold tabular-nums text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label for="tx-wallet" class="block text-xs text-[#A1A1AA] mb-1">
                {mode === 'TRANSFER' ? 'Dari Dompet' : 'Dompet'}
              </label>
              <select
                id="tx-wallet"
                bind:value={walletId}
                required
                class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
              >
                {#each wallets as w}
                  <option value={String(w.id)}>{w.name}</option>
                {/each}
              </select>
            </div>

            {#if mode === 'TRANSFER'}
              <div>
                <label for="tx-target-wallet" class="block text-xs text-[#A1A1AA] mb-1">Ke Dompet</label>
                <select
                  id="tx-target-wallet"
                  bind:value={targetWalletId}
                  required
                  class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
                >
                  {#each wallets as w}
                    <option value={String(w.id)}>{w.name}</option>
                  {/each}
                </select>
              </div>
            {:else}
              <div>
                <label for="tx-category" class="block text-xs text-[#A1A1AA] mb-1">Kategori</label>
                <select
                  id="tx-category"
                  bind:value={categoryId}
                  class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
                >
                  {#each filteredCategories as cat}
                    <option value={String(cat.id)}>{cat.name}</option>
                  {/each}
                </select>
              </div>
            {/if}
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label for="tx-date" class="block text-xs text-[#A1A1AA] mb-1">Tanggal</label>
              <input
                id="tx-date"
                type="date"
                bind:value={txDateOnly}
                required
                class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] [color-scheme:dark] focus:outline-none focus:border-[#F05A4F]"
              />
            </div>
            <div>
              <label for="tx-time" class="block text-xs text-[#A1A1AA] mb-1">Waktu</label>
              <input
                id="tx-time"
                type="time"
                bind:value={txTimeOnly}
                required
                class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] [color-scheme:dark] focus:outline-none focus:border-[#F05A4F]"
              />
            </div>
          </div>

          <div>
            <label for="tx-note" class="block text-xs text-[#A1A1AA] mb-1">Catatan (Opsional)</label>
            <input
              id="tx-note"
              type="text"
              bind:value={note}
              placeholder="Contoh: Makan siang, Gaji bulanan..."
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2.5 text-sm text-[#F4F4F5] focus:outline-none focus:border-[#F05A4F]"
            />
          </div>
        {/if}

        <button
          type="submit"
          disabled={loading}
          class="w-full py-3 rounded-xl font-semibold text-sm text-white transition-colors disabled:opacity-50 {mode === 'EVENT'
            ? 'bg-[#10B981] hover:bg-[#059669]'
            : 'bg-[#F05A4F] hover:bg-[#D9483E]'}"
        >
          {loading ? 'Menyimpan...' : editingTx ? 'Perbarui Transaksi' : 'Simpan'}
        </button>
      </form>
    </div>
  </div>
{/if}
