import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../../repositories/repository";
import {BackupSchedule} from "../BackupSchedule";

@Component({
  selector: 'app-view-schedules',
  templateUrl: './view-schedules.component.html',
  styleUrls: ['./view-schedules.component.css']
})
export class ViewSchedulesComponent implements OnInit {
  @Input() repository: Repository | undefined;
  disabled = true;
  constructor() { }

  ngOnInit(): void {
    console.log(this.repository);
    if (this.repository?.BackupSchedule === null) {
      this.repository.BackupSchedule = new BackupSchedule();
    }
  }

}
