<script>
  import { page } from '$app/state';
  import { ReceiptText, BarChart3, Wallet, Ellipsis } from 'lucide-svelte';

  const currentPath = $derived(page.url.pathname);

  const navItems = [
    { href: '/transaksi', label: 'Transaksi', icon: ReceiptText },
    { href: '/statistik', label: 'Statistik', icon: BarChart3 },
    { href: '/akun', label: 'Akun', icon: Wallet },
    { href: '/lainnya', label: 'Lainnya', icon: Ellipsis }
  ];

  function isActive(href) {
    return currentPath === href || currentPath.startsWith(href + '/');
  }
</script>

<nav data-tour="bottom-nav" class="md:hidden fixed bottom-0 left-0 right-0 z-30 bg-[#1E1E22] border-t border-[#2E2E34]">
  <div class="max-w-lg mx-auto grid grid-cols-4 h-16">
    {#each navItems as item}
      {@const Icon = item.icon}
      {@const active = isActive(item.href)}
      <a
        href={item.href}
        class="flex flex-col items-center justify-center gap-1 transition-colors {active
          ? 'text-[#F05A4F]'
          : 'text-[#A1A1AA] hover:text-[#F4F4F5]'}"
      >
        <Icon class="w-5 h-5" />
        <span class="text-[11px] font-medium truncate max-w-full px-1">{item.label}</span>
      </a>
    {/each}
  </div>
</nav>
