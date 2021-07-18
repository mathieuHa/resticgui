import { Component, OnInit } from '@angular/core';
import { User } from './user';
import { USERS } from './mock-users';
import { faEye, faEdit, faTrash } from '@fortawesome/free-solid-svg-icons';


@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrls: ['./users.component.css']
})
export class UsersComponent implements OnInit {
  users = USERS;
  selectedUser?: User;
  faEye = faEye;
  faEdit = faEdit;
  faTrash = faTrash;
  constructor() { }

  ngOnInit(): void {
  }

  onSelect(user: User): void {
    this.selectedUser = user;
  }

}
