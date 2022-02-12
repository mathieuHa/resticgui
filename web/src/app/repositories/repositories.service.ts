import { Injectable } from '@angular/core';
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {Observable} from "rxjs";
import {Repository} from "./repository";

@Injectable({
  providedIn: 'root'
})
export class RepositoriesService {

  private _repositoryUrl = 'api/v1/repositories';  // URL to web api

  httpOptions = {
    headers: new HttpHeaders({ 'Content-Type': 'application/json' })
  };
  constructor(
    private http: HttpClient,
  ) { }

  getRepositories(): Observable<Repository[]> {
    return this.http.get<Repository[]>(this._repositoryUrl);
  }

  getRepository(id: number): Observable<Repository> {
    const url = `${this._repositoryUrl}/${id}`;
    return this.http.get<Repository>(url);
  }

  updateRepository(id: number, repository: Repository): Observable<any> {
    const url = `${this._repositoryUrl}/${id}`;
    return this.http.patch<Repository>(url, repository, this.httpOptions);
  }

  deleteRepository(id: number): Observable<any> {
    const url = `${this._repositoryUrl}/${id}`;
    return this.http.delete<Repository>(url);
  }

  createRepository(repository: Repository): Observable<any> {
    return this.http.post(this._repositoryUrl, repository, this.httpOptions);
  }


}
