import {Entity} from "../entities/entity";
import {BackupSchedule} from "../schedules/BackupSchedule";

export interface Repository {
  Entities: Entity[];
  Schedule: BackupSchedule;
  BackupSchedule: BackupSchedule;
  BackupPolicy: string;
  id?: string;
  name: string;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Repository implements Repository, Serializable<Repository>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.id = input.id;
    this.name = input.name;
    this.Entities = input.Entities;
    this.Schedule = input.BackupSchedule;
    this.BackupSchedule = input.BackupSchedule;
    this.BackupPolicy = input.BackupPolicy;
    return this;
  }

}
