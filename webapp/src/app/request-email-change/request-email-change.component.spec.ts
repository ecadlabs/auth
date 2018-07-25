import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { RequestEmailChangeComponent } from './request-email-change.component';

describe('RequestEmailChangeComponent', () => {
  let component: RequestEmailChangeComponent;
  let fixture: ComponentFixture<RequestEmailChangeComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ RequestEmailChangeComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RequestEmailChangeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
