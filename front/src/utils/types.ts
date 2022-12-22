export interface Repository {
  id: number
  name?: string
  password?: string
  active?: boolean
  provider?: string
  edges?: {
    backupPolicy?: BackupPolicy
    backupSchedule?: BackupSchedule
    pruneSchedule?: PruneSchedule
    localProvider?: LocalProvider
    MinioS3Provider?: MinioS3Provider
    B2Provider?: B2Provider
    AWSS3Provider?: AWSS3Provider
  }
}

export interface BackupPolicy {
  nbLastBackup?: number
  oneByHour?: number
  oneByDay?: number
  oneByMonth?: number
  keepAll?: boolean
}

export interface BackupSchedule {
  choice?: string
  interval?: string
  cronSyntax?: string
}

export interface PruneSchedule {
  choice?: string
  interval?: string
  cronSyntax?: string
}

export interface LocalProvider {
  repoPath?: string
}

export interface MinioS3Provider {
  accessKeyID?: string
  bucketName?: string
  port?: number
  secretAccessKey?: string
  scheme?: string
}

export interface B2Provider {
  accountID?: string
  accountKey?: string
  bucketName?: string
  repoName?: string
}

export interface AWSS3Provider {
  accessKeyID?: string
  secretAccessKey?: string
  bucketName?: string
  defaultRegion?: string
}
