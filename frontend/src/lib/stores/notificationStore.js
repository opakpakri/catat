import { writable, get } from 'svelte/store';

const STORAGE_KEY = 'catat_notif_v1';
const FIRED_KEY = 'catat_notif_fired_v1';

function loadEnabled() {
  try {
    return localStorage.getItem(STORAGE_KEY) === 'true';
  } catch {
    return false;
  }
}

function loadFiredMap() {
  try {
    return JSON.parse(localStorage.getItem(FIRED_KEY) || '{}');
  } catch {
    return {};
  }
}

function saveFiredMap(map) {
  try {
    // Keep only recent entries (last 100 keys)
    const entries = Object.entries(map).slice(-100);
    localStorage.setItem(FIRED_KEY, JSON.stringify(Object.fromEntries(entries)));
  } catch {
    // ignore
  }
}

function createNotificationStore() {
  const supported = typeof window !== 'undefined' && 'Notification' in window;
  const initialPermission = supported ? Notification.permission : 'unsupported';
  const initialEnabled = supported && initialPermission === 'granted' && loadEnabled();

  const state = writable({
    supported,
    permission: initialPermission,
    enabled: initialEnabled
  });

  let swReg = null;

  async function ensureSW() {
    if (typeof navigator === 'undefined' || !('serviceWorker' in navigator)) return null;
    if (swReg) return swReg;
    try {
      swReg = await navigator.serviceWorker.register('/sw.js');
      return swReg;
    } catch {
      return null;
    }
  }

  async function showNotification(title, body, tag = 'catat-reminder') {
    if (!supported || Notification.permission !== 'granted') return false;
    const reg = await ensureSW();
    if (reg && reg.showNotification) {
      await reg.showNotification(title, {
        body,
        tag,
        renotify: true
      });
      return true;
    }
    new Notification(title, { body, tag });
    return true;
  }

  return {
    subscribe: state.subscribe,

    async init() {
      if (!supported) return;
      await ensureSW();
      const perm = Notification.permission;
      state.update((s) => ({
        ...s,
        permission: perm,
        enabled: perm === 'granted' && loadEnabled()
      }));
    },

    async enable() {
      if (!supported) return { ok: false, reason: 'unsupported' };
      const perm = await Notification.requestPermission();
      if (perm !== 'granted') {
        try {
          localStorage.setItem(STORAGE_KEY, 'false');
        } catch {
          // ignore
        }
        state.set({ supported, permission: perm, enabled: false });
        return { ok: false, reason: perm };
      }

      await ensureSW();
      try {
        localStorage.setItem(STORAGE_KEY, 'true');
      } catch {
        // ignore
      }
      state.set({ supported, permission: 'granted', enabled: true });
      return { ok: true };
    },

    disable() {
      try {
        localStorage.setItem(STORAGE_KEY, 'false');
      } catch {
        // ignore
      }
      state.update((s) => ({ ...s, enabled: false }));
    },

    async sendTest() {
      return showNotification(
        'Pengingat Acara CATAT Aktif',
        'Notifikasi reminder acara akan muncul otomatis sesuai jadwal Anda.',
        'catat-test'
      );
    },

    async checkUpcomingEvents(events = []) {
      const curr = get(state);
      if (!curr.enabled || curr.permission !== 'granted' || !Array.isArray(events)) return;

      const now = new Date();
      const pad = (n) => String(n).padStart(2, '0');
      const todayStr = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`;
      const nowMinutes = now.getHours() * 60 + now.getMinutes();

      const fired = loadFiredMap();
      let changed = false;

      for (const ev of events) {
        if (ev.is_completed || ev.event_date !== todayStr) continue;

        let shouldFire = false;
        let fireKey = `ev-${ev.id}-${ev.event_date}`;

        if (ev.event_time) {
          const [hh, mm] = ev.event_time.split(':').map(Number);
          if (!Number.isNaN(hh) && !Number.isNaN(mm)) {
            const eventMinutes = hh * 60 + mm;
            // Fire starting 15 minutes before up to 60 minutes after scheduled time
            if (nowMinutes >= eventMinutes - 15 && nowMinutes <= eventMinutes + 60) {
              shouldFire = true;
              fireKey = `ev-${ev.id}-${ev.event_date}-${ev.event_time}`;
            }
          }
        } else {
          // All-day event on today's date
          shouldFire = true;
        }

        if (shouldFire && !fired[fireKey]) {
          const timeLabel = ev.event_time ? ` (${ev.event_time})` : ' (Hari ini)';
          const body = ev.description
            ? `${ev.description}`
            : `Jadwal acara${timeLabel} belum ditandai selesai.`;
          await showNotification(`Reminder: ${ev.title}${timeLabel}`, body, fireKey);
          fired[fireKey] = Date.now();
          changed = true;
        }
      }

      if (changed) {
        saveFiredMap(fired);
      }
    }
  };
}

export const notificationStore = createNotificationStore();
