import { Injectable } from '@angular/core';
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {Observable} from "rxjs";
import {Repository} from "../repositories/repository";
import {Entity} from "./entity";

@Injectable({
  providedIn: 'root'
})
export class EntitiesService {

  private _repositoryUrl = 'api/v1/repositories';  // URL to web api

  httpOptions = {
    headers: new HttpHeaders({ 'Content-Type': 'application/json' })
  };
  constructor(
    private http: HttpClient,
  ) { }


  updateEntities(id: string, entities: Entity[]): Observable<any> {
    const url = `${this._repositoryUrl}/${id}/entities`;
    return this.http.put<Repository>(url, entities, this.httpOptions);
  }
}
