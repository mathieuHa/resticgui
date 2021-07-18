import { Injectable } from '@angular/core';
import {User} from "./user";
import {USERS} from "./mock-users";
import {Observable, of} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class UsersService {

  constructor() { }

  getUsers(): Observable<User[]> {
    const users = of(USERS)
    return users;
  }

  getUser(id: number): Observable<User> {
    // For now, assume that a user with the specified `id` always exists.
    const user = USERS.find(h => h.id === id)!;
    return of(user);
  }
}
