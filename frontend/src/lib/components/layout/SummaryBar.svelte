<script>
  import { Wallet, ArrowDownLeft, ArrowUpRight } from 'lucide-svelte';
  import { formatIDR } from '../../utils/formatCurrency.js';

  let { income = 0, expense = 0, total = 0, netWorth = 0, userName = '' } = $props();

  let initials = $derived(
    (userName || 'C')
      .split(' ')
      .filter(Boolean)
      .slice(0, 2)
      .map((w) => w[0].toUpperCase())
      .join('')
  );
  let isHealthy = $derived(total >= 0);
</script>

<div data-tour="summary-bar" class="px-4 pt-3 pb-3 space-y-2.5 bg-[#121212] border-b border-[#2E2E34]">
  <!-- Compact Greeting -->
  <div class="flex items-center gap-2.5">
    <div class="w-8 h-8 rounded-full bg-[#F05A4F]/20 border border-[#F05A4F]/40 flex items-center justify-center text-xs font-bold text-[#F05A4F] shrink-0">
      {initials}
    </div>
    <div class="min-w-0 leading-tight">
      <p class="text-[11px] text-[#A1A1AA] truncate">Halo, {userName || 'Pengguna'}!</p>
      <p class="text-xs font-semibold text-[#F4F4F5]">Kelola keuanganmu hari ini</p>
    </div>
  </div>

  <!-- Coral-Red Hero Card (Saldo Bersih) -->
  <div class="rounded-2xl bg-gradient-to-br from-[#F05A4F] via-[#E54E43] to-[#C83B31] p-4 shadow-lg shadow-[#F05A4F]/20 text-white">
    <div class="flex items-center justify-between">
      <span class="text-xs font-medium text-white/85">Saldo Bersih</span>
      <div class="w-7 h-7 rounded-lg bg-white/15 flex items-center justify-center">
        <Wallet class="w-4 h-4 text-white" />
      </div>
    </div>
    <p class="text-2xl font-bold tracking-tight tabular-nums mt-1 text-white">
      {formatIDR(netWorth)}
    </p>
    <div class="flex items-center justify-between mt-2 pt-2 border-t border-white/15 text-[11px]">
      <span class="font-semibold {isHealthy ? 'text-[#A7F3D0]' : 'text-[#FDE68A]'}">
        {isHealthy ? 'Keuangan Sehat' : 'Pengeluaran Melebihi Pendapatan'}
      </span>
      <span class="text-white/75 tabular-nums">
        Arus kas: {total >= 0 ? '+' : ''}{formatIDR(total, true)}
      </span>
    </div>
  </div>

  <!-- Compact 2-Column Pendapatan & Pengeluaran -->
  <div class="grid grid-cols-2 gap-2.5">
    <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-xl px-3 py-2.5 flex items-center gap-2.5">
      <div class="w-8 h-8 rounded-full bg-[#3B82F6]/15 flex items-center justify-center shrink-0">
        <ArrowDownLeft class="w-4 h-4 text-[#3B82F6]" />
      </div>
      <div class="min-w-0">
        <p class="text-[10px] uppercase tracking-wider text-[#A1A1AA] font-medium">Pendapatan</p>
        <p class="text-xs sm:text-sm font-bold text-[#3B82F6] tabular-nums truncate">
          {formatIDR(income)}
        </p>
      </div>
    </div>

    <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-xl px-3 py-2.5 flex items-center gap-2.5">
      <div class="w-8 h-8 rounded-full bg-[#EF4444]/15 flex items-center justify-center shrink-0">
        <ArrowUpRight class="w-4 h-4 text-[#EF4444]" />
      </div>
      <div class="min-w-0">
        <p class="text-[10px] uppercase tracking-wider text-[#A1A1AA] font-medium">Pengeluaran</p>
        <p class="text-xs sm:text-sm font-bold text-[#EF4444] tabular-nums truncate">
          {formatIDR(expense)}
        </p>
      </div>
    </div>
  </div>
</div>
