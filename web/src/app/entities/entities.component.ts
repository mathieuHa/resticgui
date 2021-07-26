import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../repositories/repository";
import {RepositoriesService} from "../repositories/repositories.service";
import {MatSnackBar} from "@angular/material/snack-bar";
import { faEye, faPlus, faEdit, faTrash } from '@fortawesome/free-solid-svg-icons';
import {EntitiesService} from "./entities.service";
import {Entity} from "./entity";

@Component({
  selector: 'app-entities',
  templateUrl: './entities.component.html',
  styleUrls: ['./entities.component.css']
})
export class EntitiesComponent implements OnInit {

  @Input() repository: Repository | undefined;
  selectedEntity?: Entity;
  faEye = faEye;
  faPlus = faPlus;
  faEdit = faEdit;
  faTrash = faTrash;
  addEntity = false;
  constructor(
    private entitiesService: EntitiesService,
    private _snackBar: MatSnackBar
  ) { }

  openSnackBar(message: string, action?: string) {
    this._snackBar.open(message, action);
  }

  ngOnInit(): void {

  }
}
