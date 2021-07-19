import { Injectable } from '@angular/core';
import {User} from "./user";
import {USERS} from "./mock-users";
import {Observable, of} from "rxjs";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {catchError} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class UsersService {
  private userUrl = 'api/users';  // URL to web api
  httpOptions = {
    headers: new HttpHeaders({ 'Content-Type': 'application/json' })
  };
  constructor(
    private http: HttpClient,
  ) { }

  getUsers(): Observable<User[]> {
    return this.http.get<User[]>(this.userUrl).pipe(
      catchError(this.handleError<User[]>('getUsers', []))
    );
  }

  getUser(id: number): Observable<User> {
    const url = `${this.userUrl}/${id}`;

    return this.http.get<User>(url).pipe(
      catchError(this.handleError<User>(`getUser id=${id}`))
    );
  }

  updateUser(user: User): Observable<any> {
    return this.http.put(this.userUrl, user, this.httpOptions).pipe(
      catchError(this.handleError<any>('updateUser'))
    );
  }

  CreateUser(user: User): Observable<any> {
    return this.http.post(this.userUrl, user, this.httpOptions).pipe(
      catchError(this.handleError<any>('createUser'))
    );
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {

      // TODO: send the error to remote logging infrastructure
      console.error(error); // log to console instead

      // Let the app keep running by returning an empty result.
      return of(result as T);
    };
  }
}
