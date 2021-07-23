import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";
import { Location } from '@angular/common';
import { faUndoAlt } from '@fortawesome/free-solid-svg-icons';


@Component({
  selector: 'app-view-users',
  templateUrl: './view-users.component.html',
  styleUrls: ['./view-users.component.css']
})
export class ViewUsersComponent implements OnInit {
  @Input() user?: User;
  faUndoAlt = faUndoAlt;
  private id: string|undefined  ;
  constructor(
    private route: ActivatedRoute,
    private userService: UsersService,
    private location: Location
  ) { }

  ngOnInit(): void {
    console.log("start view-users");
    this.getUser();
  }

  getUser(): void {
    const _id = this.route.snapshot.paramMap.get('id');
    if (_id !== null){
      this.id = _id;
      this.userService.getUser(this.id).subscribe(
        response => this.user = response
      );
    }
  }


  goBack() {
    this.location.back();
  }
}
