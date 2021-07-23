import {Component, Input, OnInit} from '@angular/core';
import {User} from "../user";
import {ActivatedRoute} from "@angular/router";
import {UsersService} from "../users.service";
import {Location} from "@angular/common";
import {faSave, faUndoAlt} from '@fortawesome/free-solid-svg-icons';

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
    private location: Location
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
  save(): void {
    if (this.id && this.user) {
      this.userService.updateUser(this.id,this.user)
        .subscribe(() => this.goBack());
    }
  }


  goBack() {
    this.location.back();
  }

}
