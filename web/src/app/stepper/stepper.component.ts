import { Component, OnInit, Input } from '@angular/core';
import {FormControl, FormGroup, Validators, FormBuilder, FormArray} from "@angular/forms";
import {Repository} from "../repositories/repository";
import {ActivatedRoute} from "@angular/router";
import {RepositoriesService} from "../repositories/repositories.service";
import {MatSnackBar} from "@angular/material/snack-bar";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {Location} from "@angular/common";
@Component({
  selector: 'app-stepper',
  templateUrl: './stepper.component.html',
  styleUrls: ['./stepper.component.css']
})
export class StepperComponent implements OnInit {

  constructor(
    private _formBuilder: FormBuilder,
    private route: ActivatedRoute,
    private repositoriesService: RepositoriesService,
    private location: Location,
    private _snackBar: MatSnackBar
    ) { }
  repositoryForm!: FormGroup;
  @Input() repository?: Repository;
  id = "";

  createFormModel() {
    this.repositoryForm = this._formBuilder.group({
      // id: [undefined, []],
      // BackupPolicy: this._formBuilder.group({}),
      // Snapshots: [],
      name: ['', 
        [
          Validators.required,
          Validators.minLength(4)
        ]
      ],
      Entities: this._formBuilder.array([
        this._formBuilder.group({
          path: ['',
            [
              Validators.required,
              Validators.minLength(4)
            ]
          ],
          exists: [false],
        })
      ]),
      BackupSchedule: this._formBuilder.group({
        interval: ['', 
          [
            Validators.required,
            Validators.minLength(4)
          ]
        ],
      }),
      Active: [false,[]],
    });
  }
  ngOnInit(): void {
    this.createFormModel()
    this.getRepository()
  }

  getRepository(): void {
    const _id = this.route.snapshot.paramMap.get('id');
    if (_id !== null){
      this.id = _id;
      this.repositoriesService.getRepository(this.id).subscribe(
        response => {
          this.repository = new Repository().deserialize(response)
          console.log(this.repository)
          this.repositoryForm.patchValue(this.repository!)
        }
      );
    } else {
      this.repository = new Repository()
    }
  }
  
  addEntity() {
      const entity = this._formBuilder.group({
        path: ['',
            [
              Validators.required,
              Validators.minLength(4)
            ]
          ],
          exists: [false],
      })
    this.Entities.push(entity);
    console.log(this.repositoryForm)
  }
  get Entities() {
    return this.repositoryForm.controls["Entities"] as FormArray;
  }

  removeEntity(index: number) {
    this.Entities.removeAt(index);
  }

  
  submit() {
    console.warn(this.repositoryForm.value);
    this.repository = this.repositoryForm.value
    console.log(this.repository)
    if (this.repository && this.id){
      this.update(this.repository)
    } else if (this.repository) {
      this.create(this.repository);
    }
  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  update(repository: Repository): void {
    if (repository) {
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

  create(repository: Repository): void {
    if (repository) {
      delete repository["id"]
      console.log(repository);
      this.repositoriesService.createRepository(repository)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("Repository Created","close")
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


}
