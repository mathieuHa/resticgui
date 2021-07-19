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
  @Input() user?: User;
  faUndoAlt = faUndoAlt;
  faSave = faSave;
  constructor(
    private route: ActivatedRoute,
    private userService: UsersService,
    private location: Location
  ) { }

  ngOnInit(): void {

  }
  save(): void {
    if (this.user) {
      this.userService.CreateUser(this.user)
        .subscribe(() => this.goBack());
    }
  }
  goBack() {
    this.location.back();
  }

}
