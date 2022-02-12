export interface User {
    ID: number;
    firstname: string;
    lastname: string;
    email: string;
}

interface Serializable<T> {
  deserialize(input: Object): T;
}
export class User implements User, Serializable<User>{

  constructor(
  ) {  }

  deserialize(input: any) {
    this.ID = input.ID;
    this.firstname = input.firstname;
    this.lastname = input.lastname;
    this.email = input.email;
    return this;
  }

}
