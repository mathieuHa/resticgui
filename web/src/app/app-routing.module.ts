import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {UsersComponent} from "./users/users.component";
import {ViewUsersComponent} from "./users/view-users/view-users.component";

const routes: Routes = [
  { path: 'users', component: UsersComponent },
  { path: 'users/:id', component: ViewUsersComponent },
  { path: 'users/:id/edit', component: ViewUsersComponent },
  { path: '', redirectTo: '/users', pathMatch: 'full' },

];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
