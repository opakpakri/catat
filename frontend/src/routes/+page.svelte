<script>
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/authStore.js';

  $effect(() => {
    if ($authStore.token) {
      goto('/transaksi', { replaceState: true });
    } else {
      goto('/login', { replaceState: true });
    }
  });
</script>
