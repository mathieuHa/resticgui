import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EditRepositoriesComponent } from './edit-repositories.component';

describe('EditRepositoriesComponent', () => {
  let component: EditRepositoriesComponent;
  let fixture: ComponentFixture<EditRepositoriesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ EditRepositoriesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(EditRepositoriesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

});
