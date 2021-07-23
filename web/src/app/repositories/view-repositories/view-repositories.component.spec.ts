import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ViewRepositoriesComponent } from './view-repositories.component';

describe('ViewRepositoriesComponent', () => {
  let component: ViewRepositoriesComponent;
  let fixture: ComponentFixture<ViewRepositoriesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ViewRepositoriesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ViewRepositoriesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
