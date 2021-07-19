import { Component, OnInit } from '@angular/core';
import { User } from "../user";

@Component({
  selector: 'app-form-users',
  templateUrl: './form-users.component.html',
  styleUrls: ['./form-users.component.css']
})
export class FormUsersComponent {

  constructor() { }

  model = new User(18, 'Dr IQ');

  submitted = false;

  onSubmit() { this.submitted = true; }

}
