export interface Repository {
  id: number
  name?: string
  password?: string
  active?: boolean
}

export interface BackupPolicy {
  nbLastBackup?: number
  oneByHour?: number
  oneByDay?: number
  oneByMonth?: number
  keepAll?: boolean
}