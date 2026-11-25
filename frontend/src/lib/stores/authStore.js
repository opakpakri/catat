import { writable } from 'svelte/store';

const STORAGE_KEY = 'catat_auth_v1';

function loadInitial() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {
    // ignore storage errors
  }
  return { token: null, user: null };
}

function createAuthStore() {
  const { subscribe, set, update } = writable(loadInitial());

  return {
    subscribe,
    setAuth(token, user) {
      const next = { token, user };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      set(next);
    },
    setUser(user) {
      update((curr) => {
        const next = { ...curr, user };
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
        return next;
      });
    },
    logout() {
      localStorage.removeItem(STORAGE_KEY);
      set({ token: null, user: null });
    }
  };
}

export const authStore = createAuthStore();
