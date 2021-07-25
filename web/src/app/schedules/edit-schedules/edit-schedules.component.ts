import {Component, Input, OnInit} from '@angular/core';
import {Repository} from "../../repositories/repository";
import {Schedule} from "../Schedule";

@Component({
  selector: 'app-edit-schedules',
  templateUrl: './edit-schedules.component.html',
  styleUrls: ['./edit-schedules.component.css']
})
export class EditSchedulesComponent implements OnInit {
  @Input() repository: Repository | undefined;

  constructor() { }

  ngOnInit(): void {
    console.log(this.repository);
    if (this.repository?.Schedule === null) {
      this.repository.Schedule = new Schedule();
    }
  }

}
