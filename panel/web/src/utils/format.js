import dayjs from 'dayjs'

export function formatBytes(bytes, decimals = 2) {
  if (bytes === null || bytes === undefined || Number.isNaN(Number(bytes))) {
    return '0 B'
  }
  const value = Number(bytes)
  if (value === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(value) / Math.log(k))
  return `${parseFloat((value / k ** i).toFixed(decimals))} ${sizes[i]}`
}

export function formatDate(time, fmt = 'YYYY-MM-DD HH:mm') {
  if (!time) return '-'
  return dayjs(time).format(fmt)
}

export function formatDateShort(time) {
  return formatDate(time, 'YYYY-MM-DD')
}

export function tunnelTypeText(type) {
  const map = {
    tcp: 'TCP',
    udp: 'UDP',
    http: 'HTTP',
    https: 'HTTPS',
    stcp: 'STCP',
    xtcp: 'XTCP',
    kcp: 'KCP',
    loadbalance: '负载均衡'
  }
  return map[type] || type || '-'
}

export function tunnelStatusText(status) {
  const map = {
    running: '运行中',
    online: '在线',
    offline: '离线',
    stopped: '已停止',
    starting: '启动中',
    stopping: '停止中',
    error: '异常',
    enabled: '已启用',
    disabled: '已禁用'
  }
  return map[status] || status || '-'
}

export function tunnelStatusType(status) {
  const map = {
    running: 'success',
    online: 'success',
    enabled: 'success',
    stopped: 'default',
    offline: 'error',
    error: 'error',
    starting: 'warning',
    stopping: 'warning',
    disabled: 'default'
  }
  return map[status] || 'default'
}

export function userStatusText(status) {
  const map = {
    1: '正常',
    0: '封禁',
    active: '正常',
    banned: '封禁'
  }
  return map[status] ?? (Number(status) === 1 ? '正常' : Number(status) === 0 ? '封禁' : String(status ?? '-'))
}

export function planExpireText(time) {
  if (!time) return '长期有效'
  const end = dayjs(time)
  if (end.isBefore(dayjs())) return '已过期'
  const days = end.diff(dayjs(), 'day')
  return `${formatDateShort(time)}（剩 ${days} 天）`
}
