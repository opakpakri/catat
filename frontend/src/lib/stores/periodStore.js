import { writable } from 'svelte/store';

const now = new Date();

function createPeriodStore() {
  const { subscribe, update, set } = writable({
    month: now.getMonth() + 1,
    year: now.getFullYear(),
    search: ''
  });

  return {
    subscribe,
    prevMonth() {
      update((p) => {
        if (p.month === 1) return { ...p, month: 12, year: p.year - 1 };
        return { ...p, month: p.month - 1 };
      });
    },
    nextMonth() {
      update((p) => {
        if (p.month === 12) return { ...p, month: 1, year: p.year + 1 };
        return { ...p, month: p.month + 1 };
      });
    },
    setPeriod(month, year) {
      update((p) => ({ ...p, month, year }));
    },
    setSearch(search) {
      update((p) => ({ ...p, search }));
    },
    resetToNow() {
      const d = new Date();
      set({ month: d.getMonth() + 1, year: d.getFullYear(), search: '' });
    }
  };
}

export const periodStore = createPeriodStore();
