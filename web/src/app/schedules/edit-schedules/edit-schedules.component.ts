import {Component, Input, OnInit, ViewChild} from '@angular/core';
import {Repository} from "../../repositories/repository";
import {Schedule} from "../Schedule";
import {SchedulesService} from "../schedules.service";
import {MatSnackBar} from "@angular/material/snack-bar";
import {MatTabGroup} from "@angular/material/tabs";

@Component({
  selector: 'app-edit-schedules',
  templateUrl: './edit-schedules.component.html',
  styleUrls: ['./edit-schedules.component.css']
})
export class EditSchedulesComponent implements OnInit {
  @Input() repository: Repository | undefined;
  @Input() disabled!: boolean;
  @ViewChild("selectedTab", { static: false }) selectedTab: MatTabGroup | undefined;


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

    if (this.repository && this.selectedTab) {
      // @ts-ignore
      if (this.selectedTab.selectedIndex == 0) {
        this.repository.Schedule.choice = "cron"
      } else if (this.selectedTab.selectedIndex == 1) {
        this.repository.Schedule.choice = "interval"
      } else if (this.selectedTab.selectedIndex == 2) {
        this.repository.Schedule.choice = "regular"
      } else if (this.selectedTab.selectedIndex == 3) {
        this.repository.Schedule.choice = "weekly"
      }
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

  updateRegularSetting(choice: string) {
    if (choice === "day" && this.repository && this.repository.Schedule.every_day) {
      this.repository.Schedule.every_week = false;
      this.repository.Schedule.every_month = false;
    } else if (choice === "week" && this.repository && this.repository.Schedule.every_week) {
      this.repository.Schedule.every_day = false;
      this.repository.Schedule.every_month = false;
    } else if (choice === "month" && this.repository && this.repository.Schedule.every_month) {
      this.repository.Schedule.every_week = false;
      this.repository.Schedule.every_day = false;
    }
  }
}
