import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../repository";
import {ActivatedRoute} from "@angular/router";
import {RepositoriesService} from "../repositories.service";
import {MatSnackBar} from "@angular/material/snack-bar";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {Location} from "@angular/common";

@Component({
  selector: 'app-edit-repositories',
  templateUrl: './edit-repositories.component.html',
  styleUrls: ['./edit-repositories.component.css']
})

export class EditRepositoriesComponent implements OnInit {
  @Input() repository?: Repository;
  faUndoAlt = faUndoAlt;
  faSave = faSave;
  id = 0;

  constructor(
    private route: ActivatedRoute,
    private repositoriesService: RepositoriesService,
    private location: Location,
    private _snackBar: MatSnackBar
  ) {
  }

  ngOnInit(): void {
    console.log("start edit-repositories");
    this.getRepository();
  }

  getRepository(): void {
    const _id = this.route.snapshot.paramMap.get('id');
    if (_id !== null){
      this.id = Number(_id);
      this.repositoriesService.getRepository(this.id).subscribe(
        response => this.repository = new Repository().deserialize(response)
      );
    }

  }
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  save(repository: Repository): void {
    if (this.id && repository) {
      console.log(repository);
      this.repositoriesService.updateRepository(this.id,repository)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("Repository Updated","close")
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
  UpdateRepository($event: Repository) {
    this.repository = $event;
    console.log($event);
    this.save(this.repository);
  }

  goBack() {
    this.location.back();
  }



}
