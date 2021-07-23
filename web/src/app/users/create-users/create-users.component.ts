import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";
import {Location} from "@angular/common";
import {MatSnackBar} from "@angular/material/snack-bar";


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
    private location: Location,
    private _snackBar: MatSnackBar
  ) { }

  ngOnInit(): void {

  }
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  save(user: User): void {
    if (user) {
      this.userService.CreateUser(user)
        .subscribe(
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("User created","close")
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
  goBack() {
    this.location.back();
  }

  CreateUser($event: User) {
    this.user = $event;
    console.log($event);
    this.save(this.user);
  }
}
