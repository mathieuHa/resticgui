export interface Entity {
  ID: number;
  RepositoryID: string;
  path: string;
  exists: boolean;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Entity implements Entity, Serializable<Entity>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.RepositoryID = input.RepositoryID;
    this.ID = input.ID;
    this.path = input.path;
    this.exists = input.exists;
    return this;
  }

}
