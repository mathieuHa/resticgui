import { Component, OnInit } from '@angular/core';
import {ActivatedRoute} from "@angular/router";
import {RepositoriesService} from "../repositories.service";
import {Location} from "@angular/common";
import {MatSnackBar} from "@angular/material/snack-bar";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {Repository} from "../repository";

@Component({
  selector: 'app-create-repositories',
  templateUrl: './create-repositories.component.html',
  styleUrls: ['./create-repositories.component.css']
})
export class CreateRepositoriesComponent implements OnInit {

  faUndoAlt = faUndoAlt;
  faSave = faSave;
  repository = new Repository();
  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoriesService,
    private location: Location,
    private _snackBar: MatSnackBar
  ) { }

  ngOnInit(): void {

  }
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  save(repository: Repository): void {
    if (repository) {
      this.repositoryService.CreateRepository(repository)
        .subscribe(
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("Repository created","close")
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
  goBack() {
    this.location.back();
  }

  CreateRepository($event: Repository) {
    this.repository = $event;
    console.log($event);
    this.save(this.repository);
  }

}
