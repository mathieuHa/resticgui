import {Entity} from "../entities/entity";

export interface Repository {
  Entities: Entity[];
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
    return this;
  }

}
