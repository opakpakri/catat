import { writable } from 'svelte/store';

function createLockStore() {
  const { subscribe, set } = writable({
    locked: false
  });

  return {
    subscribe,
    lock() {
      set({ locked: true });
    },
    unlock() {
      set({ locked: false });
    }
  };
}

export const lockStore = createLockStore();
