import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ServiceAccountListComponent } from './service-account-list.component';

describe('ServiceAccountListComponent', () => {
  let component: ServiceAccountListComponent;
  let fixture: ComponentFixture<ServiceAccountListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ServiceAccountListComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ServiceAccountListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
