import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../../repositories/repository";
import {Schedule} from "../Schedule";
import {SchedulesService} from "../schedules.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-edit-schedules',
  templateUrl: './edit-schedules.component.html',
  styleUrls: ['./edit-schedules.component.css']
})
export class EditSchedulesComponent implements OnInit {
  @Input() repository: Repository | undefined;
  @Input() disabled!: boolean;

  constructor(
    private scheduleService: SchedulesService,
    private _snackBar: MatSnackBar
  ) { }

  ngOnInit(): void {
    console.log(this.repository);
    if (this.repository?.Schedule === null) {
      this.repository.Schedule = new Schedule();
    }
  }
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  save(): void {
    if (this.repository) {
      this.scheduleService.updateSchedule(this.repository.id,this.repository.Schedule)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("Schedule Updated","close")
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
