export interface Entity {
  id: string;
  path: string;
  exists: boolean
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class Entity implements Entity, Serializable<Entity>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.id = input.id;
    this.path = input.path;
    this.exists = input.exists;
    return this;
  }

}
