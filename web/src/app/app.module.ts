import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { MenuComponent } from './menu/menu.component';
import { UsersComponent } from './users/users.component';
import {FormsModule} from "@angular/forms";
import { EditUsersComponent } from './users/edit-users/edit-users.component';
import { CreateUsersComponent } from './users/create-users/create-users.component';
import { DeleteUsersComponent } from './users/delete-users/delete-users.component';
import { ViewUsersComponent } from './users/view-users/view-users.component';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';

@NgModule({
  declarations: [
    AppComponent,
    MenuComponent,
    UsersComponent,
    EditUsersComponent,
    CreateUsersComponent,
    DeleteUsersComponent,
    ViewUsersComponent
  ],
    imports: [
        BrowserModule,
        AppRoutingModule,
        FormsModule,
        FontAwesomeModule
    ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
