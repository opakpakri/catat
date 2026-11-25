/**
 * Builds a 42-cell (6 rows x 7 cols, Sunday-Saturday) calendar grid for (year, month).
 * month is 1-indexed (1..12).
 */
export function buildCalendarGrid(year, month) {
  const firstDay = new Date(year, month - 1, 1);
  const startDow = firstDay.getDay(); // 0 = Sunday
  const daysInMonth = new Date(year, month, 0).getDate();
  const prevMonthDays = new Date(year, month - 1, 0).getDate();

  const cells = [];

  // Previous month trailing days
  for (let i = startDow - 1; i >= 0; i--) {
    const day = prevMonthDays - i;
    const prevM = month === 1 ? 12 : month - 1;
    const prevY = month === 1 ? year - 1 : year;
    cells.push({
      day,
      inCurrentMonth: false,
      dateStr: `${prevY}-${String(prevM).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    });
  }

  // Current month days
  for (let day = 1; day <= daysInMonth; day++) {
    cells.push({
      day,
      inCurrentMonth: true,
      dateStr: `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    });
  }

  // Next month leading days to fill 35 or 42 cells
  const totalCells = cells.length <= 35 ? 35 : 42;
  let nextDay = 1;
  while (cells.length < totalCells) {
    const nextM = month === 12 ? 1 : month + 1;
    const nextY = month === 12 ? year + 1 : year;
    cells.push({
      day: nextDay,
      inCurrentMonth: false,
      dateStr: `${nextY}-${String(nextM).padStart(2, '0')}-${String(nextDay).padStart(2, '0')}`
    });
    nextDay++;
  }

  return cells;
}
