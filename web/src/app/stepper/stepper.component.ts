import { Component, OnInit, Input } from '@angular/core';
import {FormControl, FormGroup, Validators, FormBuilder, FormArray} from "@angular/forms";
import {Repository} from "../repositories/repository";
import {ActivatedRoute, Router} from "@angular/router";
import {RepositoriesService} from "../repositories/repositories.service";
import {MatSnackBar} from "@angular/material/snack-bar";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {Location} from "@angular/common";
import { Entity } from '../entities/entity';
import { BackupSchedule } from '../schedules/BackupSchedule';
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
    private router: Router,
    private _snackBar: MatSnackBar
    ) {
      this.readonly = false;
     }
  repositoryForm!: FormGroup;
  repository!: Repository;
  @Input() readonly: boolean;
  id = 0;

  createFormModel() {
    this.repositoryForm = this._formBuilder.group({
      // BackupPolicy: this._formBuilder.group({}),
      // Snapshots: [],
      name: ['', 
        [
          Validators.required,
          Validators.minLength(4)
        ]
      ],
      Entities: this._formBuilder.array([]),
      BackupSchedule: this._formBuilder.group({
        interval: ['', 
          []
        ],
        every_day: [false, []], 
        every_week: [false, []],
        every_month: [false, []],
        every_monday: [false, []],
        every_tuesday: [false, []],
        every_wednesday: [false, []],
        every_thursday: [false, []],
        every_friday: [false, []],
        every_saturday: [false, []],
        every_sunday: [false, []],
        hour: ['', 
          []
        ],
        regular_hour: ['', 
          []
        ],
        week_hour: ['', 
          []
        ],
        day: ['', 
          []
        ],
        cron_syntax: ['', 
          []
        ],
        choice: ['', 
          []
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
      this.id = Number(_id);
      this.repositoriesService.getRepository(this.id).subscribe(
        response => {
          this.repository = new Repository().deserialize(response)
          this.repositoryForm.patchValue(this.repository!)
          for (let ent of this.repository.Entities) {
            this.addEntity(ent)
          }
          if (this.readonly){
            this.repositoryForm.disable()
          }
        }
      );
    } else {
      this.repository = new Repository()
      this.repository.Entities.push(new Entity())
      this.addEntity()
      console.log(this.repositoryForm)
    }
  }
  
  addEntity(entity?: Entity) {
      const newEntity = this._formBuilder.group({
        path: [entity?.path,
            []
          ],
          exists: [entity?.exists],
          ID: [entity?.ID],
          RepositoryID: [entity?.RepositoryID],
      })
    this.Entities.push(newEntity);
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
            if (!result.error) {
              this.openSnackBar("Repository Updated","close")
              this.repository = new Repository().deserialize(result)
              this.repositoryForm.patchValue(this.repository)
            } else {
              this.openSnackBar(result.error,"close")
            }
          },
          error => {
            console.log("err")
            console.log(error);
            if (error.error && error.error.error) {
              this.openSnackBar(error.error.error,"close")
            }
          });
    }
  }

  create(repository: Repository): void {
    if (repository) {
      delete repository["ID"]
      console.log(repository);
      this.repositoriesService.createRepository(repository)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error) {
              this.openSnackBar("Repository Created","close")
              this.repository = new Repository().deserialize(result)

              this.router.navigate(['repositories', this.repository.ID]);

              this.repositoryForm.patchValue(this.repository)
              
            } else {
              this.openSnackBar(result.error,"close")
            }
          },
          error => {
            console.log("err")
            console.log(error);
            if (error.error && error.error.error) {
              this.openSnackBar(error.error.error,"close")
            }
          });
    }
  }


}
