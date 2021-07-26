import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { MenuComponent } from './menu/menu.component';
import { UsersComponent } from './users/users.component';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import { EditUsersComponent } from './users/edit-users/edit-users.component';
import { CreateUsersComponent } from './users/create-users/create-users.component';
import { ViewUsersComponent } from './users/view-users/view-users.component';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import {HttpClientModule} from "@angular/common/http";
import { FormUsersComponent } from './users/form-users/form-users.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {MatSnackBarModule} from "@angular/material/snack-bar";
import { RepositoriesComponent } from './repositories/repositories.component';
import { CreateRepositoriesComponent } from './repositories/create-repositories/create-repositories.component';
import { EditRepositoriesComponent } from './repositories/edit-repositories/edit-repositories.component';
import { FormRepositoriesComponent } from './repositories/form-repositories/form-repositories.component';
import { ViewRepositoriesComponent } from './repositories/view-repositories/view-repositories.component';
import { EntitiesComponent } from './entities/entities.component';
import { EditEntitiesComponent } from './entities/edit-entities/edit-entities.component';
import { EditSchedulesComponent } from './schedules/edit-schedules/edit-schedules.component';
import { ViewSchedulesComponent } from './schedules/view-schedules/view-schedules.component';
import {MatDividerModule} from "@angular/material/divider";
import {MatTabsModule} from "@angular/material/tabs";
import {MatGridListModule} from "@angular/material/grid-list";

@NgModule({
  declarations: [
    AppComponent,
    MenuComponent,
    UsersComponent,
    EditUsersComponent,
    CreateUsersComponent,
    ViewUsersComponent,
    FormUsersComponent,
    RepositoriesComponent,
    CreateRepositoriesComponent,
    EditRepositoriesComponent,
    FormRepositoriesComponent,
    ViewRepositoriesComponent,
    EntitiesComponent,
    EditEntitiesComponent,
    EditSchedulesComponent,
    ViewSchedulesComponent
  ],
    imports: [
        BrowserModule,
        AppRoutingModule,
        FormsModule,
        ReactiveFormsModule,
        FontAwesomeModule,
        HttpClientModule,
        BrowserAnimationsModule,
        MatSnackBarModule,
        MatDividerModule,
        MatTabsModule,
        MatGridListModule
    ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
