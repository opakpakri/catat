import { writable } from 'svelte/store';

const STORAGE_KEY = 'catat_theme_v1';

function getInitialTheme() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'light' || saved === 'dark' || saved === 'system') return saved;
  } catch {
    // ignore storage errors
  }
  return 'dark';
}

function resolveTheme(mode) {
  if (mode === 'system' && typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return mode === 'light' ? 'light' : 'dark';
}

function applyTheme(mode) {
  if (typeof document === 'undefined') return;
  const effective = resolveTheme(mode);
  const root = document.documentElement;
  if (effective === 'light') {
    root.classList.remove('dark');
    root.classList.add('light');
  } else {
    root.classList.remove('light');
    root.classList.add('dark');
  }
}

function createThemeStore() {
  let currentMode = getInitialTheme();
  applyTheme(currentMode);
  const { subscribe, set } = writable(currentMode);

  if (typeof window !== 'undefined' && window.matchMedia) {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    mq.addEventListener('change', () => {
      if (currentMode === 'system') {
        applyTheme('system');
      }
    });
  }

  return {
    subscribe,
    setTheme(mode) {
      const next = mode === 'light' || mode === 'dark' || mode === 'system' ? mode : 'dark';
      currentMode = next;
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        // ignore
      }
      applyTheme(next);
      set(next);
    }
  };
}

export const themeStore = createThemeStore();
