import {Component, Input, OnInit, Output, EventEmitter} from '@angular/core';
import { User } from "../user";
import {FormBuilder, FormControl, FormGroup, Validators} from "@angular/forms";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-form-users',
  templateUrl: './form-users.component.html',
  styleUrls: ['./form-users.component.css']
})
export class FormUsersComponent implements OnInit{
  faUndoAlt = faUndoAlt;
  faSave = faSave;
  constructor() { }
  @Input() user?: User;
  userForm= new FormGroup({
    firstname: new FormControl('', [
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
  @Output() updateUserForm = new EventEmitter<User>();

  submitted = false;

  onSubmit() {
    this.submitted = true;
    this.updateUserForm.emit(new User().deserialize(this.userForm.value));
  }

  ngOnInit(): void {
    // This is for edit
    this.userForm.patchValue(
      {
          'firstname': this.user?.firstname,
          'lastname': this.user?.lastname,
          'email': this.user?.email
      }
    )
  }


}
