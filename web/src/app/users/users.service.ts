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
  get userUrl(): string {
    return this._userUrl;
  }

  set userUrl(value: string) {
    this._userUrl = value;
  }
  private _userUrl = 'api/v1/users';  // URL to web api

  httpOptions = {
    headers: new HttpHeaders({ 'Content-Type': 'application/json' })
  };
  constructor(
    private http: HttpClient,
  ) { }

  getUsers(): Observable<User[]> {
    return this.http.get<User[]>(this._userUrl).pipe(
      catchError(this.handleError<User[]>('getUsers', []))
    );
  }

  getUser(id: string): Observable<User> {
    const url = `${this._userUrl}/${id}`;

    return this.http.get<User>(url).pipe(
      catchError(this.handleError<User>(`getUser id=${id}`))
    );
  }

  updateUser(id: string, user: User): Observable<any> {
    const url = `${this._userUrl}/${id}`;
    return this.http.patch<User>(url, user, this.httpOptions);
  }

  deleteUser(id: string): Observable<any> {
    const url = `${this._userUrl}/${id}`;
    return this.http.delete<User>(url);
  }

  CreateUser(user: User): Observable<any> {
    return this.http.post(this._userUrl, user, this.httpOptions);
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
