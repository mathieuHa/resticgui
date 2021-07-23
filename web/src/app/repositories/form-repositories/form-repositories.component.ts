import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormControl, FormGroup, Validators} from "@angular/forms";
import {Repository} from "../repository";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-form-repositories',
  templateUrl: './form-repositories.component.html',
  styleUrls: ['./form-repositories.component.css']
})
export class FormRepositoriesComponent implements OnInit {

  faUndoAlt = faUndoAlt;
  faSave = faSave;
  constructor() { }
  @Input() repository?: Repository;
  repositoryForm= new FormGroup({
    name: new FormControl('', [
      Validators.required,
      Validators.minLength(3)
    ])
  });
  @Output() updateRepositoryForm = new EventEmitter<Repository>();

  submitted = false;

  onSubmit() {
    this.submitted = true;
    this.updateRepositoryForm.emit(new Repository().deserialize(this.repositoryForm.value));
  }

  ngOnInit(): void {
    // This is for edit
    this.repositoryForm.patchValue(
      {
        'name': this.repository?.name,
      }
    )
  }


}
