export function generateRandomPort(min = 10000, max = 65535): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export function generateRandomName(): string {
  const adjectives = [
    'Fast', 'Secure', 'Quick', 'Rapid', 'Stable', 'Strong', 'Swift', 'Super',
    'Turbo', 'Hyper', 'Ultra', 'Mega', 'Giga', 'Tera', 'Peta', 'Exa'
  ]
  const nouns = [
    'Tunnel', 'Proxy', 'Gateway', 'Channel', 'Portal', 'Link', 'Bridge', 'Path',
    'Route', 'Way', 'Road', 'Track', 'Line', 'Connection', 'Network', 'Stream'
  ]
  const adj = adjectives[Math.floor(Math.random() * adjectives.length)]
  const noun = nouns[Math.floor(Math.random() * nouns.length)]
  const num = Math.floor(Math.random() * 1000)
  return `${adj}${noun}${num}`
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function formatSpeed(bytesPerSecond: number): string {
  return formatBytes(bytesPerSecond) + '/s'
}
