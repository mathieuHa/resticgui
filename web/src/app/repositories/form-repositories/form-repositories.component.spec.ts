import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FormRepositoriesComponent } from './form-repositories.component';

describe('FormRepositoriesComponent', () => {
  let component: FormRepositoriesComponent;
  let fixture: ComponentFixture<FormRepositoriesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FormRepositoriesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(FormRepositoriesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
