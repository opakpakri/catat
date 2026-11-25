<script>
  import { Plus, Wallet, HandCoins, Check } from 'lucide-svelte';
  import { api } from '$lib/api.js';
  import { formatIDR } from '$lib/utils/formatCurrency.js';

  let { wallets = [], netWorth = 0, onRefresh = () => {} } = $props();

  let subTab = $state('wallets'); // 'wallets' | 'debts'
  let debts = $state([]);

  // Wallet Form
  let showWalletForm = $state(false);
  let wName = $state('');
  let wType = $state('CASH');
  let wBalance = $state('0');

  // Debt Form
  let showDebtForm = $state(false);
  let dWalletId = $state('');
  let dPerson = $state('');
  let dType = $state('DEBT');
  let dTotal = $state('');
  let dDueDate = $state('');
  let dNote = $state('');
  let dReminder = $state(true);

  // Payment input
  let payingId = $state(null);
  let payAmount = $state('');
  let payWalletId = $state('');
  let payError = $state('');

  async function loadDebts() {
    try {
      const res = await api.get('/debts');
      debts = res.debts || [];
    } catch {
      // ignore
    }
  }

  $effect(() => {
    loadDebts();
  });

  async function createWallet(e) {
    e.preventDefault();
    await api.post('/wallets', {
      name: wName.trim(),
      type: wType,
      balance: Number(wBalance) || 0,
      color_hex: wType === 'BANK' ? '#3B82F6' : wType === 'EWALLET' ? '#8B5CF6' : '#10B981'
    });
    wName = '';
    wBalance = '0';
    showWalletForm = false;
    onRefresh();
  }

  async function createDebt(e) {
    e.preventDefault();
    if (!dWalletId || !dDueDate) return;
    await api.post('/debts', {
      wallet_id: Number(dWalletId),
      person_name: dPerson.trim(),
      type: dType,
      total_amount: Number(dTotal),
      due_date: dDueDate,
      note: dNote.trim(),
      create_reminder: dReminder
    });
    dPerson = '';
    dTotal = '';
    dDueDate = '';
    dNote = '';
    showDebtForm = false;
    loadDebts();
    onRefresh();
  }

  async function submitPay(d) {
    payError = '';
    const amountNum = Number(payAmount);
    const remaining = Math.max(0, d.total_amount - d.paid_amount);
    if (!amountNum || amountNum <= 0 || !payWalletId) {
      payError = 'Pilih dompet dan masukkan nominal yang valid.';
      return;
    }
    if (amountNum > remaining) {
      payError = `Nominal tidak boleh melebihi sisa tagihan (${formatIDR(remaining)}).`;
      payAmount = String(remaining);
      return;
    }
    try {
      await api.patch(`/debts/${d.id}/pay`, {
        amount: amountNum,
        wallet_id: Number(payWalletId)
      });
      payingId = null;
      payAmount = '';
      payWalletId = '';
      loadDebts();
      onRefresh();
    } catch (err) {
      payError = err.message || 'Gagal mencatat pembayaran';
    }
  }
</script>

