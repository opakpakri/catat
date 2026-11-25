<script>
  import { CheckCircle2, Circle, Trash2, Calendar, Clock, Plus } from 'lucide-svelte';
  import { api } from '../../api.js';

  let { events = [], search = '', onAddEvent = () => {}, onRefresh = () => {} } = $props();

  let filteredEvents = $derived.by(() => {
    const q = (search || '').trim().toLowerCase();
    if (!q) return events;
    return events.filter(
      (ev) =>
        (ev.title || '').toLowerCase().includes(q) ||
        (ev.description || '').toLowerCase().includes(q)
    );
  });

  async function toggleEvent(ev) {
    if (ev.is_completed) return;
    await api.patch(`/events/${ev.id}/toggle`);
    onRefresh();
  }

  async function deleteEvent(id) {
    if (!confirm('Hapus acara / reminder ini?')) return;
    await api.del(`/events/${id}`);
    onRefresh();
  }

  function formatEventDate(dateStr) {
    if (!dateStr) return '';
    const d = new Date(`${dateStr}T00:00:00`);
    const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
    return `${days[d.getDay()]}, ${String(d.getDate()).padStart(2, '0')} ${months[d.getMonth()]} ${d.getFullYear()}`;
  }
</script>

<div class="p-4 space-y-3">
  <div class="flex items-center justify-between">
    <h3 class="text-xs font-semibold uppercase tracking-wider text-[#A1A1AA]">
      Daftar Acara & Catatan ({filteredEvents.length})
    </h3>
    <button
      onclick={onAddEvent}
      class="flex items-center gap-1 px-2.5 py-1.5 rounded-lg bg-[#10B981] hover:bg-[#059669] text-white text-xs font-medium"
    >
      <Plus class="w-3.5 h-3.5" />
      Tambah Acara
    </button>
  </div>

  {#if filteredEvents.length === 0}
    <div class="py-14 text-center px-6 bg-[#1E1E22] border border-[#2E2E34] rounded-2xl">
      <Calendar class="w-8 h-8 text-[#10B981]/70 mx-auto mb-2" />
      <p class="text-sm text-[#A1A1AA]">Belum ada acara atau reminder pada periode ini.</p>
      <p class="text-xs text-[#A1A1AA]/70 mt-1">
        Tekan tombol <strong>Tambah Acara</strong> atau <strong>+</strong> untuk menambahkan catatan acara.
      </p>
    </div>
  {:else}
    <div class="space-y-2.5">
      {#each filteredEvents as ev}
        <div class="bg-[#1E1E22] border border-[#2E2E34] rounded-2xl p-3.5 flex items-start justify-between gap-3">
          <button
            onclick={() => toggleEvent(ev)}
            disabled={ev.is_completed}
            class="flex items-start gap-3 text-left flex-1 min-w-0 {ev.is_completed ? 'cursor-default' : 'cursor-pointer'}"
          >
            <div class="mt-0.5 shrink-0">
              {#if ev.is_completed}
                <CheckCircle2 class="w-5 h-5 text-[#10B981]" />
              {:else}
                <Circle class="w-5 h-5 text-[#A1A1AA]" />
              {/if}
            </div>
            <div class="min-w-0 flex-1 space-y-1">
              <p
                class="text-sm font-semibold break-words {ev.is_completed
                  ? 'line-through text-[#A1A1AA]'
                  : 'text-[#F4F4F5]'}"
              >
                {ev.title}
              </p>

              <div class="flex flex-wrap items-center gap-2 text-[11px] text-[#A1A1AA]">
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-[#10B981]/15 text-[#34D399] font-medium">
                  <Calendar class="w-3 h-3" />
                  {formatEventDate(ev.event_date)}
                </span>
                {#if ev.event_time}
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-[#121212] border border-[#2E2E34]">
                    <Clock class="w-3 h-3" />
                    {ev.event_time}
                  </span>
                {/if}
              </div>

              {#if ev.description}
                <p class="text-xs text-[#A1A1AA] break-words pt-0.5">
                  {ev.description}
                </p>
              {/if}
            </div>
          </button>

          <button
            onclick={() => deleteEvent(ev.id)}
            class="p-1.5 text-[#A1A1AA] hover:text-[#EF4444] rounded-lg shrink-0"
            title="Hapus acara"
            aria-label="Hapus acara"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>
