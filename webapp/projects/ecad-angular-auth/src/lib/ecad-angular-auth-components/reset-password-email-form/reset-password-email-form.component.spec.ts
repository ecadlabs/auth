import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ResetPasswordEmailFormComponent } from './reset-password-email-form.component';

describe('ResetPasswordEmailFormComponent', () => {
  let component: ResetPasswordEmailFormComponent;
  let fixture: ComponentFixture<ResetPasswordEmailFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ResetPasswordEmailFormComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ResetPasswordEmailFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
