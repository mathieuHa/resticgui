import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateRepositoriesComponent } from './create-repositories.component';

describe('CreateRepositoriesComponent', () => {
  let component: CreateRepositoriesComponent;
  let fixture: ComponentFixture<CreateRepositoriesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateRepositoriesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateRepositoriesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
