export interface User {
  id: number
  username: string
  email?: string
  role: string
  status?: string
  maxTunnels?: number
  maxTraffic?: number
  twoFactorEnabled?: boolean
  licenseKey?: string
  licenseExpire?: string | null
}

export interface Outbound {
  id: number
  tunnelId: number
  name: string
  protocol: string
  address: string
  port: number
  config: string
  weight: number
  healthCheckEnabled: boolean
  healthCheckUrl: string
  healthCheckInterval: number
  isHealthy: boolean
}

export interface Tunnel {
  id: number
  name: string
  remark: string
  enabled: boolean
  inboundProtocol: string
  inboundPort: number
  inboundListen: string
  inboundAuth: boolean
  inboundUsername?: string
  inboundPassword?: string
  udpEnabled: boolean
  outbounds: Outbound[]
  routingStrategy: string
  uploadBytes: number
  downloadBytes: number
  connections: number
  trafficLimit: number
  trafficLimitUpload: number
  trafficLimitDownload: number
  trafficResetCycle: string
  speedLimit: number
  speedLimitUpload: number
  speedLimitDownload: number
  expireTime: string | null
  isRunning?: boolean
  aclEnabled: boolean
  aclMode: 'blacklist' | 'whitelist'
  allowDomains: string
  allowIps: string
  denyDomains: string
  denyIps: string
}
