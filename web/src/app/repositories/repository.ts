import {Entity} from "../entities/entity";
import {Schedule} from "../schedules/Schedule";

export interface Repository {
  Entities: Entity[];
  Schedule: Schedule;
  BackupPolicy: string;
  id: string;
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
    this.BackupPolicy = input.BackupPolicy;
    return this;
  }

}
