import { Injectable } from '@angular/core';
import {Repository} from "../repositories/repository";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {Observable} from "rxjs";
import {BackupSchedule} from "./BackupSchedule";

@Injectable({
  providedIn: 'root'
})
export class SchedulesService {

  private _repositoryUrl = 'api/v1/repositories';  // URL to web api

  httpOptions = {
    headers: new HttpHeaders({ 'Content-Type': 'application/json' })
  };
  constructor(
    private http: HttpClient,
  ) { }



  updateSchedule(id: number, repository: BackupSchedule): Observable<any> {
    const url = `${this._repositoryUrl}/${id}/schedules`;
    return this.http.put<BackupSchedule>(url, repository, this.httpOptions);
  }


}