<div class="p-4 space-y-4">
  <!-- Net Worth Banner -->
  <div data-tour="accounts-networth" class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 flex items-center justify-between">
    <div>
      <span class="text-xs uppercase tracking-wider text-[#A1A1AA]">Total Kekayaan Bersih</span>
      <p class="text-xl font-bold text-[#F8FAFC] tabular-nums mt-0.5">{formatIDR(netWorth)}</p>
    </div>
    <div class="w-10 h-10 rounded-xl bg-[#3B82F6]/20 flex items-center justify-center text-[#3B82F6]">
      <Wallet class="w-5 h-5" />
    </div>
  </div>

  <!-- Sub Tabs -->
  <div data-tour="accounts-subtabs" class="grid grid-cols-2 border-b border-[#2E2E34]">
    <button
      onclick={() => (subTab = 'wallets')}
      class="py-2.5 text-xs font-semibold transition-colors {subTab === 'wallets'
        ? 'border-b-2 border-[#F05A4F] text-white'
        : 'text-[#A1A1AA]'}"
    >
      Dompet ({wallets.length})
    </button>
    <button
      onclick={() => (subTab = 'debts')}
      class="py-2.5 text-xs font-semibold transition-colors {subTab === 'debts'
        ? 'border-b-2 border-[#F05A4F] text-white'
        : 'text-[#A1A1AA]'}"
    >
      Utang & Piutang ({debts.length})
    </button>
  </div>

  <div data-tour="accounts-content" class="space-y-4">
  {#if subTab === 'wallets'}
    <div class="flex justify-end">
      <button
        onclick={() => (showWalletForm = !showWalletForm)}
        class="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-medium"
      >
        <Plus class="w-3.5 h-3.5" />
        Tambah Dompet
      </button>
    </div>

    {#if showWalletForm}
      <form onsubmit={createWallet} class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
        <input
          type="text"
          bind:value={wName}
          placeholder="Nama Dompet (misal: BCA, GoPay)"
          required
          class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
        />
        <div class="grid grid-cols-2 gap-3">
          <select bind:value={wType} class="bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]">
            <option value="CASH">Tunai (CASH)</option>
            <option value="BANK">Bank</option>
            <option value="EWALLET">E-Wallet</option>
            <option value="INVESTMENT">Investasi</option>
          </select>
          <input
            type="number"
            bind:value={wBalance}
            placeholder="Saldo Awal"
            class="bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
          />
        </div>
        <button type="submit" class="w-full py-2.5 rounded-xl bg-[#10B981] text-white text-xs font-semibold">
          Simpan Dompet
        </button>
      </form>
    {/if}

    <div class="space-y-2.5">
      {#each wallets as w}
        <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-3 h-8 rounded-full" style="background-color: {w.color_hex}"></div>
            <div>
              <p class="text-sm font-semibold text-[#F4F4F5]">{w.name}</p>
              <p class="text-[11px] uppercase tracking-wider text-[#A1A1AA]">{w.type}</p>
            </div>
          </div>
          <span class="text-sm font-bold tabular-nums {w.balance >= 0 ? 'text-[#F8FAFC]' : 'text-[#EF4444]'}">
            {formatIDR(w.balance)}
          </span>
        </div>
      {/each}
    </div>
  {:else}
    <div class="flex justify-end">
      <button
        onclick={() => {
          showDebtForm = !showDebtForm;
          if (showDebtForm && !dWalletId && wallets[0]) {
            dWalletId = String(wallets[0].id);
          }
        }}
        class="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-[#F05A4F] hover:bg-[#D9483E] text-white text-xs font-medium"
      >
        <Plus class="w-3.5 h-3.5" />
        Catat Utang / Piutang
      </button>
    </div>

    {#if showDebtForm}
      <form onsubmit={createDebt} class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-3">
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            onclick={() => (dType = 'DEBT')}
            class="py-2 rounded-lg text-xs font-semibold border {dType === 'DEBT'
              ? 'bg-[#EF4444]/20 border-[#EF4444] text-[#EF4444]'
              : 'border-[#2E2E34] text-[#A1A1AA]'}"
          >
            Utang (Saya Pinjam)
          </button>
          <button
            type="button"
            onclick={() => (dType = 'RECEIVABLE')}
            class="py-2 rounded-lg text-xs font-semibold border {dType === 'RECEIVABLE'
              ? 'bg-[#3B82F6]/20 border-[#3B82F6] text-[#3B82F6]'
              : 'border-[#2E2E34] text-[#A1A1AA]'}"
          >
            Piutang (Dipinjam Orang)
          </button>
        </div>
        <input
          type="text"
          bind:value={dPerson}
          placeholder="Nama Orang / Pihak"
          required
          class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
        />
        <div>
          <label for="debt-wallet" class="block text-xs text-[#A1A1AA] mb-1">
            {dType === 'DEBT' ? 'Masuk ke Dompet' : 'Keluar dari Dompet'}
          </label>
          <select
            id="debt-wallet"
            bind:value={dWalletId}
            required
            class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
          >
            <option value="" disabled>Pilih Dompet</option>
            {#each wallets as w}
              <option value={String(w.id)}>{w.name}</option>
            {/each}
          </select>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label for="debt-total" class="block text-xs text-[#A1A1AA] mb-1">Nominal (IDR)</label>
            <input
              id="debt-total"
              type="number"
              min="1"
              bind:value={dTotal}
              placeholder="Nominal (IDR)"
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
            />
          </div>
          <div>
            <label for="debt-due" class="block text-xs text-[#A1A1AA] mb-1">Tanggal Jatuh Tempo</label>
            <input
              id="debt-due"
              type="date"
              bind:value={dDueDate}
              required
              class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5] [color-scheme:dark]"
            />
          </div>
        </div>
        <input
          type="text"
          bind:value={dNote}
          placeholder="Catatan (Opsional)"
          class="w-full bg-[#121212] border border-[#2E2E34] rounded-xl px-3 py-2 text-sm text-[#F4F4F5]"
        />
        <label class="flex items-center gap-2 text-xs text-[#A1A1AA]">
          <input type="checkbox" bind:checked={dReminder} class="accent-[#10B981]" />
          Buat pengingat otomatis di kalender Monthly & Note
        </label>
        <button type="submit" class="w-full py-2.5 rounded-xl bg-[#F05A4F] text-white text-xs font-semibold">
          Simpan Catatan
        </button>
      </form>
    {/if}

    <div class="space-y-2.5">
      {#each debts as d}
        {@const pct = Math.min(100, Math.round((d.paid_amount / d.total_amount) * 100))}
        <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-4 space-y-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <HandCoins class="w-4 h-4 {d.type === 'DEBT' ? 'text-[#EF4444]' : 'text-[#3B82F6]'}" />
              <span class="text-sm font-semibold text-[#F4F4F5]">{d.person_name}</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-[#121212] text-[#A1A1AA]">
                {d.type === 'DEBT' ? 'Utang' : 'Piutang'}
              </span>
            </div>
            <span
              class="text-[10px] px-2 py-0.5 rounded-full font-semibold {d.status === 'PAID'
                ? 'bg-[#10B981]/20 text-[#34D399]'
                : 'bg-[#F59E0B]/20 text-[#F59E0B]'}"
            >
              {d.status}
            </span>
          </div>

          <div class="flex justify-between text-xs text-[#A1A1AA] tabular-nums">
            <span>Terbayar: {formatIDR(d.paid_amount)}</span>
            <span>Total: {formatIDR(d.total_amount)}</span>
          </div>
          <div class="flex flex-wrap gap-2 text-[11px] text-[#A1A1AA]">
            {#if d.wallet_name}
              <span class="px-2 py-0.5 rounded bg-[#121212] border border-[#2E2E34]">Dompet: {d.wallet_name}</span>
            {/if}
            {#if d.due_date}
              <span class="px-2 py-0.5 rounded bg-[#121212] border border-[#2E2E34]">Jatuh tempo: {d.due_date}</span>
            {/if}
          </div>
          <div class="h-1.5 rounded-full bg-[#121212] overflow-hidden">
            <div class="h-full bg-[#10B981]" style="width: {pct}%"></div>
          </div>

          {#if d.status !== 'PAID'}
            {@const remaining = Math.max(0, d.total_amount - d.paid_amount)}
            {#if payingId === d.id}
              <div class="space-y-2 pt-1">
                {#if payError}
                  <p class="text-[11px] text-[#EF4444]">{payError}</p>
                {/if}
                <div class="flex flex-col sm:flex-row gap-2">
                  <select
                    bind:value={payWalletId}
                    required
                    class="bg-[#121212] border border-[#2E2E34] rounded-lg px-2.5 py-1.5 text-xs text-[#F4F4F5]"
                  >
                    {#each wallets as w}
                      <option value={String(w.id)}>{w.name}</option>
                    {/each}
                  </select>
                  <input
                    type="number"
                    min="1"
                    max={remaining}
                    value={payAmount}
                    oninput={(e) => {
                      payError = '';
                      const val = Number(e.target.value);
                      if (val > remaining) {
                        payAmount = String(remaining);
                        e.target.value = String(remaining);
                      } else {
                        payAmount = e.target.value;
                      }
                    }}
                    placeholder="Maks {formatIDR(remaining)}"
                    required
                    class="flex-1 bg-[#121212] border border-[#2E2E34] rounded-lg px-2.5 py-1.5 text-xs text-[#F4F4F5]"
                  />
                  <button
                    onclick={() => submitPay(d)}
                    class="px-3 py-1.5 rounded-lg bg-[#10B981] text-white text-xs font-semibold flex items-center justify-center gap-1"
                  >
                    <Check class="w-3.5 h-3.5" /> Bayar
                  </button>
                </div>
              </div>
            {:else}
              <button
                onclick={() => {
                  payError = '';
                  payingId = d.id;
                  payAmount = String(remaining);
                  payWalletId = d.wallet_id ? String(d.wallet_id) : (wallets[0]?.id ? String(wallets[0].id) : '');
                }}
                class="text-xs text-[#F05A4F] hover:underline font-medium pt-1"
              >
                + Catat Pembayaran / Cicilan
              </button>
            {/if}
          {/if}
        </div>
      {/each}
    </div>
  {/if}
  </div>
</div>
