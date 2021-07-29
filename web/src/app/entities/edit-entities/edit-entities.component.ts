import { Component, OnInit } from '@angular/core';
import {ActivatedRoute} from "@angular/router";
import {RepositoriesService} from "../../repositories/repositories.service";
import {Location} from "@angular/common";
import {Repository} from "../../repositories/repository";
import {Entity} from "../entity";
import { faSave, faTrash, faPlus, faTimes, faCheck } from '@fortawesome/free-solid-svg-icons';
import {FormArray, FormControl, FormGroup} from "@angular/forms";
import {EntitiesService} from "../entities.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-edit-entities',
  templateUrl: './edit-entities.component.html',
  styleUrls: ['./edit-entities.component.css']
})
export class EditEntitiesComponent implements OnInit {

  repository?: Repository;
  faTrash = faTrash;
  faPlus = faPlus;
  faSave = faSave;
  faTimes = faTimes;
  faCheck    = faCheck;
  private id: string|undefined  ;
  constructor(
    private route: ActivatedRoute,
    private repositoriesService: RepositoriesService,
    private location: Location,
    private entitiesService: EntitiesService,
    private _snackBar: MatSnackBar
  ) {
  }

  ngOnInit(): void {
    console.log("start edit-entities");
    this.getRepository();
  }

  getRepository(): void {
    const _id = this.route.snapshot.paramMap.get('id');
    if (_id !== null){
      this.id = _id;
      this.repositoriesService.getRepository(this.id).subscribe(
        response => {
          this.repository = response
        }
      );
    }
  }

  deleteEntity(entity: Entity) {
    if (this.repository){
    const index = this.repository?.Entities.indexOf(entity);
    if (index >= 0 && index < this.repository?.Entities.length) {
      this.repository.Entities.splice(index, 1);
      }
    }
  }
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  addEntity() {
    console.log("addEntity");
    console.log(this.repository);
    if (this.repository) {
      if (this.repository.Entities === undefined) {
        this.repository.Entities = [];
      }
      const newEntity = new Entity();
      newEntity.path = "";
      this.repository.Entities.push(newEntity)
    }

  }

  saveEntities() {
      if (this.id && this.repository) {
        if (!this.repository.Entities){
          this.repository.Entities = [];
        }
        console.log("Save");
      this.entitiesService.updateEntities(this.id,this.repository?.Entities)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("Entities Updated","close")
            } else {
              this.openSnackBar(result.error,"close")
            }

          },
          error => {
            console.log("err")
            console.log(error);
            if (error.error && error.error.error){
              this.openSnackBar(error.error.error,"close")
            }
          });
    }
  }
}
