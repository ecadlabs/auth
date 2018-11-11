import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ResetPasswordEmailFormComponent } from './reset-password-email-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { PASSWORD_RESET, AUTH_CONFIG } from '../../ecad-angular-auth/tokens';
import { MatInputModule, MatCardModule } from '@angular/material';
import { RouterTestingModule } from '@angular/router/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('ResetPasswordEmailFormComponent', () => {
  let component: ResetPasswordEmailFormComponent;
  let fixture: ComponentFixture<ResetPasswordEmailFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ReactiveFormsModule,
        MatInputModule,
        RouterTestingModule,
        MatCardModule,
        NoopAnimationsModule
      ],
      providers: [
        { provide: PASSWORD_RESET, useValue: {} },
        { provide: AUTH_CONFIG, useValue: {} }
      ],
      declarations: [ResetPasswordEmailFormComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
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
