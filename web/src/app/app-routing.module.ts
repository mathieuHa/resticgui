import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {UsersComponent} from "./users/users.component";
import {ViewUsersComponent} from "./users/view-users/view-users.component";
import {EditUsersComponent} from "./users/edit-users/edit-users.component";
import {RepositoriesComponent} from "./repositories/repositories.component";
import {ViewRepositoriesComponent} from "./repositories/view-repositories/view-repositories.component";
import {EditRepositoriesComponent} from "./repositories/edit-repositories/edit-repositories.component";

const routes: Routes = [
  { path: 'users', component: UsersComponent },
  { path: 'users/:id', component: ViewUsersComponent },
  { path: 'users/:id/edit', component: EditUsersComponent },
  { path: 'repositories', component: RepositoriesComponent },
  { path: 'repositories/:id', component: ViewRepositoriesComponent },
  { path: 'repositories/:id/edit', component: EditRepositoriesComponent },

  { path: '', redirectTo: '/users', pathMatch: 'full' },

];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
