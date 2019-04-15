import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ServiceAccountEditFormComponent } from './service-account-edit-form.component';

describe('ServiceAccountEditFormComponent', () => {
  let component: ServiceAccountEditFormComponent;
  let fixture: ComponentFixture<ServiceAccountEditFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ServiceAccountEditFormComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ServiceAccountEditFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
