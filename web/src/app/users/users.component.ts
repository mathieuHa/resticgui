import { Component, OnInit } from '@angular/core';
import { User } from './user';
import { faEye, faEdit, faTrash, faPlus } from '@fortawesome/free-solid-svg-icons';
import {UsersService} from "./users.service";
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrls: ['./users.component.css']
})
export class UsersComponent implements OnInit {
  users: User[] = [];
  selectedUser?: User;
  faEye = faEye;
  faPlus = faPlus;
  faEdit = faEdit;
  faTrash = faTrash;
  addUser = false;
  constructor(
    private userService: UsersService,
    private _snackBar: MatSnackBar
  ) { }

  openSnackBar(message: string, action?: string) {
    this._snackBar.open(message, action);
  }

  ngOnInit(): void {
    this.getUsers();

  }

  getUsers(): void {
    this.userService.getUsers().subscribe(
      response => this.users = response
    )
  }

  deleteUsers(id: string): void {
    this.userService.deleteUser(id).subscribe(
      response => {
        console.log(response);
          this.openSnackBar("User deleted");
          this.getUsers();
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

  onSelect(user: User): void {
    this.selectedUser = user;
  }

}
