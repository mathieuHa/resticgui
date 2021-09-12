import { Component, OnInit } from '@angular/core';
import {MatSnackBar} from "@angular/material/snack-bar";
import {Repository} from "./repository";
import {RepositoriesService} from "./repositories.service";
import { faEye, faEdit, faTrash, faPlus } from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-repositories',
  templateUrl: './repositories.component.html',
  styleUrls: ['./repositories.component.css']
})
export class RepositoriesComponent implements OnInit {

  repositories: Repository[] = [];
  selectedRepository?: Repository;
  faEye = faEye;
  faPlus = faPlus;
  faEdit = faEdit;
  faTrash = faTrash;
  addRepository = false;
  constructor(
    private repositoryService: RepositoriesService,
    private _snackBar: MatSnackBar
  ) { }

  openSnackBar(message: string, action?: string) {
    this._snackBar.open(message, action);
  }

  ngOnInit(): void {
    this.getRepositories();

  }

  getRepositories(): void {
    this.repositoryService.getRepositories().subscribe(
      response => {
        this.repositories= []
        for (let repo of response){
          this.repositories.push(new Repository().deserialize(response));
        }
        this.repositories = response
      }
    )
  }

  deleteRepositories(id: number): void {
    this.repositoryService.deleteRepository(id).subscribe(
      response => {
        console.log(response);
        this.openSnackBar("Repository deleted");
        this.getRepositories();
      },
      error => {
        console.log("err")
        console.log(error);
        if (error.error && error.error.error){
          this.openSnackBar(error.error.error,"close")
        }
      }
    )
  }

  onSelect(repository: Repository): void {
    this.selectedRepository = repository;
  }

}
