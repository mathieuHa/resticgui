import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";

@Component({
  selector: 'app-view-users',
  templateUrl: './view-users.component.html',
  styleUrls: ['./view-users.component.css']
})
export class ViewUsersComponent implements OnInit {
  @Input() user?: User;
  constructor(
    private route: ActivatedRoute,
    private location: Location,
    private userService: UsersService
  ) { }

  ngOnInit(): void {
    this.getUser();
  }

  getUser(): void {
    const id = Number(this.route.snapshot.paramMap.get('id'));
    this.userService.getUser(id).subscribe(
        response => this.user = response
    );
  }

}
