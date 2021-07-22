import {Component, Input, OnInit} from '@angular/core';
import { User } from "../user";
import {FormBuilder, FormControl, FormGroup, Validators} from "@angular/forms";

@Component({
  selector: 'app-form-users',
  templateUrl: './form-users.component.html',
  styleUrls: ['./form-users.component.css']
})
export class FormUsersComponent implements OnInit{

  constructor() { }

  @Input() user?: User;
  userForm= new FormGroup({});
  model = new User(1, "", "", "");

  submitted = false;

  onSubmit() { this.submitted = true; }
  get firstname() { return this.userForm.get('firstname'); }
  get lastname() { return this.userForm.get('lastname'); }
  get email() { return this.userForm.get('email'); }

  ngOnInit(): void {
    this.userForm = new FormGroup({
      firstname: new FormControl(this.model.firstname, [
        Validators.required,
        Validators.minLength(3)
      ]),
      lastname: new FormControl('', [
        Validators.required,
        Validators.minLength(3)
      ]),
      email: new FormControl('', [
        Validators.required,
        Validators.minLength(6)
      ]),
    });

  }


}
