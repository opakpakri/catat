export function formatIDR(amount, compact = false) {
  const num = Number(amount) || 0;
  if (compact && Math.abs(num) >= 1_000_000) {
    const jt = (num / 1_000_000).toFixed(1).replace(/\.0$/, '');
    return `${jt}jt`;
  }
  if (compact && Math.abs(num) >= 1_000) {
    const rb = Math.round(num / 1_000);
    return `${rb}rb`;
  }
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(num);
}

const MONTH_NAMES_ID = [
  'Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun',
  'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'
];

export function formatPeriodLabel(month, year) {
  return `${MONTH_NAMES_ID[month - 1] || month} ${year}`;
}

export function formatDayHeader(dateStr) {
  const d = new Date(dateStr);
  const days = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
  const dd = String(d.getDate()).padStart(2, '0');
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  return {
    dayNum: dd,
    dayName: days[d.getDay()],
    monthYear: `${mm}.${d.getFullYear()}`
  };
}
