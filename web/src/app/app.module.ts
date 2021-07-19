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
import {HttpClientModule} from "@angular/common/http";
import { FormUsersComponent } from './users/form-users/form-users.component';

// For dev purposes
import { HttpClientInMemoryWebApiModule } from 'angular-in-memory-web-api';
import { InMemoryDataService } from './in-memory-data.service';

// The HttpClientInMemoryWebApiModule module intercepts HTTP requests
// and returns simulated server responses.
// Remove it when a real server is ready to receive requests.
HttpClientInMemoryWebApiModule.forRoot(
  InMemoryDataService, { dataEncapsulation: false }
)

@NgModule({
  declarations: [
    AppComponent,
    MenuComponent,
    UsersComponent,
    EditUsersComponent,
    CreateUsersComponent,
    DeleteUsersComponent,
    ViewUsersComponent,
    FormUsersComponent
  ],
    imports: [
        BrowserModule,
        AppRoutingModule,
        FormsModule,
        FontAwesomeModule,
        HttpClientModule,
    ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
