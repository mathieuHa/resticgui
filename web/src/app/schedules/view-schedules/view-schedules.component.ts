import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../../repositories/repository";
import {Schedule} from "../Schedule";

@Component({
  selector: 'app-view-schedules',
  templateUrl: './view-schedules.component.html',
  styleUrls: ['./view-schedules.component.css']
})
export class ViewSchedulesComponent implements OnInit {
  @Input() repository: Repository | undefined;

  constructor() { }

  ngOnInit(): void {
    console.log(this.repository);
    if (this.repository?.Schedule === null) {
      this.repository.Schedule = new Schedule();
    }
  }

}
