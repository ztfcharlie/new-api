import dayjs from "dayjs";
// import "dayjs/locale/zh-cn";
import relativeTime from "dayjs/plugin/relativeTime";
dayjs.extend(relativeTime);
// dayjs.locale("zh-cn");

// 格式化日期
export function formatDateTime(date = undefined, formatText = "YYYY-MM-DD") {
  return dayjs(date).format(formatText);
}
