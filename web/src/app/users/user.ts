export interface User {
    id: number;
    firstname: string;
    lastname: string;
    email: string;
}

export class User implements User{

  constructor(
    public id: number,
    firstname: string,
    lastname: string,
    email: string,
  ) {  }

}
