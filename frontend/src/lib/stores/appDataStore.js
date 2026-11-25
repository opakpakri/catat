import { writable, get } from 'svelte/store';
import { api } from '../api.js';
import { authStore } from './authStore.js';
import { periodStore } from './periodStore.js';

function createAppDataStore() {
  const { subscribe, update, set } = writable({
    transactions: [],
    summary: { income: 0, expense: 0, total: 0 },
    calendarDays: [],
    calendarEvents: [],
    wallets: [],
    netWorth: 0,
    categories: [],
    budgets: [],
    loading: false
  });

  const modalState = writable({
    quickAddOpen: false,
    quickAddMode: 'EXPENSE',
    quickAddDate: '',
    editingTx: null
  });

  return {
    subscribe,
    modal: {
      subscribe: modalState.subscribe,
      openQuickAdd(dateStr = '', tx = null, mode = 'EXPENSE') {
        modalState.set({
          quickAddOpen: true,
          quickAddMode: mode,
          quickAddDate: dateStr,
          editingTx: tx
        });
      },
      closeQuickAdd() {
        modalState.update((m) => ({ ...m, quickAddOpen: false }));
      }
    },
    async refreshAll() {
      const auth = get(authStore);
      if (!auth.token) return;
      const { month, year, search } = get(periodStore);
      const q = `month=${month}&year=${year}${search ? `&search=${encodeURIComponent(search)}` : ''}`;
      update((s) => ({ ...s, loading: true }));
      try {
        const [txRes, calRes, walRes, catRes, bdgRes] = await Promise.all([
          api.get(`/transactions?${q}`),
          api.get(`/calendar/monthly?month=${month}&year=${year}`),
          api.get('/wallets'),
          api.get('/categories'),
          api.get(`/budgets?month=${month}&year=${year}`)
        ]);
        set({
          transactions: txRes.transactions || [],
          summary: txRes.summary || { income: 0, expense: 0, total: 0 },
          calendarDays: calRes.days || [],
          calendarEvents: calRes.events || [],
          wallets: walRes.wallets || [],
          netWorth: walRes.net_worth || 0,
          categories: catRes.categories || [],
          budgets: bdgRes.budgets || [],
          loading: false
        });
      } catch {
        update((s) => ({ ...s, loading: false }));
      }
    }
  };
}

export const appDataStore = createAppDataStore();
