import dayjs from "dayjs";

export function formatTime(date?: Date | string): string {
  if (!date) return "";
  return dayjs(date).format("YYYY-MM-DD HH:mm:ss");
}

export function formatTimeShort(date?: Date | string): string {
  if (!date) return "";
  // if today return time, else return date
  const d = dayjs(date);
  const today = dayjs();
  if (d.isSame(today, "day")) {
    return d.format("HH:mm:ss");
  } else {
    return d.format("YYYY-MM-DD");
  }
}

export function formatDuration(start: Date | string, end: Date | string): string {
  const d = dayjs(end).diff(dayjs(start), "second")
  if (d > 3600) {
    return (`${Math.floor(d / 3600)} 小时 ${Math.floor(d / 60) % 60} 分`)
  } else if (d > 60) {
    return (`${Math.floor(d / 60)} 分 ${d % 60} 秒`)
  } else {
    return (`${d} 秒`);
  }
}