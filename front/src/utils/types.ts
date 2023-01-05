export enum Provider {
  LocalProvider = 'localProvider',
  MinioS3Provider = 'MinioS3Provider',
  B2Provider = 'B2Provider',
  AWSS3Provider = 'AWSS3Provider',
}

export interface Repository {
  id?: number
  name?: string
  password?: string
  active?: boolean
  pruneActive?: boolean
  provider?: string
  repositorySize?: number
  repositoryFileCount?: string
  snapshotCount?: string
  version?: string
  initiated?: boolean
  valid?: boolean
  running?: boolean
  createdAt?: string
  edges?: {
    [Provider.LocalProvider]?: LocalProvider
    [Provider.MinioS3Provider]?: MinioS3Provider
    [Provider.B2Provider]?: B2Provider
    [Provider.AWSS3Provider]?: AWSS3Provider
    backupPolicy?: BackupPolicy
    backupSchedule?: BackupSchedule
    pruneSchedule?: PruneSchedule
    paths: Path[]
  }
}

export interface BackupPolicy {
  nbKeepLast?: number
  nbKeepHours?: number
  nbKeepDays?: number
  nbKeepWeeks?: number
  nbKeepMonths?: number
  keepAll?: boolean
}

export interface BackupSchedule {
  choice?: string
  interval?: string
  cronSyntax?: string
  lastRun?: string
  nextRun?: string
}

export interface PruneSchedule {
  choice?: string
  interval?: string
  cronSyntax?: string
  lastRun?: string
  nextRun?: string
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
  certificateAuthority?: string
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

export interface Path {
  name?: string
  absPath?: string
  relPath?: string
  included?: boolean
  isDir?: boolean
}

export interface Tree {
  __fetched: boolean
  __visible: boolean
  [key: string]: Tree | string | boolean
}
