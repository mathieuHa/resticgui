import {Component, Input, OnInit} from '@angular/core';
import {ActivatedRoute} from "@angular/router";
import {Location} from "@angular/common";
import { faUndoAlt } from '@fortawesome/free-solid-svg-icons';
import {RepositoriesService} from "../repositories.service";
import {Repository} from "../repository";

@Component({
  selector: 'app-view-repositories',
  templateUrl: './view-repositories.component.html',
  styleUrls: ['./view-repositories.component.css']
})
export class ViewRepositoriesComponent implements OnInit {

  @Input() repository?: Repository;
  faUndoAlt = faUndoAlt;
  private id: string|undefined  ;
  constructor(
    private route: ActivatedRoute,
    private repositoriesService: RepositoriesService,
    private location: Location
  ) { }

  ngOnInit(): void {
    console.log("start view-repositories");
    this.getRepository();
  }

  getRepository(): void {
    const _id = this.route.snapshot.paramMap.get('id');
    if (_id !== null){
      this.id = _id;
      this.repositoriesService.getRepository(this.id).subscribe(
        response => this.repository = response
      );
    }
  }


  goBack() {
    this.location.back();
  }
}
