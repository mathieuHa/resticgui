import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";
import {Location} from "@angular/common";


@Component({
  selector: 'app-create-users',
  templateUrl: './create-users.component.html',
  styleUrls: ['./create-users.component.css']
})
export class CreateUsersComponent implements OnInit {
  faUndoAlt = faUndoAlt;
  faSave = faSave;
  user = new User();
  constructor(
    private route: ActivatedRoute,
    private userService: UsersService,
    private location: Location
  ) { }

  ngOnInit(): void {

  }
  save(user: User): void {
    if (user) {
      this.userService.CreateUser(user)
        .subscribe(
          result => console.log(result),
          error => console.log(error));
    }
  }
  goBack() {
    this.location.back();
  }

  updateUser($event: User) {
    this.user = $event;
    console.log($event);
    this.save(this.user);
  }
}
