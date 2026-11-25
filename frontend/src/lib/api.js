import { authStore } from "./stores/authStore.js";
import { get } from "svelte/store";

const API_ORIGIN = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '');
const API_BASE = `${API_ORIGIN}/api`;

export async function apiRequest(path, options = {}) {
  const auth = get(authStore);
  const headers = {
    ...(options.body instanceof FormData
      ? {}
      : { "Content-Type": "application/json" }),
    ...(options.headers || {}),
  };

  if (auth.token) {
    headers["Authorization"] = `Bearer ${auth.token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (
    res.status === 401 &&
    !path.startsWith("/auth/login") &&
    !path.startsWith("/auth/pin/verify")
  ) {
    authStore.logout();
    throw new Error("Sesi telah berakhir. Silakan login kembali.");
  }

  const contentType = res.headers.get("content-type") || "";
  if (contentType.includes("text/csv")) {
    if (!res.ok) throw new Error("Gagal mengunduh CSV");
    return await res.blob();
  }

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `Terjadi kesalahan (${res.status})`);
  }
  return data;
}

export const api = {
  get: (path) => apiRequest(path, { method: "GET" }),
  post: (path, body) =>
    apiRequest(path, { method: "POST", body: JSON.stringify(body) }),
  put: (path, body) =>
    apiRequest(path, { method: "PUT", body: JSON.stringify(body) }),
  patch: (path, body = {}) =>
    apiRequest(path, { method: "PATCH", body: JSON.stringify(body) }),
  del: (path) => apiRequest(path, { method: "DELETE" }),
};
