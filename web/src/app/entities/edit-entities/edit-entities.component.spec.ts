import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EditEntitiesComponent } from './edit-entities.component';

describe('EditEntitiesComponent', () => {
  let component: EditEntitiesComponent;
  let fixture: ComponentFixture<EditEntitiesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ EditEntitiesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(EditEntitiesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
