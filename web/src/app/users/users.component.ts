import { Component, OnInit } from '@angular/core';
import { User } from './user';
import { faEye, faEdit, faTrash, faPlus } from '@fortawesome/free-solid-svg-icons';
import {UsersService} from "./users.service";
import {FormControl, FormGroup, Validators} from "@angular/forms";


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
  constructor(private userService: UsersService) { }



  ngOnInit(): void {
    this.getUsers();

  }

  getUsers(): void {
    this.userService.getUsers().subscribe(
      response => this.users = response
    )
  }

  onSelect(user: User): void {
    this.selectedUser = user;
  }

}
