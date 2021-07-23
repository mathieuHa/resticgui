import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";
import {Location} from "@angular/common";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-edit-users',
  templateUrl: './edit-users.component.html',
  styleUrls: ['./edit-users.component.css']
})
export class EditUsersComponent implements OnInit {
  @Input() user?: User;
  faUndoAlt = faUndoAlt;
  faSave = faSave;
  id = "";

  constructor(
    private route: ActivatedRoute,
    private userService: UsersService,
    private location: Location,
    private _snackBar: MatSnackBar
  ) {
  }

  ngOnInit(): void {
    console.log("start edit-users");
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
  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }
  save(user: User): void {
    if (this.id && user) {
      console.log(user);
      this.userService.updateUser(this.id,user)
        .subscribe (
          result => {
            console.log("res");
            console.log(result);
            if (!result.error){
              this.openSnackBar("User Updated","close")
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
  UpdateUser($event: User) {
    this.user = $event;
    console.log($event);
    this.save(this.user);
  }



    goBack() {
    this.location.back();
  }

}
