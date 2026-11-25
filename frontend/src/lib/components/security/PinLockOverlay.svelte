<script>
  import { Lock, Delete } from 'lucide-svelte';
  import { api } from '../../api.js';
  import { lockStore } from '../../stores/lockStore.js';
  import { authStore } from '../../stores/authStore.js';

  let pin = $state('');
  let error = $state('');
  let loading = $state(false);

  async function submitPin(currentPin) {
    if (currentPin.length !== 6) return;
    loading = true;
    error = '';
    try {
      const res = await api.post('/auth/pin/verify', { pin: currentPin });
      if (res.verified) {
        pin = '';
        lockStore.unlock();
      } else {
        error = 'PIN salah';
        pin = '';
      }
    } catch (e) {
      error = e.message || 'PIN salah';
      pin = '';
    } finally {
      loading = false;
    }
  }

  function pressDigit(d) {
    if (loading || pin.length >= 6) return;
    error = '';
    pin += String(d);
    if (pin.length === 6) {
      submitPin(pin);
    }
  }

  function backspace() {
    if (loading || pin.length === 0) return;
    pin = pin.slice(0, -1);
  }
</script>

<div class="fixed inset-0 z-50 bg-[#121212] flex flex-col items-center justify-center px-6">
  <div class="w-full max-w-xs flex flex-col items-center">
    <div class="w-14 h-14 rounded-2xl bg-[#1E1E22] border border-[#2E2E34] flex items-center justify-center mb-4">
      <Lock class="w-7 h-7 text-[#F05A4F]" />
    </div>
    <h2 class="text-lg font-semibold text-[#F4F4F5]">CATAT Terkunci</h2>
    <p class="text-xs text-[#A1A1AA] mt-1 mb-6">Masukkan 6-digit PIN keamanan Anda</p>

    <div class="flex items-center gap-3 mb-4">
      {#each Array(6) as _, i}
        <div
          class="w-3.5 h-3.5 rounded-full border transition-all {i < pin.length
            ? 'bg-[#F05A4F] border-[#F05A4F] scale-110'
            : 'border-[#2E2E34] bg-[#1E1E22]'}"
        ></div>
      {/each}
    </div>

    <div class="h-5 mb-4">
      {#if error}
        <span class="text-xs text-[#EF4444] font-medium">{error}</span>
      {:else if loading}
        <span class="text-xs text-[#A1A1AA]">Memverifikasi...</span>
      {/if}
    </div>

    <div class="grid grid-cols-3 gap-3 w-full">
      {#each [1, 2, 3, 4, 5, 6, 7, 8, 9] as num}
        <button
          onclick={() => pressDigit(num)}
          class="h-14 rounded-xl bg-[#1E1E22] border border-[#2E2E34] text-lg font-semibold text-[#F4F4F5] hover:bg-[#2E2E34] active:scale-95 transition-all"
        >
          {num}
        </button>
      {/each}
      <button
        onclick={() => authStore.logout()}
        class="h-14 rounded-xl text-xs text-[#A1A1AA] hover:text-[#EF4444] transition-colors"
      >
        Keluar
      </button>
      <button
        onclick={() => pressDigit(0)}
        class="h-14 rounded-xl bg-[#1E1E22] border border-[#2E2E34] text-lg font-semibold text-[#F4F4F5] hover:bg-[#2E2E34] active:scale-95 transition-all"
      >
        0
      </button>
      <button
        onclick={backspace}
        class="h-14 rounded-xl bg-[#1E1E22] border border-[#2E2E34] flex items-center justify-center text-[#A1A1AA] hover:text-[#F4F4F5] active:scale-95 transition-all"
        aria-label="Hapus digit"
      >
        <Delete class="w-5 h-5" />
      </button>
    </div>
  </div>
</div>
